package emailverifier

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Gravatar is detail about the Gravatar
type Gravatar struct {
	HasGravatar bool   `json:"has_gravatar"` // whether it has gravatar
	GravatarUrl string `json:"gravatar_url"` // gravatar url
}

// CheckGravatar will return the Gravatar records for the given email.
// Might return nil,nil on success when gravatar is disabled.
func (v *Verifier) CheckGravatar(ctx context.Context, email string) (*Gravatar, error) {
	if !v.gravatarCheckEnabled {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	emailMd5, err := getMD5Hash(trimLower(email))
	if err != nil {
		return nil, err
	}
	gravatarUrl := gravatarBaseUrl + emailMd5 + "?d=404"
	req, err := http.NewRequest("GET", gravatarUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
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
	if md5Body == gravatarDefaultMd5 || resp.StatusCode != 200 {
		return &Gravatar{}, nil
	}
	return &Gravatar{
		HasGravatar: true,
		GravatarUrl: gravatarUrl,
	}, nil
}
