#!/usr/bin/env python3
"""Convert Telegram MTProto Schema JSON to OpenAPI 3.1 specification.
   refer: https://core.telegram.org/schema/json
Usage:
  cat telegram-v20260715.schema.json | ./telegram-convert-oas.py > telegram.oas.json
  ./telegram-convert-oas.py -i schema.json -o telegram.oas.json
"""

import json
import re
import sys
from argparse import ArgumentParser
from textwrap import dedent

# Mapping of MTProto scalar types to JSON Schema types.
TYPE_MAP = {
    "int": "integer",
    "long": "integer",
    "double": "number",
    "float": "number",
    "string": "string",
    "Bool": "boolean",
    "bool": "boolean",
    "bytes": "string",
    "true": "boolean",
    "#": "integer",
    "int32": "integer",
    "int53": "integer",
    "int64": "integer",
    "int128": "string",
    "int256": "string",
    "int512": "string",
    "long": "integer",
    "double": "number",
    "string": "string",
    "bytes": "string",
}

# MTProto type suffixes that indicate a raw JSON/bytes payload
UNSTRUCTURED_TYPES = {"JSONValue", "bytes"}


def mtproto_type_to_json_schema(mttype: str) -> dict:
    """Convert an MTProto type expression to a JSON Schema fragment."""
    mttype = mttype.strip()

    # Conditional flags: flags.N?Type — extract the Type part
    m_flag = re.match(r"^flags\.\d+\?(.+)$", mttype)
    if m_flag:
        return mtproto_type_to_json_schema(m_flag.group(1))

    # Generic vector
    if mttype.startswith("Vector<") and mttype.endswith(">"):
        inner = mttype[7:-1]
        return {"type": "array", "items": mtproto_type_to_json_schema(inner)}

    # Generic type parameter
    m_generic = re.match(r"^([A-Za-z.]+)<(.+)>$", mttype)
    if m_generic:
        base = m_generic.group(1)
        if base == "Vector":
            return {"type": "array", "items": mtproto_type_to_json_schema(m_generic.group(2))}
        return {"type": "object", "description": f"Generic {base}"}

    # Bare type variable
    if mttype.startswith("!"):
        return {"type": "object", "description": "Polymorphic type parameter"}

    # Scalar types
    if mttype in TYPE_MAP:
        schema = {"type": TYPE_MAP[mttype]}
        if TYPE_MAP[mttype] == "string" and mttype == "bytes":
            schema["format"] = "byte"
        return schema

    # Identifier types (e.g. InputPeer, Message, etc.) — treat as object
    return {"type": "object", "description": f"MTProto type: {mttype}"}


def sanitize_operation_id(method_name: str) -> str:
    """Convert dotted method name to camelCase operationId."""
    parts = method_name.split(".")
    return parts[0] + "".join(p.capitalize() for p in parts[1:])


def build_parameters(params: list) -> list:
    """Build OpenAPI Parameter objects from MTProto params."""
    openapi_params = []
    for p in params:
        name = p["name"]
        mttype = p["type"]

        # Handle flag parameters — omit as query params, they go in body
        if name == "flags" and mttype == "#":
            continue
        if mttype.startswith("flags."):
            continue

        param = {
            "name": name,
            "in": "query",
            "required": True,
            "description": f"Type: {mttype}",
            "schema": {"type": "string"},
        }
        openapi_params.append(param)
    return openapi_params


def build_request_body(params: list) -> dict | None:
    """Build an OpenAPI Request Body for POST methods with complex params."""
    props = {}
    required = []
    for p in params:
        name = p["name"]
        mttype = p["type"]

        if name == "flags" and mttype == "#":
            continue

        schema = mtproto_type_to_json_schema(mttype)
        props[name] = schema

        # Only mark as required if not optional/conditional
        if not mttype.startswith("flags."):
            required.append(name)

    if not props:
        return None

    body_schema: dict = {"type": "object", "properties": props}
    if required:
        body_schema["required"] = required

    return {
        "required": True,
        "content": {"application/json": {"schema": body_schema}},
    }


def build_response_schema(return_type: str, constructors_by_type: dict) -> dict:
    """Build a response schema from the return type using known constructors."""
    if return_type in ("X", "!X"):
        return {"type": "object", "description": "Polymorphic response"}

    schemas = []
    if return_type in constructors_by_type:
        for ctor in constructors_by_type[return_type]:
            props = {}
            for p in ctor.get("params", []):
                props[p["name"]] = mtproto_type_to_json_schema(p["type"])
            # Prefer the most common / simplest constructor
            schemas.append({
                "type": "object",
                "title": ctor["predicate"],
                "properties": props,
            })

    if len(schemas) == 1:
        return schemas[0]
    return {"type": "object", "description": f"Response type: {return_type}"}


