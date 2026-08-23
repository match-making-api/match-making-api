package common

import (
	"fmt"

	"github.com/google/uuid"
)

// ErrMissingResourceOwnership is returned when required ownership fields are absent.
var ErrMissingResourceOwnership = fmt.Errorf("resource ownership requires tenant_id and client_id")

// ValidateTenantClient requires TenantID and ClientID (matchmaking integration baseline).
func (ro ResourceOwner) ValidateTenantClient() error {
	if ro.TenantID == uuid.Nil || ro.ClientID == uuid.Nil {
		return ErrMissingResourceOwnership
	}
	return nil
}

// ParseUUIDOrNil parses s as UUID; empty or invalid returns uuid.Nil (no error).
func ParseUUIDOrNil(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// ResourceOwnerFromFlat builds ResourceOwner from string IDs used in event payloads / legacy docs.
func ResourceOwnerFromFlat(tenantID, clientID, groupID, userID string) ResourceOwner {
	return ResourceOwner{
		TenantID: ParseUUIDOrNil(tenantID),
		ClientID: ParseUUIDOrNil(clientID),
		GroupID:  ParseUUIDOrNil(groupID),
		UserID:   ParseUUIDOrNil(userID),
	}
}

// FlatStrings returns tenant/client/group/user as strings (empty when Nil).
func (ro ResourceOwner) FlatStrings() (tenantID, clientID, groupID, userID string) {
	if ro.TenantID != uuid.Nil {
		tenantID = ro.TenantID.String()
	}
	if ro.ClientID != uuid.Nil {
		clientID = ro.ClientID.String()
	}
	if ro.GroupID != uuid.Nil {
		groupID = ro.GroupID.String()
	}
	if ro.UserID != uuid.Nil {
		userID = ro.UserID.String()
	}
	return
}
