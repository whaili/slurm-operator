// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	corev1 "k8s.io/api/core/v1"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/labels"
	"github.com/SlinkyProject/slurm-operator/internal/utils"
	"github.com/SlinkyProject/slurm-operator/internal/utils/config"
)

func (b *Builder) BuildAccountingConfig(accounting *slinkyv1alpha1.Accounting) (*corev1.ConfigMap, error) {
	opts := ConfigMapOpts{
		Key:      accounting.ConfigKey(),
		Metadata: accounting.Spec.Template.PodMetadata,
		Data: map[string]string{
			slurmdbdConfFile: buildSlurmdbdConf(accounting),
		},
	}

	opts.Metadata.Labels = utils.MergeMaps(opts.Metadata.Labels, labels.NewBuilder().WithAccountingLabels(accounting).Build())

	return b.BuildConfigMap(opts, accounting)
}

// https://slurm.schedmd.com/slurmdbd.conf.html
func buildSlurmdbdConf(accounting *slinkyv1alpha1.Accounting) string {
	dbdHost := accounting.PrimaryName()
	storageHost := accounting.Spec.StorageConfig.Host
	storagePort := accounting.Spec.StorageConfig.Port
	storageLoc := accounting.Spec.StorageConfig.Database
	storageUser := accounting.Spec.StorageConfig.Username
	storagePass := "$" + storagePassEnv

	conf := config.NewBuilder()

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### GENERAL ###"))
	conf.AddPropery(config.NewProperty("DbdHost", dbdHost))
	conf.AddPropery(config.NewProperty("DbdPort", SlurmdbdPort))
	conf.AddPropery(config.NewProperty("SlurmUser", slurmUser))

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### PLUGINS & PARAMETERS ###"))
	conf.AddPropery(config.NewProperty("AuthType", authType))
	conf.AddPropery(config.NewProperty("AuthAltTypes", authAltTypes))
	conf.AddPropery(config.NewProperty("AuthAltParameters", authAltParameters))
	conf.AddPropery(config.NewProperty("AuthInfo", authInfo))

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### STORAGE ###"))
	conf.AddPropery(config.NewProperty("StorageType", "accounting_storage/mysql"))
	conf.AddPropery(config.NewProperty("StorageHost", storageHost))
	conf.AddPropery(config.NewProperty("StoragePort", storagePort))
	conf.AddPropery(config.NewProperty("StorageUser", storageUser))
	conf.AddPropery(config.NewProperty("StorageLoc", storageLoc))
	conf.AddPropery(config.NewProperty("StoragePass", storagePass))

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### LOGGING ###"))
	conf.AddPropery(config.NewProperty("LogFile", devNull))
	conf.AddPropery(config.NewProperty("LogTimeFormat", logTimeFormat))

	extraConf := accounting.Spec.ExtraConf
	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### EXTRA CONFIG ###"))
	conf.AddPropery(config.NewPropertyRaw(extraConf))

	return conf.Build()
}
