# Ratings Updated Flow — MatchResultsCalculated → RatingsUpdated

Fluxo completo de MatchResultsCalculated até RatingsUpdated: consumo, cálculo de MMR (Elo), persistência com resource ownership e publicação.

## Fluxo

1. **MatchResultsCalculated** (produtor: match-making-api via MatchCompleted): emitido após persistir resultados da partida.
2. **Tópico**: `matchmaking.matches.results`
3. **Consumer**: `consumer-ratings-updated` consome, valida resource ownership e payload.
4. **Handler**: calcula deltas de rating (Elo), persiste em MongoDB, publica **RatingsUpdated**.
5. **Tópico de saída**: `matchmaking.ratings.updated` — consumido por replay-api para leaderboards e skill-based matchmaking (#30).

## Payload MatchResultsCalculated (Proto)

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `match_id` | string | Sim | Identificador da partida |
| `player_ids` | string[] | Sim | IDs dos jogadores na partida |
| `winner_team_id` | string | Não | ID do time vencedor; vazio se empate. Para 1v1, pode ser player_id |
| `is_draw` | bool | Sim | true se empate |
| `completed_at_epoch_ms` | int64 | Sim | Timestamp de conclusão (epoch ms) |
| `calculated_at_epoch_ms` | int64 | Sim | Quando os resultados foram calculados |
| `tenant_id` | string | Sim | Tenant para multi-tenancy |
| `client_id` | string | Sim | Cliente para acesso controlado |
| `resource_owner_id` | string | Sim | RID para autorização |
| `game_id` | string | Não | Para ratings por jogo; vazio se não disponível |

## Payload RatingsUpdated (Proto)

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `match_id` | string | Sim | Identificador da partida |
| `deltas` | PlayerRatingDelta[] | Sim | Delta de MMR por jogador |
| `updated_at_epoch_ms` | int64 | Sim | Quando as ratings foram atualizadas |
| `tenant_id` | string | Sim | Tenant para multi-tenancy |
| `client_id` | string | Sim | Cliente para acesso controlado |
| `resource_owner_id` | string | Sim | Para autorização |
| `game_id` | string | Não | Ratings por jogo |
| `audit_trail` | RatingAuditTrail | Sim | algorithm_version, reason, updated_by |

### PlayerRatingDelta

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `player_id` | string | ID do jogador |
| `mmr_before` | int32 | MMR antes da partida |
| `mmr_after` | int32 | MMR após a partida |
| `delta` | int32 | Variação (positivo ou negativo) |

## Regras de validação

| Regra | Descrição |
|-------|-----------|
| ResourceOwnershipRule | `resource_owner_id`, `match_id`, `tenant_id`, `client_id` obrigatórios |
| MatchIDRule | `match_id` não vazio |
| Idempotência | Se ratings já processados para `match_id`, não reprocessa nem republica |

## Algoritmo de rating (Elo)

- **Versão**: elo-v1
- **K factor**: 32
- **MMR padrão**: 1000 (novos jogadores)
- **Fórmula**: E_a = 1/(1+10^((R_b-R_a)/400)); delta = K * (S - E_a)
- Suporta 2 jogadores (1v1) e N jogadores (média de oponentes)

## Persistência (MongoDB)

- **player_ratings**: MMR por player_id+game_id+tenant_id+client_id
- **rating_audit**: Histórico de alterações (audit trail)
- **ratings_processed_matches**: Idempotência por match_id

## Modos de falha

| Falha | Tratamento |
|-------|------------|
| Validação de resource ownership | Log, skip. Mensagem não commitada. |
| Payload inválido | Log, skip. Não retry. |
| Erro ao salvar rating | Retorna erro → consumer retry. |
| Erro ao publicar RatingsUpdated | Retorna erro → consumer retry. |
| Já processado (idempotência) | HasProcessed retorna true → skip. |

## Referências

- `.docs/EVENT_SCHEMAS.md` — schema MatchResultsCalculated e RatingsUpdated
- `pkg/infra/kafka/match_results_calculated_consumer.go` — consumer
- `pkg/domain/pairing/usecases/ratings_updated_handler.go` — handler
- `pkg/domain/pairing/usecases/rating_algorithm.go` — algoritmo Elo
