-- Analysis layer: exposes data for analysis.
-- Column rename: ssn (PII) -> student_document.
-- The lineage validator must detect that this model descends from PII (source) and
-- check for proper masking before allowing it into a public layer.
select
    ssn as student_document,  -- ssn -> student_document mapping (column is still PII at the source)
    name
from {{ ref('stg_customers') }}
