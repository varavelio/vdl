#!/usr/bin/env python3
"""
Plugin that exits with a non-zero exit code to test error handling.
"""
import json
import sys

def main():
    # Read input (but ignore it)
    input_data = json.load(sys.stdin)
    
    # Print error to stderr
    print("ERROR: Simulated plugin failure!", file=sys.stderr)
    
    # Exit with non-zero code
    sys.exit(1)

if __name__ == "__main__":
    main()
