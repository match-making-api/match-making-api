# build
FROM golang:1.25.4 AS build
WORKDIR /app
COPY . .

# Copy .env file explicitly (it's in .gitignore)
COPY .env .env

RUN CGO_ENABLED=0 go build -v -o match-making-api-http-service ./cmd/rest-api/main.go
RUN CGO_ENABLED=0 go build -v -o matchmaking-worker ./cmd/workers/matchmaking/main.go
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

# runtime — Matchmaking Worker (consumer + ticker)
FROM scratch AS matchmaking-worker
COPY --from=build /app/matchmaking-worker ./app/
COPY --from=build /app/.env ./.env

ENV GODEBUG=stackguard=99999000000000

CMD ["./app/matchmaking-worker"]
