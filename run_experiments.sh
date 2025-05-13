#!/bin/bash

goodagproportion=0.0
goodpt=0.9

for ((i=0; i<110; i++)); do
    if (( i % 10 == 0 && i != 0 )); then
        goodagproportion=$(awk "BEGIN {printf \"%.1f\", $goodagproportion + 0.1}")
    fi

    go run main.go -goodpt=$goodpt -goodagproportion=$goodagproportion
done
