package blaise

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) Routing(ctx context.Context, from, to, timeStr string, departure bool) (Itinerary, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/routing", nil)

	q := req.URL.Query()
	q.Add("from", from)
	q.Add("to", to)
	if departure {
		q.Add("departure_at", timeStr)
	} else {
		q.Add("arrive_at", timeStr)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Itinerary{}, err
	}
	defer resp.Body.Close()

	var itenirary Itinerary
	if err := json.NewDecoder(resp.Body).Decode(&itenirary); err != nil {
		return Itinerary{}, err
	}
	return itenirary, nil
}
