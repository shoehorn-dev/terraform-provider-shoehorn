package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// NotificationSubscription represents a Shoehorn notification subscription.
// It routes platform events to a channel for a team or user.
type NotificationSubscription struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenant_id,omitempty"`
	Scope         string          `json:"scope"`
	ScopeID       string          `json:"scope_id"`
	Name          string          `json:"name"`
	Enabled       bool            `json:"enabled"`
	EventTypes    []string        `json:"event_types"`
	MinSeverity   string          `json:"min_severity"`
	EntityFilter  json.RawMessage `json:"entity_filter,omitempty"`
	ChannelType   string          `json:"channel_type"`
	ChannelConfig json.RawMessage `json:"channel_config,omitempty"`
	Cadence       string          `json:"cadence"`
	CadenceConfig json.RawMessage `json:"cadence_config,omitempty"`
	CreatedAt     string          `json:"created_at,omitempty"`
	UpdatedAt     string          `json:"updated_at,omitempty"`
}

// NotificationSubscriptionRequest is the request body for creating or updating
// a notification subscription.
type NotificationSubscriptionRequest struct {
	Scope         string          `json:"scope"`
	ScopeID       string          `json:"scope_id"`
	Name          string          `json:"name"`
	Enabled       bool            `json:"enabled"`
	EventTypes    []string        `json:"event_types"`
	MinSeverity   string          `json:"min_severity"`
	EntityFilter  json.RawMessage `json:"entity_filter,omitempty"`
	ChannelType   string          `json:"channel_type"`
	ChannelConfig json.RawMessage `json:"channel_config,omitempty"`
	Cadence       string          `json:"cadence"`
	CadenceConfig json.RawMessage `json:"cadence_config,omitempty"`
}

// notificationSubscriptionsResponse wraps a list of subscriptions in the API response.
type notificationSubscriptionsResponse struct {
	Subscriptions []NotificationSubscription `json:"subscriptions"`
	Total         int                        `json:"total"`
}

// CreateNotificationSubscription creates a new notification subscription.
func (c *Client) CreateNotificationSubscription(ctx context.Context, req NotificationSubscriptionRequest) (*NotificationSubscription, error) {
	body, err := c.Post(ctx, "/api/v1/notifications/subscriptions", req)
	if err != nil {
		return nil, fmt.Errorf("create notification subscription: %w", err)
	}

	var sub NotificationSubscription
	if err := json.Unmarshal(body, &sub); err != nil {
		return nil, fmt.Errorf("unmarshal create notification subscription response: %w", err)
	}

	return &sub, nil
}

// ListNotificationSubscriptions retrieves all subscriptions for a given scope and scope ID.
func (c *Client) ListNotificationSubscriptions(ctx context.Context, scope, scopeID string) ([]NotificationSubscription, error) {
	q := url.Values{}
	q.Set("scope", scope)
	q.Set("scope_id", scopeID)

	body, err := c.Get(ctx, "/api/v1/notifications/subscriptions?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("list notification subscriptions for %s %s: %w", scope, scopeID, err)
	}

	var resp notificationSubscriptionsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal notification subscriptions response: %w", err)
	}

	return resp.Subscriptions, nil
}

// UpdateNotificationSubscription updates an existing notification subscription.
func (c *Client) UpdateNotificationSubscription(ctx context.Context, id string, req NotificationSubscriptionRequest) (*NotificationSubscription, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/notifications/subscriptions/%s", id), req)
	if err != nil {
		return nil, fmt.Errorf("update notification subscription %s: %w", id, err)
	}

	var sub NotificationSubscription
	if err := json.Unmarshal(body, &sub); err != nil {
		return nil, fmt.Errorf("unmarshal update notification subscription response: %w", err)
	}

	return &sub, nil
}

// DeleteNotificationSubscription deletes a notification subscription by ID.
func (c *Client) DeleteNotificationSubscription(ctx context.Context, id string) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/notifications/subscriptions/%s", id)); err != nil {
		return fmt.Errorf("delete notification subscription %s: %w", id, err)
	}

	return nil
}
