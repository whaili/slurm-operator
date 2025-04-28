#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

function main() {
	echo "[$(date --rfc-3339="seconds")] START"

	# Reattempt reconfigure until successful
	until scontrol reconfigure; do
		sleep 2
	done

	# Record completion data
	echo "[$(date --rfc-3339="seconds")] DONE"
}
main
