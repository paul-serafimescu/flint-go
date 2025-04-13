#!/usr/bin/env bash

set -euo pipefail

mkdir -p generated_data/flat/appA
mkdir -p generated_data/flat/appB

A_FILE="generated_data/flat/appA/events.log.jsonl"
B_FILE="generated_data/flat/appB/events.log.jsonl"

echo "Checking if appA and appB data files exist..."

# Generate appA only if doesn't exist
if [[ ! -f "$A_FILE" ]]; then
    echo "Generating $A_FILE (200MB)..."
    python3 scripts/util/generate_jsonl.py --filename "$A_FILE" --size 200 &
else
    echo "$A_FILE already exists, skipping."
fi

# Generate appB only if doesn't exist
if [[ ! -f "$B_FILE" ]]; then
    echo "Generating $B_FILE (500MB)..."
    python3 scripts/util/generate_jsonl.py --filename "$B_FILE" --size 500
else
    echo "$B_FILE already exists, skipping."
fi

wait


rm -rf generated_data/distributed
mkdir -p generated_data/distributed

echo -e "\nGenerating distributed filesystem from flat data..."
python3 scripts/util/generate_fs.py \
    --in_dir 'generated_data/flat' \
    --out_dir 'generated_data/distributed' \
    --partitions 3 \
    --copies 2 \
    --chunk-size 64
