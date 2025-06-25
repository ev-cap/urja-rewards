//go:build !encore
// +build !encore

package redemption

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedeemRequest_Validation(t *testing.T) {
	req := &RedeemRequest{
		UserID:   "550e8400-e29b-41d4-a716-446655440000", // Valid UUID
		RewardID: "660e8400-e29b-41d4-a716-446655440000", // Valid UUID
	}

	assert.NotEmpty(t, req.UserID)
	assert.NotEmpty(t, req.RewardID)
}

func TestRedeemResponse_Structure(t *testing.T) {
	response := &RedeemResponse{
		RedemptionID: "redemption-1",
		PointsSpent:  500,
		Status:       "PENDING",
	}

	assert.NotEmpty(t, response.RedemptionID)
	assert.Greater(t, response.PointsSpent, int32(0))
	assert.NotEmpty(t, response.Status)
}

func TestRedemptionCreated_EventStructure(t *testing.T) {
	event := &RedemptionCreated{
		RedemptionID: "redemption-1",
		UserID:       "user-1",
		RewardID:     "reward-1",
		PointsSpent:  500,
		Status:       "PENDING",
	}

	assert.NotEmpty(t, event.RedemptionID)
	assert.NotEmpty(t, event.UserID)
	assert.NotEmpty(t, event.RewardID)
	assert.Greater(t, event.PointsSpent, int32(0))
	assert.NotEmpty(t, event.Status)
}
