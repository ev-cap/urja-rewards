# Urja Rewards Service Documentation

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Getting Started](#getting-started)
4. [API Reference](#api-reference)
5. [Configuration](#configuration)
6. [Database Schema](#database-schema)
7. [Rules Engine](#rules-engine)
8. [Authentication & Security](#authentication--security)
9. [Deployment](#deployment)
10. [Monitoring & Observability](#monitoring--observability)
11. [Testing](#testing)
12. [Troubleshooting](#troubleshooting)

## Overview

The Urja Rewards Service is a comprehensive loyalty and rewards platform built with Encore (Go) that enables EV charging companies to implement gamified user engagement through points, rewards, and real-time notifications.

### Key Features
- **Points Accrual**: Earn points for charging sessions, referrals, and ratings
- **Dynamic Rules Engine**: Configurable earning rules without code deployment
- **Rewards Catalog**: Manage and redeem rewards with point-based pricing
- **Real-time Notifications**: Push notifications via Firebase Cloud Messaging
- **Admin Dashboard**: Complete management interface for rules and rewards
- **Event-Driven Architecture**: Pub/Sub based communication between services
- **Audit Trail**: Immutable ledger of all point transactions

### Service Components
- **Accrual Service**: Handles point earning events
- **Redemption Service**: Manages reward catalog and redemptions
- **Admin Service**: Administrative interface for configuration
- **Notifications Service**: Handles push notifications
- **Rules Engine**: Dynamic rule evaluation for point calculations

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Mobile App    │    │   Admin Panel   │    │   Charging      │
│                 │    │                 │    │   Stations      │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │      Encore Gateway       │
                    └─────────────┬─────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
┌───────▼────────┐    ┌──────────▼──────────┐    ┌────────▼────────┐
│  Accrual       │    │   Redemption        │    │   Admin         │
│  Service       │    │   Service           │    │   Service       │
└───────┬────────┘    └──────────┬──────────┘    └────────┬────────┘
        │                        │                        │
        └────────────────────────┼────────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │    Pub/Sub Topics         │
                    │  • UserPointsUpdated      │
                    │  • RedemptionCreated      │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │   Notifications           │
                    │   Service                 │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │   Firebase Cloud          │
                    │   Messaging (FCM)         │
                    └───────────────────────────┘
```

## Getting Started

### Prerequisites

1. **Install Required Tools**:
   ```bash
   # Install Encore
   brew install encoredev/tap/encore
   
   # Install SQLC for database code generation
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   
   # Install OpenAPI code generator
   go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
   
   # Install Go Feature Flag CLI
   go install github.com/thomaspoignant/go-feature-flag/cmd/...
   ```

2. **Verify Installation**:
   ```bash
   encore version
   sqlc version
   oapi-codegen -h
   ```

### Local Development Setup

1. **Clone and Initialize**:
   ```bash
   git clone <repository-url>
   cd urja-rewards
   ```

2. **Generate Database Code**:
   ```bash
   sqlc generate
   ```

3. **Start Local Development**:
   ```bash
   encore run
   ```

4. **Access Services**:
   - **API Gateway**: http://localhost:4000
   - **Encore Dashboard**: http://localhost:4000/encore
   - **Admin Swagger UI**: http://localhost:4000/admin/swagger

## API Reference

### Accrual Service

#### POST /v1/events/charge
Records a charging session and awards points to the user.

**Request Body**:
```json
{
  "session_id": "session_123",
  "kwh": 7.5,
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response**:
```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440001",
  "points": 75,
  "session_id": "session_123"
}
```

**Example**:
```bash
curl -X POST http://localhost:4000/v1/events/charge \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{
    "session_id": "session_123",
    "kwh": 7.5,
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

#### POST /v1/events/referral
Records a successful referral and awards bonus points.

**Request Body**:
```json
{
  "referrer_id": "550e8400-e29b-41d4-a716-446655440000",
  "referee_phone": "+1234567890"
}
```

#### POST /v1/events/rating
Records a user rating and awards points.

**Request Body**:
```json
{
  "session_id": "session_123",
  "rating": 5,
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Redemption Service

#### GET /v1/rewards
Retrieves the available rewards catalog.

**Response**:
```json
{
  "rewards": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "name": "Free Charging Session",
      "description": "30 minutes of free charging",
      "cost": 500,
      "segment": ""
    }
  ]
}
```

#### POST /v1/redeem
Redeems a reward using points.

**Request Body**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "reward_id": "550e8400-e29b-41d4-a716-446655440002"
}
```

**Response**:
```json
{
  "redemption_id": "550e8400-e29b-41d4-a716-446655440003",
  "points_spent": 500,
  "status": "PENDING"
}
```

#### GET /v1/redemptions/{id}
Retrieves redemption status.

**Response**:
```json
{
  "redemption_id": "550e8400-e29b-41d4-a716-446655440003",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "reward_id": "550e8400-e29b-41d4-a716-446655440002",
  "points_spent": 500,
  "status": "FULFILLED",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Admin Service

#### Authentication
All admin endpoints require JWT authentication with `product-admin` role.

**Headers**:
```
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

#### Rules Management

**GET /admin/rules** - List all rules
**POST /admin/rules** - Create a new rule
**GET /admin/rules/{id}** - Get specific rule
**PUT /admin/rules/{id}** - Update rule
**DELETE /admin/rules/{id}** - Delete rule

**Example Rule Creation**:
```bash
curl -X POST http://localhost:4000/admin/rules \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Weekend Bonus",
    "description": "Double points on weekends",
    "config": {
      "multiplier": 2.0,
      "days": ["saturday", "sunday"]
    }
  }'
```

#### Rewards Management

**GET /admin/rewards** - List all rewards
**POST /admin/rewards** - Create a new reward
**GET /admin/rewards/{id}** - Get specific reward
**PUT /admin/rewards/{id}** - Update reward

**Example Reward Creation**:
```bash
curl -X POST http://localhost:4000/admin/rewards \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium Parking",
    "description": "Reserved parking spot for 1 hour",
    "cost": 1000,
    "segment": {
      "feature_flag": "premium_users"
    }
  }'
```

#### Segments Management

**GET /admin/segments** - List all segments
**POST /admin/segments** - Create a new segment
**GET /admin/segments/{id}** - Get specific segment
**PUT /admin/segments/{id}** - Update segment

## Configuration

### Rules Configuration (`rules.yaml`)

The rules engine configuration defines how points are calculated for different events:

```yaml
rules:
  # Points per kWh for charging events
  charge_kwh:
    points_per_kwh: 10
    description: "Points earned per kWh charged"
    
  # Points for referral events
  referral:
    points: 300
    description: "Points earned for successful referral"
    
  # Points for rating events
  rating:
    points: 50
    description: "Points earned for leaving a rating"
    
  # Points for first charge bonus
  first_charge:
    points: 100
    description: "Bonus points for first charge session"
    
  # Points for daily login streak
  daily_login:
    base_points: 10
    streak_multiplier: 1.5
    max_streak_days: 7
    description: "Points for daily login with streak bonus"

settings:
  max_points_per_day: 1000
  max_points_per_event: 500
  enable_streak_bonus: true
  enable_first_charge_bonus: true
```

### Environment Variables

```bash
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/urja_rewards

# JWT Configuration
JWT_SECRET=your-jwt-secret-key
JWT_ISSUER=urja-rewards

# Firebase Configuration
FCM_PROJECT_ID=your-firebase-project-id
FCM_PRIVATE_KEY_ID=your-private-key-id
FCM_PRIVATE_KEY=your-private-key
FCM_CLIENT_EMAIL=your-client-email
FCM_CLIENT_ID=your-client-id

# Feature Flags
GOFF_BACKEND_TYPE=yaml
GOFF_BACKEND_YAML_PATH=flags.yaml
```

## Database Schema

### Core Tables

#### users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### points_events (Immutable Ledger)
```sql
CREATE TABLE points_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL, -- CHARGE_KWH / REFERRAL / RATING / MANUAL_ADJUST
    ref_id TEXT, -- session-id, friend-id etc.
    points INT NOT NULL,
    meta JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### rewards_catalog
```sql
CREATE TABLE rewards_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    cost INT NOT NULL,
    segment JSONB, -- user attributes or feature-flag keys
    active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### redemptions
```sql
CREATE TABLE redemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_id UUID NOT NULL REFERENCES rewards_catalog(id) ON DELETE CASCADE,
    points_spent INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING', -- PENDING / FULFILLED / EXPIRED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### rules
```sql
CREATE TABLE rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    config JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Rules Engine

The rules engine provides dynamic point calculation without code deployment. It supports:

### Rule Types

1. **Fixed Multiplier Rules**: Points per kWh, fixed bonuses
2. **Conditional Rules**: Time-based, user-segment based
3. **Complex Rules**: Multi-step calculations with conditions

### Rule Configuration Examples

**Simple Charge Rule**:
```json
{
  "type": "charge_kwh",
  "points_per_kwh": 10,
  "max_kwh_per_session": 50
}
```

**Time-based Bonus**:
```json
{
  "type": "time_bonus",
  "base_points": 10,
  "peak_hours_multiplier": 1.5,
  "peak_hours": ["18:00", "22:00"]
}
```

**Segment-based Rule**:
```json
{
  "type": "segment_bonus",
  "segment": "premium_users",
  "multiplier": 2.0,
  "description": "Double points for premium users"
}
```

## Authentication & Security

### JWT Authentication

All endpoints require JWT authentication. The JWT should contain:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "user",
  "exp": 1642234567,
  "iat": 1642230967
}
```

### Role-Based Access Control (RBAC)

- **user**: Can earn points and redeem rewards
- **station-operator**: Can record charging sessions
- **product-admin**: Can manage rules, rewards, and segments

### Security Headers

```bash
# Idempotency key for POST requests
X-Idempotency-Key: unique-request-id

# Rate limiting headers
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1642234567
```

## Deployment

### Encore Cloud Deployment

1. **Initialize Encore App**:
   ```bash
   encore app init urja-rewards
   ```

2. **Deploy to Production**:
   ```bash
   encore deploy --env prod
   ```

3. **Set Production Secrets**:
   ```bash
   encore secret set --env prod DATABASE_URL="postgresql://..."
   encore secret set --env prod JWT_SECRET="your-secret"
   encore secret set --env prod FCM_PROJECT_ID="your-project"
   ```

### Environment Configuration

**Development**:
```bash
encore run
```

**Staging**:
```bash
encore deploy --env staging
```

**Production**:
```bash
encore deploy --env prod
```

### Database Migrations

Encore automatically handles database migrations. To run manually:

```bash
# Generate migration
sqlc generate

# Apply migration
encore db migrate
```

## Monitoring & Observability

### Metrics

The service automatically emits metrics for:

- **Points Earned**: Total points earned per user/event type
- **Points Redeemed**: Total points spent on redemptions
- **Redemption Success Rate**: Percentage of successful redemptions
- **API Response Times**: P95, P99 response times
- **Error Rates**: 4xx and 5xx error percentages

### Logging

Structured logging is enabled for all services:

```json
{
  "level": "info",
  "service": "accrual",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "CHARGE_KWH",
  "points": 75,
  "session_id": "session_123",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Health Checks

**Service Health**:
```bash
curl http://localhost:4000/health
```

**Database Health**:
```bash
curl http://localhost:4000/admin/health
```

### Alerting

Configure alerts for:
- High error rates (>5%)
- Slow response times (>500ms P95)
- Low redemption success rates (<90%)
- Database connection issues

## Testing

### Unit Tests

Run unit tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

### Integration Tests

Run integration tests:
```bash
go test ./services/... -tags=integration
```

### End-to-End Tests

```bash
# Start test environment
encore test

# Run E2E tests
go test ./tests/e2e/...
```

### Load Testing

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Run load test
hey -n 1000 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"session_id":"test","kwh":5,"user_id":"user123"}' \
  http://localhost:4000/v1/events/charge
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Errors

**Symptoms**: 500 errors, "connection refused"
**Solution**: Check DATABASE_URL and network connectivity

```bash
# Test database connection
psql $DATABASE_URL -c "SELECT 1"
```

#### 2. JWT Authentication Failures

**Symptoms**: 401 Unauthorized errors
**Solution**: Verify JWT token format and signature

```bash
# Decode JWT (without verification)
echo "your.jwt.token" | cut -d. -f2 | base64 -d | jq
```

#### 3. Rules Engine Not Loading

**Symptoms**: Default point calculation used
**Solution**: Check rules.yaml file and permissions

```bash
# Verify rules file
cat rules.yaml | yq eval
```

#### 4. Pub/Sub Message Delivery Failures

**Symptoms**: Events not processed, notifications not sent
**Solution**: Check subscription configuration and handler errors

```bash
# Check service logs
encore logs --service notifications
```

### Debug Mode

Enable debug logging:

```bash
export ENCORE_DEBUG=1
encore run
```

### Performance Tuning

1. **Database Indexes**: Ensure proper indexes on frequently queried columns
2. **Connection Pooling**: Configure appropriate connection pool sizes
3. **Caching**: Implement Redis caching for frequently accessed data
4. **Rate Limiting**: Configure appropriate rate limits for API endpoints

### Support

For additional support:

1. **Documentation**: Check the README.md and inline code comments
2. **Issues**: Create GitHub issues for bugs or feature requests
3. **Discussions**: Use GitHub Discussions for questions and ideas
4. **Encore Support**: Contact Encore support for platform-specific issues

---

**Version**: 1.0.0  
**Last Updated**: January 2024  
**Maintainer**: Urja Rewards Team 