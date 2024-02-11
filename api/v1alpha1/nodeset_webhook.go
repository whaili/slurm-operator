// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// log is for logging in this package.
var nodesetlog = logf.Log.WithName("nodeset-resource")

func (r *NodeSet) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-slinky-slurm-net-v1alpha1-nodeset,mutating=true,failurePolicy=fail,sideEffects=None,groups=slinky.slurm.net,resources=nodesets,verbs=create;update,versions=v1alpha1,name=mnodeset.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &NodeSet{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *NodeSet) Default() {
	nodesetlog.Info("default", "name", r.Name)

	if r.Spec.RevisionHistoryLimit == nil {
		r.Spec.RevisionHistoryLimit = ptr.To[int32](0)
	}
	if r.Spec.UpdateStrategy.Type == "" {
		r.Spec.UpdateStrategy.Type = RollingUpdateNodeSetStrategyType
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-slinky-slurm-net-v1alpha1-nodeset,mutating=false,failurePolicy=fail,sideEffects=None,groups=slinky.slurm.net,resources=nodesets,verbs=create;update,versions=v1alpha1,name=vnodeset.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &NodeSet{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NodeSet) ValidateCreate() (admission.Warnings, error) {
	nodesetlog.Info("validate create", "name", r.Name)

	warns, errs := validateNodeSet(r)

	return warns, utilerrors.NewAggregate(errs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NodeSet) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	nodesetlog.Info("validate update", "name", r.Name)

	warns, errs := validateNodeSet(r)

	oldObj := old.(*NodeSet)
	errMsgStub := "Mutatable fields include: 'Replicas', 'Selector', 'RevisionHistoryLimit', 'UpdateStrategy', 'PersistentVolumeClaimRetentionPolicy', 'MinReadySeconds'"
	if r.Spec.ClusterName != oldObj.Spec.ClusterName {
		errs = append(errs, fmt.Errorf("updates to `NodeSet.Spec.ClusterName` is forbidden. %v", errMsgStub))
	}
	if r.Spec.ServiceName != oldObj.Spec.ServiceName {
		errs = append(errs, fmt.Errorf("updates to `NodeSet.Spec.ServiceName` is forbidden. %v", errMsgStub))
	}
	if !apiequality.Semantic.DeepEqual(r.Spec.VolumeClaimTemplates, oldObj.Spec.VolumeClaimTemplates) {
		errs = append(errs, fmt.Errorf("updates to `NodeSet.Spec.VolumeClaimTemplates` is forbidden. %v", errMsgStub))
	}

	return warns, utilerrors.NewAggregate(errs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NodeSet) ValidateDelete() (admission.Warnings, error) {
	nodesetlog.Info("validate delete", "name", r.Name)

	return nil, nil
}

func validateNodeSet(r *NodeSet) (admission.Warnings, []error) {
	var warns admission.Warnings
	var errs []error

	if r.Spec.ServiceName == "" {
		errs = append(errs, fmt.Errorf("`NodeSet.Spec.ServiceName` cannot be empty"))
	}

	switch r.Spec.UpdateStrategy.Type {
	case RollingUpdateNodeSetStrategyType:
		// valid
	case OnDeleteNodeSetStrategyType:
		// valid
	default:
		errs = append(errs, fmt.Errorf("`NodeSet.Spec.UpdateStrategy.Type` is not valid. Got: %v. Expected of: %s; %s",
			r.Spec.UpdateStrategy.Type, RollingUpdateNodeSetStrategyType, OnDeleteNodeSetStrategyType))
	}

	if r.Spec.PersistentVolumeClaimRetentionPolicy != nil {
		switch r.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted {
		case RetainPersistentVolumeClaimRetentionPolicyType:
			// valid
		case DeletePersistentVolumeClaimRetentionPolicyType:
			// valid
		default:
			errs = append(errs, fmt.Errorf("`NodeSet.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted` is not valid. Got: %v. Expected of: %s; %s",
				r.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted, RetainPersistentVolumeClaimRetentionPolicyType, DeletePersistentVolumeClaimRetentionPolicyType))
		}
	}

	return warns, errs
}
