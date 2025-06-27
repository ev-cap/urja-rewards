//go:build encore
// +build encore

package notifications

import (
	"context"
	"log"

	"encore.app/services/accrual"
	"encore.app/services/redemption"
	"encore.dev/pubsub"
)

//encore:service
type Service struct {
	// Service will be initialized by Encore
}

// init initializes the notifications service
func init() {
	// Service will be initialized by Encore
}

// HandleUserPointsUpdated processes UserPointsUpdated events
//
//encore:api private
func HandleUserPointsUpdated(ctx context.Context, event *accrual.UserPointsUpdated) error {
	log.Printf("📊 UserPointsUpdated: User %s earned %d points for %s (Event: %s)",
		event.UserID, event.Points, event.EventType, event.EventID)

	// TODO: Send FCM notification to user's device
	// This will be implemented in the next task

	return nil
}

// HandleRedemptionCreated processes RedemptionCreated events
//
//encore:api private
func HandleRedemptionCreated(ctx context.Context, event *redemption.RedemptionCreated) error {
	log.Printf("🎁 RedemptionCreated: User %s redeemed reward %s for %d points (Status: %s)",
		event.UserID, event.RewardID, event.PointsSpent, event.Status)

	// TODO: Send FCM notification to user's device
	// This will be implemented in the next task

	return nil
}

// Subscribe to UserPointsUpdated events
var _ = pubsub.NewSubscription(
	accrual.UserPointsUpdatedTopic,
	"notifications-user-points-updated",
	pubsub.SubscriptionConfig[*accrual.UserPointsUpdated]{
		Handler: HandleUserPointsUpdated,
	},
)

// Subscribe to RedemptionCreated events
var _ = pubsub.NewSubscription(
	redemption.RedemptionCreatedTopic,
	"notifications-redemption-created",
	pubsub.SubscriptionConfig[*redemption.RedemptionCreated]{
		Handler: HandleRedemptionCreated,
	},
)
