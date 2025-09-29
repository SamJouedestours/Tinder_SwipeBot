package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

type Candidate struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Age       int      `json:"age"`
	Interests []string `json:"interests"`
}
type SwipeRequest struct {
	CandidateID string `json:"candidate_id"`
	Action      string `json:"action"`
}
type SwipeResponse struct {
	Matched bool   `json:"matched"`
	Message string `json:"message"`
}

type Client struct {
	base   *url.URL
	token  string
	http   *http.Client
	limit  *rate.Limiter
	retryN int
}

func NewClient(baseURL, token string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		base:   u,
		token:  token,
		http:   &http.Client{Timeout: 15 * time.Second},
		limit:  rate.NewLimiter(rate.Every(500*time.Millisecond), 1),
		retryN: 3,
	}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any) error {
	if err := c.limit.Wait(ctx); err != nil {
		return err
	}
	u := c.base.ResolveReference(&url.URL{Path: path})

	var req *http.Request
	var err error
	if in != nil {
		pr, pw := http.Pipe()
		go func() {
			defer pw.Close()
			enc := json.NewEncoder(pw)
			enc.SetEscapeHTML(true)
			_ = enc.Encode(in)
		}()
		req, err = http.NewRequestWithContext(ctx, method, u.String(), pr)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, u.String(), nil)
	}
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	var resp *http.Response
	for attempt := 0; attempt <= c.retryN; attempt++ {
		resp, err = c.http.Do(req)
		if err != nil {
			if attempt == c.retryN {
				return err
			}
		} else {
			if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
				resp.Body.Close()
				if attempt == c.retryN {
					return fmt.Errorf("server error %d after retries", resp.StatusCode)
				}
			} else {
				break
			}
		}
		time.Sleep(time.Duration(1<<attempt) * 300 * time.Millisecond)
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) GetCandidates(ctx context.Context, limit int) ([]Candidate, error) {
	type respT struct {
		Results []Candidate `json:"results"`
	}
	var r respT
	path := fmt.Sprintf("/candidates?limit=%d", limit)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &r); err != nil {
		return nil, err
	}
	return r.Results, nil
}

func (c *Client) Swipe(ctx context.Context, id string, action string) (*SwipeResponse, error) {
	req := SwipeRequest{CandidateID: id, Action: action}
	var resp SwipeResponse
	if err := c.doJSON(ctx, http.MethodPost, "/swipe", &req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
