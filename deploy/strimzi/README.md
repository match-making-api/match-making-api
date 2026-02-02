# Strimzi Kafka — Leet Matchmaking 2026

Story **2501-001: Kafka Infrastructure Setup**. Deploys a Strimzi Kafka cluster with matchmaking topics for event-driven integration between replay-api and match-making-api.

## Prerequisites

- **Kubernetes cluster** with cluster-admin access
- **Strimzi Operator** installed ([deploy guide](https://strimzi.io/docs/operators/latest/full/deploying.html))
- Sufficient **PVC** capacity for broker (100Gi × 3) and ZooKeeper (10Gi × 3) storage

## Namespace

All resources use the **`matchmaking`** namespace:

```bash
kubectl apply -f namespace.yaml
```

## Deployment Order

1. **Namespace**
   ```bash
   kubectl apply -f namespace.yaml
   ```

2. **Kafka cluster** (3 brokers, plain 9092 + TLS 9093)
   ```bash
   kubectl apply -f kafka.yaml
   ```

3. **Topics** (apply after Kafka is ready — status `Ready`)
   ```bash
   kubectl apply -f kafka-topics.yaml
   ```

4. **Verify**
   ```bash
   kubectl get kafka -n matchmaking
   kubectl get kafkatopic -n matchmaking
   ```

## Bootstrap Servers

| Listener | Port | URL (from within cluster) |
|----------|------|---------------------------|
| **plain** | 9092 | `matchmaking-cluster-kafka-bootstrap.matchmaking.svc.cluster.local:9092` |
| **tls**   | 9093 | `matchmaking-cluster-kafka-bootstrap.matchmaking.svc.cluster.local:9093` |

From pods in the **same namespace** (`matchmaking`):
```
matchmaking-cluster-kafka-bootstrap:9092
matchmaking-cluster-kafka-bootstrap:9093
```

From other namespaces:
```
matchmaking-cluster-kafka-bootstrap.matchmaking.svc.cluster.local:9092
matchmaking-cluster-kafka-bootstrap.matchmaking.svc.cluster.local:9093
```

## Topics

| Topic | Partitions | Replicas | Retention |
|-------|------------|----------|-----------|
| `matchmaking.commands` | 12 | 3 | 7 days |
| `matchmaking.events`   | 12 | 3 | 30 days |
| `matchmaking.matches`  | 12 | 3 | 90 days |

## Smoke Test

### Option 1: Kafka client pod (kubectl)

```bash
# Create a temporary Kafka client pod
kubectl run kafka-client -n matchmaking --rm -i --restart=Never --image=quay.io/strimzi/kafka:0.39.0-kafka-3.6.0 -- \
  bin/kafka-console-producer.sh \
  --bootstrap-server matchmaking-cluster-kafka-bootstrap:9092 \
  --topic matchmaking.commands

# Type a message and press Enter, then Ctrl+C to exit
```

In another terminal:

```bash
kubectl run kafka-consumer -n matchmaking --rm -i --restart=Never --image=quay.io/strimzi/kafka:0.39.0-kafka-3.6.0 -- \
  bin/kafka-console-consumer.sh \
  --bootstrap-server matchmaking-cluster-kafka-bootstrap:9092 \
  --topic matchmaking.commands \
  --from-beginning
```

### Option 2: kcat (outside cluster)

If you have a NodePort or LoadBalancer for Kafka (or port-forward):

```bash
kubectl port-forward svc/matchmaking-cluster-kafka-bootstrap 9092:9092 -n matchmaking
```

```bash
# Produce
echo '{"type":"test","payload":"smoke"}' | kcat -P -b localhost:9092 -t matchmaking.commands -k test-key

# Consume
kcat -C -b localhost:9092 -t matchmaking.commands -f 'Key: %k\nValue: %s\n'
```

### Option 3: Docker Compose (local dev)

For local development without Kubernetes:

```bash
cp docker-compose.yml.example docker-compose.yml
docker compose up -d
```

The `topic-init` service creates `matchmaking.commands`, `matchmaking.events`, and `matchmaking.matches` on startup. Bootstrap server: `localhost:29092` (from host) or `kafka:9092` (from containers).

Optional script for manual topic creation: `deploy/strimzi/scripts/create-topics-local.sh`

## Configuration Reference

- **Epic §10** — Strimzi Kafka Integration Design, Topic Design for Scalability
- **Broker config**: `num.partitions`, `default.replication.factor`, `min.insync.replicas`, offsets/transaction replication
- **Retention**: 7d commands, 30d events, 90d matches (per epic Topic Design)
