// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	corev1 "k8s.io/api/core/v1"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/labels"
	"github.com/SlinkyProject/slurm-operator/internal/utils"
)

func (b *Builder) BuildControllerScripts(controller *slinkyv1alpha1.Controller) (*corev1.ConfigMap, error) {
	opts := ConfigMapOpts{
		Key:      controller.ScriptsKey(),
		Metadata: controller.Spec.Template.PodMetadata,
		Data:     buildScriptMap(controller),
	}

	opts.Metadata.Labels = utils.MergeMaps(opts.Metadata.Labels, labels.NewBuilder().WithControllerLabels(controller).Build())

	return b.BuildConfigMap(opts, controller)
}

func buildScriptMap(controller *slinkyv1alpha1.Controller) map[string]string {
	prologScripts := controller.Spec.PrologScripts
	epilogScripts := controller.Spec.EpilogScripts

	m := make(map[string]string)
	for filename, text := range prologScripts {
		name := prologPrefix + filename
		m[name] = text
	}
	for filename, text := range epilogScripts {
		name := epilogPrefix + filename
		m[name] = text
	}

	return m
}
