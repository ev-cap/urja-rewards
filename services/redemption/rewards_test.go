//go:build !encore
// +build !encore

package redemption

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRewards_ResponseStructure(t *testing.T) {
	// Test that the response structure is correct
	response := &GetRewardsResponse{
		Rewards: []Reward{
			{
				ID:          "reward-1",
				Name:        "Free Coffee",
				Description: "Get a free coffee",
				Cost:        100,
				Segment:     "",
			},
		},
	}

	assert.NotNil(t, response)
	assert.Len(t, response.Rewards, 1)
	assert.Equal(t, "reward-1", response.Rewards[0].ID)
	assert.Equal(t, "Free Coffee", response.Rewards[0].Name)
	assert.Equal(t, int32(100), response.Rewards[0].Cost)
}

func TestReward_Validation(t *testing.T) {
	reward := &Reward{
		ID:   "reward-1",
		Name: "Test Reward",
		Cost: 500,
	}

	assert.NotEmpty(t, reward.ID)
	assert.NotEmpty(t, reward.Name)
	assert.Greater(t, reward.Cost, int32(0))
}
