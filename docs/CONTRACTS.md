# N1 contracts and authority

## One-way hash direction

N1 has one valid direction:

1. M binds immutable pre-TRAFFIC inputs.
2. UniRouter and X-Insight independently verify only their product component
   and emit canonical acknowledgements bound to M.
3. A binds M, both exact acknowledgement byte digests, the accepted actions,
   decision identity, confirmation digest and authorization window.
4. A detached purpose-specific signature binds canonical A.
5. Later P binds M and A; X records observed P; later R binds the terminal
   receipts.

M and A must not contain P, X import-receipt or R fields. This removes the
package/manifest hash cycle and prevents an observed execution result from
being relabelled as a preauthorized input.

## Manifest M

foigent.integration-campaign-manifest.v1 is limited to
poc_source_evidence. Its builder reads exact bytes for:

- the Coordinator decision artifact;
- typed Source binding and frozen request set;
- product-owned Source-evidence trust set;
- neutral authorization trust set;
- each product's immutable component contract, release identity and scope
  receipt.

The builder checks declared schema identities where the reference is typed and
computes every digest. A draft containing any pre-filled derived digest fails
closed. There is no target_identity_sha256: product scope receipts are the
typed immutable authorities and no redundant operator-entered digest is
accepted.

For the accepted UniRouter component, Source binding schema
unirouter.source-evidence.source-scope.v1 identifies product, the
request-set-derived source_instance, neutral_campaign_authorization as the
identity authority, runtime_observation as not_claimed, and
unirouter.completed-execution.v1 as the Feed snapshot schema. It deliberately
contains no runtime tree, config digest or reference-mode assertion. The
installed release input is likewise limited to the product receipt facts:
sequence, version, canonical commit, package digest, signing-key identity,
database schema and observation time; it does not claim a runtime tree.

## Component acknowledgement boundary

The common acknowledgement envelope is defined under
$defs.component_acknowledgement in the authorization schema, not as a sixth
schema. Neutral code parses only these common facts:

- product and acknowledgement schema identity;
- exact M digest;
- exact component contract and scope-receipt references from M;
- product-owned verification-receipt reference;
- product verifier-release reference;
- accepted status and canonical verification time.

Neutral verification checks canonical acknowledgement bytes, M binding,
component reference equality, exact acknowledgement digest and chronology from
M freeze through user confirmation. It does not parse the product verification
receipt, query product state or claim the product check is substantively true.
Each product remains the sole authority for its receipt.

## Authorization A and detached signature

foigent.integration-campaign-authorization.v1 stores only the minimum user
decision facts: authorization and decision identifiers, M digest, two
acknowledgement references, exact actions and window, confirmed_at, and
confirmation_sha256. The original user confirmation text remains in the
Coordinator's restricted ledger and is never copied into A, fixtures or a
product package.

The signed bytes are:

    foigent.integration-campaign-authorization.v1 NUL
    integration_campaign_authorization NUL
    canonical_authorization_bytes

Only Ed25519 public verification is implemented. This repository contains no
private key and exposes no signing command.

## Two independent trust purposes

M contains two independently named bindings:

- source_evidence_signing: purpose source_evidence, domain
  unirouter.source-evidence.component.v1, producer operational key reference,
  cryptographic public-key ID and product trust-set digest;
- authorization_signing: purpose integration_campaign_authorization, domain
  foigent.integration-campaign-authorization.v1, and the neutral
  authorization trust-set digest.

A.sig.trust_set_sha256 and the authorization verifier may match only
M.authorization_signing.trust_set_sha256. A future UniRouter package verifier
may use only M.source_evidence_signing.trust_set_sha256. Trust-set equality,
wrong-purpose domains, Source key references, upgrade/release authorities,
disabled keys, out-of-window keys and ambiguous key selection all fail closed.
Even when identical public-key bytes are intentionally present under both
purposes, neither trust set confers the other purpose.

Authorization and trust-key validity windows are half-open:
not_before is inclusive and not_after or valid_until is exclusive.

## Canonical and schema authority

The normative executable authority is:

1. valid UTF-8 and bounded size;
2. recursive duplicate-key rejection;
3. strict typed decode with unknown fields rejected;
4. semantic validation;
5. schema-specific Go-struct JSON canonicalization.

Five JSON Schema documents are deterministic standard-library projections:

- manifest;
- authorization, including acknowledgement definitions;
- detached signature;
- authorization trust set;
- release metadata.

Tests compare projected bytes with every committed schema and compare the
top-level typed field set with projected properties and required fields. JSON
Schema is documentation and ecosystem aid; it cannot override the typed
validator.

## DATA_RESET reuse boundary

Future DATA_RESET may reuse only these narrow primitives: immutable manifest,
component acknowledgement references, one exact authorization, detached
purpose-specific verification and forward-only result binding. DATA_RESET
requires its own phase-specific scope and product-owned PREPARE/DELETE
receipts. N1 neither defines that scope nor creates a generic phase
orchestrator, reset engine or journal.

## Crash and replay behavior

Neutral operations have no external side effects. Canonical output is a pure
function of exact input bytes. Output paths are create-only:

- a crash before creation is recovered by rerunning;
- a completed artifact is verified by exact bytes and digest;
- an existing path is never silently replaced;
- a different input produces a different M or A digest.

Product-side crash recovery remains in each product's existing durable receipt
and journal mechanisms.
