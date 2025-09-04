// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ObjectReference is a reference to an object.
// +structType=atomic
type ObjectReference struct {
	// Namespace of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +optional
	Name string `json:"name,omitempty"`
}

func (o *ObjectReference) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
}

func (o *ObjectReference) IsMatch(key types.NamespacedName) bool {
	switch {
	case o.Name != key.Name:
		return false
	case o.Namespace != key.Namespace:
		return false
	default:
		return true
	}
}

type SecretKeySelector struct {
	// SecretKeySelector selects a key of a Secret.
	// +structType=atomic
	corev1.SecretKeySelector `json:",inline"`

	// Generate indicates whether the Secret should be generated if the Secret referenced is not present.
	// +optional
	// +default:=false
	Generate bool `json:"generate,omitzero"`
}

// PodTemplate describes a template for creating copies of a predefined pod.
type PodTemplate struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	PodMetadata Metadata `json:"metadata,omitempty"`

	// PodSpec is a description of a pod.
	// +optional
	PodSpecWrapper PodSpecWrapper `json:"spec,omitempty"`
}

// Metadata defines the metadata to added to resources.
type Metadata struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PodSpecWrapper is a wrapper around corev1.PodSpec with a custom implementation
// of MarshalJSON and UnmarshalJSON which delegate to the underlying Spec to avoid CRD pollution.
// +kubebuilder:pruning:PreserveUnknownFields
type PodSpecWrapper struct {
	corev1.PodSpec `json:"-"`
}

// MarshalJSON defers JSON encoding data from the wrapper.
func (m *PodSpecWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.PodSpec)
}

// UnmarshalJSON will decode the data into the wrapper.
func (m *PodSpecWrapper) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.PodSpec)
}

func (m *PodSpecWrapper) DeepCopy() *PodSpecWrapper {
	return &PodSpecWrapper{
		PodSpec: m.PodSpec,
	}
}

// ContainerWrapper is a wrapper around corev1.Container with a custom implementation
// of MarshalJSON and UnmarshalJSON which delegate to the underlying Spec to avoid CRD pollution.
// +kubebuilder:pruning:PreserveUnknownFields
type ContainerWrapper struct {
	corev1.Container `json:"-"`
}

// MarshalJSON defers JSON encoding data from the wrapper.
func (m *ContainerWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Container)
}

// UnmarshalJSON will decode the data into the wrapper.
func (m *ContainerWrapper) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.Container)
}

func (m *ContainerWrapper) DeepCopy() *ContainerWrapper {
	return &ContainerWrapper{
		Container: m.Container,
	}
}

// SideCar defines container used as a side-car.
type SideCar struct {
	// Image URI.
	// More info: https://kubernetes.io/docs/concepts/containers/images
	// +optional
	Image string `json:"image,omitempty"`

	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Compute Resources required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitzero"`
}

// ServiceSpec defines a template to customize Service objects.
type ServiceSpec struct {
	// ServiceSpec describes the attributes that a user creates on a service.
	// +optional
	ServiceSpecWrapper `json:",inline"`

	// The external service port number.
	// +optional
	Port int `json:"port"`
}

// ServiceSpecWrapper is a wrapper around corev1.Container with a custom implementation
// of MarshalJSON and UnmarshalJSON which delegate to the underlying Spec to avoid CRD pollution.
// +kubebuilder:pruning:PreserveUnknownFields
type ServiceSpecWrapper struct {
	corev1.ServiceSpec `json:"-"`
}

// MarshalJSON defers JSON encoding data from the wrapper.
func (m *ServiceSpecWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.ServiceSpec)
}

// UnmarshalJSON will decode the data into the wrapper.
func (m *ServiceSpecWrapper) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.ServiceSpec)
}

func (m *ServiceSpecWrapper) DeepCopy() *ServiceSpecWrapper {
	return &ServiceSpecWrapper{
		ServiceSpec: m.ServiceSpec,
	}
}
