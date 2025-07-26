package common

import "github.com/google/uuid"

var (
	// Default TenantID for TeamPRO (random and valid UUID)
	TeamPROTenantID = uuid.MustParse("a3a80810-f91c-4391-9eff-6d47a13bebde")

	// Default ClientID for TeamPRO (random and valid UUID)
	TeamPROAppClientID = uuid.MustParse("ff96c01f-a741-4429-a0cd-2868d408c42f")

	// Default ClientID for the server (random and valid UUID)
	ServerClientID = uuid.MustParse("ff96c01f-a741-4429-a0cd-2868d408c42f")
)
