#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT=$1
LCSC_ID=$2

if [ -z "$LCSC_ID" ]; then
    echo "Usage: $0 <project> <lcsc_id>"
    exit 1
fi

if [ ! -d "$PROJECT" ]; then
    echo "Project $PROJECT does not exist"
    exit 1
fi

cd "${SCRIPT_DIR}/${PROJECT}"
easyeda2kicad --full --project-relative --output ./$PROJECT --lcsc_id $LCSC_ID
