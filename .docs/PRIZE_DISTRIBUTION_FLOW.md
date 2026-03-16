# Prize Distribution Flow — MatchResultsCalculated → PrizeDistributed

Fluxo completo de MatchResultsCalculated até PrizeDistributed: consumo, identificação de vencedores, resolução de prêmios e publicação. A match-making-api **não executa transferências**; a Wallet API consome o evento e executa as transferências (#31).

## Fluxo

1. **MatchResultsCalculated** (produtor: match-making-api via MatchCompleted): emitido após persistir resultados da partida.
2. **Tópico**: `matchmaking.matches.results`
3. **Consumer**: `consumer-prize-distribution` consome, valida resource ownership e payload.
4. **Handler**: identifica vencedores (1v1: winner_team_id em player_ids), resolve valores via PrizeAmountResolver, persiste idempotência em MongoDB, publica **PrizeDistributed**.
5. **Tópico de saída**: `matchmaking.prizes.distributed` — consumido pela Wallet API para executar transferências.

## Payload MatchResultsCalculated (Proto) — campos relevantes

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `match_id` | string | Sim | Identificador da partida |
| `player_ids` | string[] | Sim | IDs dos jogadores na partida |
| `winner_team_id` | string | Não | ID do time vencedor; vazio se empate. Para 1v1, pode ser player_id |
| `is_draw` | bool | Sim | true se empate |
| `tenant_id` | string | Sim | Tenant para multi-tenancy |
| `client_id` | string | Sim | Cliente para acesso controlado |
| `resource_owner_id` | string | Sim | RID para autorização |
| `lobby_id` | string | Não | Referência ao lobby (para prize pool) |
| `prize_pool_id` | string | Não | Referência ao prize pool |

## Payload PrizeDistributed (Proto)

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `match_id` | string | Sim | Identificador da partida |
| `winner_details` | WinnerPrizeDetail[] | Sim | player_id, amount_cents, currency por vencedor |
| `distributed_at_epoch_ms` | int64 | Sim | Timestamp de publicação (audit) |
| `tenant_id` | string | Sim | Tenant para multi-tenancy |
| `client_id` | string | Sim | Cliente para acesso controlado |
| `resource_owner_id` | string | Sim | Para autorização; Wallet API valida antes da transferência |
| `lobby_id` | string | Não | Referência ao lobby |
| `prize_pool_id` | string | Não | Referência ao prize pool |
| `currency` | string | Sim | Ex.: "USD" — quando valores são uniformes |

### WinnerPrizeDetail

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `player_id` | string | ID do jogador vencedor |
| `amount_cents` | int64 | Valor em menor unidade (ex.: centavos) |
| `currency` | string | Moeda; pode sobrescrever a do payload |

## Regras de validação

| Regra | Descrição |
|-------|-----------|
| ResourceOwnershipRule | `resource_owner_id`, `match_id`, `tenant_id`, `client_id` obrigatórios |
| Idempotência | Se prêmios já distribuídos para `match_id`, não republica |
| Sem vencedor | draw ou winner_team_id não mapeável → skip |
| Sem prêmios | PrizeAmountResolver retorna nil/vazio → não publica |

## PrizeAmountResolver

Interface para obter valores de prêmio por partida. Implementações podem consultar lobby/prize pool, replay-api ou config. **Implementação atual**: `NoOpPrizeAmountResolver` — retorna nil (nenhum prêmio); substituir por implementação real quando lobby/prize pool estiver disponível.

## Persistência (MongoDB)

- **prizes_distributed_matches**: Idempotência por `match_id` — evita duplicatas

## Modos de falha

| Falha | Tratamento |
|-------|------------|
| Validação de resource ownership | Log, skip. Mensagem não commitada. |
| Payload inválido | Log, skip. Não retry. |
| Erro ao publicar PrizeDistributed | Retorna erro → consumer retry. |
| Já distribuído (idempotência) | HasDistributed retorna true → skip. |
| Resolver retorna nil | Skip; não publica evento. |

## Referências

- `.docs/EVENT_SCHEMAS.md` — schema MatchResultsCalculated e PrizeDistributed
- `pkg/infra/kafka/match_results_calculated_consumer.go` — consumer base
- `pkg/domain/pairing/usecases/prize_distribution_handler.go` — handler
- `pkg/domain/pairing/ports/out/prize_amount_resolver.go` — interface
