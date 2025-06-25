package redemption

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"encore.app/internal/db"
	"github.com/google/uuid"
)

//encore:api public method=POST path=/v1/redeem
func (s *Service) Redeem(ctx context.Context, req *RedeemRequest) (*RedeemResponse, error) {
	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse reward ID
	rewardID, err := uuid.Parse(req.RewardID)
	if err != nil {
		return nil, fmt.Errorf("invalid reward ID: %w", err)
	}

	// Get the reward to check its cost
	reward, err := s.db.GetReward(ctx, rewardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("reward not found")
		}
		return nil, fmt.Errorf("failed to get reward: %w", err)
	}

	if !reward.Active {
		return nil, fmt.Errorf("reward is not active")
	}

	// Get user's current points balance
	balance, err := s.db.GetUserPointsBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	// Check if user has enough points
	if balance < int64(reward.Cost) {
		return nil, fmt.Errorf("insufficient points: need %d, have %d", reward.Cost, balance)
	}

	// Create the redemption
	redemption, err := s.db.CreateRedemption(ctx, db.CreateRedemptionParams{
		UserID:      userID,
		RewardID:    rewardID,
		PointsSpent: reward.Cost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create redemption: %w", err)
	}

	// Deduct points by creating a negative points event
	_, err = s.db.CreatePointsEvent(ctx, db.CreatePointsEventParams{
		UserID:    userID,
		EventType: "REDEMPTION",
		RefID:     sql.NullString{String: redemption.ID.String(), Valid: true},
		Points:    -reward.Cost, // Negative points to deduct
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deduct points: %w", err)
	}

	// Publish RedemptionCreated event
	_, err = RedemptionCreatedTopic.Publish(ctx, &RedemptionCreated{
		RedemptionID: redemption.ID.String(),
		UserID:       req.UserID,
		RewardID:     req.RewardID,
		PointsSpent:  redemption.PointsSpent,
		Status:       redemption.Status,
	})
	if err != nil {
		// Log error but don't fail the request
		// In production, you might want to handle this differently
	}

	return &RedeemResponse{
		RedemptionID: redemption.ID.String(),
		PointsSpent:  redemption.PointsSpent,
		Status:       redemption.Status,
	}, nil
}
