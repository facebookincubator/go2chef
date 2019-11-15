#!/bin/bash

if [[ -f contents ]]; then
        echo "found contents!"
        exit 0
else
        echo "didn't find contents!"
        exit 1
fi
