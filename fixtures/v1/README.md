# N1 conformance fixtures

The valid directory is a fully offline, synthetic preauthorization chain. It
contains only public verification material and a detached fixture signature.
It is not a user authorization, Target fact, product release or TRAFFIC
authority.

The normative identities exercised by the positive fixture are:

- unirouter.source-evidence.source-scope.v1;
- unirouter.source-evidence.request-set.v1;
- unirouter.source-evidence.component-package.v1.

The Source input deliberately claims no runtime observation. Product-owned
release, scope and verification documents contain synthetic values solely to
exercise neutral byte binding; their schemas and bytes are not interoperable
product authority and must later be replaced by each Owner's immutable
accepted artifacts.

The positive fixture intentionally uses the same public-key identity in the
product-owned Source-evidence metadata and the neutral authorization trust set,
while pinning different trust-set digests and purposes. This proves that key
bytes alone do not confer cross-purpose authority.

The invalid directory contains stable fail-closed examples for wrong Source
and request-set contract identities, wrong-purpose trust, Source-key reuse in
an authorization signature, non-accepted product status and duplicate JSON
keys. The broader mutation matrix is generated in-memory by tests without
storing private keys or user confirmation text.
