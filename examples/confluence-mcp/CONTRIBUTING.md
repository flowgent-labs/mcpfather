# Contributing to confluence-mcp

Thank you for your interest in contributing!

## How to Contribute

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Make your changes and add tests if applicable.
4. Run `make test` to ensure all tests pass.
5. Submit a pull request with a clear description.

## Development

```sh
# Build the binary
make build

# Run tests
make test

# Build Docker image
make build-image

# Regenerate DSL schema (after config changes)
make gen-dsl-schema
```

## Code Style

- Use `gofmt` to format your code.
- Write clear, concise commit messages.

## Reporting Issues

- Use the [GitHub Issues](../../issues) page.
- Choose the appropriate issue template (bug report or feature request).
- Include as much detail as possible: environment, configuration, steps to reproduce, and logs.

## Security

See [SECURITY.md](SECURITY.md) for our security policy and vulnerability reporting process.
