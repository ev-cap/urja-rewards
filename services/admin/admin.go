//encore:service
package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"encore.app/internal/db"
	"encore.dev/pubsub"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sqlc-dev/pqtype"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"gopkg.in/yaml.v3"
)

// Service configuration
var (
	// Database queries
	queries *db.Queries
)

// Pub/Sub topics for admin events
var (
	RuleUpdated    = pubsub.NewTopic[*RuleUpdateEvent]("rule-updated", pubsub.TopicConfig{DeliveryGuarantee: pubsub.AtLeastOnce})
	RewardUpdated  = pubsub.NewTopic[*RewardUpdateEvent]("reward-updated", pubsub.TopicConfig{DeliveryGuarantee: pubsub.AtLeastOnce})
	SegmentUpdated = pubsub.NewTopic[*SegmentUpdateEvent]("segment-updated", pubsub.TopicConfig{DeliveryGuarantee: pubsub.AtLeastOnce})
)

// Event types for pub/sub
type RuleUpdateEvent struct {
	RuleID    uuid.UUID `json:"rule_id"`
	Action    string    `json:"action"` // "created", "updated", "deleted"
	RuleName  string    `json:"rule_name"`
	UpdatedBy uuid.UUID `json:"updated_by"`
	Timestamp time.Time `json:"timestamp"`
}

type RewardUpdateEvent struct {
	RewardID   uuid.UUID `json:"reward_id"`
	Action     string    `json:"action"` // "created", "updated"
	RewardName string    `json:"reward_name"`
	UpdatedBy  uuid.UUID `json:"updated_by"`
	Timestamp  time.Time `json:"timestamp"`
}

