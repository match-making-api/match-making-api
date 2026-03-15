package templates

import (
	"fmt"
	"strings"

	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
)

// Language codes
const (
	LangEN = "en"
	LangPT = "pt"
	LangES = "es"
)

// DefaultLanguage is the fallback locale when the requested one isn't available.
var DefaultLanguage = LangEN

// SupportedLanguages lists all supported locales for notification content.
var SupportedLanguages = []string{LangEN, LangPT, LangES}

// NotificationUrgency controls the visual urgency level on the frontend.
type NotificationUrgency string

const (
	UrgencyLow      NotificationUrgency = "low"
	UrgencyMedium   NotificationUrgency = "medium"
	UrgencyHigh     NotificationUrgency = "high"
	UrgencyCritical NotificationUrgency = "critical"
)

// NotificationTemplate holds a localised title + message with placeholder support.
// Placeholders use {key} syntax and are resolved via RenderParams().
//
// Extended fields drive richer frontend rendering:
//   - Emoji:        displayed alongside the notification title
//   - ShortMessage: compact one-liner for toasts / push notifications
//   - Urgency:      controls visual intensity (glow, animation, sound)
//   - ActionURL:    deeplink for the notification's primary action
//   - Icon:         Solar icon name for the frontend (e.g., "solar:shield-check-bold")
type NotificationTemplate struct {
	Title        string
	Message      string
	Emoji        string              // e.g., "⚔️", "✅", "🎮"
	ShortMessage string              // concise toast/push text with placeholders
	Urgency      NotificationUrgency // low | medium | high | critical
	ActionURL    string              // e.g., "/lobby/{lobby_id}"
	Icon         string              // Solar icon name for frontend rendering
}

