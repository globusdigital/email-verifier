package emailverifier

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Gravatar is detail about the Gravatar
type Gravatar struct {
	HasGravatar bool `json:"has_gravatar"` // whether it has gravatar
	// GravatarUrl keeps its non-idiomatic spelling because it is an exported
	// field: renaming it to GravatarURL would break every caller.
	GravatarUrl string `json:"gravatar_url"` //nolint:staticcheck // ST1003
}

// CheckGravatar will return the Gravatar records for the given email.
// Might return nil,nil on success when gravatar is disabled.
func (v *Verifier) CheckGravatar(ctx context.Context, email string) (*Gravatar, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	emailMd5, err := getMD5Hash(trimLower(email))
	if err != nil {
		return nil, err
	}
	gravatarURL := gravatarBaseURL + emailMd5 + "?d=404"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gravatarURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// check body
	md5Body, err := getMD5Hash(string(body))
	if err != nil {
		return nil, err
	}
	if md5Body == gravatarDefaultMd5 || resp.StatusCode != http.StatusOK {
		return &Gravatar{}, nil
	}
	return &Gravatar{
		HasGravatar: true,
		GravatarUrl: gravatarURL,
	}, nil
}
