package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Identity holds the verified user info from Bastion.
type Identity struct {
	Sub         string `json:"sub"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

// Validator calls the Bastion auth server to validate a session cookie.
type Validator struct {
	bastionURL string
	client     *http.Client
}

func NewValidator(bastionURL string) *Validator {
	return &Validator{
		bastionURL: bastionURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

// Validate calls auth.shadovx.me/session/me with the session cookie and returns
// the verified identity. Returns an error if the session is invalid or expired.
func (v *Validator) Validate(sessionCookie string) (*Identity, error) {
	req, err := http.NewRequest("GET", v.bastionURL+"/session/me", nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: "auth_session", Value: sessionCookie})

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bastion unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid session (status %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var id Identity
	if err := json.Unmarshal(body, &id); err != nil {
		return nil, err
	}
	if id.Sub == "" {
		return nil, fmt.Errorf("empty subject in session response")
	}
	return &id, nil
}
