// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package annotations

const (
	// AnnotationPodCordon indicates NodeSet Pods that should be DRAIN[ING|ED] in Slurm.
	PodCordon = "slinky.slurm.net/pod-cordon"

	// LabelPodDeletionCost can be used to set to an int32 that represent the cost of deleting
	// a pod compared to other pods belonging to the same ReplicaSet. Pods with lower
	// deletion cost are preferred to be deleted before pods with higher deletion cost.
	// Note that this is honored on a best-effort basis, and so it does not offer guarantees on
	// pod deletion order.
	// The implicit deletion cost for pods that don't set the annotation is 0, negative values are permitted.
	PodDeletionCost = "slinky.slurm.net/pod-deletion-cost"

	// PodDeadline stores a time.RFC3339 timestamp, indicating when the Slurm node should complete its running
	// workload by. Pods an earlier daedline are preferred to be deleted before pods with a later deadline.
	// NOTE: this is honored on a best-effort basis, and does not offer guarantees on pod deletion order.
	PodDeadline = "slinky.slurm.net/pod-deadline"
)
