#!/usr/bin/env bash

REPO_DIR="$(git rev-parse --show-toplevel)"
PROJECT=$1

if [ -z "$PROJECT" ] || [ ! -d "${REPO_DIR}/hardware/${PROJECT}" ]; then
    echo "Project '$PROJECT' does not exist"
    exit 1
fi

cd "${REPO_DIR}/hardware/${PROJECT}"
kicad-cli pcb render --quality high --background opaque --perspective --rotate "'-15,20,15'" -o "${REPO_DIR}/images/${PROJECT}-top.png" "${PROJECT}.kicad_pcb"
kicad-cli pcb render --quality high --background opaque --perspective --rotate "'-15,200,-15'" -o "${REPO_DIR}/images/${PROJECT}-bottom.png" "${PROJECT}.kicad_pcb"