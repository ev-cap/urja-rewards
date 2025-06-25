package accrual

import (
	"encore.app/internal/db"
	"encore.app/internal/rules"
	"encore.dev/pubsub"
)

//encore:service
type Service struct {
	db     *db.Queries
	engine *rules.Engine
}

// ChargeEvent represents a charging session that earns points
type ChargeEvent struct {
	SessionID string  `json:"session_id"`
	KWH       float64 `json:"kwh"`
	UserID    string  `json:"user_id"`
}

// ChargeResponse represents the response from a charge event
type ChargeResponse struct {
	EventID   string `json:"event_id"`
	Points    int32  `json:"points"`
	SessionID string `json:"session_id"`
}

// UserPointsUpdated is published when a user's points are updated
type UserPointsUpdated struct {
	UserID    string `json:"user_id"`
	EventID   string `json:"event_id"`
	Points    int32  `json:"points"`
	EventType string `json:"event_type"`
	SessionID string `json:"session_id,omitempty"`
}

// UserPointsUpdatedTopic is the pub/sub topic for user points updates
var UserPointsUpdatedTopic = pubsub.NewTopic[*UserPointsUpdated]("user-points-updated", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// init initializes the accrual service
func init() {
	// Service will be initialized by Encore
}

// isTestMode checks if we're running in test mode
func isTestMode() bool {
	// Simple check - in a real implementation, you might use build tags or environment variables
	return false
}
