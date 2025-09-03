// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/metadata"
	"github.com/SlinkyProject/slurm-operator/internal/utils/structutils"
)

type ServiceOpts struct {
	Key      types.NamespacedName
	Metadata slinkyv1alpha1.Metadata
	corev1.ServiceSpec
	Selector map[string]string
	Headless bool
}

func (b *Builder) BuildService(opts ServiceOpts, owner metav1.Object) (*corev1.Service, error) {
	if len(opts.Ports) > 1 {
		if err := validateServicePorts(opts.Ports); err != nil {
			return nil, fmt.Errorf("error validating Ports: %w", err)
		}
	}

	objectMeta := metadata.NewBuilder(opts.Key).
		WithMetadata(opts.Metadata).
		Build()

	o := &corev1.Service{
		ObjectMeta: objectMeta,
		Spec:       opts.ServiceSpec,
	}

	o.Spec.Selector = structutils.MergeMaps(o.Spec.Selector, opts.Selector)

	if opts.Headless {
		o.Spec.ClusterIP = "None"
		o.Spec.PublishNotReadyAddresses = true
	}

	if err := controllerutil.SetControllerReference(owner, o, b.client.Scheme()); err != nil {
		return nil, fmt.Errorf("failed to set owner controller: %w", err)
	}

	return o, nil
}

func validateServicePorts(ports []corev1.ServicePort) error {
	nameMap := make(map[string]bool, len(ports))
	portMap := make(map[int32]bool, len(ports))
	for _, p := range ports {
		if nameMap[p.Name] {
			return fmt.Errorf("port name '%s' is already taken by another port", p.Name)
		}
		nameMap[p.Name] = true
		if portMap[p.Port] {
			return fmt.Errorf("port number '%d' is already taken by another port", p.Port)
		}
		portMap[p.Port] = true
	}
	return nil
}
