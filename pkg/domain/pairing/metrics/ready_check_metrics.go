package metrics

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// ReadyCheckMetrics provides operational visibility into the readiness confirmation flow.
// It tracks counters and latencies using atomic operations for thread safety.
//
// When prometheus/client_golang is added to go.mod, these counters can be
// replaced with promauto.NewCounterVec / promauto.NewHistogramVec while
// preserving the same public API.
type ReadyCheckMetrics struct {
	// Counters
	readyChecksStarted  atomic.Int64
	readinessConfirmed  atomic.Int64
	readinessDeclined   atomic.Int64
	readyChecksTimedOut atomic.Int64
	allPlayersReady     atomic.Int64
	connInfoDelivered   atomic.Int64

	// Notification counters by channel
	notifSent   sync.Map // key: "channel:type" -> *atomic.Int64
	notifFailed sync.Map // key: "channel:type" -> *atomic.Int64

	// Latency tracking (in milliseconds)
	confirmLatencySum   atomic.Int64
	confirmLatencyCount atomic.Int64
}

// Global singleton - safe to use from any goroutine.
var Global = &ReadyCheckMetrics{}

// RecordReadyCheckStarted increments the counter for new ready checks.
func (m *ReadyCheckMetrics) RecordReadyCheckStarted(lobbyID string, playerCount int) {
	m.readyChecksStarted.Add(1)
	slog.Info("metrics.ready_check_started",
		"lobby_id", lobbyID,
		"player_count", playerCount,
		"total_started", m.readyChecksStarted.Load())
}

// RecordReadinessConfirmed increments the confirmed counter and records latency.
func (m *ReadyCheckMetrics) RecordReadinessConfirmed(lobbyID, playerID string, latency time.Duration) {
	m.readinessConfirmed.Add(1)
	ms := latency.Milliseconds()
	m.confirmLatencySum.Add(ms)
	m.confirmLatencyCount.Add(1)
	slog.Info("metrics.readiness_confirmed",
		"lobby_id", lobbyID,
		"player_id", playerID,
		"latency_ms", ms,
		"total_confirmed", m.readinessConfirmed.Load())
}

// RecordReadinessDeclined increments the declined counter.
func (m *ReadyCheckMetrics) RecordReadinessDeclined(lobbyID, playerID string) {
	m.readinessDeclined.Add(1)
	slog.Info("metrics.readiness_declined",
		"lobby_id", lobbyID,
		"player_id", playerID,
		"total_declined", m.readinessDeclined.Load())
}

// RecordReadyCheckTimedOut increments the timeout counter.
func (m *ReadyCheckMetrics) RecordReadyCheckTimedOut(lobbyID string, timedOutCount int) {
	m.readyChecksTimedOut.Add(1)
	slog.Warn("metrics.ready_check_timed_out",
		"lobby_id", lobbyID,
		"timed_out_players", timedOutCount,
		"total_timed_out", m.readyChecksTimedOut.Load())
}

// RecordAllPlayersReady increments the all-ready counter.
func (m *ReadyCheckMetrics) RecordAllPlayersReady(lobbyID string) {
	m.allPlayersReady.Add(1)
	slog.Info("metrics.all_players_ready",
		"lobby_id", lobbyID,
		"total_all_ready", m.allPlayersReady.Load())
}

// RecordConnectionInfoDelivered increments the connection-info delivery counter.
func (m *ReadyCheckMetrics) RecordConnectionInfoDelivered(lobbyID, matchID string, playerCount int) {
	m.connInfoDelivered.Add(1)
	slog.Info("metrics.connection_info_delivered",
		"lobby_id", lobbyID,
		"match_id", matchID,
		"player_count", playerCount,
		"total_delivered", m.connInfoDelivered.Load())
}

// RecordNotificationSent records a successful notification delivery.
func (m *ReadyCheckMetrics) RecordNotificationSent(channel, notifType string) {
	counter := getOrCreate(&m.notifSent, notifKey(channel, notifType))
	counter.Add(1)
	slog.Debug("metrics.notification_sent",
		"channel", channel,
		"type", notifType,
		"total", counter.Load())
}

// RecordNotificationFailed records a failed notification delivery.
func (m *ReadyCheckMetrics) RecordNotificationFailed(channel, notifType, reason string) {
	counter := getOrCreate(&m.notifFailed, notifKey(channel, notifType))
	counter.Add(1)
	slog.Warn("metrics.notification_failed",
		"channel", channel,
		"type", notifType,
		"reason", reason,
		"total_failed", counter.Load())
}

func notifKey(channel, notifType string) string { return channel + ":" + notifType }

func getOrCreate(m *sync.Map, key string) *atomic.Int64 {
	if v, ok := m.Load(key); ok {
		return v.(*atomic.Int64)
	}
	v := &atomic.Int64{}
	actual, _ := m.LoadOrStore(key, v)
	return actual.(*atomic.Int64)
}

// Snapshot returns a point-in-time view of all counters.
type Snapshot struct {
	ReadyChecksStarted  int64   `json:"ready_checks_started"`
	ReadinessConfirmed  int64   `json:"readiness_confirmed"`
	ReadinessDeclined   int64   `json:"readiness_declined"`
	ReadyChecksTimedOut int64   `json:"ready_checks_timed_out"`
	AllPlayersReady     int64   `json:"all_players_ready"`
	ConnInfoDelivered   int64   `json:"connection_info_delivered"`
	AvgConfirmLatencyMs float64 `json:"avg_confirm_latency_ms"`
}

// Snapshot captures the current counter values.
func (m *ReadyCheckMetrics) Snapshot() Snapshot {
	count := m.confirmLatencyCount.Load()
	var avgMs float64
	if count > 0 {
		avgMs = float64(m.confirmLatencySum.Load()) / float64(count)
	}

	return Snapshot{
		ReadyChecksStarted:  m.readyChecksStarted.Load(),
		ReadinessConfirmed:  m.readinessConfirmed.Load(),
		ReadinessDeclined:   m.readinessDeclined.Load(),
		ReadyChecksTimedOut: m.readyChecksTimedOut.Load(),
		AllPlayersReady:     m.allPlayersReady.Load(),
		ConnInfoDelivered:   m.connInfoDelivered.Load(),
		AvgConfirmLatencyMs: avgMs,
	}
}
