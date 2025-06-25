package redemption

import (
	"context"
)

//encore:api public method=GET path=/v1/rewards
func (s *Service) GetRewards(ctx context.Context) (*GetRewardsResponse, error) {
	// Get all active rewards from the catalog
	rewards, err := s.db.GetRewardsCatalog(ctx)
	if err != nil {
		return nil, err
	}

	// Convert database models to API response
	response := &GetRewardsResponse{
		Rewards: make([]Reward, len(rewards)),
	}

	for i, reward := range rewards {
		response.Rewards[i] = Reward{
			ID:      reward.ID.String(),
			Name:    reward.Name,
			Cost:    reward.Cost,
			Segment: "", // TODO: Implement segment matching logic
		}

		// Add description if available
		if reward.Description.Valid {
			response.Rewards[i].Description = reward.Description.String
		}
	}

	return response, nil
}
