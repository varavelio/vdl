#!/usr/bin/env python3
"""
Basic plugin that validates the IR structure and generates a simple output file.
This plugin verifies that:
1. The IR contains expected types, enums, rpcs, procedures, etc.
2. Options are passed correctly (empty in this case)
3. The protocol is working (stdin/stdout JSON)
"""
import json
import sys

def main():
    # Read input from stdin
    input_data = json.load(sys.stdin)
    
    # Validate structure
    ir = input_data.get("ir", {})
    options = input_data.get("options")
    
    # Basic validation
    assert "types" in ir, "IR must contain 'types'"
    assert "enums" in ir, "IR must contain 'enums'"
    assert "constants" in ir, "IR must contain 'constants'"
    assert "patterns" in ir, "IR must contain 'patterns'"
    assert "rpcs" in ir, "IR must contain 'rpcs'"
    assert "procedures" in ir, "IR must contain 'procedures'"
    assert "streams" in ir, "IR must contain 'streams'"
    assert "docs" in ir, "IR must contain 'docs'"
    
    # Verify User type exists
    types = ir["types"]
    user_type = next((t for t in types if t["name"] == "User"), None)
    assert user_type is not None, "User type not found"
    assert len(user_type["fields"]) == 3, "User should have 3 fields"
    
    # Verify UserService RPC exists
    rpcs = ir["rpcs"]
    user_service = next((r for r in rpcs if r["name"] == "UserService"), None)
    assert user_service is not None, "UserService RPC not found"
    
    # Verify GetUser procedure
    procedures = ir["procedures"]
    get_user = next((p for p in procedures if p["name"] == "GetUser"), None)
    assert get_user is not None, "GetUser procedure not found"
    assert get_user["rpcName"] == "UserService", "GetUser should belong to UserService"
    
    # Options should be empty dict in this test (no options configured)
    assert options == {} or options is None, f"Expected empty options, got {options}"
    
    # Generate output
    output = {
        "files": [
            {
                "path": "output.json",
                "content": json.dumps({
                    "success": True,
                    "types_count": len(ir["types"]),
                    "rpcs_count": len(ir["rpcs"]),
                    "procedures_count": len(ir["procedures"]),
                    "user_type": user_type,
                    "get_user_proc": get_user
                }, indent=2)
            }
        ]
    }
    
    # Write output to stdout
    json.dump(output, sys.stdout)

if __name__ == "__main__":
    main()
