# Capability boundary

The neutral asset can:

- read explicitly named local files;
- strictly decode, canonicalize and hash N1 artifacts;
- verify Ed25519 signatures with explicitly supplied public trust-set bytes;
- write a new local artifact without overwriting an existing path;
- emit payload-neutral verification diagnostics.

It cannot:

- use network, HTTP, RPC, database, product SDK or process-execution adapters;
- read or mutate UniRouter or X-Insight repositories or runtime state;
- access Target, cloud APIs, credentials, private keys or system trust stores;
- deploy, migrate, reset, BOOTSTRAP, create Cursor/checkpoint/epoch/credential,
  control a worker or send Provider/Feed/upstream/Gateway traffic;
- sign an authorization;
- generate product facts or validate the contents of product-owned receipts;
- generate terminal result R in N1.

Automated tests parse non-test Go imports, reject transport/database/process
capabilities, reject signing APIs and private-key material, enforce the
standard-library-only go.mod, and verify the positive/negative offline
fixtures.

The fixture signature is verification-only. The committed repository contains
its public key and detached signature, not the temporary fixture private key.
