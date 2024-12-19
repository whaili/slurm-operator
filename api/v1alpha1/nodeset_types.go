// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	NodeSetRevisionLabel = appsv1.ControllerRevisionHashLabelKey
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NodeSetSpec defines the desired state of NodeSet
type NodeSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// clusterName is the name of the Slurm cluster to which this NodeSet
	// belongs to. This will be matched with the name in Cluster CRD.
	ClusterName string `json:"clusterName"`

	// replicas is the desired number of replicas of the given Template.
	// These are replicas in the sense that they are instantiations of the
	// same Template, but individual replicas also have a consistent identity.
	// If unspecified, defaults to 0.
	// TODO: Consider a rename of this field.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// If empty, defaulted to labels on Pod Template.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector"`

	// serviceName is the name of the service that governs this NodeSet.
	// This service must exist before the NodeSet, and is responsible for the
	// network identity of the NodeSet. Pods get DNS/hostnames that follow the
	// pattern: pod-specific-string.serviceName.default.svc.cluster.local
	// where "pod-specific-string" is managed by the NodeSet controller.
	ServiceName string `json:"serviceName"`

	// template is the object that describes the pod that will be created.
	// The NodeSet will create exactly one copy of this pod on every node
	// that matches the template's node selector (or on every node if no node
	// selector is specified).
	// The only allowed template.spec.restartPolicy value is "Always".
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#pod-template
	Template corev1.PodTemplateSpec `json:"template"`

	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The NodeSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`

	// updateStrategy indicates the NodeSetUpdateStrategy that will be
	// employed to update Pods in the NodeSet when a revision is made to
	// Template.
	UpdateStrategy NodeSetUpdateStrategy `json:"updateStrategy,omitempty"`

	// revisionHistoryLimit is the maximum number of revisions that will
	// be maintained in the StatefulSet's revision history. The revision history
	// consists of all revisions not represented by a currently applied
	// StatefulSetSpec version. The default value is 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`

	// PersistentVolumeClaimRetentionPolicy describes the policy used for PVCs
	// created from the NodeSet VolumeClaimTemplates. This requires the
	// NodeSetAutoDeletePVC feature gate to be enabled, which is alpha.
	// +optional
	PersistentVolumeClaimRetentionPolicy *NodeSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty"`

	// minReadySeconds is the minimum number of seconds for which a newly
	// created NodeSet Pod should be ready without any of its container crashing,
	// for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready).
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`
}

// NodeSetUpdateStrategy indicates the strategy that the NodeSet
// controller will be used to perform updates. It includes any additional
// parameters necessary to perform the update for the indicated strategy.
type NodeSetUpdateStrategy struct {
	// Type indicates the type of the NodeSetUpdateStrategy.
	// Default is RollingUpdate.
	// +optional
	Type NodeSetUpdateStrategyType `json:"type,omitempty"`

	// RollingUpdate is used to communicate parameters when Type is
	// RollingUpdateNodeSetStrategyType.
	// +optional
	RollingUpdate *RollingUpdateNodeSetStrategy `json:"rollingUpdate,omitempty"`
}

// PersistentVolumeClaimRetentionPolicyType is a string enumeration of the policies that will determine
// when volumes from the VolumeClaimTemplates will be deleted when the controlling NodeSet is
// deleted or scaled down.
type PersistentVolumeClaimRetentionPolicyType string

const (
	// RetainPersistentVolumeClaimRetentionPolicyType is the default
	// PersistentVolumeClaimRetentionPolicy and specifies that
	// PersistentVolumeClaims associated with NodeSet VolumeClaimTemplates
	// will not be deleted.
	RetainPersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Retain"
	// DeletePersistentVolumeClaimRetentionPolicyType specifies that
	// PersistentVolumeClaims associated with NodeSet VolumeClaimTemplates
	// will be deleted in the scenario specified in
	// NodeSetPersistentVolumeClaimPolicy.
	DeletePersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Delete"
)

// NodeSetPersistentVolumeClaimRetentionPolicy describes the policy used for PVCs
// created from the NodeSet VolumeClaimTemplates.
type NodeSetPersistentVolumeClaimRetentionPolicy struct {
	// WhenDeleted specifies what happens to PVCs created from NodeSet
	// VolumeClaimTemplates when the NodeSet is deleted. The default policy
	// of `Retain` causes PVCs to not be affected by NodeSet deletion. The
	// `Delete` policy causes those PVCs to be deleted.
	WhenDeleted PersistentVolumeClaimRetentionPolicyType `json:"whenDeleted,omitempty"`
}

// NodeSetUpdateStrategyType is a string enumeration type that enumerates
// all possible update strategies for the NodeSet controller.
// +enum
type NodeSetUpdateStrategyType string

const (
	// RollingUpdateNodeSetStrategyType indicates that NodeSet pods will replace
	// the old pods by new ones using a rolling update method
	// (i.e replace pods on each node one after the other).
	RollingUpdateNodeSetStrategyType NodeSetUpdateStrategyType = "RollingUpdate"

	// OnDeleteNodeSetStrategyType indicates that NodeSet pods will only be
	// replaced when the old pod is killed for any reason.
	OnDeleteNodeSetStrategyType NodeSetUpdateStrategyType = "OnDelete"
)

// RollingUpdateNodeSetStrategy is used to communicate parameters for
// RollingUpdateNodeSetStrategyType.
type RollingUpdateNodeSetStrategy struct {
	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up. This can not be 0.
	// Defaults to 1. This field is alpha-level and is only honored by servers that enable the
	// MaxUnavailableStatefulSet feature. The field applies to all pods in the range 0 to
	// Replicas-1. That means if there is any unavailable pod in the range 0 to Replicas-1, it
	// will be counted towards MaxUnavailable.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`

	// The number of NodeSet pods remained to be old version.
	// Maximum value is status.desiredNumberScheduled, which means no pod will be updated.
	// Default value is 0.
	// +optional
	Partition *int32 `json:"partition,omitempty"`

	// Indicates that the nodeset is paused and will not be processed by the
	// nodeset controller.
	// +optional
	Paused *bool `json:"paused,omitempty"`
}

