# Match Results Flow — MatchCompleted → MatchResultsCalculated

Fluxo completo de MatchCompleted (game server) até MatchResultsCalculated (match-making-api): consumo, persistência com resource ownership e publicação.

## Fluxo

1. **MatchCompleted** (produtor: game server / replay-api): emitido quando a partida termina.
2. **Tópico**: `matchmaking.matches.completed`
3. **Consumer**: `consumer-match-completed` consome, valida resource ownership e payload.
4. **Handler**: calcula resultados (winner, is_draw, etc.), persiste em MongoDB, publica **MatchResultsCalculated**.
5. **Tópico de saída**: `matchmaking.matches.results` — consumido por replay-api para ratings (#30) e prêmios (#31).

## Payload MatchCompleted (Proto)

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `match_id` | string | Sim | Identificador da partida |
| `player_ids` | string[] | Sim | IDs dos jogadores na partida |
| `winner_team_id` | string | Não | ID do time vencedor; vazio se empate |
| `is_draw` | bool | Sim | true se empate |
| `completed_at_epoch_ms` | int64 | Sim | Timestamp de conclusão (epoch ms) |
| `tenant_id` | string | Sim | Tenant para multi-tenancy |
| `client_id` | string | Sim | Cliente para acesso controlado |

## Payload MatchResultsCalculated (Proto)

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `match_id` | string | Sim | Identificador da partida |
| `player_ids` | string[] | Sim | IDs dos jogadores na partida |
| `winner_team_id` | string | Não | ID do time vencedor; vazio se empate |
| `is_draw` | bool | Sim | true se empate |
| `completed_at_epoch_ms` | int64 | Sim | Timestamp de conclusão (epoch ms) |
| `calculated_at_epoch_ms` | int64 | Sim | Quando os resultados foram calculados (auditoria) |
| `tenant_id` | string | Sim | Tenant para multi-tenancy |
| `client_id` | string | Sim | Cliente para acesso controlado |
| `resource_owner_id` | string | Sim | RID para autorização (do envelope) |

## Regras de validação

| Regra | Descrição |
|-------|-----------|
| ResourceOwnershipRule | `resource_owner_id` obrigatório no envelope; `match_id`, `tenant_id`, `client_id` no payload |
| MatchIDRule | `match_id` não vazio |
| Idempotência | Se MatchResult já existe para `match_id`, não reprocessa nem republica |

## Persistência (MongoDB)

- Coleção: `match_results` (ou equivalente via `config.MongoDB.DBName`)
- Documento keyed por `match_id` para idempotência
- Campos armazenados: `match_id`, `player_ids`, `winner_team_id`, `is_draw`, `completed_at_epoch_ms`, `calculated_at_epoch_ms`, `tenant_id`, `client_id`, `resource_owner_id`
- Acesso: restrito por tenant/client/owner (resource ownership)

## Modos de falha

| Falha | Tratamento |
|-------|------------|
| Validação de resource ownership | Log, skip. Mensagem não commitada (consumer pode retry). |
| Payload inválido | Log, skip. Não retry — dados inválidos. |
| Erro de persistência | Retorna erro → consumer retry. |
| Erro ao publicar MatchResultsCalculated | Retorna erro → consumer retry. |
| Já processado (idempotência) | GetByMatchID retorna existente → não reprocessa nem republica. |

## Idempotência

- Antes de persistir, `GetByMatchID(match_id)` verifica se já existe.
- Se existir, handler retorna nil (sem erro) — não republica MatchResultsCalculated.
- Chave MongoDB `_id = match_id` garante unicidade.

## Ordenação

- MatchCompleted → persistência → MatchResultsCalculated.
- Recomenda-se particionar por `match_id` no tópico de entrada se necessário para ordem por partida.

## Referências

- `.docs/EVENT_SCHEMAS.md` — schema MatchCompleted e MatchResultsCalculated
- `pkg/infra/kafka/match_completed_consumer.go` — consumer
- `pkg/domain/pairing/usecases/match_completed_handler.go` — handler
- `pkg/domain/pairing/entities/match_result.go` — entidade
- `pkg/infra/db/mongodb/match_result_mongodb.go` — repositório
