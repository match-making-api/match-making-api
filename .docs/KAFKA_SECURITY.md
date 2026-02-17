# Kafka Security — Matchmaking Cluster

Configuration guide for **TLS**, **authentication (SASL)**, and **authorization (ACLs)** on the matchmaking Kafka cluster managed by **Strimzi**.

## Bootstrap Endpoints

| Listener | Protocol | Port | Use case |
|----------|----------|------|----------|
| **plain** | `PLAINTEXT` | `9092` | Development only (Docker Compose). No TLS, no auth. |
| **tls** | `SSL` | `9093` | TLS-encrypted, no SASL. Used when mTLS alone is sufficient. |
| **sasl-tls** | `SASL_SSL` | `9094` | TLS + SCRAM-SHA-512 auth. **Recommended for production.** |

### Strimzi Bootstrap Addresses

```
# Internal (same namespace)
matchmaking-kafka-bootstrap:9092   # plain
matchmaking-kafka-bootstrap:9093   # tls
matchmaking-kafka-bootstrap:9094   # sasl-tls

# External (via Ingress/LoadBalancer — configure in Kafka CR listeners)
matchmaking-kafka-0.example.com:9094
```

## Security Protocols

The `KAFKA_SECURITY_PROTOCOL` environment variable controls which protocol the client uses:

| Value | TLS | SASL | Description |
|-------|-----|------|-------------|
| `PLAINTEXT` | No | No | Development/Docker Compose only |
| `SSL` | Yes | No | TLS encryption, optional mTLS client certs |
| `SASL_PLAINTEXT` | No | Yes | SASL over plaintext — **not recommended** |
| `SASL_SSL` | Yes | Yes | SASL + TLS — **recommended for production** |

## Environment Variables

```bash
# Protocol & SASL
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_MECHANISM=SCRAM-SHA-512
KAFKA_SASL_USERNAME=matchmaking-commands-consumer
KAFKA_SASL_PASSWORD=<from KafkaUser Secret>

# TLS certificates
KAFKA_TLS_CA_CERT_PATH=/etc/kafka/certs/ca.crt
KAFKA_TLS_CERT_PATH=/etc/kafka/certs/user.crt
KAFKA_TLS_KEY_PATH=/etc/kafka/certs/user.key
KAFKA_TLS_SKIP_VERIFY=false
```

## KafkaUser Resources

Strimzi KafkaUser CRDs manage credentials and ACLs. Manifests are in `deploy/strimzi/`.

### matchmaking-commands-consumer

**Used by:** `matchmaking-worker` (PlayerQueuedConsumer)

| Resource | Type | Operations |
|----------|------|------------|
| Topic `matchmaking.commands` | literal | Read, Describe |
| Group `match-making-api-commands` | literal | Read |

```bash
kubectl apply -f deploy/strimzi/kafka-user-matchmaking-commands-consumer.yaml
```

### matchmaking-producer

**Used by:** `match-making-api` (REST API) and `matchmaking-worker` (QueueStatusTicker)

| Resource | Type | Operations |
|----------|------|------------|
| Topics `matchmaking.*` | prefix | Write, Describe, Create |
| Topic `websocket.broadcasts` | literal | Write, Describe |

```bash
kubectl apply -f deploy/strimzi/kafka-user-matchmaking-producer.yaml
```

### matchmaking-match-results-consumer

**Used by:** `match-making-api` (REST API — MatchResultConsumer)

| Resource | Type | Operations |
|----------|------|------------|
| Topic `matchmaking.matches.results` | literal | Read, Describe |
| Group `match-making-group` | literal | Read |

```bash
kubectl apply -f deploy/strimzi/kafka-user-matchmaking-match-results-consumer.yaml
```

## Extracting Credentials from Strimzi Secrets

When a `KafkaUser` is created, Strimzi generates a Secret with the same name containing the password (for SCRAM) or certificates (for mTLS).

### SCRAM password

```bash
# Get password for a KafkaUser
kubectl get secret matchmaking-commands-consumer \
  -o jsonpath='{.data.password}' | base64 -d

# Get SASL/JAAS config string
kubectl get secret matchmaking-commands-consumer \
  -o jsonpath='{.data.sasl\.jaas\.config}' | base64 -d
```

### Cluster CA certificate

```bash
# The cluster CA cert is in a Secret named <cluster>-cluster-ca-cert
kubectl get secret matchmaking-cluster-ca-cert \
  -o jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
```

### Client certificates (if mTLS)

```bash
kubectl get secret matchmaking-commands-consumer \
  -o jsonpath='{.data.user\.crt}' | base64 -d > user.crt

kubectl get secret matchmaking-commands-consumer \
  -o jsonpath='{.data.user\.key}' | base64 -d > user.key
```

## Mounting Certs in Kubernetes Pods

Use Kubernetes volume mounts to make certs available to the application pods:

