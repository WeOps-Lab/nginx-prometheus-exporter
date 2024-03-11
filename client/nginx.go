package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

const templateMetrics string = `Active connections: %d
server accepts handled requests
%d %d %d
Reading: %d Writing: %d Waiting: %d
`

const templateTengineMetrics string = `Active connections: %d
server accepts handled requests request_time 
%d %d %d %d 
Reading: %d Writing: %d Waiting: %d
`

// NginxClient allows you to fetch NGINX metrics from the stub_status page.
type NginxClient struct {
	apiEndpoint string
	httpClient  *http.Client
}

// StubStats represents NGINX stub_status metrics.
type StubStats struct {
	Connections  StubConnections
	Requests     int64
	RequestsTime int64
}

// StubConnections represents connections related metrics.
type StubConnections struct {
	Active   int64
	Accepted int64
	Handled  int64
	Reading  int64
	Writing  int64
	Waiting  int64
}

// NewNginxClient creates an NginxClient.
func NewNginxClient(httpClient *http.Client, apiEndpoint string) (*NginxClient, error) {
	client := &NginxClient{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
	}

	_, err := client.GetStubStats()
	return client, err
}

// GetStubStats fetches the stub_status metrics.
func (client *NginxClient) GetStubStats() (*StubStats, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.apiEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a get request: %w", err)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %w", client.apiEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %w", err)
	}

	r := bytes.NewReader(body)
	stats, err := parseStubStats(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body %q: %w", string(body), err)
	}

	return stats, nil
}

func parseStubStats(r io.Reader) (*StubStats, error) {
	var s StubStats

	// 读取输入流到缓冲区
	buf1 := new(bytes.Buffer)
	if _, err := buf1.ReadFrom(r); err != nil {
		return nil, err
	}

	// 复制一份输入流到第二个缓冲区
	buf2 := bytes.NewBuffer(buf1.Bytes())

	if _, err := fmt.Fscanf(buf1, templateMetrics,
		&s.Connections.Active,
		&s.Connections.Accepted,
		&s.Connections.Handled,
		&s.Requests,
		&s.Connections.Reading,
		&s.Connections.Writing,
		&s.Connections.Waiting); err == nil {
		return &s, nil
	}

	if _, err := fmt.Fscanf(buf2, templateTengineMetrics,
		&s.Connections.Active,
		&s.Connections.Accepted,
		&s.Connections.Handled,
		&s.Requests,
		&s.RequestsTime,
		&s.Connections.Reading,
		&s.Connections.Writing,
		&s.Connections.Waiting); err == nil {
		return &s, nil
	}

	return nil, fmt.Errorf("failed to parse template metrics")
}
