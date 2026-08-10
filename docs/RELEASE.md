# Local release process

The accepted first-batch sequence is deliberately non-self-referential:

1. Baseline B establishes the local no-remote repository and capability
   boundary.
2. Implementation checkpoint C contains contracts, validator, canonicalizer,
   CLI, fixtures, tests and documentation.
3. Coordinator review accepts or rejects exact C.
4. Only after acceptance, two clean builds from C must produce byte-identical
   artifacts.
5. Release metadata commit R records C's source commit/tree, five-schema
   aggregate, valid/invalid fixture aggregates, exact artifact hashes, full
   gates and zero external effects.
6. A local candidate tag points to R.

R is not part of its own source or artifact digest. C never claims to be a
release, and local main is not advanced to C before Coordinator review. There
is no remote or push in N1.

Release metadata is constrained to go1.25.0 plus toolchain go1.25.12,
offline_files_only, sorted artifacts/gates, double-build equality and zero
Target/product/key/worker/traffic effects.

N2 terminal-result V1 is a separate reviewed batch after immutable UniRouter
package and X-Insight receipt schemas exist. N2 must be frozen before any real
TRAFFIC baton; it is not implemented or simulated by N1 release metadata.
