#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PCB="$1"
PCB_DIR="${SCRIPT_DIR}/${PCB}"
OUTPUT=${PCB_DIR}/production

if [ -z "$PCB" ]; then
    echo "Usage: $0 <pcb-name>"
    exit 1
fi

mkdir -p "$OUTPUT"
kikit fab jlcpcb --no-drc --assembly --schematic "${PCB_DIR}/${PCB}.kicad_sch" "${PCB_DIR}/${PCB}.kicad_pcb" "$OUTPUT/"
