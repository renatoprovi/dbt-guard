-- Camada Staging: espelha a source raw.raw_clientes.
-- Não altera nomes de colunas; dependência direta da source (depends_on no manifest).
select *
from {{ source('raw', 'raw_clientes') }}
