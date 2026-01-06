package common

import (
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

var (
	// Default TenantID for TeamPRO (random and valid UUID)
	TeamPROTenantID = replay_common.TeamPROTenantID

	// Default ClientID for TeamPRO (random and valid UUID)
	TeamPROAppClientID = replay_common.TeamPROAppClientID

	// Default ClientID for the server (random and valid UUID)
	ServerClientID = replay_common.ServerClientID
)

const (
	// semantic aliases
	ALLOW = true
	DENY  = false
)
