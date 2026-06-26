# Roadmap

Implementation status for dbt-guard governance capabilities.

---

## Phase 1: Manifest graph — Done

Parse `manifest.json` (dbt v10+); expose nodes, sources, and `depends_on`.

| Deliverable | Location |
|-------------|----------|
| `LoadManifest`, `Manifest`, `ManifestNode`, `SourceDef` | `internal/parser/manifest.go` |
| `NodeIDsWithPII`, `SourceIDsWithPII` | `internal/parser/manifest.go` |
| CLI | `dbt-guard manifest <path>` |
| Tests | `internal/parser/manifest_test.go`, `testdata/manifest_minimal.json` |

---

## Phase 2: Sensitivity propagation (DFS) — Done

Walk lineage parents; mark nodes that declare or inherit PII.

| Deliverable | Location |
|-------------|----------|
| `IsSensitive`, `LineagePathToPII` | `internal/parser/lineage.go` |
| CLI | `dbt-guard sensitive <path>` |
| Tests | `internal/parser/lineage_test.go` |

---

## Phase 3: Validate gatekeeper — Done

Fail CI when restricted-layer models carry unmasked PII lineage.

| Deliverable | Location |
|-------------|----------|
| `RunValidate`, `Violation` | `internal/validator/validate.go` |
| `LayerPolicy`, `LoadLayerPolicy` | `internal/config/layers.go` |
| `RestrictedNodeIDs` | `internal/parser/layers.go` |
| CLI flags | `--config`, `--allowed`, `--restricted` |
| CLI | `dbt-guard validate <path> [--config dbt-guard.yml]` |
| Tests | `validate_test.go`, `layers_test.go`, `manifest_with_confidential.json` |

**Default policy:** `/analysis/` restricted (backward compatible).

**Custom policy:** `pii_allowed` / `pii_restricted` path patterns in `dbt-guard.yml`.

---

## Future (not scheduled)

- Schema- or FQN-based matching (in addition to folder paths).
- Column-level masking validation (today: model-level `meta.masked`).
- Integration with warehouse query logs for usage-based governance.
- Policy-as-code in dbt `meta` (e.g. `pii_policy: block|allow`) as alternative to YAML.
