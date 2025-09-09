// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package conditions

import (
	corev1 "k8s.io/api/core/v1"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

const (
	StatePrefix = "SlurmNodeState"

	// Slurm Base States
	PodConditionAllocated corev1.PodConditionType = StatePrefix + "Allocated"
	PodConditionDown      corev1.PodConditionType = StatePrefix + "Down"
	PodConditionError     corev1.PodConditionType = StatePrefix + "Error"
	PodConditionFuture    corev1.PodConditionType = StatePrefix + "Future"
	PodConditionIdle      corev1.PodConditionType = StatePrefix + "Idle"
	PodConditionMixed     corev1.PodConditionType = StatePrefix + "Mixed"
	PodConditionUnknown   corev1.PodConditionType = StatePrefix + "Unknown"

	// Slurm Flag States
	PodConditionCompleting    corev1.PodConditionType = StatePrefix + "Completing"
	PodConditionDrain         corev1.PodConditionType = StatePrefix + "Drain"
	PodConditionFail          corev1.PodConditionType = StatePrefix + "Fail"
	PodConditionInvalid       corev1.PodConditionType = StatePrefix + "Invalid"
	PodConditionInvalidReg    corev1.PodConditionType = StatePrefix + "InvalidReg"
	PodConditionMaintenance   corev1.PodConditionType = StatePrefix + "Maintenance"
	PodConditionNotResponding corev1.PodConditionType = StatePrefix + "NotResponding"
	PodConditionUndrain       corev1.PodConditionType = StatePrefix + "Undrain"
)

func IsConditionTrue(status *corev1.PodStatus, condType corev1.PodConditionType) bool {
	_, cond := podutil.GetPodCondition(status, condType)
	return cond != nil && cond.Status == corev1.ConditionTrue
}

func IsNodeBusy(status *corev1.PodStatus) bool {
	isBusy := IsConditionTrue(status, PodConditionAllocated) ||
		IsConditionTrue(status, PodConditionMixed)
	return isBusy || IsConditionTrue(status, PodConditionCompleting)
}

func IsNodeDrained(status *corev1.PodStatus) bool {
	return IsNodeDrain(status) && !IsNodeBusy(status)
}

func IsNodeDraining(status *corev1.PodStatus) bool {
	return IsNodeDrain(status) && IsNodeBusy(status)
}

func IsNodeDrain(status *corev1.PodStatus) bool {
	return IsConditionTrue(status, PodConditionDrain) &&
		!IsConditionTrue(status, PodConditionUndrain)
}
