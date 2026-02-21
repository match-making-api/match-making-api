# build
FROM golang:1.25.4 AS build
WORKDIR /app
COPY . .

# Copy .env file explicitly (it's in .gitignore)
COPY .env .env

RUN CGO_ENABLED=0 go build -v -o match-making-api-http-service ./cmd/rest-api/main.go
RUN CGO_ENABLED=0 go build -v -o consumer-matchmaking-commands ./cmd/consumers/matchmaking-commands/main.go
RUN CGO_ENABLED=0 go build -v -o consumer-server-allocated ./cmd/consumers/server-allocated/main.go
RUN CGO_ENABLED=0 go build -v -o worker-queue-status ./cmd/workers/queue-status/main.go
RUN mkdir -p /app/match_making_files
RUN mkdir -p /app/coverage

# runtime — REST API (default)
FROM scratch AS runtime
COPY --from=build /app/match-making-api-http-service ./app/
COPY --from=build /app/coverage ./app/coverage
COPY --from=build /app/.env ./.env

ENV GODEBUG=stackguard=99999000000000

EXPOSE 4991
CMD ["./app/match-making-api-http-service"]

# consumer — Matchmaking Commands (Kafka consumer for PlayerQueued events)
FROM scratch AS consumer-matchmaking-commands
COPY --from=build /app/consumer-matchmaking-commands ./app/
COPY --from=build /app/.env ./.env

ENV GODEBUG=stackguard=99999000000000

CMD ["./app/consumer-matchmaking-commands"]

# consumer — Server Allocated (Kafka consumer for ServerAllocated events, broadcasts MatchReady)
FROM scratch AS consumer-server-allocated
COPY --from=build /app/consumer-server-allocated ./app/
COPY --from=build /app/.env ./.env

ENV GODEBUG=stackguard=99999000000000

CMD ["./app/consumer-server-allocated"]

# worker — Queue Status (periodic ticker for queue position updates)
FROM scratch AS worker-queue-status
COPY --from=build /app/worker-queue-status ./app/
COPY --from=build /app/.env ./.env

ENV GODEBUG=stackguard=99999000000000

CMD ["./app/worker-queue-status"]
