# Foigent integration-control

Product-neutral, offline control artifacts for joint UniRouter/X-Insight
integration campaigns.

N1 freezes this one-way preauthorization chain:

    manifest M -> product acknowledgements -> authorization A -> detached signature

A later signed UniRouter package binds M and A. X-Insight then binds M, A and
the observed package in its import receipt. Terminal result R is a later N2
artifact and can only point forward across those completed facts; M and A never
contain package, import or result digests.

The executable authority is the strict Go decoder, validator and
schema-specific canonicalizer. The five JSON Schema files are deterministic
projections checked byte-for-byte in tests.

## Offline CLI

The repository is pinned to go 1.25.0 with toolchain go1.25.12.

    go run ./cmd/foigent-integration-control schemas check --dir contracts/v1
    go run ./cmd/foigent-integration-control manifest verify --manifest MANIFEST
    go run ./cmd/foigent-integration-control ack verify --manifest MANIFEST --ack PRODUCT_ACK
    go run ./cmd/foigent-integration-control trust-set verify --trust-set AUTHORIZATION_TRUST_SET
    go run ./cmd/foigent-integration-control authorization verify [all exact N1 artifacts]

manifest build consumes exact decision, Source binding, request-set, trust,
component-contract, product-release and product-scope files and computes every
content digest itself. authorization build consumes the exact canonical
acknowledgements and a confirmation input; only the confirmation SHA-256 is
emitted. Build outputs use create-only semantics and never overwrite an
existing artifact.

There is deliberately no signing command, private key, product adapter,
network transport, database client, deployment control, worker control,
BOOTSTRAP action or TRAFFIC action.

See docs/CONTRACTS.md, docs/CAPABILITY-BOUNDARY.md and docs/RELEASE.md.