// NodeSetStatus defines the observed state of NodeSet
type NodeSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The number of NodeSet pods that are running and are supposed to be
	// running.
	CurrentNumberScheduled int32 `json:"currentNumberScheduled"`

	// The number of NodeSet pods that are running but are not supposed to be
	// running.
	NumberMisscheduled int32 `json:"numberMisscheduled"`

	// The total number of NodeSet pods that should be running, including
	// NodeSet pods that are correctly running.
	DesiredNumberScheduled int32 `json:"desiredNumberScheduled"`

	// The number of NodeSet pods that are running, supposed to be running, and
	// have a Ready Condition.
	NumberReady int32 `json:"numberReady"`

	// The total number of NodeSet pods that are running, supposed to be
	// running, and are running updated pods.
	// +optional
	UpdatedNumberScheduled int32 `json:"updatedNumberScheduled,omitempty"`

	// The number of nodes that should be running the
	// daemon pod and have one or more of the daemon pod running and
	// available (ready for at least spec.minReadySeconds)
	// +optional
	NumberAvailable int32 `json:"numberAvailable,omitempty"`

	// The number of nodes that should be running the
	// daemon pod and have none of the daemon pod running and available
	// (ready for at least spec.minReadySeconds)
	// +optional
	NumberUnavailable int32 `json:"numberUnavailable,omitempty"`

	// The number of NodeSet pods that are running and are in the Slurm IDLE
	// state. IDLE means the Slurm nodes is not ALLOCATED or MIXED, hence is not
	// allocated any Slurm jobs, nor doing work.
	// +optional
	NumberIdle int32 `json:"numberIdle,omitempty"`

	// The number of NodeSet pods that are running and are in the Slurm
	// ALLOCATED or MIXED state. ALLOCATED/MIXED means the Slurm node is
	// allocated one or more Slurm jobs and is doing work.
	// +optional
	NumberAllocated int32 `json:"numberAllocated,omitempty"`

	// The number of NodeSet pods that are running and are in the Slurm
	// DOWN state. DOWN means the Slurm node is unavailable for use.
	// +optional
	NumberDown int32 `json:"numberDown,omitempty"`

	// The number of NodeSet pods that are running and are in the Slurm DRAIN
	// state. DRAIN means the Slurm node becomes unschedulable but allocated
	// Slurm jobs will not be evicted and can continue running until completion.
	// +optional
	NumberDrain int32 `json:"numberDrain,omitempty"`

	// observedGeneration is the most recent generation observed for this NodeSet. It corresponds to the
	// NodeSet's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// NodeSetHash is the "controller-revision-hash", which represents the
	// latest version of the NodeSet.
	NodeSetHash string `json:"nodeSetHash"`

	// Count of hash collisions for the NodeSet. The NodeSet controller
	// uses this field as a collision avoidance mechanism when it needs to
	// create the name for the newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`

	// Represents the latest available observations of a NodeSet's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Add Selector to status for HPA support in the scale subresource.
	Selector string `json:"selector"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=nodesets;nss
//+kubebuilder:subresource:scale:specpath=".spec.replicas",statuspath=".status.currentNumberScheduled",selectorpath=".status.selector"
//+kubebuilder:printcolumn:name="DESIRED",type="integer",JSONPath=".status.desiredNumberScheduled",priority=0,description="The desired number of pods."
//+kubebuilder:printcolumn:name="CURRENT",type="integer",JSONPath=".status.currentNumberScheduled",priority=0,description="The current number of pods."
//+kubebuilder:printcolumn:name="UPDATED",type="integer",JSONPath=".status.updatedNumberScheduled",priority=0,description="The number of pods updated."
//+kubebuilder:printcolumn:name="READY",type="integer",JSONPath=".status.numberAvailable",priority=0,description="The number of pods ready."
//+kubebuilder:printcolumn:name="IDLE",type="integer",JSONPath=".status.numberIdle",priority=1,description="The number of IDLE slurm nodes."
//+kubebuilder:printcolumn:name="ALLOCATED",type="integer",JSONPath=".status.numberAllocated",priority=1,description="The number of ALLOCATED/MIXED slurm nodes."
//+kubebuilder:printcolumn:name="DOWN",type="integer",JSONPath=".status.numberDown",priority=1,description="The number of DOWN slurm nodes."
//+kubebuilder:printcolumn:name="DRAIN",type="integer",JSONPath=".status.numberDrain",priority=1,description="The number of DRAIN slurm nodes."
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",priority=0,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// NodeSet is the Schema for the nodesets API
type NodeSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeSetSpec   `json:"spec,omitempty"`
	Status NodeSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeSetList contains a list of NodeSet
type NodeSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeSet{}, &NodeSetList{})
}
