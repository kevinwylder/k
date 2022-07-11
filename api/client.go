package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kevinwylder/k/fs"
)

type Client struct {
	http  http.Client
	addr  string
	data *fs.StorageDir
}

func NewClient(addr string, data *fs.StorageDir) *Client {
	return &Client{
		addr: addr,
		data: data,
	}
}

func (c *Client) day(ctx context.Context, t time.Time, segment io.Reader) error {
	url := fmt.Sprintf("http://%s/day?t=%d", c.addr, t.Unix())
	method := "GET"
	if segment != nil {
		method = "POST"
	}
	req, err := http.NewRequestWithContext(ctx, method, url, segment)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	err = c.data.Write(t, res.Body, true)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func (c *Client) DownloadDay(ctx context.Context, t time.Time) error {
	return c.day(ctx, t, nil)
}

func (c *Client) UploadSegment(ctx context.Context, t time.Time, segment io.Reader) error {
	return c.day(ctx, t, segment)
}