```yaml
# In Deployment spec
volumes:
  - name: kafka-ca-cert
    secret:
      secretName: matchmaking-cluster-ca-cert
      items:
        - key: ca.crt
          path: ca.crt
  - name: kafka-user-certs
    secret:
      secretName: matchmaking-commands-consumer
      items:
        - key: user.crt
          path: user.crt
        - key: user.key
          path: user.key

containers:
  - name: matchmaking-worker
    volumeMounts:
      - name: kafka-ca-cert
        mountPath: /etc/kafka/certs
        readOnly: true
      - name: kafka-user-certs
        mountPath: /etc/kafka/certs
        readOnly: true
    env:
      - name: KAFKA_SECURITY_PROTOCOL
        value: "SASL_SSL"
      - name: KAFKA_SASL_MECHANISM
        value: "SCRAM-SHA-512"
      - name: KAFKA_SASL_USERNAME
        value: "matchmaking-commands-consumer"
      - name: KAFKA_SASL_PASSWORD
        valueFrom:
          secretKeyRef:
            name: matchmaking-commands-consumer
            key: password
      - name: KAFKA_TLS_CA_CERT_PATH
        value: "/etc/kafka/certs/ca.crt"
```

## Certificate Rotation Runbook

### Strimzi auto-rotation (cluster CA)

Strimzi renews the cluster CA automatically before expiry (default: 365 days, renewal 30 days before). To force rotation:

```bash
# Annotate the cluster CA Secret to trigger renewal
kubectl annotate secret matchmaking-cluster-ca-cert \
  strimzi.io/force-renew=true

# Monitor rollout
kubectl get pods -l strimzi.io/cluster=matchmaking -w
```

After rotation:
1. **Pods using volume mounts**: Kubernetes auto-updates the Secret volume (may take up to 60s).
2. **Pods using env vars**: Restart required to pick up new certificates.
3. **External clients**: Manually extract and distribute the new `ca.crt`.

### KafkaUser credential rotation

To rotate a SCRAM user password:

```bash
# Delete and recreate the KafkaUser to regenerate password
kubectl delete kafkauser matchmaking-commands-consumer
kubectl apply -f deploy/strimzi/kafka-user-matchmaking-commands-consumer.yaml

# Wait for the Secret to be recreated
kubectl get secret matchmaking-commands-consumer -w

# Extract the new password
kubectl get secret matchmaking-commands-consumer \
  -o jsonpath='{.data.password}' | base64 -d
```

For zero-downtime rotation:
1. Create a **new** KafkaUser with a different name (e.g., `matchmaking-commands-consumer-v2`).
2. Update the application deployment to use the new credentials.
3. Verify connectivity.
4. Delete the old KafkaUser.

### mTLS client certificate rotation

Strimzi renews user certificates automatically. To force:

```bash
kubectl annotate secret matchmaking-commands-consumer \
  strimzi.io/force-renew=true
```

## Access Troubleshooting

### Common errors

| Error | Cause | Fix |
|-------|-------|-----|
| `TOPIC_AUTHORIZATION_FAILED` | Missing ACL for topic | Add topic ACL to KafkaUser |
| `GROUP_AUTHORIZATION_FAILED` | Missing ACL for consumer group | Add group ACL to KafkaUser; verify `groupID` matches ACL |
| `SASL authentication failed` | Wrong username/password | Re-extract password from KafkaUser Secret |
| `x509: certificate signed by unknown authority` | Missing or wrong CA cert | Mount cluster CA cert; set `KAFKA_TLS_CA_CERT_PATH` |
| `tls: bad certificate` | Client cert rejected by broker | Verify KafkaUser cert is not expired; force-renew |
| `i/o timeout` on `9093`/`9094` | Firewall or wrong bootstrap | Check network policy; verify bootstrap address |

### Diagnostic commands

```bash
# Test TLS connectivity
openssl s_client -connect matchmaking-kafka-bootstrap:9093 \
  -CAfile ca.crt -showcerts

# Test SASL + TLS with kafkacat (kcat)
kcat -b matchmaking-kafka-bootstrap:9094 \
  -X security.protocol=SASL_SSL \
  -X sasl.mechanism=SCRAM-SHA-512 \
  -X sasl.username=matchmaking-commands-consumer \
  -X sasl.password=$(kubectl get secret matchmaking-commands-consumer -o jsonpath='{.data.password}' | base64 -d) \
  -X ssl.ca.location=ca.crt \
  -L

# List ACLs for a user
kubectl exec matchmaking-kafka-0 -- bin/kafka-acls.sh \
  --bootstrap-server localhost:9092 \
  --list --principal User:matchmaking-commands-consumer

# Describe consumer group
kubectl exec matchmaking-kafka-0 -- bin/kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group match-making-api-commands
```

### Verify KafkaUser status

```bash
# Check if KafkaUser is ready
kubectl get kafkauser matchmaking-commands-consumer -o yaml

# Look for conditions
kubectl get kafkauser matchmaking-commands-consumer \
  -o jsonpath='{.status.conditions[*].type}'
# Expected: Ready
```

## Development Setup

For **local development** (Docker Compose), Kafka runs without security:

```bash
KAFKA_BOOTSTRAP_SERVERS=localhost:29092,localhost:39092
KAFKA_SECURITY_PROTOCOL=PLAINTEXT
```

No TLS, no SASL, no ACLs. This is intentional for developer experience.

## References

- [Strimzi KafkaUser Documentation](https://strimzi.io/docs/operators/latest/configuring.html#type-KafkaUser-reference)
- [Strimzi TLS / Authentication](https://strimzi.io/docs/operators/latest/configuring.html#assembly-securing-access-str)
- [Strimzi Certificate Renewal](https://strimzi.io/docs/operators/latest/deploying.html#proc-renewing-ca-certs-manually-str)
- Epic §10 — Consumer Group Patterns
- Migration Phase 1 — Configure security (TLS, authentication)
