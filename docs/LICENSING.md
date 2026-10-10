# Licensing and preserved distributions

On 2026-10-10 the project owner approved Apache-2.0 for the project-owned source,
documentation, schemas, fixtures and test tools in this repository, matching
UniRouter. This grant includes the project-owned material in the preserved
historical sources below; it does not claim that older archives already
contained license files or change any historical artifact bytes.

- N1 source `12f1ede8ae85e9d311a8150694eea54ff6f02950`, tree
  `e6945fa071afd6a0b9317d3ec862cbcbf47a0154`.
- N1 release metadata `c9e5a77dd0416654d8e46efb09f18a6ecca026f8`, tag
  `v0.1.0-candidate.1`.
- Consumer qualifier `db10730d64a9302887ad59c53921e98c757927e7`.

Read the full [Apache-2.0 license](../LICENSE), [NOTICE](../NOTICE) and
[third-party notices](../THIRD_PARTY_NOTICES.md). Material identified as third
party retains its own license and attribution. This grant does not relicense
dependencies, grant trademark rights or provide a hosted-service commitment.

## Supply the license materials with historical artifacts

A newer repository LICENSE is not inserted into an older source archive,
binary or pinned action snapshot. When redistributing a preserved artifact,
provide these files alongside it, retaining their relative layout in a separate
license-materials directory:

- `LICENSE`, `NOTICE`, `THIRD_PARTY_NOTICES.md`, `docs/LICENSING.md`;
- `LICENSES/Go-LICENSE`;
- the unchanged vendored license files listed in `THIRD_PARTY_NOTICES.md`,
  where applicable.

Bind that accompanying directory to the exact licensed repository commit and
the historical artifact identity in the distribution description. Do not claim
that it changes the original release manifest, signature or archive digest.
For the N1 distribution, keep the strict five-file qualification directory
separate; do not add a sixth asset to that historical release or its verifier
input. A public download must make the complete accompanying license materials
available and verify the association before announcement.

New CI carrier commits can include these files while preserving all released
protocol and artifact bytes. Carrier identity and canonical release identity
are different: updating a checkout/action pin does not rewrite the protocol
version, product pin, historical source identity or release digest.

This preparation changes licensing and community documentation only. Actual
public access and license-material delivery require publication verification.

## Preserved N1 asset identities

These are the original qualification inputs, not a new release manifest.
The separately supplied license materials do not change these bytes.

| Asset | SHA-256 |
| --- | --- |
| `foigent-integration-control-v0.1.0-candidate.1-darwin-arm64` | `ba574ac25abd063914d785a34d097b14a1cf09021a953c40803414900c1f3dc1` |
| `foigent-integration-control-v0.1.0-candidate.1-linux-amd64` | `b37a9b92ef237844327911b9a4b1a1266b2970f3acbc50de428aa7d518a1fbb8` |
| `foigent-integration-control-v0.1.0-candidate.1-source.tar` | `360cbc70cbbe84361c23e9f89f3aeed2689b235197b2ebcb7b436187c9aedfef` |
| `foigent.integration-control.release.v1.json` | `0604100b8cb8c8a728fc091d1c9269eafbc3cb70449222ab6c10657bca33d3e5` |
| `SHA256SUMS` | `d1536bf3f68bf35293470491c57cfc78f784c61f4d6291a3b9b9ba1b3f34e848` |
