package jmap

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
)

type JMAPClient struct {
    url      string
    username string
    password string
    http     *http.Client
}

func NewJMAPClient(url, username, password string) *JMAPClient {
    return &JMAPClient{
        url:      url,
        username: username,
        password: password,
        http:     &http.Client{},
    }
}

func (c *JMAPClient) Call(ctx context.Context, payload interface{}) (map[string]interface{}, error) {
    body, _ := json.Marshal(payload)

    req, _ := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(body))
    req.SetBasicAuth(c.username, c.password)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    return result, nil
}
