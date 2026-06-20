package mail

import (
	"context"
	"net/http"
	"time"
)

type JMAPBackend struct {
	SessionURL string
}

func NewJMAPBackend(url string) *JMAPBackend {
	return &JMAPBackend{SessionURL: url}
}

func (b *JMAPBackend) VerifyCredentials(ctx context.Context, email, password string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", b.SessionURL, nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(email, password)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil
}