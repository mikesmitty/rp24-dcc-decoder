#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PCB="$1"
PCB_DIR="${SCRIPT_DIR}/${PCB}"
OUTPUT=${PCB_DIR}/production

if [ -z "$PCB" ]; then
    echo "Usage: $0 <pcb-name> [panel-option]"
    exit 1
fi

if [ -n "$2" ]; then
    BATCH_MOD="-p ${PCB_DIR}/panelize-${2}.json"
fi

mkdir -p "$OUTPUT"
if [ -e "${PCB_DIR}/panelize.json" ]; then
    kikit panelize -p "${PCB_DIR}/panelize.json" $BATCH_MOD -p :jlcTooling "${PCB_DIR}/${PCB}.kicad_pcb" "${OUTPUT}/${PCB}.kicad_pcb"
    kikit fab jlcpcb --no-drc --assembly --autoname --schematic "${PCB_DIR}/${PCB}.kicad_sch" "${OUTPUT}/${PCB}.kicad_pcb" "$OUTPUT/"
else
    kikit fab jlcpcb --no-drc --assembly --autoname --schematic "${PCB_DIR}/${PCB}.kicad_sch" "${PCB_DIR}/${PCB}.kicad_pcb" "$OUTPUT/"
fi
