// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuilder_BuildControllerScripts(t *testing.T) {
	type fields struct {
		client client.Client
	}
	type args struct {
		controller *slinkyv1alpha1.Controller
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				client: fake.NewFakeClient(),
			},
			args: args{
				controller: &slinkyv1alpha1.Controller{
					ObjectMeta: metav1.ObjectMeta{
						Name: "slurm",
					},
					Spec: slinkyv1alpha1.ControllerSpec{
						PrologScripts: map[string]string{
							"00-exit.sh": "#!/usr/bin/sh\nexit0",
						},
						EpilogScripts: map[string]string{
							"00-exit.sh": "#!/usr/bin/sh\nexit0",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.fields.client)
			got, err := b.BuildControllerScripts(tt.args.controller)
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.BuildControllerScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch {
			case err != nil:
				return

			case len(got.Data)+len(got.BinaryData) != len(tt.args.controller.Spec.PrologScripts)+len(tt.args.controller.Spec.EpilogScripts):
				t.Errorf("len(got.Data) = %d , len(got.BinaryData) = %d , len(PrologScripts)+len(EpilogScripts) = %d", len(got.Data), len(got.BinaryData), len(tt.args.controller.Spec.PrologScripts)+len(tt.args.controller.Spec.EpilogScripts))
			}
			got2, err := b.BuildControllerConfig(tt.args.controller)
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.BuildControllerScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			slurmConf := strings.Split(got2.Data[slurmConfFile], "\n")
			for filename := range got.Data {
				found := slices.ContainsFunc(slurmConf, func(item string) bool {
					match1, _ := regexp.Match(fmt.Sprintf("^Prolog=%s", filename), []byte(item))
					match2, _ := regexp.Match(fmt.Sprintf("^Epilog=%s", filename), []byte(item))
					return match1 || match2
				})
				if !found {
					t.Errorf("%s is missing Prolog/Epilog configuration", slurmConfFile)
				}
			}
		})
	}
}
