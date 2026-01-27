#!/usr/bin/env python3
"""
Plugin that generates actual Python code from VDL IR.
The generated code is syntactically valid and can be imported/verified.
This tests the full pipeline: VDL -> IR -> Plugin -> Generated Code -> Verification.
"""
import json
import sys
from typing import Any


def generate_python_type(type_ref: dict[str, Any]) -> str:
    """Convert VDL type reference to Python type annotation."""
    kind = type_ref.get("kind", "")
    
    if kind == "primitive":
        prim = type_ref.get("primitive", "")
        type_map = {
            "string": "str",
            "int": "int",
            "float": "float",
            "bool": "bool",
            "datetime": "datetime",
            "bytes": "bytes",
            "any": "Any",
        }
        return type_map.get(prim, "Any")
    elif kind == "type":
        return type_ref.get("type", "Any")
    elif kind == "enum":
        return type_ref.get("enum", "str")
    elif kind == "array":
        element_type = generate_python_type(type_ref.get("element", {}))
        return f"list[{element_type}]"
    elif kind == "map":
        value_type = generate_python_type(type_ref.get("value", {}))
        return f"dict[str, {value_type}]"
    elif kind == "object":
        return "dict[str, Any]"
    else:
        return "Any"


def generate_enum(enum_def: dict[str, Any]) -> str:
    """Generate Python Enum class."""
    name = enum_def["name"]
    members = enum_def.get("members", [])
    value_type = enum_def.get("valueType", "string")
    
    base_class = "str, Enum" if value_type == "string" else "IntEnum"
    
    lines = [f"class {name}({base_class}):"]
    
    doc = enum_def.get("doc", "")
    if doc:
        lines.append(f'    """{doc}"""')
    
    for member in members:
        member_name = member["name"]
        member_value = member["value"]
        if value_type == "string":
            lines.append(f'    {member_name} = "{member_value}"')
        else:
            lines.append(f'    {member_name} = {member_value}')
    
    return "\n".join(lines)


def generate_type(type_def: dict[str, Any]) -> str:
    """Generate Python dataclass."""
    name = type_def["name"]
    fields = type_def.get("fields", [])
    
    lines = ["@dataclass"]
    lines.append(f"class {name}:")
    
    doc = type_def.get("doc", "")
    if doc:
        lines.append(f'    """{doc}"""')
    
    if not fields:
        lines.append("    pass")
        return "\n".join(lines)
    
    # Sort fields: required fields first, optional fields last (for dataclass compatibility)
    required_fields = [f for f in fields if not f.get("optional", False)]
    optional_fields = [f for f in fields if f.get("optional", False)]
    sorted_fields = required_fields + optional_fields
    
    for field in sorted_fields:
        field_name = field["name"]
        field_type = generate_python_type(field.get("type", {}))
        optional = field.get("optional", False)
        
        if optional:
            lines.append(f"    {field_name}: Optional[{field_type}] = None")
        else:
            lines.append(f"    {field_name}: {field_type}")
    
    return "\n".join(lines)


def generate_proc(proc_def: dict[str, Any], rpc_name: str) -> str:
    """Generate procedure method stub."""
    name = proc_def["name"]
    
    # Generate input type - input is directly a list of fields
    input_fields = proc_def.get("input", []) or []
    input_params = []
    for field in input_fields:
        field_name = field["name"]
        field_type = generate_python_type(field.get("type", {}))
        optional = field.get("optional", False)
        if optional:
            input_params.append(f"{field_name}: Optional[{field_type}] = None")
        else:
            input_params.append(f"{field_name}: {field_type}")
    
    params_str = ", ".join(["self"] + input_params)
    
    doc = proc_def.get("doc", "")
    doc_line = f'\n        """{doc}"""' if doc else ""
    
    return f'''    def {name}({params_str}) -> dict[str, Any]:{doc_line}
        raise NotImplementedError("{rpc_name}.{name}")'''


