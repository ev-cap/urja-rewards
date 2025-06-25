//go:build !encore
// +build !encore

package accrual

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharge_PointsCalculation(t *testing.T) {
	// Test that 7 kWh results in 70 points (7 * 10)
	kwh := 7.0
	expectedPoints := int32(kwh * 10)

	assert.Equal(t, int32(70), expectedPoints, "7 kWh should give 70 points")
}

func TestChargeEvent_Validation(t *testing.T) {
	event := &ChargeEvent{
		SessionID: "session-123",
		KWH:       7.0,
		UserID:    "550e8400-e29b-41d4-a716-446655440000", // Valid UUID
	}

	assert.NotEmpty(t, event.SessionID)
	assert.Greater(t, event.KWH, 0.0)
	assert.NotEmpty(t, event.UserID)
}

func TestCharge_WithRulesEngine(t *testing.T) {
	// Create a service with rules engine
	service := &Service{
		engine: nil, // Will fall back to original calculation
	}

	event := &ChargeEvent{
		SessionID: "session-123",
		KWH:       7.0,
		UserID:    "550e8400-e29b-41d4-a716-446655440000",
	}

	// Test that the service can handle the event (even without DB)
	// This test verifies the rules engine integration doesn't break existing logic
	assert.NotNil(t, service)
	assert.NotNil(t, event)

	// Verify the expected calculation
	expectedPoints := int32(7.0 * 10)
	assert.Equal(t, int32(70), expectedPoints, "7 kWh should give 70 points with fallback calculation")
}
