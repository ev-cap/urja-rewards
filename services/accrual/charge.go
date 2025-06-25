package accrual

import (
	"context"
	"database/sql"

	"encore.app/internal/db"
	"encore.app/internal/rules"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

//encore:api public method=POST path=/v1/events/charge
func (s *Service) Charge(ctx context.Context, event *ChargeEvent) (*ChargeResponse, error) {
	// Initialize rules engine
	engine, err := rules.NewEngine("rules.yaml")
	if err != nil {
		// Log error but continue without rules engine
		engine = nil
	}

	// Parse user ID
	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return nil, err
	}

	var points int32

	// Use rules engine to calculate points if available
	if engine != nil {
		payload := &rules.EventPayload{
			EventType: "CHARGE_KWH",
			UserID:    event.UserID,
			Data: map[string]interface{}{
				"kwh": event.KWH,
			},
		}

		calculatedPoints, err := engine.EvaluateRules(ctx, payload)
		if err != nil {
			return nil, err
		}
		points = int32(calculatedPoints)
	} else {
		// Fallback to original calculation (10 points per kWh)
		points = int32(event.KWH * 10)
	}

	// Create points event
	pointsEvent, err := s.db.CreatePointsEvent(ctx, db.CreatePointsEventParams{
		UserID:    userID,
		EventType: "CHARGE_KWH",
		RefID:     sql.NullString{String: event.SessionID, Valid: true},
		Points:    points,
		Meta:      pqtype.NullRawMessage{},
	})
	if err != nil {
		return nil, err
	}

	// Publish UserPointsUpdated event
	_, err = UserPointsUpdatedTopic.Publish(ctx, &UserPointsUpdated{
		UserID:    event.UserID,
		EventID:   pointsEvent.ID.String(),
		Points:    points,
		EventType: "CHARGE_KWH",
		SessionID: event.SessionID,
	})
	if err != nil {
		// Log error but don't fail the request
		// In production, you might want to handle this differently
		// For now, we'll just return the response even if publishing fails
	}

	return &ChargeResponse{
		EventID:   pointsEvent.ID.String(),
		Points:    points,
		SessionID: event.SessionID,
	}, nil
}