def build_methods(methods: list) -> dict:
    """Build OpenAPI paths from MTProto methods."""
    paths: dict = {}
    for m in methods:
        method_name = m["method"]
        # Determine namespace (e.g., auth, account, messages, etc.)
        ns = method_name.split(".")[0] if "." in method_name else "root"
        op_id = sanitize_operation_id(method_name)

        operation: dict = {
            "operationId": op_id,
            "summary": f"Invoke {method_name}",
            "description": f"MTProto RPC method: {method_name}\n\nReturns: {m['type']}",
            "tags": [ns],
            "responses": {
                "200": {
                    "description": f"Successful response ({m['type']})",
                    "content": {
                        "application/json": {
                            "schema": {"type": "object", "description": f"MTProto return type: {m['type']}"}
                        }
                    },
                }
            },
        }

        request_body = build_request_body(m.get("params", []))
        if request_body:
            operation["requestBody"] = request_body
        if method_name.startswith("destroy_") or method_name.endswith("delete") or method_name.endswith("Delete"):
            operation["x-deprecated"] = True

        url_path = f"/rpc/{method_name}"
        # Avoid duplicate operations
        if url_path not in paths:
            paths[url_path] = {}
        paths[url_path]["post"] = operation

    return paths


def build_tags(methods: list) -> list:
    """Build tag groups from method namespaces."""
    namespaces: dict = {}
    for m in methods:
        name = m["method"]
        ns = name.split(".")[0] if "." in name else "root"
        if ns not in namespaces:
            namespaces[ns] = 0
        namespaces[ns] += 1

    tags = []
    for ns, count in sorted(namespaces.items(), key=lambda x: -x[1]):
        tags.append({
            "name": ns,
            "description": f"{ns}.* methods ({count} RPC calls)",
        })
    return tags


def build_openapi(data: dict, server_url: str) -> dict:
    """Build the full OpenAPI specification from MTProto schema JSON."""
    methods = data.get("methods", [])
    constructors = data.get("constructors", [])

    # Index constructors by type for response schema generation
    constructors_by_type: dict = {}
    for c in constructors:
        t = c.get("type", "unknown")
        constructors_by_type.setdefault(t, []).append(c)

    return {
        "openapi": "3.1.0",
        "info": {
            "title": "Telegram MTProto API",
            "description": (
                "OpenAPI specification generated from the Telegram MTProto schema. "
                "Each RPC method is exposed as a POST endpoint. "
                "This is a low-level protocol mapping; see the "
                "[Telegram API documentation](https://core.telegram.org/api) for protocol details."
            ),
            "version": "1.0.0",
        },
        "servers": [
            {"url": server_url, "description": "Telegram API endpoint"},
        ],
        "tags": build_tags(methods),
        "paths": build_methods(methods),
        "components": {
            "schemas": {},
        },
    }


def main():
    parser = ArgumentParser(
        description="Convert Telegram MTProto Schema JSON to OpenAPI 3.1 specification",
        epilog=dedent("""\
            Examples:
              %(prog)s -i telegram-v20260715.schema.json -o telegram.oas.3.1.0.json
              cat telegram-v20260715.schema.json | %(prog)s > telegram.oas.3.1.0.json"""),
    )
    parser.add_argument(
        "-i", "--input",
        default=None,
        help="Input schema JSON file (default: stdin)",
    )
    parser.add_argument(
        "-o", "--output",
        default=None,
        help="Output OpenAPI JSON file (default: stdout)",
    )
    parser.add_argument(
        "--server-url",
        default="https://api.telegram.org/bot<token>",
        help="Telegram API server URL (default: https://api.telegram.org/bot<token>)",
    )
    parser.add_argument(
        "--compact",
        action="store_true",
        default=False,
        help="Compact (single-line) JSON output",
    )

    args = parser.parse_args()

    # Read input
    if args.input:
        with open(args.input) as f:
            data = json.load(f)
    else:
        if sys.stdin.isatty():
            parser.print_help()
            sys.exit(1)
        data = json.load(sys.stdin)

    methods = data.get("methods", [])
    if not methods:
        print("Error: No 'methods' found in input JSON", file=sys.stderr)
        sys.exit(1)

    openapi = build_openapi(data, args.server_url)

    # Write output
    indent = None if args.compact else 2
    output_json = json.dumps(openapi, indent=indent, ensure_ascii=False)

    if args.output:
        with open(args.output, "w") as f:
            f.write(output_json)
            f.write("\n")
    else:
        print(output_json)


if __name__ == "__main__":
    main()
