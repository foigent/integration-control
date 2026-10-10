# Contributing

English is the primary language for documentation, Issue/PR titles and release
summaries. Simplified Chinese questions and contributions are welcome.
Use `.zh-CN.md` for a Chinese translation and link both versions.

For a small documentation correction, describe the problem and your link or
readback check. For a parser defect or negative fixture, include a synthetic
reproduction and the expected rejection. Discuss changes to canonicalization,
authorization or release identity before implementation. Follow
[the neutral boundary](AGENTS.md) and [Code of Conduct](CODE_OF_CONDUCT.md).

## Development and validation

Use Go 1.25.12. A contributor's fork can run these offline checks:

```sh
GOTOOLCHAIN=local go test ./... -count=1
GOTOOLCHAIN=local go test -race ./... -count=1
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local go run ./cmd/foigent-integration-control schemas check --dir contracts/v1
```

No service, provider account, signing key or product checkout is needed.
Report what ran and any skipped checks. Keep unrelated work separate.

The full distribution gate intentionally checks the official Git origin,
historical tags, two identical builds and all five frozen assets. It is run by
upstream PR CI using `N1_REMOTE_POLICY=github-origin`; an ordinary fork's origin
does not satisfy that authority check. Do not change origin validation or call
a partial local run a full release qualification. See
[the existing distribution guide](docs/DURABLE-DISTRIBUTION.md).

## Contribution license

Project-owned contributions are submitted under [Apache-2.0](LICENSE).
Retain third-party licenses and attribution. Sign off your own commits under
the [Developer Certificate of Origin 1.1](https://developercertificate.org/):

```sh
git commit -s -m "docs: explain a validation error"
```

This certifies your right to submit the contribution; it does not transfer
copyright. Do not manufacture someone else's sign-off. Maintainers review
sign-offs; no automated DCO gate is claimed here.

Use [Issues](https://github.com/foigent/integration-control/issues) for ordinary
questions and bounded proposals. Follow [Security](SECURITY.md) for sensitive
reports. Community support depends on maintainer availability.
