package rules

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RulesConfig represents the configuration loaded from rules.yaml
type RulesConfig struct {
	Rules    map[string]Rule `yaml:"rules"`
	Settings Settings        `yaml:"settings"`
}

// Rule represents a single rule configuration
type Rule struct {
	PointsPerKWH     int     `yaml:"points_per_kwh,omitempty"`
	Points           int     `yaml:"points,omitempty"`
	BasePoints       int     `yaml:"base_points,omitempty"`
	StreakMultiplier float64 `yaml:"streak_multiplier,omitempty"`
	MaxStreakDays    int     `yaml:"max_streak_days,omitempty"`
	Description      string  `yaml:"description"`
}

// Settings represents global rule evaluation settings
type Settings struct {
	MaxPointsPerDay        int  `yaml:"max_points_per_day"`
	MaxPointsPerEvent      int  `yaml:"max_points_per_event"`
	EnableStreakBonus      bool `yaml:"enable_streak_bonus"`
	EnableFirstChargeBonus bool `yaml:"enable_first_charge_bonus"`
}

// EventPayload represents the data passed to rule evaluation
type EventPayload struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
}

// Engine represents the rules engine
type Engine struct {
	config *RulesConfig
}

// NewEngine creates a new rules engine instance
func NewEngine(configPath string) (*Engine, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules config: %w", err)
	}

	var config RulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse rules config: %w", err)
	}

	return &Engine{config: &config}, nil
}

// EvaluateRules evaluates rules for a given event payload
func (e *Engine) EvaluateRules(ctx context.Context, payload *EventPayload) (int, error) {
	switch payload.EventType {
	case "CHARGE_KWH":
		return e.evaluateChargeKWH(payload)
	case "REFERRAL":
		return e.evaluateReferral(payload)
	case "RATING":
		return e.evaluateRating(payload)
	case "FIRST_CHARGE":
		return e.evaluateFirstCharge(payload)
	case "DAILY_LOGIN":
		return e.evaluateDailyLogin(payload)
	default:
		return 0, fmt.Errorf("unknown event type: %s", payload.EventType)
	}
}

// evaluateChargeKWH calculates points for charging events
func (e *Engine) evaluateChargeKWH(payload *EventPayload) (int, error) {
	rule, exists := e.config.Rules["charge_kwh"]
	if !exists {
		return 0, fmt.Errorf("charge_kwh rule not found")
	}

	kwh, ok := payload.Data["kwh"].(float64)
	if !ok {
		return 0, fmt.Errorf("kwh value not found or invalid in payload")
	}

	points := int(kwh * float64(rule.PointsPerKWH))

	// Apply max points per event limit
	if points > e.config.Settings.MaxPointsPerEvent {
		points = e.config.Settings.MaxPointsPerEvent
	}

	return points, nil
}

// evaluateReferral calculates points for referral events
func (e *Engine) evaluateReferral(payload *EventPayload) (int, error) {
	rule, exists := e.config.Rules["referral"]
	if !exists {
		return 0, fmt.Errorf("referral rule not found")
	}

	return rule.Points, nil
}

// evaluateRating calculates points for rating events
func (e *Engine) evaluateRating(payload *EventPayload) (int, error) {
	rule, exists := e.config.Rules["rating"]
	if !exists {
		return 0, fmt.Errorf("rating rule not found")
	}

	return rule.Points, nil
}

// evaluateFirstCharge calculates points for first charge bonus
func (e *Engine) evaluateFirstCharge(payload *EventPayload) (int, error) {
	if !e.config.Settings.EnableFirstChargeBonus {
		return 0, nil
	}

	rule, exists := e.config.Rules["first_charge"]
	if !exists {
		return 0, fmt.Errorf("first_charge rule not found")
	}

	return rule.Points, nil
}

// evaluateDailyLogin calculates points for daily login with streak bonus
func (e *Engine) evaluateDailyLogin(payload *EventPayload) (int, error) {
	if !e.config.Settings.EnableStreakBonus {
		return 0, nil
	}

	rule, exists := e.config.Rules["daily_login"]
	if !exists {
		return 0, fmt.Errorf("daily_login rule not found")
	}

	streakDays, ok := payload.Data["streak_days"].(int)
	if !ok {
		streakDays = 1
	}

	points := rule.BasePoints

	// Apply streak multiplier
	if streakDays > 1 && streakDays <= rule.MaxStreakDays {
		multiplier := rule.StreakMultiplier
		for i := 2; i <= streakDays; i++ {
			points = int(float64(points) * multiplier)
		}
	}

	return points, nil
}

// GetConfig returns the current rules configuration
func (e *Engine) GetConfig() *RulesConfig {
	return e.config
}
