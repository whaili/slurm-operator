// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
)

func fqdn(name, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
}

func fqdnShort(name, namespace string) string {
	return fmt.Sprintf("%s.%s", name, namespace)
}
