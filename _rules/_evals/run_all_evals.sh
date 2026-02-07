#!/bin/bash

# BeadsLog Master Eval Runner
# Usage: ./run_all_evals.sh <results_dir>

RESULTS_DIR=${1:-"_rules/_evals/results"}
mkdir -p "$RESULTS_DIR"

echo "### BeadsLog Always-On Protocol Evaluation ###"
echo "Results will be saved to: $RESULTS_DIR"
echo ""

# Summary file
SUMMARY_FILE="$RESULTS_DIR/summary.jsonl"
rm -f "$SUMMARY_FILE"

# Iterate over all scenarios in scenarios.json
scenarios=$(grep -o '"id": "[^"]*"' _rules/_evals/scenarios.json | cut -d'"' -f4)

for id in $scenarios; do
    echo "Processing Scenario $id..."
    
    # Try to find a trace file for this ID if it exists
    trace_file="_rules/_evals/traces/trace_$id.log"
    if [ ! -f "$trace_file" ]; then
        # Use dummy/empty trace if not found
        trace_file="/tmp/empty_trace.log"
        touch "$trace_file"
    fi
    
    # Run eval.go
    go run _rules/_evals/eval.go --id "$id" --trace "$trace_file" >> "$SUMMARY_FILE"
done

echo ""
echo "### EVALUATION COMPLETE ###"

# Generate human-readable table if possible
if command -v column &> /dev/null; then
    echo ""
    echo "### PROTOCOL PERFORMANCE & TOKEN IMPACT ###"
    printf "ID\tPASS\tSCORE\tEST_TOKENS\tSAVINGS\tPROMPT\n"
    while read -r line; do
        id=$(echo "$line" | sed -n 's/.*"scenario_id":"\([^"]*\)".*/\1/p')
        pass=$(echo "$line" | sed -n 's/.*"pass":\([^,]*\),.*/\1/p')
        score=$(echo "$line" | sed -n 's/.*"score":\([^,]*\),.*/\1/p')
        tokens=$(echo "$line" | sed -n 's/.*"estimated_tokens":\([0-9]*\).*/\1/p')
        savings=$(echo "$line" | sed -n 's/.*"savings_vs_vanilla":\([-0-9]*\).*/\1/p')
        prompt=$(echo "$line" | sed -n 's/.*"prompt":"\([^"]*\)".*/\1/p')
        printf "$id\t$pass\t$score\t$tokens\t$savings\t$prompt\n"
    done < "$SUMMARY_FILE" | column -t -s $'\t'
fi
