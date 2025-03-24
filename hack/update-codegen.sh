#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail


SCRIPT_ROOT=$( cd "$(dirname "${BASH_SOURCE[0]}")/.." ; pwd -P )
CODEGEN_PKG=${CODEGEN_PKG:-$(
  go list -m -f "{{.Dir}}" k8s.io/code-generator
)}

function codegen::join() {
  local IFS="$1"
  shift
  echo "$*"
}

PKG_NAME="github.com/tlscert/backend"
OUTPUT_PKG="pkg/generated"
BOILERPLATE="${SCRIPT_ROOT}"/hack/boilerplate.go.txt

source "${CODEGEN_PKG}"/kube_codegen.sh

# go get sigs.k8s.io/controller-tools/cmd/controller-gen
go install \
    sigs.k8s.io/controller-tools/cmd/controller-gen \
    k8s.io/code-generator/cmd/deepcopy-gen \
    k8s.io/code-generator/cmd/defaulter-gen \
    k8s.io/code-generator/cmd/register-gen \
    k8s.io/code-generator/cmd/applyconfiguration-gen \
    k8s.io/code-generator/cmd/client-gen \
    k8s.io/code-generator/cmd/lister-gen \
    k8s.io/code-generator/cmd/informer-gen


echo "Generating helpers..." >&2
kube::codegen::gen_helpers --boilerplate "$BOILERPLATE" "$SCRIPT_ROOT"

echo "Generating scheme registration..." >&2
kube::codegen::gen_register --boilerplate "$BOILERPLATE" "${SCRIPT_ROOT}"

echo "Generating clientset..." >&2
kube::codegen::gen_client \
  --with-watch \
  --with-applyconfig \
  --output-dir "${SCRIPT_ROOT}/${OUTPUT_PKG}"\
  --output-pkg "${PKG_NAME}/${OUTPUT_PKG}" \
  --boilerplate "$BOILERPLATE" \
  "${SCRIPT_ROOT}"

# This should be covered by gen_helpers, but it won't do it for some reason
echo "Generating deepcopy..." >&2
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  object \
  paths="./..." \
  output:dir="${SCRIPT_ROOT}/api/v1alpha1"
echo "Deepcopy generated" >&2

pushd "${SCRIPT_ROOT}" >/dev/null

# Generate CRD manifests for all types using controller-gen
echo "Generating crd manifests..." >&2
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  crd \
  paths="./..." \
  output:dir="${SCRIPT_ROOT}/crds"
echo "CRD manifests generated" >&2

echo "Done" >&2
popd >/dev/null