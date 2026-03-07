#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 2 ]]; then
  echo "Usage: $0 <version> <artifact-dir>"
  exit 1
fi

version="$1"
artifact_dir="$2"

if [[ ! -d "$artifact_dir" ]]; then
  echo "ERROR: artifact directory does not exist: $artifact_dir"
  exit 1
fi

expected=(
  "prr_${version}_darwin_arm64"
  "prr_${version}_linux_amd64"
  "prr_${version}_linux_arm64"
  "prr_${version}_windows_amd64.exe"
)

missing=0
for artifact in "${expected[@]}"; do
  if [[ ! -f "$artifact_dir/$artifact" ]]; then
    echo "ERROR: missing expected artifact: $artifact"
    missing=1
  fi
done

if [[ "$missing" -ne 0 ]]; then
  echo "Artifacts found in $artifact_dir:"
  ls -la "$artifact_dir"
  exit 1
fi

found_count="$(find "$artifact_dir" -maxdepth 1 -type f -name "prr_${version}_*" | wc -l | tr -d ' ')"
if [[ "$found_count" -ne 4 ]]; then
  echo "ERROR: expected exactly 4 release artifacts for version $version, found $found_count"
  echo "Artifacts found in $artifact_dir:"
  ls -la "$artifact_dir"
  exit 1
fi

echo "Artifact contract verified for $version"
printf '%s\n' "${expected[@]}"
