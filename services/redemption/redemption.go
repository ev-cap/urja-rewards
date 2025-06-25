//go:build encore
// +build encore

package redemption

import (
	"encore.app/internal/db"
	"encore.dev/pubsub"
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

// RedemptionCreatedTopic is the pub/sub topic for redemption creation events
var RedemptionCreatedTopic = pubsub.NewTopic[*RedemptionCreated]("redemption-created", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// init initializes the redemption service
func init() {
	// Service will be initialized by Encore
}
