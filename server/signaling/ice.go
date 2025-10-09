package signaling

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username"`
	Credential string   `json:"credential"`
}

type ICEServersResponse struct {
	ICEServers []ICEServer `json:"iceServers"`
}

type GenerateICEServersRequest struct {
	TTL int `json:"ttl"`
}

// GenerateICEServers calls the Cloudflare TURN API to generate ICE server credentials
func GenerateICEServers(ctx context.Context, turnKeyID string, apiToken string, ttl int) (*ICEServersResponse, error) {
	url := fmt.Sprintf("https://rtc.live.cloudflare.com/v1/turn/keys/%s/credentials/generate-ice-servers", turnKeyID)

	reqBody := GenerateICEServersRequest{
		TTL: ttl,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var iceServers ICEServersResponse
	if err := json.NewDecoder(resp.Body).Decode(&iceServers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &iceServers, nil
}
