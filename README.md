## Architecture

### Services

This repository produces **three** independent binaries:

| Binary | Type | Path | Description |
|---|---|---|---|
| `match-making-api` | REST API | `cmd/rest-api/` | HTTP REST API (port 4991) |
| `consumer-matchmaking-commands` | Consumer | `cmd/consumers/matchmaking-commands/` | Kafka consumer for `PlayerQueued` events |
| `worker-queue-status` | Worker | `cmd/workers/queue-status/` | Periodic ticker for queue position updates |

The **REST API** handles HTTP requests (games, lobbies, invitations, etc.).

The **Consumer** (`consumer-matchmaking-commands`) consumes `PlayerQueued` events from the `matchmaking.commands` Kafka topic and adds players to the matchmaking pool via Dragonfly.

The **Worker** (`worker-queue-status`) runs a periodic ticker (every 5s) that reads active queue entries from Dragonfly, refreshes positions from pool state, and publishes `QueueStatusUpdated` events to `websocket.broadcasts`.

All three share the same codebase (`pkg/`) but use different DI injection:
- API: `infra.Inject` + `domain.Inject` (full stack: MongoDB, Kafka, Redis, IAM, billing, etc.)
- Consumer / Worker: `infra.InjectWorker` + `domain.InjectWorker` (slim: MongoDB, Kafka, Redis, game, pairing, schedules)

### Infrastructure

| Service | Purpose | Default Port |
|---|---|---|
| MongoDB | Persistent storage (games, regions, pools) | 37019 |
| Kafka | Event streaming (commands, events, broadcasts) | 29092 |
| Dragonfly | Distributed queue state (`ActiveQueueStore`) | 6379 |

### Quick Start

```bash
# Start all services (API + consumer + worker + infra)
docker-compose -f docker-compose.dev.yml up -d

# Or build locally (binaries output to bin/)
make build-all
make start-rest-api               # terminal 1
make start-consumer-matchmaking   # terminal 2
make start-worker-queue-status    # terminal 3
```

### Environment Variables

See `.env.example` for all available settings. Key additions:

```env
REDIS_ADDR=localhost:6379    # Dragonfly/Redis address
REDIS_PASSWORD=              # Optional auth
```

---

Tenancy Scheme:
```mermaid
graph TD
    Tenant[Tenant]
    Client[Client]
    Group[Group]
    User[User]
    
    Tenant --> Client
    Tenant --> Group
    Client --> User
    Group --> User
```

## Domain-Driven Design for a Team-vs-Team Matchmaking System

### Bounded Contexts

**1. Matchmaking**
* **Core Responsibilities:**
   - Managing player and team queues.
   - Applying matchmaking algorithms to pair suitable teams.
   - Scheduling matches.
   - Handling matchmaking preferences and constraints.
* **Upstream:**
   - Game Server: Receives match details and player information.
   - Player Profile Service: Fetches player data (skill rating, preferences, etc.).
* **Downstream:**
   - Game Server: Sends match details to initiate the game.
   - Player Profile Service: Updates player statistics and preferences.

**2. Player Profile**
* **Core Responsibilities:**
   - Storing and managing player profiles.
   - Calculating and updating player skill ratings.
   - Tracking player preferences and settings.
* **Upstream:**
   - Matchmaking: Fetches player data for matchmaking decisions.
* **Downstream:**
   - Matchmaking: Updates player statistics and preferences after matches.

### Domain Entities and Responsibilities

**Matchmaking Domain:**

* **Lobby:**
   (?) Possible Region(Locale), Place(), 
* **Pool:**
   √ A subset of players or teams within a lobby, often filtered by specific criteria.
* **Pair:**
   √ Represents a pair of teams matched for a game.
   √ Stores information about the matched teams, the scheduled time, and other relevant details.
* **Inquiry:**
   √ A request from a player or team to be matched.
   (?) Includes information about the player's or team's preferences and constraints.
* **Commitment:**
   √ A player or team's commitment to a specific match.
   √ Involves accepting a match offer or declining it.
* **Notification:**
   √ A message sent to players or teams about match-related information.
* **Party:**
   √ A group of players who queue together as a team.
* **Peer:**
   √ An individual player within a party.
* **Schedule:**
   X A plan for a sequence of matches.
   X Includes information about match times, durations, and locations.
* **Appointment:**
   X A specific match scheduled for a particular time and duration.
* **Availability:**
   √ The times when a player or team is available to play.
* **Constraint:**
   √ A limitation or restriction on a player's or team's availability or preferences.
* **Invitation:**
   √ A manual invitation for a user to join a match or event.
   √ Supports status tracking (Pending, Accepted, Declined, Expired, Revoked).
* **ExternalInvitation:**
   √ A manual invitation for an external user (not yet on the platform) to join a match or event.
   √ Includes email validation, registration token generation, and automatic platform membership upon acceptance.
* **Notification:**
   √ A notification sent to users via multiple channels (in-app, email, SMS).
   √ Supports status tracking (Pending, Sent, Failed, Retrying), retry policies, and user preferences.
* **NotificationTemplate:**
   √ A reusable template for notifications with support for variables and multiple languages.
* **UserNotificationPreferences:**
   √ User preferences for notifications including enabled/disabled channels, do not disturb times, and type preferences.
