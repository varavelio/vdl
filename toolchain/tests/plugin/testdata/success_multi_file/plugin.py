#!/usr/bin/env python3
"""
Plugin that validates multi-file schema processing.
Verifies that includes are correctly resolved and all definitions
from all files are present in the merged IR.
"""
import json
import sys

def main():
    input_data = json.load(sys.stdin)
    ir = input_data.get("ir", {})
    
    # === Validate Types ===
    types = {t["name"]: t for t in ir.get("types", [])}
    
    # From types.vdl
    assert "BaseEntity" in types, "BaseEntity type not found (from types.vdl)"
    assert "User" in types, "User type not found (from types.vdl)"
    assert "UserProfile" in types, "UserProfile type not found (from types.vdl)"
    assert "Order" in types, "Order type not found (from types.vdl)"
    assert "OrderItem" in types, "OrderItem type not found (from types.vdl)"
    
    # Verify User has spread from BaseEntity
    user = types["User"]
    user_fields = [f["name"] for f in user["fields"]]
    assert "username" in user_fields, "User should have username field"
    assert "email" in user_fields, "User should have email field"
    
    # === Validate Enums ===
    enums = {e["name"]: e for e in ir.get("enums", [])}
    
    # From enums.vdl
    assert "UserRole" in enums, "UserRole enum not found (from enums.vdl)"
    assert "OrderStatus" in enums, "OrderStatus enum not found (from enums.vdl)"
    assert "Priority" in enums, "Priority enum not found (from enums.vdl)"
    
    # Verify Priority has integer values (stored as strings in IR)
    priority = enums["Priority"]
    priority_members = {m["name"]: m.get("value") for m in priority["members"]}
    assert priority_members.get("Low") == "1", "Priority.Low should be 1"
    assert priority_members.get("Critical") == "4", "Priority.Critical should be 4"
    
    # === Validate RPCs ===
    rpcs = {r["name"]: r for r in ir.get("rpcs", [])}
    
    # From services.vdl
    assert "UserService" in rpcs, "UserService not found (from services.vdl)"
    assert "OrderService" in rpcs, "OrderService not found (from services.vdl)"
    
    # === Validate Procedures ===
    procedures = {p["name"]: p for p in ir.get("procedures", [])}
    
    assert "GetUser" in procedures, "GetUser procedure not found"
    assert "CreateUser" in procedures, "CreateUser procedure not found"
    assert "GetOrder" in procedures, "GetOrder procedure not found"
    assert "CreateOrder" in procedures, "CreateOrder procedure not found"
    
    # Verify procedure belongs to correct RPC
    assert procedures["GetUser"]["rpcName"] == "UserService"
    assert procedures["GetOrder"]["rpcName"] == "OrderService"
    
    # === Validate Streams ===
    streams = {s["name"]: s for s in ir.get("streams", [])}
    
    assert "UserActivity" in streams, "UserActivity stream not found"
    assert "OrderUpdates" in streams, "OrderUpdates stream not found"
    
    # === Validate Constants ===
    constants = {c["name"]: c for c in ir.get("constants", [])}
    
    # From main.vdl
    assert "APP_NAME" in constants, "APP_NAME constant not found (from main.vdl)"
    assert "APP_VERSION" in constants, "APP_VERSION constant not found (from main.vdl)"
    assert constants["APP_NAME"]["value"] == "MultiFileTest"
    assert constants["APP_VERSION"]["value"] == "2.0.0"
    
    # === Generate summary output ===
    summary = {
        "success": True,
        "multi_file_test": True,
        "files_merged": ["main.vdl", "types.vdl", "enums.vdl", "services.vdl"],
        "counts": {
            "types": len(types),
            "enums": len(enums),
            "rpcs": len(rpcs),
            "procedures": len(procedures),
            "streams": len(streams),
            "constants": len(constants)
        },
        "validated": {
            "types": list(types.keys()),
            "enums": list(enums.keys()),
            "rpcs": list(rpcs.keys()),
            "procedures": list(procedures.keys()),
            "streams": list(streams.keys()),
            "constants": list(constants.keys())
        }
    }
    
    output = {
        "files": [
            {
                "path": "summary.json",
                "content": json.dumps(summary, indent=2)
            }
        ]
    }
    
    json.dump(output, sys.stdout)

if __name__ == "__main__":
    main()
