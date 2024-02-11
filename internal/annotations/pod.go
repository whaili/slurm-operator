// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package annotations

const (
	// PodProgressiveCreate indicates nodeset pods created in manage phase will
	// be controlled by partition.
	// This annotation will be added to NodeSet when it is created, and removed
	// if partition is set to 0.
	PodProgressiveCreate = "slinky.slurm.net/pod-progressive-create"

	// PodDeletionCost can be used to set to an int32 that represents the cost
	// of deleting a NodeSet Pod compared to other pods belonging to the same
	// NodeSet.
	// Note that this is honored on a best-effort basis, and so it does not
	// offer guarantees on pod deletion order.
	// Ref: https://kubernetes.io/docs/reference/labels-annotations-taints/#pod-deletion-cost
	PodDeletionCost = "slinky.slurm.net/pod-deletion-cost"

	// PodCordon indicates NodeSet Pods that should be DRAIN[ING|ED] in Slurm.
	PodCordon = "slinky.slurm.net/pod-cordon"

	// PodDelete indicates NodeSet Pods that should be deleted in Kubernetes and
	// ignored in Slurm. Usually because the Slurm Node no longer exists.
	PodDelete = "slinky.slurm.net/pod-delete"
)
