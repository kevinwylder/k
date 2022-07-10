package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/kevinwylder/k/files"
)

type Client struct {
	http  http.Client
	addr  string
	cache *files.CacheDir
}

func NewClient(addr string, cache *files.CacheDir) *Client {
	return &Client{addr: addr}
}

func (c *Client) get(ctx context.Context, path string, dst interface{}) error {
	url := fmt.Sprintf("http://%s/%s", c.addr, path)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	switch t := dst.(type) {
	case nil:
	case *os.File:
		_, err = io.Copy(t, req.Body)
		if err != nil {
			return fmt.Errorf("read body: %w", err)
		}
	default:
		err := json.NewDecoder(res.Body).Decode(t)
		if err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}
	return nil
}

func (c *Client) Check(ctx context.Context) error {
	return c.get(ctx, "/ping", nil)
}

func (c *Client) Upload(ctx context.Context, t time.Time) error {

}
