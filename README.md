# Proquint

[![codecov](https://codecov.io/gh/iilei/proquint/branch/master/graph/badge.svg)](https://codecov.io/gh/iilei/proquint)

Proquint generates and decodes human-recognisable identifiers: short five-letter words ("proquints") that represent arbitrary-size integers.

Details at the [Proposal for Proquints.](https://arxiv.org/html/0901.4016)

Why use proquints?

- They are easier to read, say, and remember than long numeric or hex strings.
- They provide a compact, hyphen-separated representation for big integers.

See [docs/API.md](docs/API.md) for additional design notes.

## Install

Build from source:

```
go build ./cmd/proquint
```

Install the CLI (recommended):

```
go install github.com/iilei/proquint/cmd/proquint@latest
```

Homebrew via tap:

```
brew install iilei/tap/proquint
```

Run the test suite:

```
go test ./...
```

## Usage

Basic CLI usage:

- Encode a decimal or 0x-prefixed hex bigint to a proquint:

```
proquint encode <number>
```

- Decode a proquint back to a decimal bigint:

```
proquint decode <proquint>
```

- `--pad-groups=N` (encode): left-pad the output with `N` groups of `babab` (zero) so the result has at least `N` groups.

Examples (these match the test-suite expectations):

- `proquint encode 42` -> `babop`
- `proquint encode --pad-groups=2 42` -> `babab-babop`
- `proquint decode babop` -> `42`
- Large values are supported; the tests verify 256-bit max values round-trip.

## Development

- The CLI entrypoint is [cmd/proquint/main.go](cmd/proquint/main.go#L1).
- The encoding/decoding implementation lives in [pkg/cli/proquint.go](pkg/cli/proquint.go#L1).

To run the binary built locally:

```
./proquint encode 12345
```

## Contributing

Contributions welcome — open an issue or submit a PR. Please include tests for new behavior.

## License

This project is licensed under the terms in the `LICENSE` file.
