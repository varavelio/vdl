#!/usr/bin/env python3
"""
Basic plugin that validates the IR structure and generates a simple output file.
This plugin verifies that:
1. The IR contains expected types, enums, rpcs, procedures, etc.
2. Options are passed correctly (empty in this case)
3. The protocol is working (stdin/stdout JSON)
4. The version field contains a valid semver string
5. The schema field contains the complete VDL schema as plain text
"""
import json
import re
import sys

def main():
    # Read input from stdin
    input_data = json.load(sys.stdin)
    
    # =========================================================================
    # VALIDATE VERSION FIELD (new protocol field)
    # =========================================================================
    version = input_data.get("version")
    assert version is not None, "Missing required field: version"
    assert isinstance(version, str), f"version must be string, got {type(version)}"
    assert len(version) > 0, "version must not be empty"
    # Version should be a valid semver (e.g., "0.0.0-dev", "1.0.0", etc.)
    semver_pattern = r'^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$'
    assert re.match(semver_pattern, version), f"version must be valid semver, got '{version}'"
    
    # =========================================================================
    # VALIDATE SCHEMA FIELD (new protocol field)
    # =========================================================================
    schema = input_data.get("schema")
    assert schema is not None, "Missing required field: schema"
    assert isinstance(schema, str), f"schema must be string, got {type(schema)}"
    assert len(schema) > 0, "schema must not be empty"
    # Schema should contain the VDL content we defined
    assert "type User" in schema, "schema should contain 'type User' definition"
    assert "rpc UserService" in schema, "schema should contain 'rpc UserService' definition"
    assert "proc GetUser" in schema, "schema should contain 'proc GetUser' definition"
    
    # =========================================================================
    # VALIDATE IR STRUCTURE
    # =========================================================================
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
                    "protocol_validated": True,
                    "version_received": version,
                    "schema_length": len(schema),
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
