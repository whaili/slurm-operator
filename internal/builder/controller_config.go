// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/labels"
	"github.com/SlinkyProject/slurm-operator/internal/utils"
	"github.com/SlinkyProject/slurm-operator/internal/utils/config"
)

const (
	slurmConfFile  = "slurm.conf"
	cgroupConfFile = "cgroup.conf"
)

func (b *Builder) BuildControllerConfig(controller *slinkyv1alpha1.Controller) (*corev1.ConfigMap, error) {
	ctx := context.TODO()

	accounting, err := b.refResolver.GetAccounting(ctx, controller.Spec.AccountingRef)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
	}

	nodesetList, err := b.refResolver.GetNodeSetsForController(ctx, controller)
	if err != nil {
		return nil, err
	}

	configFilesList := &corev1.ConfigMapList{
		Items: make([]corev1.ConfigMap, 0, len(controller.Spec.ConfigFileRefs)),
	}
	for _, ref := range controller.Spec.ConfigFileRefs {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{
			Namespace: controller.Namespace,
			Name:      ref.Name,
		}
		if err := b.client.Get(ctx, key, cm); err != nil {
			return nil, err
		}
		configFilesList.Items = append(configFilesList.Items, *cm)
	}
	cgroupEnabled := true
	hasCgroupConfFile := false
	for _, configMap := range configFilesList.Items {
		if contents, ok := configMap.Data[cgroupConfFile]; ok {
			hasCgroupConfFile = true
			cgroupEnabled = isCgroupEnabled(contents)
		}
	}

	prologScripts := []string{}
	for _, ref := range controller.Spec.PrologScriptRefs {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{
			Namespace: controller.Namespace,
			Name:      ref.Name,
		}
		if err := b.client.Get(ctx, key, cm); err != nil {
			return nil, err
		}
		filenames := utils.Keys(cm.Data)
		sort.Strings(filenames)
		prologScripts = filenames
	}

	epilogScripts := []string{}
	for _, ref := range controller.Spec.EpilogScriptRefs {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{
			Namespace: controller.Namespace,
			Name:      ref.Name,
		}
		if err := b.client.Get(ctx, key, cm); err != nil {
			return nil, err
		}
		filenames := utils.Keys(cm.Data)
		sort.Strings(filenames)
		epilogScripts = filenames
	}

	opts := ConfigMapOpts{
		Key:      controller.ConfigKey(),
		Metadata: controller.Spec.Template.PodMetadata,
		Data: map[string]string{
			slurmConfFile: buildSlurmConf(controller, accounting, nodesetList, prologScripts, epilogScripts, cgroupEnabled),
		},
	}
	if !hasCgroupConfFile {
		opts.Data[cgroupConfFile] = buildCgroupConf()
	}

	opts.Metadata.Labels = utils.MergeMaps(opts.Metadata.Labels, labels.NewBuilder().WithControllerLabels(controller).Build())

	return b.BuildConfigMap(opts, controller)
}

