# Security Policy

## Supported Versions

| Version | Supported |
|---|---|
| latest | :white_check_mark: |
| < latest | :x: |

## Reporting a Vulnerability

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, report them privately via GitHub Security Advisories:

1. Go to the [Security tab](../../security) of this repository.
2. Click **"Report a vulnerability"**.
3. Provide a detailed description of the vulnerability.

We aim to acknowledge reports within 48 hours and provide an initial assessment within 5 business days.

### Scope

- The MCP server's authentication and authorization mechanisms
- Token handling, credential forwarding, and OIDC/LDAP interaction
- Configuration injection and secret management
- Dependency supply chain vulnerabilities

### Out of Scope

- Issues in the upstream API that this MCP server proxies
- Misconfiguration by the operator (documented in README)
- Denial-of-service via unbounded tool invocations (rate-limiting is the operator's responsibility)
