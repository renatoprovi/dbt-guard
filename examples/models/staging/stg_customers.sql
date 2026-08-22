-- Staging layer: mirrors the raw.raw_customers source.
-- Does not rename columns; direct dependency on the source (depends_on in the manifest).
select
    ssn,
    name
from {{ source('raw', 'raw_customers') }}
