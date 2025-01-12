#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PCB="$(basename $SCRIPT_DIR)"
OUTPUT=$SCRIPT_DIR/production

if [ $1 == "small" ]; then
    BATCH_MOD="-p ${SCRIPT_DIR}/panelize-small.json"
fi

mkdir -p "$OUTPUT"
kikit panelize -p "${SCRIPT_DIR}/panelize.json" $BATCH_MOD -p :jlcTooling "${SCRIPT_DIR}/${PCB}.kicad_pcb" "$OUTPUT/panel.kicad_pcb"
kikit fab jlcpcb --no-drc --assembly --schematic "${SCRIPT_DIR}/${PCB}.kicad_sch" "${OUTPUT}/panel.kicad_pcb" "$OUTPUT/"
