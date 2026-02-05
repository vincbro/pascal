package blaise

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (c *Client) GetAge(ctx context.Context) (uint32, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/gtfs/age", nil)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	val, err := strconv.ParseUint(strings.TrimSpace(string(bodyBytes)), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse age: %w", err)
	}
	return uint32(val), nil
}

func (c *Client) TriggerRefresh(ctx context.Context, url string) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/gtfs/fetch-url", nil)
	q := req.URL.Query()
	q.Add("q", url)
	req.URL.RawQuery = q.Encode()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to trigger refresh, status: %d", resp.StatusCode)
	}
	return nil
}
