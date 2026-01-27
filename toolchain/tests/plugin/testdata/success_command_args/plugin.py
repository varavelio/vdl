#!/usr/bin/env python3
"""
Plugin that validates command-line arguments are passed correctly.
Tests that VDL properly passes extra args defined in the command array.
"""
import json
import sys

def main():
    # Read input from stdin
    input_data = json.load(sys.stdin)
    
    errors = []
    
    # Validate command-line arguments
    # sys.argv should be: ["plugin.py", "--format", "json", "--verbose"]
    expected_args = ["--format", "json", "--verbose"]
    actual_args = sys.argv[1:]  # Skip script name
    
    if actual_args != expected_args:
        errors.append(f"Expected args {expected_args}, got {actual_args}")
    
    # Validate options from vdl.yaml
    options = input_data.get("options", {})
    
    if options.get("target") != "client":
        errors.append(f"Expected options.target='client', got {options.get('target')}")
    
    expected_features = ["streaming", "validation"]
    if options.get("features") != expected_features:
        errors.append(f"Expected options.features={expected_features}, got {options.get('features')}")
    
    if errors:
        for err in errors:
            print(f"ERROR: {err}", file=sys.stderr)
        output = {
            "files": [{
                "path": "result.json",
                "content": json.dumps({"success": False, "errors": errors}, indent=2)
            }]
        }
        json.dump(output, sys.stdout)
        sys.exit(1)
    
    output = {
        "files": [{
            "path": "result.json",
            "content": json.dumps({
                "success": True,
                "received_args": actual_args,
                "received_options": options,
                "message": "Command arguments and options passed correctly"
            }, indent=2)
        }]
    }
    json.dump(output, sys.stdout)

if __name__ == "__main__":
    main()
