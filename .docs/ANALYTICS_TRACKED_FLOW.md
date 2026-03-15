# Analytics Tracked Flow — MatchResultsCalculated → AnalyticsTracked

Fluxo completo de MatchResultsCalculated até AnalyticsTracked: consumo, produção de eventos para pipeline de analytics (#32). Analytics service consome para dashboards, reporting e decisões de produto.

## Fluxo

1. **MatchResultsCalculated** (produtor: match-making-api via MatchCompleted): emitido após persistir resultados da partida.
2. **Tópico**: `matchmaking.matches.results`
3. **Consumer**: `consumer-analytics-tracked` consome, valida resource ownership e payload.
4. **Handler**: produz **AnalyticsTracked** com match_id, outcome, player_ids, completed_at, resource ownership.
5. **Tópico de saída**: `matchmaking.analytics.tracked` — consumido por Analytics Service para dashboards e reporting.

## Payload AnalyticsTracked (Proto)

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `match_id` | string | Sim | Identificador da partida |
| `player_ids` | string[] | Sim | IDs dos jogadores na partida |
| `winner_team_id` | string | Não | ID do time vencedor; vazio se empate |
| `is_draw` | bool | Sim | true se empate |
| `completed_at_epoch_ms` | int64 | Sim | Timestamp de conclusão |
| `tracked_at_epoch_ms` | int64 | Sim | Quando o evento foi produzido (audit) |
| `duration_ms` | int64 | Sim | 0 se desconhecido; MatchResultsCalculated não tem start time |
| `tenant_id` | string | Sim | Tenant para multi-tenancy |
| `client_id` | string | Sim | Cliente para acesso controlado |
| `resource_owner_id` | string | Sim | Isolamento tenant/client; consumidores devem respeitar |
| `game_id` | string | Não | Para analytics por jogo |
| `lobby_id` | string | Não | Referência ao lobby/torneio |
| `prize_pool_id` | string | Não | Referência ao prize pool |

## Regras de validação

| Regra | Descrição |
|-------|-----------|
| ResourceOwnershipRule | `resource_owner_id`, `match_id`, `tenant_id`, `client_id` obrigatórios |
| Idempotência | Se analytics já trackeado para `match_id`, não republica |

## Enriquecimento de dados

- **Rating deltas** e **prizes**: Analytics Service pode consumir também `RatingsUpdated` e `PrizeDistributed` e fazer join por `match_id` para enriquecer dashboards.
- **Duration**: 0 no payload inicial; pode ser calculado se MatchStarted estiver disponível (start_timestamp_epoch_ms).

## Persistência (MongoDB)

- **analytics_tracked_matches**: Idempotência por `match_id`

## Modos de falha

| Falha | Tratamento |
|-------|------------|
| Validação de resource ownership | Log, skip. Mensagem não commitada. |
| Payload inválido | Log, skip. Não retry. |
| Erro ao publicar AnalyticsTracked | Retorna erro → consumer retry. |
| Já trackeado (idempotência) | HasTracked retorna true → skip. |

## Referências

- `.docs/EVENT_SCHEMAS.md` — schema MatchResultsCalculated e AnalyticsTracked
- `pkg/domain/pairing/usecases/analytics_tracked_handler.go` — handler
