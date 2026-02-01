#!/bin/bash
# Create matchmaking topics for local development (Docker Compose)
# Story 2501-001: Kafka Infrastructure Setup
#
# Usage: Run after docker-compose up (e.g. docker-compose up -d && sleep 10 && ./create-topics-local.sh)
# Or: docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --create ...
#
BOOTSTRAP="${KAFKA_BOOTSTRAP_SERVERS:-localhost:29092}"
REPLICATION=1
PARTITIONS=12

topics=("matchmaking.commands:604800000" "matchmaking.events:2592000000" "matchmaking.matches:7776000000")

for t in "${topics[@]}"; do
  name="${t%%:*}"
  retention="${t##*:}"
  echo "Creating topic: $name (retention=${retention}ms)"
  docker run --rm --network host \
    confluentinc/cp-kafka:latest \
    kafka-topics --bootstrap-server "$BOOTSTRAP" \
    --create --topic "$name" \
    --partitions "$PARTITIONS" --replication-factor "$REPLICATION" \
    --config retention.ms="$retention" \
    --if-not-exists 2>/dev/null || true
done

echo "Done. Verify: kafka-topics --bootstrap-server $BOOTSTRAP --list"
