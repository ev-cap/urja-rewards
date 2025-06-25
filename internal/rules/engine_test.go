package rules

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	// Create a temporary rules file for testing
	tempRules := `rules:
  charge_kwh:
    points_per_kwh: 10
    description: "Points earned per kWh charged"
  referral:
    points: 300
    description: "Points earned for successful referral"
  rating:
    points: 50
    description: "Points earned for leaving a rating"
settings:
  max_points_per_day: 1000
  max_points_per_event: 500
  enable_streak_bonus: true
  enable_first_charge_bonus: true`

	tmpfile, err := os.CreateTemp("", "rules-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(tempRules))
	require.NoError(t, err)
	tmpfile.Close()

	engine, err := NewEngine(tmpfile.Name())
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.config)
}

func TestEvaluateRules_ChargeKWH(t *testing.T) {
	// Create a temporary rules file for testing
	tempRules := `rules:
  charge_kwh:
    points_per_kwh: 10
    description: "Points earned per kWh charged"
settings:
  max_points_per_day: 1000
  max_points_per_event: 500
  enable_streak_bonus: true
  enable_first_charge_bonus: true`

	tmpfile, err := os.CreateTemp("", "rules-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(tempRules))
	require.NoError(t, err)
	tmpfile.Close()

	engine, err := NewEngine(tmpfile.Name())
	require.NoError(t, err)

	// Test 7 kWh should give 70 points (7 * 10)
	payload := &EventPayload{
		EventType: "CHARGE_KWH",
		UserID:    "test-user",
		Data: map[string]interface{}{
			"kwh": 7.0,
		},
	}

	points, err := engine.EvaluateRules(context.Background(), payload)
	require.NoError(t, err)
	assert.Equal(t, 70, points, "7 kWh should give 70 points")

	// Test 5 kWh should give 50 points (5 * 10)
	payload.Data["kwh"] = 5.0
	points, err = engine.EvaluateRules(context.Background(), payload)
	require.NoError(t, err)
	assert.Equal(t, 50, points, "5 kWh should give 50 points")
}

func TestEvaluateRules_Referral(t *testing.T) {
	// Create a temporary rules file for testing
	tempRules := `rules:
  referral:
    points: 300
    description: "Points earned for successful referral"
settings:
  max_points_per_day: 1000
  max_points_per_event: 500
  enable_streak_bonus: true
  enable_first_charge_bonus: true`

	tmpfile, err := os.CreateTemp("", "rules-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(tempRules))
	require.NoError(t, err)
	tmpfile.Close()

	engine, err := NewEngine(tmpfile.Name())
	require.NoError(t, err)

	payload := &EventPayload{
		EventType: "REFERRAL",
		UserID:    "test-user",
		Data:      map[string]interface{}{},
	}

	points, err := engine.EvaluateRules(context.Background(), payload)
	require.NoError(t, err)
	assert.Equal(t, 300, points, "Referral should give 300 points")
}

func TestEvaluateRules_Rating(t *testing.T) {
	// Create a temporary rules file for testing
	tempRules := `rules:
  rating:
    points: 50
    description: "Points earned for leaving a rating"
settings:
  max_points_per_day: 1000
  max_points_per_event: 500
  enable_streak_bonus: true
  enable_first_charge_bonus: true`

	tmpfile, err := os.CreateTemp("", "rules-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(tempRules))
	require.NoError(t, err)
	tmpfile.Close()

	engine, err := NewEngine(tmpfile.Name())
	require.NoError(t, err)

	payload := &EventPayload{
		EventType: "RATING",
		UserID:    "test-user",
		Data:      map[string]interface{}{},
	}

	points, err := engine.EvaluateRules(context.Background(), payload)
	require.NoError(t, err)
	assert.Equal(t, 50, points, "Rating should give 50 points")
}

func TestEvaluateRules_FirstCharge(t *testing.T) {
	// Create a temporary rules file for testing
	tempRules := `rules:
  first_charge:
    points: 100
    description: "Bonus points for first charge session"
settings:
  max_points_per_day: 1000
  max_points_per_event: 500
  enable_streak_bonus: true
  enable_first_charge_bonus: true`

	tmpfile, err := os.CreateTemp("", "rules-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(tempRules))
	require.NoError(t, err)
	tmpfile.Close()

	engine, err := NewEngine(tmpfile.Name())
	require.NoError(t, err)

	payload := &EventPayload{
		EventType: "FIRST_CHARGE",
		UserID:    "test-user",
		Data:      map[string]interface{}{},
	}

	points, err := engine.EvaluateRules(context.Background(), payload)
	require.NoError(t, err)
	assert.Equal(t, 100, points, "First charge should give 100 points")
}

func TestEvaluateRules_DailyLogin(t *testing.T) {
	// Create a temporary rules file for testing
	tempRules := `rules:
  daily_login:
    base_points: 10
    streak_multiplier: 1.5
    max_streak_days: 7
    description: "Points for daily login with streak bonus"
settings:
  max_points_per_day: 1000
  max_points_per_event: 500
  enable_streak_bonus: true
  enable_first_charge_bonus: true`

	tmpfile, err := os.CreateTemp("", "rules-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(tempRules))
	require.NoError(t, err)
	tmpfile.Close()

	engine, err := NewEngine(tmpfile.Name())
	require.NoError(t, err)

	// Test first day login (no streak)
	payload := &EventPayload{
		EventType: "DAILY_LOGIN",
		UserID:    "test-user",
		Data: map[string]interface{}{
			"streak_days": 1,
		},
	}

	points, err := engine.EvaluateRules(context.Background(), payload)
	require.NoError(t, err)
	assert.Equal(t, 10, points, "First day login should give 10 points")

	// Test 3-day streak
	payload.Data["streak_days"] = 3
	points, err = engine.EvaluateRules(context.Background(), payload)
	require.NoError(t, err)
	// 10 * 1.5 * 1.5 = 22.5, rounded to 22
	assert.Equal(t, 22, points, "3-day streak should give 22 points")
}

func TestEvaluateRules_UnknownEventType(t *testing.T) {
	// Create a temporary rules file for testing
	tempRules := `rules:
  charge_kwh:
    points_per_kwh: 10
    description: "Points earned per kWh charged"
settings:
  max_points_per_day: 1000
  max_points_per_event: 500
  enable_streak_bonus: true
  enable_first_charge_bonus: true`

	tmpfile, err := os.CreateTemp("", "rules-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(tempRules))
	require.NoError(t, err)
	tmpfile.Close()

	engine, err := NewEngine(tmpfile.Name())
	require.NoError(t, err)

	payload := &EventPayload{
		EventType: "UNKNOWN_EVENT",
		UserID:    "test-user",
		Data:      map[string]interface{}{},
	}

	points, err := engine.EvaluateRules(context.Background(), payload)
	assert.Error(t, err)
	assert.Equal(t, 0, points)
	assert.Contains(t, err.Error(), "unknown event type")
}
