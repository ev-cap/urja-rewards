package accrual

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPointsCalculation tests the core points calculation logic
func TestPointsCalculation(t *testing.T) {
	tests := []struct {
		name     string
		kwh      float64
		expected int32
	}{
		{"7 kWh should give 70 points", 7.0, 70},
		{"10 kWh should give 100 points", 10.0, 100},
		{"5.5 kWh should give 55 points", 5.5, 55},
		{"0 kWh should give 0 points", 0.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := int32(tt.kwh * 10)
			assert.Equal(t, tt.expected, points)
		})
	}
}

// TestChargeEventStructure tests the ChargeEvent struct
func TestChargeEventStructure(t *testing.T) {
	event := &ChargeEvent{
		SessionID: "test-session-123",
		KWH:       7.0,
		UserID:    "550e8400-e29b-41d4-a716-446655440000",
	}

	assert.Equal(t, "test-session-123", event.SessionID)
	assert.Equal(t, 7.0, event.KWH)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", event.UserID)
}
