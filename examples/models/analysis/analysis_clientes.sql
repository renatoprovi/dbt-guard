-- Camada Analysis: expõe dados para análise.
-- Transformação de nome de coluna: cpf (PII) → documento_aluno.
-- O validador de linhagem deve detectar que este modelo descende de PII (source) e
-- verificar se há mascaramento adequado antes de permitir em camada pública.
select
    cpf as documento_aluno,  -- mapeamento cpf → documento_aluno (coluna ainda PII na origem)
    nome
from {{ ref('stg_clientes') }}
