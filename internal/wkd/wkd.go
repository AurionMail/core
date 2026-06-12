package wkd

import (
    "context"
    "crypto/sha1"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/tv42/zbase32"
)

func WKDHash(email string) string {
    normalized := strings.ToLower(email)
    sum := sha1.Sum([]byte(normalized))
    return zbase32.EncodeToString(sum[:])
}

// LookupWKD returns the armored public key for an email if WKD is available.
// If no key is found, it returns "" and no error.
func LookupWKD(email string) (string, error) {
    email = strings.ToLower(strings.TrimSpace(email))

    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "", fmt.Errorf("invalid email")
    }

    local := parts[0]
    domain := parts[1]

    // WKD hash = SHA1(local-part) → z-base-32
    sum := sha1.Sum([]byte(local))
    hu := zbase32.EncodeToString(sum[:])

    url := fmt.Sprintf(
        "https://%s/.well-known/openpgpkey/hu/%s?l=%s",
        domain,
        hu,
        local,
    )

    // --- Timeout / Deadline ---
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return "", err
    }

    client := &http.Client{
        Timeout: 4 * time.Second, // hard timeout
    }

    resp, err := client.Do(req)
    if err != nil {
        // Timeout or network error → treat as "no key"
        return "", nil
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return "", nil
    }

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(data), nil
}