def generate_stream(stream_def: dict[str, Any], rpc_name: str) -> str:
    """Generate stream method stub."""
    name = stream_def["name"]
    
    # input is directly a list of fields
    input_fields = stream_def.get("input", []) or []
    input_params = []
    for field in input_fields:
        field_name = field["name"]
        field_type = generate_python_type(field.get("type", {}))
        input_params.append(f"{field_name}: {field_type}")
    
    params_str = ", ".join(["self"] + input_params)
    
    doc = stream_def.get("doc", "")
    doc_line = f'\n        """{doc}"""' if doc else ""
    
    return f'''    def {name}({params_str}) -> Iterator[dict[str, Any]]:{doc_line}
        raise NotImplementedError("{rpc_name}.{name} stream")'''


def generate_rpc(rpc_def: dict[str, Any]) -> str:
    """Generate RPC service class."""
    name = rpc_def["name"]
    procs = rpc_def.get("procs", [])
    streams = rpc_def.get("streams", [])
    
    lines = [f"class {name}:"]
    
    doc = rpc_def.get("doc", "")
    if doc:
        lines.append(f'    """{doc}"""')
    
    if not procs and not streams:
        lines.append("    pass")
        return "\n".join(lines)
    
    for proc in procs:
        lines.append(generate_proc(proc, name))
        lines.append("")
    
    for stream in streams:
        lines.append(generate_stream(stream, name))
        lines.append("")
    
    return "\n".join(lines)


def main():
    input_data = json.load(sys.stdin)
    ir = input_data.get("ir", {})
    options = input_data.get("options", {})
    
    # Generate Python code
    code_lines = [
        '"""',
        'Generated by VDL Plugin.',
        'This file is auto-generated - do not edit manually.',
        '"""',
        'from __future__ import annotations',
        '',
        'from dataclasses import dataclass',
        'from datetime import datetime',
        'from enum import Enum, IntEnum',
        'from typing import Any, Iterator, Optional',
        '',
        '',
        '# ============================================================================',
        '# Enums',
        '# ============================================================================',
        '',
    ]
    
    for enum_def in ir.get("enums", []):
        code_lines.append(generate_enum(enum_def))
        code_lines.append("")
        code_lines.append("")
    
    code_lines.extend([
        '# ============================================================================',
        '# Types',
        '# ============================================================================',
        '',
    ])
    
    for type_def in ir.get("types", []):
        code_lines.append(generate_type(type_def))
        code_lines.append("")
        code_lines.append("")
    
    code_lines.extend([
        '# ============================================================================',
        '# RPC Services',
        '# ============================================================================',
        '',
    ])
    
    for rpc_def in ir.get("rpcs", []):
        code_lines.append(generate_rpc(rpc_def))
        code_lines.append("")
        code_lines.append("")
    
    # Add metadata
    code_lines.extend([
        '# ============================================================================',
        '# Metadata',
        '# ============================================================================',
        '',
        'GENERATED_TYPES = [',
    ])
    
    for type_def in ir.get("types", []):
        code_lines.append(f'    "{type_def["name"]}",')
    
    code_lines.extend([
        ']',
        '',
        'GENERATED_ENUMS = [',
    ])
    
    for enum_def in ir.get("enums", []):
        code_lines.append(f'    "{enum_def["name"]}",')
    
    code_lines.extend([
        ']',
        '',
        'GENERATED_RPCS = [',
    ])
    
    for rpc_def in ir.get("rpcs", []):
        code_lines.append(f'    "{rpc_def["name"]}",')
    
    code_lines.extend([
        ']',
        '',
    ])
    
    code_content = "\n".join(code_lines)
    
    # Also generate a manifest file for verification
    manifest = {
        "success": True,
        "generated_from": "schema.vdl",
        "language": options.get("language", "python"),
        "include_typing": options.get("includeTyping", False),
        "statistics": {
            "types": len(ir.get("types", [])),
            "enums": len(ir.get("enums", [])),
            "rpcs": len(ir.get("rpcs", [])),
            "procedures": len(ir.get("procedures", [])),
            "streams": len(ir.get("streams", [])),
        }
    }
    
    output = {
        "files": [
            {
                "path": "generated.py",
                "content": code_content
            },
            {
                "path": "manifest.json",
                "content": json.dumps(manifest, indent=2)
            }
        ]
    }
    
    json.dump(output, sys.stdout)


if __name__ == "__main__":
    main()
