#!/usr/bin/env bash

REPO_DIR="$(git rev-parse --show-toplevel)"
PROJECT=$1
LCSC_ID=$2

if [ -z "$LCSC_ID" ]; then
    echo "Usage: $0 <project> <lcsc_id>"
    exit 1
fi

if [ ! -d "${REPO_DIR}/hardware/${PROJECT}" ]; then
    echo "Project $PROJECT does not exist"
    exit 1
fi

cd "${REPO_DIR}/hardware/${PROJECT}"
easyeda2kicad --full --project-relative --output ./$PROJECT --lcsc_id $LCSC_ID
