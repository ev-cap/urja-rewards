//go:build !encore
// +build !encore

package notifications

import (
	"context"
	"testing"

	"encore.app/services/accrual"
	"encore.app/services/redemption"
)

func TestHandleUserPointsUpdated(t *testing.T) {
	event := &accrual.UserPointsUpdated{
		UserID:    "user123",
		EventID:   "event456",
		Points:    70,
		EventType: "CHARGE_KWH",
		SessionID: "session789",
	}

	err := HandleUserPointsUpdated(context.Background(), event)
	if err != nil {
		t.Errorf("HandleUserPointsUpdated failed: %v", err)
	}
}

func TestHandleRedemptionCreated(t *testing.T) {
	event := &redemption.RedemptionCreated{
		RedemptionID: "redemption123",
		UserID:       "user456",
		RewardID:     "reward789",
		PointsSpent:  500,
		Status:       "PENDING",
	}

	err := HandleRedemptionCreated(context.Background(), event)
	if err != nil {
		t.Errorf("HandleRedemptionCreated failed: %v", err)
	}
}
