package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// K8sAgent represents a registered K8s agent.
type K8sAgent struct {
	ID            int    `json:"id,omitempty"`
	ClusterID     string `json:"clusterId"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	TokenPrefix   string `json:"tokenPrefix,omitempty"`
	Status        string `json:"status,omitempty"`
	OnlineStatus  string `json:"onlineStatus,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
	ExpiresAt     string `json:"expiresAt,omitempty"`
	LastHeartbeat string `json:"lastHeartbeat,omitempty"`
}

// RegisterK8sAgentRequest is the request body for registering a K8s agent.
type RegisterK8sAgentRequest struct {
	ClusterID   string                 `json:"clusterId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	ExpiresIn   *int                   `json:"expiresIn,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RegisterK8sAgentResponse is the response from registering a K8s agent.
// The token is only returned once on creation.
type RegisterK8sAgentResponse struct {
	Token       string `json:"token"`
	TokenPrefix string `json:"tokenPrefix"`
	ClusterID   string `json:"clusterId"`
	Name        string `json:"name"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

// k8sAgentListResponse wraps the list response.
type k8sAgentListResponse struct {
	Agents []K8sAgent `json:"agents"`
	Total  int        `json:"total"`
}

// ListK8sAgents retrieves all K8s agents.
func (c *Client) ListK8sAgents(ctx context.Context) ([]K8sAgent, error) {
	body, err := c.Get(ctx, "/api/v1/k8s/agents")
	if err != nil {
		return nil, fmt.Errorf("list k8s agents: %w", err)
	}

	var resp k8sAgentListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal k8s agents response: %w", err)
	}

	return resp.Agents, nil
}

// GetK8sAgent retrieves a K8s agent by cluster ID.
func (c *Client) GetK8sAgent(ctx context.Context, clusterID string) (*K8sAgent, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/k8s/agents/%s", clusterID))
	if err != nil {
		return nil, fmt.Errorf("get k8s agent %s: %w", clusterID, err)
	}

	var agent K8sAgent
	if err := json.Unmarshal(body, &agent); err != nil {
		return nil, fmt.Errorf("unmarshal k8s agent response: %w", err)
	}

	return &agent, nil
}

// RegisterK8sAgent registers a new K8s agent.
func (c *Client) RegisterK8sAgent(ctx context.Context, req RegisterK8sAgentRequest) (*RegisterK8sAgentResponse, error) {
	body, err := c.Post(ctx, "/api/v1/k8s/agents/register", req)
	if err != nil {
		return nil, fmt.Errorf("register k8s agent: %w", err)
	}

	var resp RegisterK8sAgentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal register k8s agent response: %w", err)
	}

	return &resp, nil
}

// RevokeK8sAgent revokes a K8s agent by cluster ID.
func (c *Client) RevokeK8sAgent(ctx context.Context, clusterID string) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/v1/k8s/agents/%s/revoke", clusterID), map[string]string{})
	if err != nil {
		return fmt.Errorf("revoke k8s agent %s: %w", clusterID, err)
	}
	return nil
}

// DeleteK8sAgent deletes a K8s agent by cluster ID.
func (c *Client) DeleteK8sAgent(ctx context.Context, clusterID string) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/k8s/agents/%s", clusterID)); err != nil {
		return fmt.Errorf("delete k8s agent %s: %w", clusterID, err)
	}
	return nil
}
