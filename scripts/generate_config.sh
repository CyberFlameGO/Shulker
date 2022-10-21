#!/bin/bash

SCRIPTPATH="$(cd -- "$(dirname "$0")" >/dev/null 2>&1; pwd -P)"
ROOTDIR="${SCRIPTPATH}/.."

controller-gen rbac:roleName=manager-role crd webhook paths="${ROOTDIR}/..." output:crd:artifacts:config=${ROOTDIR}/config/crd/bases
controller-gen object:headerFile="${ROOTDIR}/hack/boilerplate.go.txt" paths="${ROOTDIR}/..."
