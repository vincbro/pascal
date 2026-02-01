package blaise

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) Routing(ctx context.Context, from string, to string) (Itenirary, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/routing", nil)

	q := req.URL.Query()
	q.Add("from", from)
	q.Add("to", to)
	req.URL.RawQuery = q.Encode()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Itenirary{}, err
	}
	defer resp.Body.Close()

	var itenirary Itenirary
	if err := json.NewDecoder(resp.Body).Decode(&itenirary); err != nil {
		return Itenirary{}, err
	}
	return itenirary, nil
}
