//go:build !encore
// +build !encore

package redemption

import (
	"context"

	"encore.app/internal/db"
)

//encore:service
type Service struct {
	db *db.Queries
}

// Reward represents a reward from the catalog
type Reward struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Cost        int32  `json:"cost"`
	Segment     string `json:"segment,omitempty"`
}

// GetRewardsResponse represents the response for getting rewards
type GetRewardsResponse struct {
	Rewards []Reward `json:"rewards"`
}

// RedeemRequest represents a redemption request
type RedeemRequest struct {
	UserID   string `json:"user_id"`
	RewardID string `json:"reward_id"`
}

// RedeemResponse represents the response from a redemption
type RedeemResponse struct {
	RedemptionID string `json:"redemption_id"`
	PointsSpent  int32  `json:"points_spent"`
	Status       string `json:"status"`
}

// RedemptionCreated is published when a redemption is created
type RedemptionCreated struct {
	RedemptionID string `json:"redemption_id"`
	UserID       string `json:"user_id"`
	RewardID     string `json:"reward_id"`
	PointsSpent  int32  `json:"points_spent"`
	Status       string `json:"status"`
}

// RedemptionCreatedTopic is a mock topic for non-Encore builds
var RedemptionCreatedTopic = &MockTopic{}

// MockTopic is a mock implementation for testing
type MockTopic struct{}

func (m *MockTopic) Publish(ctx context.Context, msg *RedemptionCreated) (string, error) {
	// Mock implementation - does nothing
	return "mock-message-id", nil
}

// init initializes the redemption service
func init() {
	// Service will be initialized by Encore
}
