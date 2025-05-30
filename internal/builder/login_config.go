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

func (b *Builder) BuildLoginSshConfig(loginset *slinkyv1alpha1.LoginSet) (*corev1.ConfigMap, error) {
	spec := loginset.Spec
	opts := ConfigMapOpts{
		Key:      loginset.SshConfigKey(),
		Metadata: loginset.Spec.Template.PodMetadata,
		Data: map[string]string{
			authorizedKeysFile: buildAuthorizedKeys(spec.RootSshAuthorizedKeys),
			sshdConfigFile:     buildSshdConfig(spec.ExtraSshdConfig),
		},
	}

	opts.Metadata.Labels = utils.MergeMaps(opts.Metadata.Labels, labels.NewBuilder().WithLoginLabels(loginset).Build())

	return b.BuildConfigMap(opts, loginset)
}

func buildAuthorizedKeys(authorizedKeys string) string {
	conf := config.NewBuilder()

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### SLINKY ###"))
	conf.AddPropery(config.NewPropertyRaw(authorizedKeys))

	return conf.Build()
}

func buildSshdConfig(extraConf string) string {
	conf := config.NewBuilder().WithSeperator(" ")

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### GENERAL ###"))
	conf.AddPropery(config.NewProperty("Include", "/etc/ssh/sshd_config.d/*.conf"))
	conf.AddPropery(config.NewProperty("UsePAM", "yes"))
	conf.AddPropery(config.NewProperty("X11Forwarding", "yes"))
	conf.AddPropery(config.NewProperty("Subsystem", "sftp internal-sftp"))

	conf.AddPropery(config.NewPropertyRaw("#"))
	conf.AddPropery(config.NewPropertyRaw("### EXTRA CONFIG ###"))
	conf.AddPropery(config.NewPropertyRaw(extraConf))

	return conf.Build()
}
