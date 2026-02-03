#!/usr/bin/env python3
"""
Plugin that validates all IR elements are correctly serialized.
Tests: types, enums, constants, patterns, rpcs, procedures, streams, docs,
       spreads, optionals, arrays, maps, inline objects, deprecation.
Also validates the protocol fields: version and schema.
"""
import json
import re
import sys

def main():
    input_data = json.load(sys.stdin)
    ir = input_data.get("ir", {})
    
    errors = []
    
    # =========================================================================
    # VALIDATE PROTOCOL FIELDS (version and schema)
    # =========================================================================
    version = input_data.get("version")
    if version is None:
        errors.append("Missing required field: version")
    elif not isinstance(version, str) or len(version) == 0:
        errors.append(f"version must be non-empty string, got: {version}")
    else:
        # Validate semver format
        semver_pattern = r'^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$'
        if not re.match(semver_pattern, version):
            errors.append(f"version must be valid semver, got: {version}")
    
    schema = input_data.get("schema")
    if schema is None:
        errors.append("Missing required field: schema")
    elif not isinstance(schema, str) or len(schema) == 0:
        errors.append(f"schema must be non-empty string")
    else:
        # Verify schema contains expected definitions
        expected_in_schema = [
            "type BaseEntity",
            "type User",
            "enum Role",
            "enum Priority",
            "const API_VERSION",
            "rpc UserService",
            "rpc OrderService",
            "proc GetUser",
            "stream UserUpdates"
        ]
        for expected in expected_in_schema:
            if expected not in schema:
                errors.append(f"schema should contain '{expected}'")
    
    # =========================================================================
    # VALIDATE CONSTANTS
    # =========================================================================
    constants = ir.get("constants", [])
    const_names = {c["name"] for c in constants}
    
    expected_consts = ["API_VERSION", "MAX_RETRIES", "TIMEOUT_SECONDS", "DEBUG_MODE"]
    for name in expected_consts:
        if name not in const_names:
            errors.append(f"Missing constant: {name}")
    
    # Check constant types
    for c in constants:
        if c["name"] == "API_VERSION":
            if c["constType"] != "string" or c["value"] != "1.0.0":
                errors.append(f"API_VERSION wrong: {c}")
        elif c["name"] == "MAX_RETRIES":
            if c["constType"] != "int" or c["value"] != "3":
                errors.append(f"MAX_RETRIES wrong: {c}")
        elif c["name"] == "TIMEOUT_SECONDS":
            if c["constType"] != "float" or c["value"] != "30.5":
                errors.append(f"TIMEOUT_SECONDS wrong: {c}")
        elif c["name"] == "DEBUG_MODE":
            if c["constType"] != "bool" or c["value"] != "true":
                errors.append(f"DEBUG_MODE wrong: {c}")
    
    # =========================================================================
    # VALIDATE PATTERNS
    # =========================================================================
    patterns = ir.get("patterns", [])
    pattern_names = {p["name"] for p in patterns}
    
    # Note: VDL formatter normalizes pattern names to PascalCase
    if "UserProfileUrl" not in pattern_names:
        errors.append("Missing pattern: UserProfileUrl")
    if "OrderItemUrl" not in pattern_names:
        errors.append("Missing pattern: OrderItemUrl")
    
    for p in patterns:
        if p["name"] == "UserProfileUrl":
            if p["template"] != "/users/{userId}/profile":
                errors.append(f"UserProfileUrl template wrong: {p['template']}")
            if p["placeholders"] != ["userId"]:
                errors.append(f"UserProfileUrl placeholders wrong: {p['placeholders']}")
        elif p["name"] == "OrderItemUrl":
            if set(p["placeholders"]) != {"orderId", "itemId"}:
                errors.append(f"OrderItemUrl placeholders wrong: {p['placeholders']}")
    
    # =========================================================================
    # VALIDATE ENUMS
    # =========================================================================
    enums = ir.get("enums", [])
    enum_map = {e["name"]: e for e in enums}
    
    if "Role" not in enum_map:
        errors.append("Missing enum: Role")
    else:
        role = enum_map["Role"]
        if role["enumType"] != "string":
            errors.append(f"Role enumType wrong: {role['enumType']}")
        # Note: VDL formatter normalizes enum member names to PascalCase
        member_names = {m["name"] for m in role["members"]}
        if member_names != {"Admin", "User", "Guest"}:
            errors.append(f"Role members wrong: {member_names}")
    
    if "Priority" not in enum_map:
        errors.append("Missing enum: Priority")
    else:
        priority = enum_map["Priority"]
        if priority["enumType"] != "int":
            errors.append(f"Priority enumType wrong: {priority['enumType']}")
    
    if "LegacyStatus" not in enum_map:
        errors.append("Missing enum: LegacyStatus")
    else:
        legacy = enum_map["LegacyStatus"]
        if legacy.get("deprecated") is None:
            errors.append("LegacyStatus should be deprecated")
    
    # =========================================================================
    # VALIDATE TYPES
    # =========================================================================
    types = ir.get("types", [])
    type_map = {t["name"]: t for t in types}
    
    expected_types = ["BaseEntity", "Address", "User", "UserWithFriends"]
    for name in expected_types:
        if name not in type_map:
            errors.append(f"Missing type: {name}")
    
    # Check User type (with spread expansion)
    if "User" in type_map:
        user = type_map["User"]
        field_names = {f["name"] for f in user["fields"]}
        
        # Should have spread fields from BaseEntity
        if "id" not in field_names:
            errors.append("User missing spread field 'id' from BaseEntity")
        if "createdAt" not in field_names:
            errors.append("User missing spread field 'createdAt' from BaseEntity")
        
        # Check optional field
        email_field = next((f for f in user["fields"] if f["name"] == "email"), None)
        if email_field and not email_field.get("optional"):
            errors.append("User.email should be optional")
        
        # Check array field
        tags_field = next((f for f in user["fields"] if f["name"] == "tags"), None)
        if tags_field:
            if tags_field["typeRef"]["kind"] != "array":
                errors.append(f"User.tags should be array, got {tags_field['typeRef']['kind']}")
        
        # Check nested array (matrix)
        scores_field = next((f for f in user["fields"] if f["name"] == "scores"), None)
        if scores_field:
            if scores_field["typeRef"]["kind"] != "array":
                errors.append(f"User.scores should be array")
            # Should be array of arrays
            if scores_field["typeRef"].get("arrayDims", 1) < 2:
                errors.append(f"User.scores should be 2D array")
        
        # Check map field
        metadata_field = next((f for f in user["fields"] if f["name"] == "metadata"), None)
        if metadata_field:
            if metadata_field["typeRef"]["kind"] != "map":
                errors.append(f"User.metadata should be map, got {metadata_field['typeRef']['kind']}")
        
        # Check inline object field
        prefs_field = next((f for f in user["fields"] if f["name"] == "preferences"), None)
        if prefs_field:
            if prefs_field["typeRef"]["kind"] != "object":
                errors.append(f"User.preferences should be object, got {prefs_field['typeRef']['kind']}")
            if prefs_field["typeRef"].get("objectFields") is None:
                errors.append("User.preferences missing inline object definition")
        
        # Check enum reference
        role_field = next((f for f in user["fields"] if f["name"] == "role"), None)
        if role_field:
            if role_field["typeRef"]["kind"] != "enum":
                errors.append(f"User.role should be enum, got {role_field['typeRef']['kind']}")
            if role_field["typeRef"].get("enumName") != "Role":
                errors.append(f"User.role should reference Role enum")
        
        # Check deprecation
        if user.get("deprecated") is None:
            errors.append("User type should be deprecated")
    
    # =========================================================================
    # VALIDATE RPCS
    # =========================================================================
    rpcs = ir.get("rpcs", [])
    rpc_map = {r["name"]: r for r in rpcs}
    
    if "UserService" not in rpc_map:
        errors.append("Missing RPC: UserService")
    
    if "OrderService" not in rpc_map:
        errors.append("Missing RPC: OrderService")
    
    # =========================================================================
    # VALIDATE PROCEDURES (flattened)
    # =========================================================================
    procedures = ir.get("procedures", [])
    proc_full_names = {f"{p['rpcName']}.{p['name']}" for p in procedures}
    
    expected_procs = [
        "UserService.GetUser",
        "UserService.CreateUser", 
        "UserService.ListUsers",
        "OrderService.GetOrder"
    ]
    for name in expected_procs:
        if name not in proc_full_names:
            errors.append(f"Missing flattened procedure: {name}")
    
    # Check CreateUser deprecation
    create_user = next((p for p in procedures if p["name"] == "CreateUser"), None)
    if create_user and create_user.get("deprecated") is None:
        errors.append("CreateUser proc should be deprecated")
    
    # =========================================================================
    # VALIDATE STREAMS (flattened)
    # =========================================================================
    streams = ir.get("streams", [])
    stream_full_names = {f"{s['rpcName']}.{s['name']}" for s in streams}
    
    expected_streams = ["UserService.UserUpdates", "OrderService.OrderStatus"]
    for name in expected_streams:
        if name not in stream_full_names:
            errors.append(f"Missing flattened stream: {name}")
    
    # =========================================================================
    # VALIDATE DOCS
    # =========================================================================
    # Schema.docs contains standalone `doc """..."""` blocks (not docstrings)
    # These are optional - our schema may not have any
    docs = ir.get("docs", [])
    
    # =========================================================================
    # OUTPUT
    # =========================================================================
    if errors:
        # Output errors to stderr for debugging
        for err in errors:
            print(f"ERROR: {err}", file=sys.stderr)
        # Still generate output but mark as failed
        output = {
            "files": [{
                "path": "validation_result.json",
                "content": json.dumps({
                    "success": False,
                    "errors": errors
                }, indent=2)
            }]
        }
        json.dump(output, sys.stdout)
        sys.exit(1)
    
    # Generate comprehensive output
    output = {
        "files": [{
            "path": "validation_result.json",
            "content": json.dumps({
                "success": True,
                "protocol": {
                    "version": version,
                    "schema_length": len(schema) if schema else 0
                },
                "summary": {
                    "types_count": len(types),
                    "enums_count": len(enums),
                    "constants_count": len(constants),
                    "patterns_count": len(patterns),
                    "rpcs_count": len(rpcs),
                    "procedures_count": len(procedures),
                    "streams_count": len(streams),
                    "docs_count": len(docs)
                }
            }, indent=2)
        }]
    }
    
    json.dump(output, sys.stdout)

if __name__ == "__main__":
    main()
