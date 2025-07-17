// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	_ "embed"
	"strings"
	"testing"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/builder/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/set"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuilder_BuildComputePodTemplate(t *testing.T) {
	type fields struct {
		client client.Client
	}
	type args struct {
		nodeset    *slinkyv1alpha1.NodeSet
		controller *slinkyv1alpha1.Controller
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "default",
			fields: fields{
				client: fake.NewFakeClient(),
			},
			args: args{
				nodeset: &slinkyv1alpha1.NodeSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "slurm-foo",
					},
					Spec: slinkyv1alpha1.NodeSetSpec{
						Template: slinkyv1alpha1.NodeSetPodTemplate{
							ExtraConf: strings.Join([]string{
								"features=bar",
								"weight=5",
							}, " "),
							PodTemplate: slinkyv1alpha1.PodTemplate{
								PodSpec: slinkyv1alpha1.PodSpec{
									Hostname: "foo-",
								},
							},
						},
					},
					Status: slinkyv1alpha1.NodeSetStatus{
						Selector: k8slabels.SelectorFromSet(k8slabels.Set(labels.NewBuilder().WithComputeSelectorLabels(&slinkyv1alpha1.NodeSet{ObjectMeta: metav1.ObjectMeta{Name: "slurm"}}).Build())).String(),
					},
				},
				controller: &slinkyv1alpha1.Controller{
					ObjectMeta: metav1.ObjectMeta{
						Name: "slurm",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.fields.client)
			got := b.BuildComputePodTemplate(tt.args.nodeset, tt.args.controller)
			selector, err := k8slabels.ConvertSelectorToLabelsMap(tt.args.nodeset.Status.Selector)
			if err != nil {
				t.Errorf("ConvertSelectorToLabelsMap() = %v", err)
			}
			switch {
			case !set.KeySet(got.Labels).HasAll(set.KeySet(selector).UnsortedList()...):
				t.Errorf("Labels = %v , Selector = %v", got.Labels, selector)

			case got.Spec.Containers[0].Name != labels.ComputeApp:
				t.Errorf("Containers[0].Name = %v , want = %v",
					got.Spec.Containers[0].Name, labels.ComputeApp)

			case got.Spec.Containers[0].Ports[0].Name != labels.ComputeApp:
				t.Errorf("Containers[0].Ports[0].Name = %v , want = %v",
					got.Spec.Containers[0].Ports[0].Name, labels.ComputeApp)

			case got.Spec.Containers[0].Ports[0].ContainerPort != SlurmdPort:
				t.Errorf("Containers[0].Ports[0].ContainerPort = %v , want = %v",
					got.Spec.Containers[0].Ports[0].Name, SlurmdPort)
			}
		})
	}
}
