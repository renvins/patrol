package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

type QueryResponse struct {
	Status string    `json:"status"`
	Data   QueryData `json:"data"`
}

type QueryData struct {
	ResultType string        `json:"resultType"`
	Result     []QueryResult `json:"result"`
}

type QueryResult struct {
	Metric map[string]string `json:"metric"`
	Value  []any             `json:"value"`
}

func New(baseURL string) *Client {
	return &Client{
		HTTPClient: &http.Client{},
		BaseURL:    baseURL,
	}
}

func (c *Client) QueryInstant(ctx context.Context, query string) (float64, error) {
	params := url.Values{}
	params.Set("query", query)

	u := c.BaseURL + "/api/v1/query?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var queryResponse QueryResponse
	err = json.Unmarshal(body, &queryResponse)
	if err != nil {
		return 0, err
	}
	if queryResponse.Status != "success" {
		return 0, fmt.Errorf("query %q was not success", query)
	}
	if len(queryResponse.Data.Result) == 0 {
		return 0, fmt.Errorf("query %q returned no results", query)
	}
	result, err := strconv.ParseFloat(queryResponse.Data.Result[0].Value[1].(string), 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}
