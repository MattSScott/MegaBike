#!/bin/bash

goodagpercent=0.0
goodpt=0.55

for ((i=1; i<=110; i++)); do
    if (( i % 10 == 1 && i != 1 )); then
        goodagpercent=$(awk "BEGIN {printf \"%.1f\", $goodagpercent + 0.1}")
    fi

    go run main.go -goodpt=$goodpt -goodagpercent=$goodagpercent
done
