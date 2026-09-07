# license

`license` is a small interactive CLI for adding a license to a project. It presents common licenses in an interactive terminal menu, asks only for the values required by the chosen template, and writes the result to `LICENSE`.

## Installation

Install with Go:

```sh
go install github.com/coalaura/license@latest
```

Prebuilt binaries for Windows, Linux, and macOS are also available from the [GitHub releases](https://github.com/coalaura/license/releases) page.

## Usage

Run the command from the project directory that should receive the license:

```sh
cd my-project
license
```

Choose a license, then complete any requested fields. Suggested values appear in brackets; press Enter to accept one or type a replacement. The tool infers:

- the current year
- the project name from the current directory
- the author from the owner of the Git `origin` remote, when available

If a license file already exists, `license` asks for confirmation before replacing it. Declining leaves the existing file unchanged.

To have the tool recommend a license, enable interactive recommendation mode:

```sh
license --interactive
license -i
```

The tool asks a short series of yes/no questions about copyleft, patent protection, network use, library linking, and file-level requirements, then selects the closest supported license. `--interactive` cannot be combined with `--license`.

Display the installed version with either form:

```sh
license --version
license -v
```

## Supported Licenses

The selection is ordered by general popularity:

| License | Summary |
| --- | --- |
| MIT | Simple and permissive, with minimal restrictions on reuse. |
| Apache 2.0 | Permissive, with an explicit patent grant from contributors. |
| GPLv3 | Strong copyleft requiring derivatives to remain open source. |
| LGPLv3 | Library-focused copyleft that permits proprietary linking. |
| AGPLv3 | Strong copyleft that also covers software used over a network. |
| MPL 2.0 | File-level copyleft that can be combined with proprietary code. |

## Development

```sh
go test ./...
go build .
```

This project itself is licensed under the [Mozilla Public License 2.0](LICENSE).
