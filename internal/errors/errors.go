// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
)

// ErrorReason is an enumeration of possible failure causes.
// +enum
type ErrorReason string

const (
	// ErrorReasonNodeNotDrained means the node is not drained yet.
	ErrorReasonNodeNotDrained ErrorReason = "NodeNotDrained"
)

func New(errorReason ErrorReason) error {
	return errors.New(string(errorReason))
}

func IsNodeNotDrained(err error) bool {
	if err != nil {
		return err.Error() == string(ErrorReasonNodeNotDrained)
	}
	return false
}
