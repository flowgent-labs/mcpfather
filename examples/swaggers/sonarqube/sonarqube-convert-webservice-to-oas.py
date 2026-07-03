#!/usr/bin/env python3
"""Convert SonarQube webservices JSON to OpenAPI 3.1 (or 3.0.1) specification.

Usage:
  ./sonarqube-convert-webservice-to-oas.py [--version 3.1|3.0.1] [--server-url URL] < input.json > output.json
  ./sonarqube-convert-webservice-to-oas.py [--version 3.1|3.0.1] [--server-url URL] -i input.json -o output.json
"""

import json
import re
import sys
from argparse import ArgumentParser
from textwrap import dedent


def strip_html(text: str) -> str:
    """Strip basic HTML tags from description text, preserving readability."""
    text = re.sub(r"<br\s*/?\s*>", " ", text, flags=re.IGNORECASE)
    text = re.sub(r"</?p\s*/?\s*>", " ", text, flags=re.IGNORECASE)
    text = re.sub(r"<li\s*>", " - ", text, flags=re.IGNORECASE)
    text = re.sub(r"</li\s*>", "", text, flags=re.IGNORECASE)
    text = re.sub(r"</?ul\s*>", "", text, flags=re.IGNORECASE)
    text = re.sub(r"<[^>]+>", "", text)
    text = text.replace("&lt;", "<").replace("&gt;", ">").replace("&amp;", "&")
    text = re.sub(r"  +", " ", text).strip()
    return text


def first_sentence(text: str) -> str:
    """Extract the first sentence or first <br/>-separated segment as summary."""
    if not text:
        return ""
    # Split on <br/> before stripping HTML, take first segment
    br_match = re.search(r"<br\s*/?\s*>", text, re.IGNORECASE)
    if br_match:
        text = text[: br_match.start()].strip()
    text = strip_html(text)
    # Try first sentence ending with period + space/end
    m = re.match(r"^([^.]+\w\.)(?:\s|$)", text)
    if m and len(m.group(1)) < 120:
        return m.group(1).strip()
    # Fallback: first period-separated segment of reasonable length
    parts = re.split(r"\.(?:\s+|$)", text)
    for p in parts:
        p = p.strip()
        if not p:
            continue
        if p == text[: len(p)] and len(p) >= 10:
            return p + "." if not p.endswith(".") else p
    return text[:120]


def build_param_schema(param: dict) -> dict:
    """Build a JSON Schema for a parameter."""
    schema: dict = {}
    if "possibleValues" in param:
        schema["type"] = "string"
        schema["enum"] = param["possibleValues"]
    elif "maximumLength" in param or "minimumLength" in param:
        schema["type"] = "string"
        if "maximumLength" in param:
            schema["maxLength"] = param["maximumLength"]
        if "minimumLength" in param:
            schema["minLength"] = param["minimumLength"]
    elif "maximumValue" in param:
        schema["type"] = "integer"
        schema["maximum"] = param["maximumValue"]
    elif param.get("maxValuesAllowed", 0) > 1:
        schema["type"] = "array"
        schema["items"] = {"type": "string"}
        schema["maxItems"] = param["maxValuesAllowed"]
    return schema


def build_parameter(param: dict) -> dict:
    """Build an OpenAPI Parameter object from a SonarQube param."""
    openapi_param: dict = {
        "name": param["key"],
        "in": "query",
        "description": strip_html(param.get("description", "")),
        "required": param.get("required", False),
    }
    schema = build_param_schema(param)
    if schema:
        openapi_param["schema"] = schema
    elif "defaultValue" in param:
        openapi_param["schema"] = {"type": "string", "default": param["defaultValue"]}
    else:
        # Always provide a default string schema so the parameter is not lost
        openapi_param["schema"] = {"type": "string"}
    if param.get("deprecatedSince"):
        openapi_param["deprecated"] = True
    if param.get("internal"):
        openapi_param["x-internal"] = True
    if "exampleValue" in param:
        openapi_param["example"] = param["exampleValue"]
    if "since" in param:
        openapi_param["x-since"] = param["since"]
    if param.get("deprecatedKey"):
        openapi_param["x-deprecatedKey"] = param["deprecatedKey"]
        if param.get("deprecatedKeySince"):
            openapi_param["x-deprecatedKeySince"] = param["deprecatedKeySince"]
    return openapi_param


def sanitize_operation_id(op_id: str) -> str:
    """Convert to camelCase, removing special chars."""
    op_id = re.sub(r"[^a-zA-Z0-9_-]", "_", op_id)
    parts = op_id.split("_")
    if len(parts) > 1:
        return parts[0] + "".join(p.capitalize() for p in parts[1:])
    return op_id


