//go:build encore
// +build encore

package redemption

import (
	"context"
	"time"

	"encore.app/internal/db"
)

//encore:api cron name=expire_pending schedule="0 0 * * *"
func ExpirePendingRedemptions(ctx context.Context) error {
	// Get database connection
	queries := db.New(nil) // Encore injects DB

	// Calculate the cutoff time (24 hours ago)
	cutoffTime := time.Now().Add(-24 * time.Hour)

	// Get all pending redemptions older than 24 hours
	oldRedemptions, err := queries.GetPendingRedemptionsOlderThan(ctx, cutoffTime)
	if err != nil {
		return err
	}

	// Update each redemption to EXPIRED status
	for _, redemption := range oldRedemptions {
		_, err := queries.UpdateRedemptionStatus(ctx, db.UpdateRedemptionStatusParams{
			ID:     redemption.ID,
			Status: "EXPIRED",
		})
		if err != nil {
			// Log error but continue processing other redemptions
			// In production, you might want to handle this differently
			continue
		}
	}

	return nil
}
