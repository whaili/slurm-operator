// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/labels"
	"github.com/SlinkyProject/slurm-operator/internal/builder/metadata"
	"github.com/SlinkyProject/slurm-operator/internal/utils"
)

const (
	SlurmrestdPort = 6820

	slurmrestdUser    = "nobody"
	slurmrestdUserUid = int64(65534)
	slurmrestdUserGid = slurmrestdUserUid
)

func (b *Builder) BuildRestapi(restapi *slinkyv1alpha1.RestApi) (*appsv1.Deployment, error) {
	key := restapi.Key()

	selectorLabels := labels.NewBuilder().
		WithRestapiSelectorLabels(restapi).
		Build()
	objectMeta := metadata.NewBuilder(key).
		WithMetadata(restapi.Spec.Template.PodMetadata).
		WithLabels(labels.NewBuilder().WithRestapiLabels(restapi).Build()).
		Build()

	podTemplate, err := b.restapiPodTemplate(restapi)
	if err != nil {
		return nil, fmt.Errorf("failed to build pod template: %w", err)
	}

	o := &appsv1.Deployment{
		ObjectMeta: objectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas:             restapi.Spec.Replicas,
			RevisionHistoryLimit: ptr.To[int32](0),
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: podTemplate,
		},
	}

	if err := controllerutil.SetControllerReference(restapi, o, b.client.Scheme()); err != nil {
		return nil, fmt.Errorf("failed to set owner controller: %w", err)
	}

	return o, nil
}

func (b *Builder) restapiPodTemplate(restapi *slinkyv1alpha1.RestApi) (corev1.PodTemplateSpec, error) {
	ctx := context.TODO()
	key := restapi.Key()

	controller, err := b.refResolver.GetController(ctx, restapi.Spec.ControllerRef)
	if err != nil {
		return corev1.PodTemplateSpec{}, err
	}

	hasAccounting := !apiequality.Semantic.DeepEqual(controller.Spec.AccountingRef, slinkyv1alpha1.ObjectReference{})

	objectMeta := metadata.NewBuilder(key).
		WithMetadata(restapi.Spec.Template.PodMetadata).
		WithLabels(labels.NewBuilder().WithRestapiLabels(restapi).Build()).
		WithAnnotations(map[string]string{
			annotationDefaultContainer: labels.RestapiApp,
		}).
		Build()

	spec := restapi.Spec
	template := spec.Template.PodSpecWrapper

	opts := PodTemplateOpts{
		Key: key,
		Metadata: slinkyv1alpha1.Metadata{
			Annotations: objectMeta.Annotations,
			Labels:      objectMeta.Labels,
		},
		base: corev1.PodSpec{
			AutomountServiceAccountToken: ptr.To(false),
			Affinity:                     template.Affinity,
			Containers: []corev1.Container{
				slurmrestdContainer(spec.Slurmrestd, hasAccounting),
			},
			Hostname:          template.Hostname,
			ImagePullSecrets:  template.ImagePullSecrets,
			NodeSelector:      template.NodeSelector,
			PriorityClassName: template.PriorityClassName,
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: ptr.To(true),
				RunAsUser:    ptr.To(slurmrestdUserUid),
				RunAsGroup:   ptr.To(slurmrestdUserGid),
				FSGroup:      ptr.To(slurmrestdUserGid),
			},
			Tolerations: template.Tolerations,
			Volumes:     utils.MergeList(restapiVolumes(controller), template.Volumes),
		},
		merge: template.PodSpec,
	}

	o := b.buildPodTemplate(opts)

	return o, nil
}

func restapiVolumes(controller *slinkyv1alpha1.Controller) []corev1.Volume {
	out := []corev1.Volume{
		{
			Name: slurmEtcVolume,
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					DefaultMode: ptr.To[int32](0o600),
					Sources: []corev1.VolumeProjection{
						{
							ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: controller.ConfigKey().Name,
								},
								Items: []corev1.KeyToPath{
									{Key: slurmConfFile, Path: slurmConfFile},
								},
							},
						},
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: controller.AuthSlurmRef().Name,
								},
								Items: []corev1.KeyToPath{
									{Key: controller.AuthSlurmRef().Key, Path: slurmKeyFile},
								},
							},
						},
					},
				},
			},
		},
	}
	return out
}

func slurmrestdContainer(containerWrapper slinkyv1alpha1.ContainerWrapper, hasAccounting bool) corev1.Container {
	container := containerWrapper.Container
	out := corev1.Container{
		Name:            labels.RestapiApp,
		Env:             slurmrestEnv(containerWrapper),
		Args:            slurmrestArgs(containerWrapper, hasAccounting),
		Image:           container.Image,
		ImagePullPolicy: container.ImagePullPolicy,
		Ports: []corev1.ContainerPort{
			{
				Name:          labels.RestapiApp,
				ContainerPort: SlurmrestdPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Resources: container.Resources,
		StartupProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(SlurmrestdPort),
				},
			},
			FailureThreshold: 6,
			PeriodSeconds:    10,
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(SlurmrestdPort),
				},
			},
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: ptr.To(true),
			RunAsUser:    ptr.To(slurmrestdUserUid),
			RunAsGroup:   ptr.To(slurmrestdUserGid),
		},
		VolumeDevices: container.VolumeDevices,
		VolumeMounts: []corev1.VolumeMount{
			{Name: slurmEtcVolume, MountPath: slurmEtcDir, ReadOnly: true},
		},
	}
	out.VolumeMounts = append(out.VolumeMounts, container.VolumeMounts...)
	return out
}

func slurmrestArgs(containerWrapper slinkyv1alpha1.ContainerWrapper, hasAccounting bool) []string {
	container := containerWrapper.Container
	args := container.Args
	if !hasAccounting {
		args = append(args, "-s")
		args = append(args, "openapi/slurmctld")
	}
	args = append(args, fmt.Sprintf("0.0.0.0:%d", SlurmrestdPort))
	return args
}

func slurmrestEnv(containerWrapper slinkyv1alpha1.ContainerWrapper) []corev1.EnvVar {
	container := containerWrapper.Container
	options := []string{
		"disable_unshare_files",
		"disable_unshare_sysv",
	}
	env := []corev1.EnvVar{
		{Name: "SLURM_JWT", Value: "daemon"},
		{Name: "SLURMRESTD_SECURITY", Value: strings.Join(options, ",")},
	}
	return mergeEnvVar(container.Env, env, ",")
}
