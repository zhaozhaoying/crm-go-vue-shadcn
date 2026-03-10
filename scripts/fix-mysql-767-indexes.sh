#!/usr/bin/env bash
set -euo pipefail

input_path="${1:-}"
output_path="${2:-}"

if [[ -z "${input_path}" || -z "${output_path}" ]]; then
  echo "usage: $0 INPUT.sql OUTPUT.sql" >&2
  exit 1
fi

if [[ ! -f "${input_path}" ]]; then
  echo "input file not found: ${input_path}" >&2
  exit 1
fi

perl -0pe '
  s/`notification_key` varchar\(255\) NOT NULL/`notification_key` varchar(150) NOT NULL/g;
  s/`token_hash` varchar\(255\) NOT NULL/`token_hash` varchar(64) NOT NULL/g;
  s/`replaced_by_hash` varchar\(255\) NOT NULL DEFAULT '\'''\''/`replaced_by_hash` varchar(64) NOT NULL DEFAULT '\'''\''/g;
  s/`source_uid` varchar\(255\) NOT NULL DEFAULT '\'''\''/`source_uid` varchar(150) NOT NULL DEFAULT '\'''\''/g;
  s/KEY `idx_resource_pool_name` \(`name`\)/KEY `idx_resource_pool_name` (`name`(191))/g;
  s/`jti` varchar\(255\) NOT NULL/`jti` varchar(64) NOT NULL/g;
' "${input_path}" > "${output_path}"

echo "wrote compatible dump: ${output_path}"
