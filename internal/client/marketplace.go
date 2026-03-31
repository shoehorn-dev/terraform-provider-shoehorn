package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// MarketplaceItem represents an item available in the Shoehorn marketplace catalog.
type MarketplaceItem struct {
	Slug        string `json:"slug"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	AuthorName  string `json:"author_name,omitempty"`
	Category    string `json:"category,omitempty"`
	Tier        string `json:"tier,omitempty"`
	Verified    bool   `json:"verified,omitempty"`
	Featured    bool   `json:"featured,omitempty"`
}

// MarketplaceInstallation represents an installed marketplace item.
type MarketplaceInstallation struct {
	ID          string                 `json:"id"`
	Slug        string                 `json:"itemSlug"`
	Kind        string                 `json:"itemKind"`
	Version     string                 `json:"itemVersion"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config,omitempty"`
	SyncStatus  string                 `json:"syncStatus,omitempty"`
	LastSyncAt  string                 `json:"lastSyncAt,omitempty"`
	InstalledBy string                 `json:"installedBy,omitempty"`
	CreatedAt   string                 `json:"createdAt,omitempty"`
	UpdatedAt   string                 `json:"updatedAt,omitempty"`
}

// marketplaceItemsResponse wraps the list catalog response.
type marketplaceItemsResponse struct {
	Items []MarketplaceItem `json:"items"`
}

// marketplaceInstallationsResponse wraps the list installations response.
type marketplaceInstallationsResponse struct {
	Installations []MarketplaceInstallation `json:"installations"`
}

// marketplaceInstallationResponse wraps a single installation response.
type marketplaceInstallationResponse struct {
	Installation MarketplaceInstallation `json:"installation"`
}

// marketplaceInstallRequest is the request body for installing a marketplace item.
type marketplaceInstallRequest struct {
	Slug string `json:"slug"`
}

// marketplaceConfigRequest is the request body for updating a marketplace item's config.
type marketplaceConfigRequest struct {
	Config map[string]interface{} `json:"config"`
}

// ListMarketplaceItems retrieves available items from the marketplace catalog.
// Optional kind and category parameters filter the results.
func (c *Client) ListMarketplaceItems(ctx context.Context, kind, category string) ([]MarketplaceItem, error) {
	path := "/api/v1/marketplace"

	params := url.Values{}
	if kind != "" {
		params.Set("kind", kind)
	}
	if category != "" {
		params.Set("category", category)
	}
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list marketplace items: %w", err)
	}

	var resp marketplaceItemsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal marketplace items response: %w", err)
	}

	return resp.Items, nil
}

// ListMarketplaceInstallations retrieves all installed marketplace items.
func (c *Client) ListMarketplaceInstallations(ctx context.Context) ([]MarketplaceInstallation, error) {
	body, err := c.Get(ctx, "/api/v1/marketplace/installed")
	if err != nil {
		return nil, fmt.Errorf("list marketplace installations: %w", err)
	}

	var resp marketplaceInstallationsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal marketplace installations response: %w", err)
	}

	return resp.Installations, nil
}

// GetMarketplaceInstallation retrieves an installed marketplace item by slug.
// The API returns the installation object directly (not wrapped in an envelope).
func (c *Client) GetMarketplaceInstallation(ctx context.Context, slug string) (*MarketplaceInstallation, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/marketplace/installed/%s", url.PathEscape(slug)))
	if err != nil {
		return nil, fmt.Errorf("get marketplace installation %q: %w", slug, err)
	}

	var installation MarketplaceInstallation
	if err := json.Unmarshal(body, &installation); err != nil {
		return nil, fmt.Errorf("unmarshal marketplace installation response: %w", err)
	}

	return &installation, nil
}

// InstallMarketplaceItem installs a marketplace item by slug.
// The API returns the installation object directly (not wrapped in an envelope).
func (c *Client) InstallMarketplaceItem(ctx context.Context, slug string) (*MarketplaceInstallation, error) {
	body, err := c.Post(ctx, "/api/v1/marketplace/install", marketplaceInstallRequest{Slug: slug})
	if err != nil {
		return nil, fmt.Errorf("install marketplace item %q: %w", slug, err)
	}

	var installation MarketplaceInstallation
	if err := json.Unmarshal(body, &installation); err != nil {
		return nil, fmt.Errorf("unmarshal install marketplace response: %w", err)
	}

	return &installation, nil
}

// UninstallMarketplaceItem uninstalls a marketplace item by slug.
func (c *Client) UninstallMarketplaceItem(ctx context.Context, slug string) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/marketplace/%s/uninstall", url.PathEscape(slug))); err != nil {
		return fmt.Errorf("uninstall marketplace item %q: %w", slug, err)
	}
	return nil
}

// EnableMarketplaceItem enables an installed marketplace item by slug.
func (c *Client) EnableMarketplaceItem(ctx context.Context, slug string) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/v1/marketplace/%s/enable", url.PathEscape(slug)), nil)
	if err != nil {
		return fmt.Errorf("enable marketplace item %q: %w", slug, err)
	}
	return nil
}

// DisableMarketplaceItem disables an installed marketplace item by slug.
func (c *Client) DisableMarketplaceItem(ctx context.Context, slug string) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/v1/marketplace/%s/disable", url.PathEscape(slug)), nil)
	if err != nil {
		return fmt.Errorf("disable marketplace item %q: %w", slug, err)
	}
	return nil
}

// UpdateMarketplaceItemConfig updates the configuration of an installed marketplace item.
func (c *Client) UpdateMarketplaceItemConfig(ctx context.Context, slug string, config map[string]interface{}) (*MarketplaceInstallation, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/marketplace/%s/config", url.PathEscape(slug)), marketplaceConfigRequest{Config: config})
	if err != nil {
		return nil, fmt.Errorf("update marketplace item config %q: %w", slug, err)
	}

	var installation MarketplaceInstallation
	if err := json.Unmarshal(body, &installation); err != nil {
		return nil, fmt.Errorf("unmarshal update marketplace config response: %w", err)
	}

	return &installation, nil
}