def build_operation(ws_path: str, action: dict) -> tuple:
    """Build an OpenAPI Operation object. Returns (method, operation)."""
    method = "post" if action.get("post") else "get"
    # Strip redundant api/ prefix for concise, elegant operationIds
    # e.g. api/issues/search → getIssuesSearch (not getApiIssuesSearch)
    clean_path = ws_path[4:] if ws_path.startswith("api/") else ws_path
    op: dict = {
        "operationId": method
        + sanitize_operation_id(
            f"_{clean_path.replace('/', '_')}_{action['key']}"
        ),
        "summary": first_sentence(action.get("description", "")),
        "description": strip_html(action.get("description", "")),
        "tags": [ws_path],
        "responses": {},
    }
    if action.get("deprecatedSince"):
        op["deprecated"] = True
        op["x-deprecatedSince"] = action["deprecatedSince"]
    if action.get("internal"):
        op["x-internal"] = True
    if action.get("since"):
        op["x-since"] = action["since"]
    if action.get("params"):
        op["parameters"] = [build_parameter(p) for p in action["params"]]
    # Build response
    if action.get("hasResponseExample"):
        op["responses"]["200"] = {
            "description": "Successful response",
            "content": {
                "application/json": {
                    "schema": {
                        "type": "object",
                        "description": "Response data (see SonarQube API documentation for schema details)",
                    }
                }
            },
        }
    elif method == "post" and not action.get("hasResponseExample"):
        op["responses"]["200"] = {
            "description": "Successful response (typically empty)"
        }
    else:
        op["responses"]["200"] = {
            "description": "Successful response",
            "content": {"application/json": {"schema": {"type": "object"}}},
        }
    op["responses"]["401"] = {"$ref": "#/components/responses/Unauthorized"}
    op["responses"]["403"] = {"$ref": "#/components/responses/Forbidden"}
    return method, op


def build_paths(web_services: list) -> dict:
    """Build OpenAPI paths from web services."""
    paths: dict = {}
    for ws in web_services:
        ws_path = ws["path"]
        for action in ws.get("actions", []):
            url_path = f"/{ws_path}/{action['key']}"
            method, operation = build_operation(ws_path, action)
            paths.setdefault(url_path, {})[method] = operation
    return paths


def build_tags(web_services: list) -> list:
    """Build tags from web services."""
    tags: list = []
    for ws in web_services:
        tag: dict = {
            "name": ws["path"],
            "description": strip_html(ws.get("description", "")),
        }
        if ws.get("since"):
            tag["x-since"] = ws["since"]
        tags.append(tag)
    return tags


def build_openapi(web_services: list, version: str, server_url: str) -> dict:
    """Build the full OpenAPI specification."""
    openapi_version = "3.1.0" if version == "3.1" else "3.0.1"
    return {
        "openapi": openapi_version,
        "info": {
            "title": "SonarQube Web API",
            "description": strip_html(
                "OpenAPI specification generated from SonarQube web services definition. "
                "See the official SonarQube API documentation for detailed response schemas."
            ),
            "version": "1.0.0",
        },
        "servers": [
            {"url": server_url, "description": "SonarQube server"},
        ],
        "tags": build_tags(web_services),
        "paths": build_paths(web_services),
        "components": {
            "responses": {
                "Unauthorized": {
                    "description": "Authentication is required and has failed or has not been provided",
                },
                "Forbidden": {
                    "description": "Authenticated user does not have the required permissions",
                },
            },
        },
    }


def main():
    parser = ArgumentParser(
        description="Convert SonarQube webservices JSON to OpenAPI/Swagger specification",
        epilog=dedent("""\
            Examples:
              %(prog)s -i sonarqube-v2026.4.0.124573.webservices.json -o sonarqube-v2026.4.0.124573.oas.3.1.0.json
              cat sonarqube-v2026.4.0.124573.webservices.json | %(prog)s > sonarqube-v2026.4.0.124573.oas.3.1.0.json"""),
    )
    parser.add_argument(
        "-i", "--input",
        default=None,
        help="Input JSON file (default: stdin)",
    )
    parser.add_argument(
        "-o", "--output",
        default=None,
        help="Output JSON file (default: stdout)",
    )
    parser.add_argument(
        "--version",
        choices=["3.1", "3.0.1"],
        default="3.1",
        help="OpenAPI version (default: 3.1)",
    )
    parser.add_argument(
        "--server-url",
        default="http://localhost:9000",
        help="SonarQube server URL (default: http://localhost:9000)",
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

    # Extract web services
    web_services = data.get("webServices", [])
    if not web_services:
        print("Error: No 'webServices' found in input JSON", file=sys.stderr)
        sys.exit(1)

    # Build OpenAPI spec
    openapi = build_openapi(web_services, args.version, args.server_url)

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
