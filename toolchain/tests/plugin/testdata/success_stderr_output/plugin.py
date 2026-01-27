#!/usr/bin/env python3
"""
Plugin that writes to stderr and still succeeds.
Tests that stderr is properly streamed to the user while plugin completes successfully.
"""
import json
import sys

def main():
    # Read input from stdin
    input_data = json.load(sys.stdin)
    
    # Write various messages to stderr (should be visible to user)
    print("[INFO] Plugin started", file=sys.stderr)
    print("[DEBUG] Processing IR with {} types".format(len(input_data.get("ir", {}).get("types", []))), file=sys.stderr)
    print("[WARN] This is a warning message", file=sys.stderr)
    print("[INFO] Plugin completed successfully", file=sys.stderr)
    
    # Generate output with info about what was written to stderr
    output = {
        "files": [{
            "path": "result.json",
            "content": json.dumps({
                "success": True,
                "stderr_messages": [
                    "[INFO] Plugin started",
                    "[DEBUG] Processing IR with types",
                    "[WARN] This is a warning message",
                    "[INFO] Plugin completed successfully"
                ],
                "message": "Plugin wrote to stderr and completed successfully"
            }, indent=2)
        }]
    }
    json.dump(output, sys.stdout)

if __name__ == "__main__":
    main()
