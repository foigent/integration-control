# Neutral integration-control boundary

- This repository owns only product-neutral, offline campaign contracts,
  canonicalization, verification, fixtures, and release metadata.
- It must not import or access UniRouter or X-Insight product code, databases,
  APIs, credentials, deployment controls, workers, BOOTSTRAP state, or traffic.
- It must not contain private keys or provide a signing command.
- N1 covers preauthorization artifacts only. Terminal result V1 is a separate
  N2 batch and must be frozen before real TRAFFIC.

