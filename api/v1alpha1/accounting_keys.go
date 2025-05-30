// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (o *Accounting) Key() types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-accounting", o.Name),
		Namespace: o.Namespace,
	}
}

func (o *Accounting) PrimaryName() string {
	key := o.Key()
	return fmt.Sprintf("%s-0", key.Name)
}

func (o *Accounting) ServiceKey() types.NamespacedName {
	key := o.Key()
	return types.NamespacedName{
		Name:      key.Name,
		Namespace: o.Namespace,
	}
}

func (o *Accounting) ServiceFQDN() string {
	s := o.ServiceKey()
	return fqdn(s.Name, s.Namespace)
}

func (o *Accounting) ServiceFQDNShort() string {
	s := o.ServiceKey()
	return fqdnShort(s.Name, s.Namespace)
}

func (o *Accounting) AuthStorageKey() types.NamespacedName {
	return types.NamespacedName{
		Name:      o.Spec.StorageConfig.PasswordKeyRef.Name,
		Namespace: o.Namespace,
	}
}

func (o *Accounting) AuthStorageRef() *SecretKeySelector {
	authKey := o.AuthStorageKey()
	return &SecretKeySelector{
		SecretKeySelector: corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: authKey.Name,
			},
			Key: o.Spec.StorageConfig.PasswordKeyRef.Key,
		},
		Generate: false,
	}
}

func (o *Accounting) AuthSlurmKey() types.NamespacedName {
	return types.NamespacedName{
		Name:      o.Spec.SlurmKeyRef.Name,
		Namespace: o.Namespace,
	}
}

func (o *Accounting) AuthSlurmRef() *SecretKeySelector {
	ref := o.Spec.SlurmKeyRef
	return &SecretKeySelector{
		SecretKeySelector: corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: ref.Name,
			},
			Key: ref.Key,
		},
		Generate: false,
	}
}

func (o *Accounting) AuthJwtHs256Key() types.NamespacedName {
	return types.NamespacedName{
		Name:      o.Spec.JwtHs256KeyRef.Name,
		Namespace: o.Namespace,
	}
}

func (o *Accounting) AuthJwtHs256Ref() *SecretKeySelector {
	ref := o.Spec.JwtHs256KeyRef
	return &SecretKeySelector{
		SecretKeySelector: corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: ref.Name,
			},
			Key: ref.Key,
		},
		Generate: false,
	}
}

func (o *Accounting) ConfigKey() types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-accounting", o.Name),
		Namespace: o.Namespace,
	}
}
