# Event Schemas - Matchmaking

Definição de schemas canônicos para eventos de matchmaking trocados via Kafka entre **replay-api** e **match-making-api**. Os schemas garantem acordo na estrutura de payload, tipos de campo e metadados de resource ownership.

## Localização dos Schemas

- **Arquivos .proto**: `pkg/infra/events/schemas/matchmaking_events.proto`
- **Código Go gerado**: `pkg/infra/events/schemas/matchmaking_events.pb.go`
- **Documentação**: `.docs/EVENT_SCHEMAS.md` (este arquivo)

## Estrutura dos Eventos

Todos os eventos usam um envelope comum (`EventEnvelope`) mais um payload tipado:

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `event_id` | string | Sim | UUID único do evento |
| `event_type` | string | Sim | Tipo do evento (ex: `PlayerQueued`, `MatchCreated`) |
| `aggregate_id` | string | Sim | ID do agregado (ex: player_id, match_id) |
| `timestamp` | Timestamp | Sim | Momento da ocorrência |
| `version` | int32 | Sim | Versão do schema (evolução) |
| `source` | string | Sim | Serviço produtor (ex: `replay-api`, `match-making-api`) |
| `resource_owner_id` | string | Sim | RID ou identificador do dono para multi-tenancy/autorização |
| `correlation_id` | string | Não | Para tracing distribuído |

## Eventos Definidos

### PlayerQueued (replay-api → match-making-api)

Emitido quando um jogador entra na fila.

**Payload:**
- `player_id`, `game_id`, `region` (obrigatórios)
- `skill_range` (opcional): min/max MMR
- `priority_boost` (opcional): prioridade na fila
- `tenant_id`, `client_id` (obrigatórios)
- `resource_permissions` (opcional): ex. `["read", "write"]`

### MatchCreated (match-making-api → replay-api)

Emitido quando uma partida é criada.

**Payload:**
- `match_id`, `lobby_id`, `tenant_id`, `client_id` (obrigatórios)
- `players[]`: `player_id`, `party_id`, `resource_permissions`
- `game_server`: `server_id`, `region`, `resource_owner_id`

### MatchCompleted (opcional)

Placeholder para uso em produtores (2501-003). Payload inclui `match_id`, `player_ids`, `winner_team_id`, `is_draw`, etc.

### RatingsUpdated (opcional)

Placeholder para o epic. Payload inclui `deltas[]` de MMR por jogador.

## Estratégia de Versionamento

### Campo `version` no envelope

- O campo `version` (int32) no envelope indica a versão do schema do payload.
- Versão inicial: `1`.
- Consumidores devem suportar múltiplas versões quando possível e registrar incompatibilidades.

### Regras de evolução (Protobuf)

1. **Não remover campos**: use `reserved` se precisar descontinuar um campo.
2. **Novos campos**: sempre adicione com número de campo novo; campos opcionais ou com valor default são seguros.
3. **Novos tipos de payload**: adicione ao `oneof` em `MatchmakingEvent` com novo número.
4. **Renomear**: Protobuf usa números de campo; renomear o nome do campo é seguro (JSON mantém compatibilidade via `json_name`).

### Onde os schemas são armazenados

- **Repositório**: schemas ficam em `pkg/infra/events/schemas/` no repo.
- **Schema Registry** (se adotado): documentar mapeamento topic → schema e processo de registro de novas versões no Confluence ou README do time.

## Mapeamento Topic → Schema

| Topic | Eventos | Schema |
|-------|---------|--------|
| `matchmaking.queue.events` | PlayerQueued | `PlayerQueuedPayload` |
| `matchmaking.matches.created` | MatchCreated | `MatchCreatedPayload` |
| `matchmaking.matches.results` | MatchCompleted | `MatchCompletedPayload` |
| (a definir) | RatingsUpdated | `RatingsUpdatedPayload` |

## Como adicionar novos eventos

1. Adicione a mensagem de payload em `matchmaking_events.proto`.
2. Adicione o payload ao `oneof` em `MatchmakingEvent`.
3. Execute `make proto-gen` para regenerar o código Go.
4. Atualize este documento com o novo evento e o mapeamento de topic.
5. Se usar Schema Registry, registre a nova versão do schema.

## Como evoluir eventos existentes

1. Adicione novos campos com números de campo não usados.
2. Para campos opcionais, use `optional` em proto3.
3. Incremente o `version` no envelope quando houver mudança incompatível (ex.: remoção de campo).
4. Documente as versões suportadas e o plano de depreciação.
5. Execute `make proto-gen` e ajuste consumidores/produtores.

## Compilação

```bash
make proto-gen
```

Requer `protoc` e `protoc-gen-go` instalados. O target `install-tools` instala o plugin Go.

## Referências

- Epic §9 — Event Payloads with Resource Ownership
- Epic §10 — Event Schemas, Schema Evolution
- Issue #34 — Distributed tracing (correlation_id)
