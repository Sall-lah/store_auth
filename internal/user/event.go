package user

import "time"

const (
	// EventUserBanned is published when an administrator suspends or bans a user account.
	EventUserBanned = "user.banned"
	// EventUserDeleted is published when a user requests account deletion.
	EventUserDeleted = "user.deleted"
)

// LifecycleEvent defines the structured domain event payload dispatched on user lifecycle Kafka topics.
// Why: Standardizes asynchronous communication contract across microservices for account lifecycle events.
type LifecycleEvent struct {
	Event     string    `json:"event"`
	UserID    string    `json:"userId"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Reason    string    `json:"reason,omitempty"`
}