// Render replaces {key} placeholders in title, message, short message, and action URL
// with the provided params.
func (t NotificationTemplate) Render(params map[string]string) NotificationTemplate {
	title := t.Title
	message := t.Message
	shortMessage := t.ShortMessage
	actionURL := t.ActionURL
	for k, v := range params {
		placeholder := "{" + k + "}"
		title = strings.ReplaceAll(title, placeholder, v)
		message = strings.ReplaceAll(message, placeholder, v)
		shortMessage = strings.ReplaceAll(shortMessage, placeholder, v)
		actionURL = strings.ReplaceAll(actionURL, placeholder, v)
	}
	return NotificationTemplate{
		Title:        title,
		Message:      message,
		Emoji:        t.Emoji,
		ShortMessage: shortMessage,
		Urgency:      t.Urgency,
		ActionURL:    actionURL,
		Icon:         t.Icon,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Registry – keyed by (NotificationType, language)
// ──────────────────────────────────────────────────────────────────────────────

type templateKey struct {
	Type     pairing_entities.NotificationType
	Language string
}

var registry = map[templateKey]NotificationTemplate{
	// ── ReadyCheck (type 6) ──────────────────────────────────────────────────
	{pairing_entities.NotificationTypeReadyCheck, LangEN}: {
		Title:        "Match Found!",
		Message:      "Your {game_name} match is ready. Confirm in {timeout}s!",
		Emoji:        "⚔️",
		ShortMessage: "Match found! Confirm in {timeout}s",
		Urgency:      UrgencyCritical,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:shield-check-bold",
	},
	{pairing_entities.NotificationTypeReadyCheck, LangPT}: {
		Title:        "Partida Encontrada!",
		Message:      "Sua partida de {game_name} está pronta. Confirme em {timeout}s!",
		Emoji:        "⚔️",
		ShortMessage: "Partida encontrada! Confirme em {timeout}s",
		Urgency:      UrgencyCritical,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:shield-check-bold",
	},
	{pairing_entities.NotificationTypeReadyCheck, LangES}: {
		Title:        "¡Partida Encontrada!",
		Message:      "Tu partida de {game_name} está lista. ¡Confirma en {timeout}s!",
		Emoji:        "⚔️",
		ShortMessage: "¡Partida encontrada! Confirma en {timeout}s",
		Urgency:      UrgencyCritical,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:shield-check-bold",
	},

	// ── ReadinessConfirmed (type 7) ──────────────────────────────────────────
	{pairing_entities.NotificationTypeReadinessConfirmed, LangEN}: {
		Title:        "Player Ready",
		Message:      "{player_name} confirmed readiness ({confirmed}/{total}).",
		Emoji:        "✅",
		ShortMessage: "{player_name} ready ({confirmed}/{total})",
		Urgency:      UrgencyMedium,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:check-circle-bold",
	},
	{pairing_entities.NotificationTypeReadinessConfirmed, LangPT}: {
		Title:        "Jogador Pronto",
		Message:      "{player_name} confirmou prontidão ({confirmed}/{total}).",
		Emoji:        "✅",
		ShortMessage: "{player_name} pronto ({confirmed}/{total})",
		Urgency:      UrgencyMedium,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:check-circle-bold",
	},
	{pairing_entities.NotificationTypeReadinessConfirmed, LangES}: {
		Title:        "Jugador Listo",
		Message:      "{player_name} confirmó que está listo ({confirmed}/{total}).",
		Emoji:        "✅",
		ShortMessage: "{player_name} listo ({confirmed}/{total})",
		Urgency:      UrgencyMedium,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:check-circle-bold",
	},

	// ── AllPlayersReady (type 8) ─────────────────────────────────────────────
	{pairing_entities.NotificationTypeAllPlayersReady, LangEN}: {
		Title:        "All Players Ready!",
		Message:      "Everyone confirmed. Your {game_name} match is starting now!",
		Emoji:        "🎮",
		ShortMessage: "All ready! {game_name} starting",
		Urgency:      UrgencyHigh,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:users-group-rounded-bold",
	},
	{pairing_entities.NotificationTypeAllPlayersReady, LangPT}: {
		Title:        "Todos Prontos!",
		Message:      "Todos confirmaram. Sua partida de {game_name} está começando agora!",
		Emoji:        "🎮",
		ShortMessage: "Todos prontos! {game_name} começando",
		Urgency:      UrgencyHigh,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:users-group-rounded-bold",
	},
	{pairing_entities.NotificationTypeAllPlayersReady, LangES}: {
		Title:        "¡Todos Listos!",
		Message:      "Todos confirmaron. ¡Tu partida de {game_name} comienza ahora!",
		Emoji:        "🎮",
		ShortMessage: "¡Todos listos! {game_name} comenzando",
		Urgency:      UrgencyHigh,
		ActionURL:    "/lobby/{lobby_id}",
		Icon:         "solar:users-group-rounded-bold",
	},

	// ── ReadyCheckTimeout (type 9) ───────────────────────────────────────────
	{pairing_entities.NotificationTypeReadyCheckTimeout, LangEN}: {
		Title:        "Ready Check Expired",
		Message:      "The ready check for your {game_name} match has timed out. You'll be re-queued automatically.",
		Emoji:        "⏰",
		ShortMessage: "Ready check expired — re-queuing",
		Urgency:      UrgencyMedium,
		ActionURL:    "/matchmaking",
		Icon:         "solar:alarm-bold",
	},
	{pairing_entities.NotificationTypeReadyCheckTimeout, LangPT}: {
		Title:        "Confirmação Expirou",
		Message:      "A confirmação da sua partida de {game_name} expirou. Você será recolocado na fila automaticamente.",
		Emoji:        "⏰",
		ShortMessage: "Confirmação expirou — reenfileirado",
		Urgency:      UrgencyMedium,
		ActionURL:    "/matchmaking",
		Icon:         "solar:alarm-bold",
	},
	{pairing_entities.NotificationTypeReadyCheckTimeout, LangES}: {
		Title:        "Confirmación Expirada",
		Message:      "La confirmación de tu partida de {game_name} ha expirado. Serás reubicado automáticamente.",
		Emoji:        "⏰",
		ShortMessage: "Confirmación expirada — reubicado",
		Urgency:      UrgencyMedium,
		ActionURL:    "/matchmaking",
		Icon:         "solar:alarm-bold",
	},

	// ── GameConnectionInfo (type 10) ─────────────────────────────────────────
	{pairing_entities.NotificationTypeGameConnectionInfo, LangEN}: {
		Title:        "Game Server Ready",
		Message:      "Your {game_name} server is ready! Connect to {server_address} now.",
		Emoji:        "🌐",
		ShortMessage: "Server ready — connect now!",
		Urgency:      UrgencyHigh,
		ActionURL:    "/lobby/{lobby_id}/connect",
		Icon:         "solar:server-bold",
	},
	{pairing_entities.NotificationTypeGameConnectionInfo, LangPT}: {
		Title:        "Servidor Pronto",
		Message:      "Seu servidor de {game_name} está pronto! Conecte-se a {server_address} agora.",
		Emoji:        "🌐",
		ShortMessage: "Servidor pronto — conecte agora!",
		Urgency:      UrgencyHigh,
		ActionURL:    "/lobby/{lobby_id}/connect",
		Icon:         "solar:server-bold",
	},
	{pairing_entities.NotificationTypeGameConnectionInfo, LangES}: {
		Title:        "Servidor Listo",
		Message:      "¡Tu servidor de {game_name} está listo! Conéctate a {server_address} ahora.",
		Emoji:        "🌐",
		ShortMessage: "¡Servidor listo — conéctate ahora!",
		Urgency:      UrgencyHigh,
		ActionURL:    "/lobby/{lobby_id}/connect",
		Icon:         "solar:server-bold",
	},

	// ── ReadinessDeclined (type 11) ──────────────────────────────────────────
	{pairing_entities.NotificationTypeReadinessDeclined, LangEN}: {
		Title:        "Player Declined",
		Message:      "{player_name} declined the {game_name} match. Finding a new match...",
		Emoji:        "❌",
		ShortMessage: "{player_name} declined — re-queuing",
		Urgency:      UrgencyMedium,
		ActionURL:    "/matchmaking",
		Icon:         "solar:close-circle-bold",
	},
	{pairing_entities.NotificationTypeReadinessDeclined, LangPT}: {
		Title:        "Jogador Recusou",
		Message:      "{player_name} recusou a partida de {game_name}. Buscando nova partida...",
		Emoji:        "❌",
		ShortMessage: "{player_name} recusou — reenfileirado",
		Urgency:      UrgencyMedium,
		ActionURL:    "/matchmaking",
		Icon:         "solar:close-circle-bold",
	},
	{pairing_entities.NotificationTypeReadinessDeclined, LangES}: {
		Title:        "Jugador Rechazó",
		Message:      "{player_name} rechazó la partida de {game_name}. Buscando nueva partida...",
		Emoji:        "❌",
		ShortMessage: "{player_name} rechazó — reubicado",
		Urgency:      UrgencyMedium,
		ActionURL:    "/matchmaking",
		Icon:         "solar:close-circle-bold",
	},
}

// ──────────────────────────────────────────────────────────────────────────────
// Lookup functions
// ──────────────────────────────────────────────────────────────────────────────

// Get returns the template for the given type and language.
// Falls back to DefaultLanguage if the requested locale is not available.
func Get(notifType pairing_entities.NotificationType, lang string) NotificationTemplate {
	if lang == "" {
		lang = DefaultLanguage
	}
	lang = normaliseLanguage(lang)

	key := templateKey{Type: notifType, Language: lang}
	if tmpl, ok := registry[key]; ok {
		return tmpl
	}
	// Fallback to English
	key.Language = DefaultLanguage
	if tmpl, ok := registry[key]; ok {
		return tmpl
	}
	// Ultimate fallback
	return NotificationTemplate{
		Title:   "Notification",
		Message: fmt.Sprintf("Notification type %d", notifType),
	}
}

// Render is a convenience function that fetches and renders a template in one call.
func Render(notifType pairing_entities.NotificationType, lang string, params map[string]string) NotificationTemplate {
	return Get(notifType, lang).Render(params)
}

// normaliseLanguage extracts the base language from a locale string (e.g. "pt-BR" → "pt").
func normaliseLanguage(lang string) string {
	parts := strings.SplitN(lang, "-", 2)
	return strings.ToLower(parts[0])
}
