# Foigent integration-control

**Verify offline authorization artifacts before products act on them.**

[简体中文](README.zh-CN.md) · [Contributing](../CONTRIBUTING.md) ·
[Contracts](../docs/CONTRACTS.md) · [License](../LICENSE)

integration-control is a small, product-neutral Go CLI and contract library.
It validates a campaign manifest, product acknowledgements and authorization
artifacts, including their exact identities, canonical bytes and signatures.
UniRouter and X-Insight consume these facts through their own reviewed paths.

It operates on local files. It does not contact either product, hold private
keys, sign authorizations, access databases or deploy services. A valid artifact
does not itself grant permission to perform an operation.

## Try it locally

Install Go 1.25.12, then clone this repository once it is public:

```sh
git clone https://github.com/foigent/integration-control.git
cd integration-control
GOTOOLCHAIN=local go test ./... -count=1
GOTOOLCHAIN=local go run ./cmd/foigent-integration-control schemas check --dir contracts/v1
```

No provider key, service or paid API call is required. The tests use local
fixtures. Read [the contract guide](../docs/CONTRACTS.md) before supplying your
own artifacts; build commands create new outputs and refuse to overwrite them.

## What is included

- Strict Go decoders, validators and schema-specific canonicalization.
- Five projected JSON Schemas with byte-for-byte drift checks.
- Positive and negative synthetic fixtures.
- Reproducible release verification and a fresh-consumer qualification path.

The preserved N1 distribution is `v0.1.0-candidate.1`. Its source, manifest,
two platform binaries and checksums retain their original identities. This
open-source preparation does not introduce N2 runtime behavior or a new release.
The [release-bound overview](../README.md), [capability boundary](../docs/CAPABILITY-BOUNDARY.md)
and [distribution record](../docs/DURABLE-DISTRIBUTION.md) retain their original
scope and chronology.

## Contribute

Start with [Contributing](../CONTRIBUTING.md). Useful changes include clearer
error explanations, reproducible parser defects and focused negative fixtures.
Discuss contract, canonicalization or release-identity changes before coding.
English is the primary language; Chinese questions and contributions are welcome.

Use [Issues](https://github.com/foigent/integration-control/issues) for ordinary
questions, [Security](../SECURITY.md) for sensitive reports, and follow the
[Code of Conduct](../CODE_OF_CONDUCT.md).

## License and release status

Project-owned material is licensed under [Apache-2.0](../LICENSE), like
UniRouter. [Third-party notices](../THIRD_PARTY_NOTICES.md) remain applicable.
[Licensing and historical distributions](../docs/LICENSING.md) explains how
the grant covers preserved consumer snapshots and how notices accompany them.

This is an open-source preparation branch. Public visibility, anonymous release
downloads and confidential-reporting availability still require publication
verification. The fixed historical release assets are not rewritten here.
