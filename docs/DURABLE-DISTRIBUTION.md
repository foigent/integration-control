# N1 durable distribution candidate

## Frozen release identity

Distribution preserves the accepted N1 candidate without rebuilding release
history or changing any accepted bytes:

- source C2: `12f1ede8ae85e9d311a8150694eea54ff6f02950`, tree
  `e6945fa071afd6a0b9317d3ec862cbcbf47a0154`;
- release R: `c9e5a77dd0416654d8e46efb09f18a6ecca026f8`, tree
  `325863188f625c07e9f10a7a3eae67f9574de198`;
- lightweight unsigned candidate tag: `v0.1.0-candidate.1`, pointing to R;
- release manifest SHA-256:
  `0604100b8cb8c8a728fc091d1c9269eafbc3cb70449222ab6c10657bca33d3e5`;
- `SHA256SUMS` SHA-256:
  `d1536bf3f68bf35293470491c57cfc78f784c61f4d6291a3b9b9ba1b3f34e848`.

The exact five release assets are the two platform executables, the source
archive, the canonical release manifest and `SHA256SUMS`. The source archive
contains all five N1 contracts and all accepted positive and negative fixtures.
No sixth mutable receipt or latest alias is an authority.

## Local qualification

All work paths must be new. Scripts retain their work directories on success
or failure and never recursively clean them.

```sh
scripts/rebuild-n1-candidate.sh /private/tmp/foigent-n1-rebuild-UNIQUE
scripts/stage-n1-release.sh \
  /private/tmp/foigent-n1-rebuild-UNIQUE/release-assets \
  /private/tmp/foigent-n1-simulated-release-UNIQUE
scripts/qualify-n1-release.sh \
  /private/tmp/foigent-n1-simulated-release-UNIQUE \
  /private/tmp/foigent-n1-consumer-UNIQUE
```

`rebuild-n1-candidate.sh` makes two independent archives and two independent
cross-builds from exact C2. It pins Go 1.25.12, disables toolchain and module
downloads, disables CGO, uses `-trimpath -ldflags=-buildid=`, compares both
builds byte-for-byte and requires the accepted sizes and hashes.

`stage-n1-release.sh` is create-only. Repeating it over byte-identical assets
is a no-op; an existing conflicting byte sequence or any unexpected file is a
hard failure. `qualify-n1-release.sh` starts from the flat five-file release,
checks every size and SHA, extracts the exact source archive, and uses only the
downloaded host binary to verify the release, schema projection and complete
M-to-acknowledgements-to-A public verification chain.

The aggregate local/CI gate is:

```sh
scripts/full-d1-gate.sh /private/tmp/foigent-n1-d1-gate-UNIQUE
```

It also proves exact restaging is idempotent, a conflicting destination is not
overwritten, and an extra release asset is rejected. The committed GitHub
workflow has read-only contents permission, does not retain checkout
credentials, performs no upload or release action, and runs the same gate on a
fresh Linux amd64 runner.

## External publication preview — not executed or authorized by D1

The recommended durable coordinate is the private GitHub repository
`foigent/integration-control` plus the exact GitHub Release tag
`v0.1.0-candidate.1`. A later external authorization must bind the
Coordinator-accepted D1 commit and these five immutable asset hashes. It must
cover, in one narrow baton:

1. create the private repository and set its only remote to the exact GitHub
   coordinate;
2. push the accepted history to `main` without force and push the existing tag
   without rewriting it;
3. create a draft Release whose target is exact R, upload exactly the five
   assets, download them into a fresh directory, and pass local qualification;
4. publish that already-qualified draft without replacing any asset;
5. provision a separate cross-repository CI identity with read-only Contents
   access to this private repository; it receives no write, release, product,
   environment, deployment or secret-administration authority.

The operator preview below is intentionally documentation, not an executable
publisher. `ACCEPTED_D1_HEAD` must be replaced by the exact checkpoint accepted
after independent review, and `FRESH_ASSET_DIR` must be the output of a fresh
successful D1 rebuild. No command may proceed if a readback differs.

```sh
git rev-parse HEAD^{commit}
git rev-parse HEAD^{tree}
git rev-parse v0.1.0-candidate.1^{commit}
git remote -v

gh repo create foigent/integration-control --private
git remote add origin https://github.com/foigent/integration-control.git
git push origin ACCEPTED_D1_HEAD:refs/heads/main
git push origin refs/tags/v0.1.0-candidate.1:refs/tags/v0.1.0-candidate.1

gh release create v0.1.0-candidate.1 \
  --repo foigent/integration-control \
  --verify-tag \
  --target c9e5a77dd0416654d8e46efb09f18a6ecca026f8 \
  --draft \
  --title 'integration-control v0.1.0-candidate.1' \
  --notes 'Exact offline N1 preauthorization control asset; candidate only.'

gh release upload v0.1.0-candidate.1 --repo foigent/integration-control \
  FRESH_ASSET_DIR/foigent-integration-control-v0.1.0-candidate.1-darwin-arm64 \
  FRESH_ASSET_DIR/foigent-integration-control-v0.1.0-candidate.1-linux-amd64 \
  FRESH_ASSET_DIR/foigent-integration-control-v0.1.0-candidate.1-source.tar \
  FRESH_ASSET_DIR/foigent.integration-control.release.v1.json \
  FRESH_ASSET_DIR/SHA256SUMS

READBACK_DIR="$(mktemp -d)"
gh release download v0.1.0-candidate.1 \
  --repo foigent/integration-control \
  --dir "$READBACK_DIR"
scripts/qualify-n1-release.sh "$READBACK_DIR" "$(mktemp -d)/consumer"
gh release edit v0.1.0-candidate.1 \
  --repo foigent/integration-control \
  --draft=false
```

The later authorization must additionally name the selected read-only GitHub
App or fine-grained token custody owner and the two exact product repositories.
The token value is never printed, stored in this repository, placed in a
fixture, or passed to the neutral binary.

## Explicit exclusions

D1 does not create a remote, push, publish, upload, sign, provision trust or
credentials, edit either product, access Target, open a product database or
API, create BOOTSTRAP state, run workers, or send Provider/Feed/upstream/
Gateway traffic. N2 terminal-result V1 is still `NOT_STARTED`; it must be
frozen after immutable released product receipts exist and before real TRAFFIC.
