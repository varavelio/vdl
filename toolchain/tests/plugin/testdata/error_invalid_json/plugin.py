#!/usr/bin/env python3
"""
Plugin that outputs invalid JSON to test error handling.
"""
import json
import sys

def main():
    # Read input
    input_data = json.load(sys.stdin)
    
    # Output invalid JSON (not proper protocol format)
    print("This is not valid JSON output {{{")

if __name__ == "__main__":
    main()
