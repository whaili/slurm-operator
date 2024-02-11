// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// log is for logging in this package.
var clusterlog = logf.Log.WithName("cluster-resource")

func (r *Cluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-slinky-slurm-net-v1alpha1-cluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=slinky.slurm.net,resources=clusters,verbs=create;update,versions=v1alpha1,name=mcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Cluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cluster) Default() {
	clusterlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-slinky-slurm-net-v1alpha1-cluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=slinky.slurm.net,resources=clusters,verbs=create;update,versions=v1alpha1,name=vcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Cluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateCreate() (admission.Warnings, error) {
	clusterlog.Info("validate create", "name", r.Name)

	warns, errs := validateCluster(r)

	return warns, utilerrors.NewAggregate(errs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	clusterlog.Info("validate update", "name", r.Name)

	warns, errs := validateCluster(r)

	return warns, utilerrors.NewAggregate(errs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateDelete() (admission.Warnings, error) {
	clusterlog.Info("validate delete", "name", r.Name)

	return nil, nil
}

func validateCluster(r *Cluster) (admission.Warnings, []error) {
	var warns admission.Warnings
	var errs []error

	if r.Spec.Server == "" {
		errs = append(errs, fmt.Errorf("`Cluster.Spec.Server` cannot be empty"))
	}
	if r.Spec.Token.SecretRef == "" {
		errs = append(errs, fmt.Errorf("`Cluster.Spec.Token.SecretRef` cannot be empty"))
	}

	return warns, errs
}
