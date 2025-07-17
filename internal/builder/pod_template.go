// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/metadata"
	"github.com/SlinkyProject/slurm-operator/internal/utils"
)

type PodTemplateOpts struct {
	Key      types.NamespacedName
	Metadata slinkyv1alpha1.Metadata
	base     corev1.PodSpec
	merge    corev1.PodSpec
}

func (b *Builder) buildPodTemplate(opts PodTemplateOpts) corev1.PodTemplateSpec {
	objectMeta := metadata.NewBuilder(opts.Key).
		WithMetadata(opts.Metadata).
		Build()

	o := corev1.PodTemplateSpec{
		ObjectMeta: objectMeta,
		Spec:       opts.base,
	}

	o.Spec.Volumes = utils.MergeList(o.Spec.Volumes, opts.merge.Volumes)
	o.Spec.InitContainers = utils.MergeList(o.Spec.InitContainers, opts.merge.InitContainers)
	o.Spec.Containers = utils.MergeList(o.Spec.Containers, opts.merge.Containers)
	o.Spec.EphemeralContainers = utils.MergeList(o.Spec.EphemeralContainers, opts.merge.EphemeralContainers)
	o.Spec.NodeSelector = utils.MergeMaps(o.Spec.NodeSelector, opts.merge.NodeSelector)
	o.Spec.ImagePullSecrets = utils.MergeList(o.Spec.ImagePullSecrets, opts.merge.ImagePullSecrets)
	o.Spec.Tolerations = utils.MergeList(o.Spec.Tolerations, opts.merge.Tolerations)
	o.Spec.ReadinessGates = utils.MergeList(o.Spec.ReadinessGates, opts.merge.ReadinessGates)
	o.Spec.TopologySpreadConstraints = utils.MergeList(o.Spec.TopologySpreadConstraints, opts.merge.TopologySpreadConstraints)
	o.Spec.SchedulingGates = utils.MergeList(o.Spec.SchedulingGates, opts.merge.SchedulingGates)
	o.Spec.ResourceClaims = utils.MergeList(o.Spec.ResourceClaims, opts.merge.ResourceClaims)

	if opts.merge.RestartPolicy != "" {
		o.Spec.RestartPolicy = opts.merge.RestartPolicy
	}
	if opts.merge.TerminationGracePeriodSeconds != nil {
		o.Spec.TerminationGracePeriodSeconds = opts.merge.TerminationGracePeriodSeconds
	}
	if opts.merge.ActiveDeadlineSeconds != nil {
		o.Spec.ActiveDeadlineSeconds = opts.merge.ActiveDeadlineSeconds
	}
	if opts.merge.DNSPolicy != "" {
		o.Spec.DNSPolicy = opts.merge.DNSPolicy
	}
	if opts.merge.ServiceAccountName != "" {
		o.Spec.ServiceAccountName = opts.merge.ServiceAccountName
	}
	if opts.merge.DeprecatedServiceAccount != "" {
		o.Spec.DeprecatedServiceAccount = opts.merge.DeprecatedServiceAccount
	}
	if opts.merge.AutomountServiceAccountToken != nil {
		o.Spec.AutomountServiceAccountToken = opts.merge.AutomountServiceAccountToken
	}
	if opts.merge.NodeName != "" {
		o.Spec.NodeName = opts.merge.NodeName
	}
	if opts.merge.HostNetwork != o.Spec.HostNetwork {
		o.Spec.HostNetwork = opts.merge.HostNetwork
	}
	if opts.merge.HostPID != o.Spec.HostPID {
		o.Spec.HostPID = opts.merge.HostPID
	}
	if opts.merge.HostIPC != o.Spec.HostIPC {
		o.Spec.HostIPC = opts.merge.HostIPC
	}
	if opts.merge.ShareProcessNamespace != nil {
		o.Spec.ShareProcessNamespace = opts.merge.ShareProcessNamespace
	}
	if opts.merge.SecurityContext != nil {
		o.Spec.SecurityContext = opts.merge.SecurityContext
	}
	if opts.merge.Hostname != "" {
		o.Spec.Hostname = opts.merge.Hostname
	}
	if opts.merge.Subdomain != "" {
		o.Spec.Subdomain = opts.merge.Subdomain
	}
	if opts.merge.Affinity != nil {
		o.Spec.Affinity = opts.merge.Affinity
	}
	if opts.merge.SchedulerName != "" {
		o.Spec.SchedulerName = opts.merge.SchedulerName
	}
	if opts.merge.PriorityClassName != "" {
		o.Spec.PriorityClassName = opts.merge.PriorityClassName
	}
	if opts.merge.Priority != nil {
		o.Spec.Priority = opts.merge.Priority
	}
	if opts.merge.DNSConfig != nil {
		o.Spec.DNSConfig = opts.merge.DNSConfig
	}
	if opts.merge.RuntimeClassName != nil {
		o.Spec.RuntimeClassName = opts.merge.RuntimeClassName
	}
	if opts.merge.EnableServiceLinks != nil {
		o.Spec.EnableServiceLinks = opts.merge.EnableServiceLinks
	}
	if opts.merge.PreemptionPolicy != nil {
		o.Spec.PreemptionPolicy = opts.merge.PreemptionPolicy
	}
	if opts.merge.Overhead != nil {
		o.Spec.Overhead = opts.merge.Overhead
	}
	if opts.merge.SetHostnameAsFQDN != nil {
		o.Spec.SetHostnameAsFQDN = opts.merge.SetHostnameAsFQDN
	}
	if opts.merge.OS != nil {
		o.Spec.OS = opts.merge.OS
	}
	if opts.merge.HostUsers != nil {
		o.Spec.HostUsers = opts.merge.HostUsers
	}
	if opts.merge.Resources != nil {
		o.Spec.Resources = opts.merge.Resources
	}

	return o
}
