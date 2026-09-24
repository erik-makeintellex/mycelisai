package invocation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type countEvidence struct {
	InvocationID  string `json:"invocation_id"`
	Counter       string `json:"counter"`
	ObservedCount int    `json:"observed_count"`
}

// The adapter never retries, follows redirects, or consults mutable process
// configuration. The endpoint is the validated pin from the invocation row.
func invokeCounting(ctx context.Context, b binding, invocationID string, input Input) (countEvidence, error) {
	if err := validBinding(b); err != nil {
		return countEvidence{}, err
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy: nil, DialContext: (&net.Dialer{Timeout: 2 * time.Second}).DialContext,
			DisableKeepAlives: true,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	requestBody := jsonBytes(map[string]any{"invocation_id": invocationID, "counter": input.Counter})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.Endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return countEvidence{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return countEvidence{}, err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 65536))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return countEvidence{}, fmt.Errorf("counting adapter returned status %d", response.StatusCode)
	}
	// A successful POST is only an observation. The non-mutating evidence read
	// can verify that exactly one increment was associated with this invocation.
	evidence := countEvidence{InvocationID: invocationID, Counter: input.Counter}
	u, err := url.Parse(b.Endpoint)
	if err != nil {
		return evidence, nil
	}
	query := u.Query()
	query.Set("invocation_id", invocationID)
	u.RawQuery = query.Encode()
	readReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return evidence, nil
	}
	readResponse, err := client.Do(readReq)
	if err != nil {
		return evidence, nil
	}
	defer readResponse.Body.Close()
	if readResponse.StatusCode != http.StatusOK {
		return evidence, nil
	}
	var readback countEvidence
	if err := json.NewDecoder(io.LimitReader(readResponse.Body, 65536)).Decode(&readback); err != nil {
		return evidence, nil
	}
	if readback.InvocationID != invocationID || readback.Counter != input.Counter || readback.ObservedCount < 0 {
		return evidence, nil
	}
	return readback, nil
}