// https://slurm.schedmd.com/slurm.conf.html
func buildSlurmConf(
	controller *slinkyv1alpha1.Controller,
	accounting *slinkyv1alpha1.Accounting,
	nodesetList *slinkyv1alpha1.NodeSetList,
	prologScripts, epilogScripts []string,
	cgroupEnabled bool,
) string {
	controllerHost := fmt.Sprintf("%s(%s)", controller.PrimaryName(), controller.PrimaryFQDN())

	conf := config.NewBuilder()

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### GENERAL ###"))
	conf.AddPropery(config.NewProperty("ClusterName", controller.ClusterName()))
	conf.AddPropery(config.NewProperty("SlurmUser", slurmUser))
	conf.AddPropery(config.NewProperty("SlurmctldHost", controllerHost))
	conf.AddPropery(config.NewProperty("SlurmctldPort", SlurmctldPort))
	conf.AddPropery(config.NewProperty("StateSaveLocation", slurmctldSpoolDir))
	conf.AddPropery(config.NewProperty("SlurmdUser", slurmdUser))
	conf.AddPropery(config.NewProperty("SlurmdPort", SlurmdPort))
	conf.AddPropery(config.NewProperty("SlurmdSpoolDir", slurmdSpoolDir))
	conf.AddPropery(config.NewProperty("ReturnToService", 2))
	conf.AddPropery(config.NewProperty("MaxNodeCount", 1024))
	conf.AddPropery(config.NewProperty("GresTypes", "gpu"))

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### LOGGING ###"))
	conf.AddPropery(config.NewProperty("SlurmctldLogFile", slurmctldLogFilePath))
	conf.AddPropery(config.NewProperty("SlurmSchedLogFile", slurmctldLogFilePath))
	conf.AddPropery(config.NewProperty("SlurmdLogFile", slurmdLogFilePath))
	conf.AddPropery(config.NewProperty("LogTimeFormat", logTimeFormat))

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### PLUGINS & PARAMETERS ###"))
	conf.AddPropery(config.NewProperty("AuthType", authType))
	conf.AddPropery(config.NewProperty("CredType", credType))
	conf.AddPropery(config.NewProperty("AuthAltTypes", authAltTypes))
	conf.AddPropery(config.NewProperty("AuthAltParameters", authAltParameters))
	conf.AddPropery(config.NewProperty("AuthInfo", authInfo))
	conf.AddPropery(config.NewProperty("CommunicationParameters", "block_null_hash"))
	conf.AddPropery(config.NewProperty("SelectTypeParameters", "CR_Core_Memory"))
	conf.AddPropery(config.NewProperty("SlurmctldParameters", "enable_configless,enable_stepmgr"))
	if cgroupEnabled {
		conf.AddPropery(config.NewProperty("ProctrackType", "proctrack/cgroup"))
		conf.AddPropery(config.NewProperty("PrologFlags", "Contain"))
		conf.AddPropery(config.NewProperty("TaskPlugin", "task/cgroup,task/affinity"))
	} else {
		conf.AddPropery(config.NewProperty("TaskPlugin", "task/affinity"))
	}

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### ACCOUNTING ###"))
	if accounting != nil {
		conf.AddPropery(config.NewProperty("AccountingStorageType", "accounting_storage/slurmdbd"))
		conf.AddPropery(config.NewProperty("AccountingStorageHost", accounting.ServiceKey().Name))
		conf.AddPropery(config.NewProperty("AccountingStoragePort", SlurmdbdPort))
		conf.AddPropery(config.NewProperty("AccountingStorageTRES", "gres/gpu"))
		if cgroupEnabled {
			conf.AddPropery(config.NewProperty("JobAcctGatherType", "jobacct_gather/cgroup"))
		}
	} else {
		conf.AddPropery(config.NewProperty("AccountingStorageType", "accounting_storage/none"))
		conf.AddPropery(config.NewProperty("JobAcctGatherType", "jobacct_gather/none"))
	}

	if len(prologScripts) > 0 || len(epilogScripts) > 0 {
		conf.AddPropery(config.NewPropertyRaw("#"))
		conf.AddPropery(config.NewPropertyRaw("### PROLOG & EPILOG ###"))
	}
	for _, filename := range prologScripts {
		conf.AddPropery(config.NewProperty("Prolog", filename))
	}
	for _, filename := range epilogScripts {
		conf.AddPropery(config.NewProperty("Epilog", filename))
	}

	if len(nodesetList.Items) > 0 {
		conf.AddPropery(config.NewPropertyRaw("#"))
		conf.AddPropery(config.NewPropertyRaw("### COMPUTE & PARTITION ###"))
	}
	for _, nodeset := range nodesetList.Items {
		partition := nodeset.Spec.Partition
		if !partition.Enabled {
			continue
		}
		name := nodeset.Name
		template := nodeset.Spec.Template.PodSpecWrapper
		if template.Hostname != "" {
			name = strings.Trim(template.Hostname, "-")
		}
		nodesetLine := []string{
			fmt.Sprintf("NodeSet=%v", name),
			fmt.Sprintf("Feature=%v", name),
		}
		nodesetLineRendered := strings.Join(nodesetLine, " ")
		conf.AddPropery(config.NewPropertyRaw(nodesetLineRendered))
		partitionLine := []string{
			fmt.Sprintf("PartitionName=%v", name),
			fmt.Sprintf("Nodes=%v", name),
			partition.Config,
		}
		partitionLineRendered := strings.Join(partitionLine, " ")
		conf.AddPropery(config.NewPropertyRaw(partitionLineRendered))
	}

	extraConf := controller.Spec.ExtraConf
	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### EXTRA CONFIG ###"))
	conf.AddPropery(config.NewPropertyRaw(extraConf))

	return conf.Build()
}

func buildCgroupConf() string {
	conf := config.NewBuilder()

	conf.AddPropery(config.NewProperty("CgroupPlugin", "autodetect"))
	conf.AddPropery(config.NewProperty("IgnoreSystemd", "yes"))

	return conf.Build()
}

func isCgroupEnabled(cgroupConf string) bool {
	r := regexp.MustCompile(`(?im)^CgroupPlugin=disabled`)
	found := r.FindStringSubmatch(cgroupConf)
	return len(found) == 0
}
