import json
import sys
import argparse
import re

def load_scenarios(path):
    with open(path, 'r') as f:
        return json.load(f)

def analyze_trace(trace_path, scenario):
    with open(trace_path, 'r') as f:
        commands = [line.strip() for line in f if line.strip()]

    required_tools = scenario.get('required_tools', [])
    anti_patterns = scenario.get('anti_patterns', [])
    
    found_required = []
    found_anti = []
    
    first_action_valid = False
    valid_action_index = -1
    anti_action_index = -1

    print(f"--- Analyzing Scenario: {scenario['prompt']} ---")
    print(f"Trace length: {len(commands)}")

    for i, cmd in enumerate(commands):
        # Check for required tools
        for req in required_tools:
            if req in cmd:
                if req not in found_required:
                    found_required.append(req)
                    if valid_action_index == -1:
                        valid_action_index = i

        # Check for anti-patterns
        for anti in anti_patterns:
            # Simple substring match for now, could be regex
            if anti in cmd:
                if anti not in found_anti:
                    found_anti.append(anti)
                    if anti_action_index == -1:
                        anti_action_index = i

    # Compliance Logic
    # 1. Did we use at least one required tool?
    used_required = len(found_required) > 0
    
    # 2. Did we use it BEFORE any anti-pattern?
    # If no anti-patterns used, sequence is perfect.
    # If anti-pattern used, valid must be < anti.
    sequence_compliant = True
    if anti_action_index != -1:
        if valid_action_index == -1:
            sequence_compliant = False # Anti used, but no valid used
        elif valid_action_index > anti_action_index:
            sequence_compliant = False # Valid used AFTER anti

    # Scoring
    score = 0
    if used_required:
        score += 50
    if sequence_compliant:
        score += 50
        
    print(f"\nResults:")
    print(f"  Required Tools Used: {found_required} (Goal: {required_tools})")
    print(f"  Anti-Patterns Used: {found_anti}")
    print(f"  Sequence Compliant: {sequence_compliant}")
    print(f"  Score: {score}/100")
    
    return {
        "score": score,
        "compliant": sequence_compliant,
        "found_required": found_required,
        "found_anti": found_anti
    }

def main():
    parser = argparse.ArgumentParser(description='Score agent traces against BeadsLog protocol')
    parser.add_argument('--scenarios', default='_rules/_evals/scenarios.json', help='Path to scenarios JSON')
    parser.add_argument('--trace', required=True, help='Path to trace file (one command per line)')
    parser.add_argument('--id', required=True, help='Scenario ID to evaluate against')
    
    args = parser.parse_args()
    
    scenarios = load_scenarios(args.scenarios)
    target_scenario = next((s for s in scenarios if s['id'] == args.id), None)
    
    if not target_scenario:
        print(f"Error: Scenario ID {args.id} not found")
        sys.exit(1)
        
    analyze_trace(args.trace, target_scenario)

if __name__ == "__main__":
    main()
