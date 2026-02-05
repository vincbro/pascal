package blaise

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

func (c *Client) SearchAreas(ctx context.Context, query string, count int) ([]Location, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/search/area", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", query)
	q.Add("count", strconv.Itoa(count))
	req.URL.RawQuery = q.Encode()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var areas []Location
	if err := json.NewDecoder(resp.Body).Decode(&areas); err != nil {
		return nil, err
	}
	return areas, nil
}
