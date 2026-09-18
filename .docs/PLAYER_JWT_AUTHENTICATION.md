# Player JWT Authentication (Sync)

Refs **2508-004**. Validates player JWTs on match-making-api **HTTP** entry points (lobbies, game modes, invitations, etc.). Kafka consumer auth remains **2501-004** (separate).

## When it applies

| Case | Behavior |
|------|----------|
| `Authorization: Bearer <jwt>` present | Validate HMAC JWT; on success set tenant/client/user in context; on failure **401/403** JSON |
| Header absent | Continue; `ResourceContextMiddleware` may authenticate via `X-Resource-Owner-ID` (RID) |
| Health `/health` | Same rules (no Bearer = no JWT check) |

## Environment

```bash
JWT_HMAC_SECRET=<shared-hmac-secret>   # required to accept Bearer tokens
JWT_ISSUER=replay-api                  # optional; enforced when set
JWT_AUDIENCE=match-making-api          # optional; enforced when set
```

If a Bearer token is sent but `JWT_HMAC_SECRET` is empty, the API returns **401** `auth_misconfigured`.

## Claims (required)

| Claim | Notes |
|-------|--------|
| `user_id` or `sub` | UUID of the player |
| `tenant_id` | UUID |
| `client_id` | UUID |
| `group_id` | optional UUID |
| `roles` | optional string array (e.g. `["player"]`) |
| `exp` / `iat` | standard JWT time claims |

Signing: **HS256** / HS384 / HS512 only.

## Error payloads

```json
{"error":"token_expired","message":"JWT expired"}
{"error":"token_invalid","message":"JWT validation failed"}
{"error":"token_claims_invalid","message":"token missing tenant_id"}
```

| HTTP | Codes |
|------|--------|
| 401 | `token_expired`, `token_invalid`, `token_missing`, `auth_misconfigured`, malformed `Authorization` |
| 403 | `token_claims_invalid` (missing/invalid ownership claims) |

## User profile

Player profile data (nickname, avatar) is **not** embedded in the JWT. When handlers need profile fields, call replay-api **PlayerProfile** gRPC (`pkg/infra/squad`) using `user_id` from context. Match-making does not issue tokens.

## Code

- Validator: `pkg/domain/iam/jwt`
- Middleware: `cmd/rest-api/middlewares/jwt_middleware.go`
- Wired in: `cmd/rest-api/routing/router.go` (before RID middleware)
