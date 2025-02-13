#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PCB="$1"
PCB_DIR="${SCRIPT_DIR}/${PCB}"
OUTPUT=${PCB_DIR}/production

if [ -z "$PCB" ]; then
    echo "Usage: $0 <pcb-name> [small]"
    exit 1
fi

if [ -n "$2" ]; then
    BATCH_MOD="-p ${PCB_DIR}/panelize-${2}.json"
fi

mkdir -p "$OUTPUT"
kikit panelize -p "${PCB_DIR}/panelize.json" $BATCH_MOD -p :jlcTooling "${PCB_DIR}/${PCB}.kicad_pcb" "$OUTPUT/panel.kicad_pcb"
kikit fab jlcpcb --no-drc --assembly --schematic "${PCB_DIR}/${PCB}.kicad_sch" "${OUTPUT}/panel.kicad_pcb" "$OUTPUT/"