type SegmentUpdateEvent struct {
	SegmentID   uuid.UUID `json:"segment_id"`
	Action      string    `json:"action"` // "created", "updated"
	SegmentName string    `json:"segment_name"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
	Timestamp   time.Time `json:"timestamp"`
}

// JWT Claims structure
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}

// AdminService implements the ServerInterface
type AdminService struct{}

// init initializes the service
func init() {
	// Initialize database queries (Encore injects DB connection)
	queries = db.New(nil)
}

// RBAC middleware to check for product-admin role
func requireProductAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get user from context (set by JWT middleware)
		user := c.Get("user")
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
		}

		claims, ok := user.(*Claims)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
		}

		// Check for product-admin role
		if claims.Role != "product-admin" {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}

		return next(c)
	}
}

// JWT middleware for authentication
func jwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
		}
		// Extract Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header format")
		}
		// TODO: Implement proper JWT validation
		// For now, create a mock user for testing
		claims := &Claims{
			UserID: uuid.New(),
			Email:  "admin@urja.com",
			Role:   "product-admin",
		}
		c.Set("user", claims)
		return next(c)
	}
}

// Health check endpoint
func (s *AdminService) GetHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"service":   "admin",
	})
}

// Rules endpoints
func (s *AdminService) GetRules(ctx echo.Context) error {
	rules, err := queries.ListRules(ctx.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve rules")
	}
	var response []Rule
	for _, rule := range rules {
		var config map[string]interface{}
		if err := json.Unmarshal(rule.Config, &config); err != nil {
			config = nil
		}
		response = append(response, Rule{
			Id:          (*openapi_types.UUID)(&rule.ID),
			Name:        &rule.Name,
			Description: &rule.Description.String,
			Config:      &config,
			Active:      &rule.Active,
			CreatedAt:   &rule.CreatedAt,
			UpdatedAt:   &rule.UpdatedAt,
		})
	}
	return ctx.JSON(http.StatusOK, response)
}

func (s *AdminService) PostRules(ctx echo.Context) error {
	var req PostRulesJSONBody
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	user := ctx.Get("user").(*Claims)
	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid config format")
	}
	desc := sql.NullString{String: "", Valid: false}
	if req.Description != nil {
		desc = sql.NullString{String: *req.Description, Valid: true}
	}
	createdBy := uuid.NullUUID{UUID: user.UserID, Valid: true}
	rule, err := queries.CreateRule(ctx.Request().Context(), db.CreateRuleParams{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: desc,
		Config:      configBytes,
		Active:      true,
		CreatedBy:   createdBy,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create rule")
	}
	var config map[string]interface{}
	_ = json.Unmarshal(rule.Config, &config)
	response := Rule{
		Id:          (*openapi_types.UUID)(&rule.ID),
		Name:        &rule.Name,
		Description: &rule.Description.String,
		Config:      &config,
		Active:      &rule.Active,
		CreatedAt:   &rule.CreatedAt,
		UpdatedAt:   &rule.UpdatedAt,
	}
	return ctx.JSON(http.StatusCreated, response)
}

func (s *AdminService) GetRulesRuleId(ctx echo.Context, ruleId openapi_types.UUID) error {
	rule, err := queries.GetRule(ctx.Request().Context(), uuid.UUID(ruleId))
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Rule not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve rule")
	}
	var config map[string]interface{}
	_ = json.Unmarshal(rule.Config, &config)
	response := Rule{
		Id:          (*openapi_types.UUID)(&rule.ID),
		Name:        &rule.Name,
		Description: &rule.Description.String,
		Config:      &config,
		Active:      &rule.Active,
		CreatedAt:   &rule.CreatedAt,
		UpdatedAt:   &rule.UpdatedAt,
	}
	return ctx.JSON(http.StatusOK, response)
}

func (s *AdminService) PutRulesRuleId(ctx echo.Context, ruleId openapi_types.UUID) error {
	var req PutRulesRuleIdJSONBody
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	desc := sql.NullString{String: "", Valid: false}
	if req.Description != nil {
		desc = sql.NullString{String: *req.Description, Valid: true}
	}
	var name string
	if req.Name != nil {
		name = *req.Name
	}
	var configBytes []byte
	if req.Config != nil {
		var err error
		configBytes, err = json.Marshal(req.Config)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid config format")
		}
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	rule, err := queries.UpdateRule(ctx.Request().Context(), db.UpdateRuleParams{
		ID:          uuid.UUID(ruleId),
		Name:        name,
		Description: desc,
		Config:      configBytes,
		Active:      active,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Rule not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update rule")
	}
	var config map[string]interface{}
	_ = json.Unmarshal(rule.Config, &config)
	response := Rule{
		Id:          (*openapi_types.UUID)(&rule.ID),
		Name:        &rule.Name,
		Description: &rule.Description.String,
		Config:      &config,
		Active:      &rule.Active,
		CreatedAt:   &rule.CreatedAt,
		UpdatedAt:   &rule.UpdatedAt,
	}
	return ctx.JSON(http.StatusOK, response)
}

func (s *AdminService) DeleteRulesRuleId(ctx echo.Context, ruleId openapi_types.UUID) error {
	// Get user from context
	user := ctx.Get("user").(*Claims)

	// Delete rule from database
	err := queries.DeleteRule(ctx.Request().Context(), uuid.UUID(ruleId))
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Rule not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete rule")
	}

	// Publish event
	RuleUpdated.Publish(ctx.Request().Context(), &RuleUpdateEvent{
		RuleID:    uuid.UUID(ruleId),
		Action:    "deleted",
		RuleName:  "deleted",
		UpdatedBy: user.UserID,
		Timestamp: time.Now(),
	})

	return ctx.NoContent(http.StatusNoContent)
}

// Rewards endpoints
func (s *AdminService) GetRewards(ctx echo.Context, params GetRewardsParams) error {
	user := ctx.Get("user").(*Claims)
	ffCtx := ffcontext.NewEvaluationContext(user.UserID.String())
	// Check early-access feature flag
	earlyAccessFlag, err := ffclient.BoolVariation("early-access", ffCtx, false)
	if err != nil {
		earlyAccessFlag = false
	}
	// Build query parameters
	active := true
	if params.Active != nil {
		active = *params.Active
	}
	rewards, err := queries.ListRewards(ctx.Request().Context(), active)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve rewards")
	}
	var filteredRewards []Reward
	for _, reward := range rewards {
		var segment map[string]interface{}
		if reward.Segment.Valid {
			_ = json.Unmarshal(reward.Segment.RawMessage, &segment)
		}
		// Skip early-access rewards if user doesn't have access
		if !earlyAccessFlag && segment != nil {
			if t, ok := segment["type"]; ok && t == "early-access" {
				continue
			}
		}
		// Apply segment filtering if specified
		if params.Segment != nil && segment != nil {
			if n, ok := segment["name"]; ok && n != *params.Segment {
				continue
			}
		}
		cost := int(reward.Cost)
		filteredRewards = append(filteredRewards, Reward{
			Id:          (*openapi_types.UUID)(&reward.ID),
			Name:        &reward.Name,
			Description: &reward.Description.String,
			Cost:        &cost,
			Segment:     &segment,
			Active:      &reward.Active,
			CreatedBy:   (*openapi_types.UUID)(&reward.CreatedBy.UUID),
			CreatedAt:   &reward.CreatedAt,
		})
	}
	return ctx.JSON(http.StatusOK, filteredRewards)
}

func (s *AdminService) PostRewards(ctx echo.Context) error {
	var req PostRewardsJSONBody
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	user := ctx.Get("user").(*Claims)
	desc := sql.NullString{String: "", Valid: false}
	if req.Description != nil {
		desc = sql.NullString{String: *req.Description, Valid: true}
	}
	var segmentRaw pqtype.NullRawMessage
	if req.Segment != nil {
		b, err := json.Marshal(req.Segment)
		if err == nil {
			segmentRaw = pqtype.NullRawMessage{RawMessage: b, Valid: true}
		}
	}
	createdBy := uuid.NullUUID{UUID: user.UserID, Valid: true}
	reward, err := queries.CreateReward(ctx.Request().Context(), db.CreateRewardParams{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: desc,
		Cost:        int32(req.Cost),
		Segment:     segmentRaw,
		Active:      true,
		CreatedBy:   createdBy,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create reward")
	}
	var segment map[string]interface{}
	if reward.Segment.Valid {
		_ = json.Unmarshal(reward.Segment.RawMessage, &segment)
	}
	cost := int(reward.Cost)
	response := Reward{
		Id:          (*openapi_types.UUID)(&reward.ID),
		Name:        &reward.Name,
		Description: &reward.Description.String,
		Cost:        &cost,
		Segment:     &segment,
		Active:      &reward.Active,
		CreatedBy:   (*openapi_types.UUID)(&reward.CreatedBy.UUID),
		CreatedAt:   &reward.CreatedAt,
	}
	return ctx.JSON(http.StatusCreated, response)
}

func (s *AdminService) GetRewardsRewardId(ctx echo.Context, rewardId openapi_types.UUID) error {
	reward, err := queries.GetReward(ctx.Request().Context(), uuid.UUID(rewardId))
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Reward not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve reward")
	}
	var segment map[string]interface{}
	if reward.Segment.Valid {
		_ = json.Unmarshal(reward.Segment.RawMessage, &segment)
	}
	cost := int(reward.Cost)
	response := Reward{
		Id:          (*openapi_types.UUID)(&reward.ID),
		Name:        &reward.Name,
		Description: &reward.Description.String,
		Cost:        &cost,
		Segment:     &segment,
		Active:      &reward.Active,
		CreatedBy:   (*openapi_types.UUID)(&reward.CreatedBy.UUID),
		CreatedAt:   &reward.CreatedAt,
	}
	return ctx.JSON(http.StatusOK, response)
}

func (s *AdminService) PutRewardsRewardId(ctx echo.Context, rewardId openapi_types.UUID) error {
	var req PutRewardsRewardIdJSONBody
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	desc := sql.NullString{String: "", Valid: false}
	if req.Description != nil {
		desc = sql.NullString{String: *req.Description, Valid: true}
	}
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	var segmentRaw pqtype.NullRawMessage
	if req.Segment != nil {
		b, err := json.Marshal(req.Segment)
		if err == nil {
			segmentRaw = pqtype.NullRawMessage{RawMessage: b, Valid: true}
		}
	}
	cost := int32(0)
	if req.Cost != nil {
		cost = int32(*req.Cost)
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	reward, err := queries.UpdateReward(ctx.Request().Context(), db.UpdateRewardParams{
		ID:          uuid.UUID(rewardId),
		Name:        name,
		Description: desc,
		Cost:        cost,
		Segment:     segmentRaw,
		Active:      active,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Reward not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update reward")
	}
	var segment map[string]interface{}
	if reward.Segment.Valid {
		_ = json.Unmarshal(reward.Segment.RawMessage, &segment)
	}
	costVal := int(reward.Cost)
	response := Reward{
		Id:          (*openapi_types.UUID)(&reward.ID),
		Name:        &reward.Name,
		Description: &reward.Description.String,
		Cost:        &costVal,
		Segment:     &segment,
		Active:      &reward.Active,
		CreatedBy:   (*openapi_types.UUID)(&reward.CreatedBy.UUID),
		CreatedAt:   &reward.CreatedAt,
	}
	return ctx.JSON(http.StatusOK, response)
}

// Segments endpoints
func (s *AdminService) GetSegments(ctx echo.Context) error {
	segments, err := queries.ListSegments(ctx.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve segments")
	}
	var response []Segment
	for _, segment := range segments {
		var criteria map[string]interface{}
		_ = json.Unmarshal(segment.Criteria, &criteria)
		response = append(response, Segment{
			Id:        (*openapi_types.UUID)(&segment.ID),
			Name:      &segment.Name,
			Criteria:  &criteria,
			Active:    &segment.Active,
			CreatedAt: &segment.CreatedAt,
		})
	}
	return ctx.JSON(http.StatusOK, response)
}

func (s *AdminService) PostSegments(ctx echo.Context) error {
	var req PostSegmentsJSONBody
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	user := ctx.Get("user").(*Claims)
	criteriaBytes, err := json.Marshal(req.Criteria)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid criteria format")
	}
	createdBy := uuid.NullUUID{UUID: user.UserID, Valid: true}
	segment, err := queries.CreateSegment(ctx.Request().Context(), db.CreateSegmentParams{
		ID:        uuid.New(),
		Name:      req.Name,
		Criteria:  criteriaBytes,
		Active:    true,
		CreatedBy: createdBy,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create segment")
	}
	var criteria map[string]interface{}
	_ = json.Unmarshal(segment.Criteria, &criteria)
	response := Segment{
		Id:        (*openapi_types.UUID)(&segment.ID),
		Name:      &segment.Name,
		Criteria:  &criteria,
		Active:    &segment.Active,
		CreatedAt: &segment.CreatedAt,
	}
	return ctx.JSON(http.StatusCreated, response)
}

func (s *AdminService) GetSegmentsSegmentId(ctx echo.Context, segmentId openapi_types.UUID) error {
	segment, err := queries.GetSegment(ctx.Request().Context(), uuid.UUID(segmentId))
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Segment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve segment")
	}
	var criteria map[string]interface{}
	_ = json.Unmarshal(segment.Criteria, &criteria)
	response := Segment{
		Id:        (*openapi_types.UUID)(&segment.ID),
		Name:      &segment.Name,
		Criteria:  &criteria,
		Active:    &segment.Active,
		CreatedAt: &segment.CreatedAt,
	}
	return ctx.JSON(http.StatusOK, response)
}

func (s *AdminService) PutSegmentsSegmentId(ctx echo.Context, segmentId openapi_types.UUID) error {
	var req PutSegmentsSegmentIdJSONBody
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	var criteriaBytes []byte
	if req.Criteria != nil {
		var err error
		criteriaBytes, err = json.Marshal(req.Criteria)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid criteria format")
		}
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	segment, err := queries.UpdateSegment(ctx.Request().Context(), db.UpdateSegmentParams{
		ID:       uuid.UUID(segmentId),
		Name:     name,
		Criteria: criteriaBytes,
		Active:   active,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Segment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update segment")
	}
	var criteria map[string]interface{}
	_ = json.Unmarshal(segment.Criteria, &criteria)
	response := Segment{
		Id:        (*openapi_types.UUID)(&segment.ID),
		Name:      &segment.Name,
		Criteria:  &criteria,
		Active:    &segment.Active,
		CreatedAt: &segment.CreatedAt,
	}
	return ctx.JSON(http.StatusOK, response)
}

// SwaggerUIResponse represents the response for Swagger UI
type SwaggerUIResponse struct {
	HTML string `json:"html"`
}

// OpenAPISpecResponse represents the response for OpenAPI specification
type OpenAPISpecResponse struct {
	Spec string `json:"spec"`
}

//encore:api public path=/v1/docs method=GET
func SwaggerUIHandler(ctx context.Context) (string, error) {
	// Read the static HTML file
	data, err := os.ReadFile("services/admin/swagger.html")
	if err != nil {
		return "", fmt.Errorf("could not read Swagger UI HTML: %w", err)
	}

	return string(data), nil
}

//encore:api public path=/v1/docs/openapi.json method=GET
func OpenAPISpecHandler(ctx context.Context) (map[string]interface{}, error) {
	// Read the admin.yaml file and convert it to JSON
	data, err := os.ReadFile("services/admin/admin.yaml")
	if err != nil {
		return nil, fmt.Errorf("could not read OpenAPI spec: %w", err)
	}

	var spec map[string]interface{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("could not parse OpenAPI YAML: %w", err)
	}

	return spec, nil
}

// initService initializes the admin service
func initService() {
	// Create Echo server
	e := echo.New()

	// Add middleware
	e.Use(jwtMiddleware)
	e.Use(requireProductAdmin)

	// Create admin service
	adminService := &AdminService{}

	// Register handlers
	RegisterHandlers(e, adminService)

	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil {
			// Remove encore.Log.Error("Failed to start admin service", "error", err)
		}
	}()

	// Remove encore.Log.Info("Admin service started on port 8080")

	// Initialize go-feature-flag
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 3 * time.Second,
	})
	if err != nil {
		panic("failed to initialize go-feature-flag: " + err.Error())
	}
}
