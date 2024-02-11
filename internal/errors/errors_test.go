// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		errorReason ErrorReason
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "error return",
			args: args{
				errorReason: ErrorReasonNodeNotDrained,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := New(tt.args.errorReason); (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsNodeNotDrained(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil error",
			args: args{
				err: nil,
			},
			want: false,
		},
		{
			name: "is not drained error",
			args: args{
				err: errors.New(string(ErrorReasonNodeNotDrained)),
			},
			want: true,
		},
		{
			name: "wrong error",
			args: args{
				err: errors.New(string("foo")),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNodeNotDrained(tt.args.err); got != tt.want {
				t.Errorf("IsNodeNotDrained() = %v, want %v", got, tt.want)
			}
		})
	}
}
