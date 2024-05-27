package emailverifier

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckGravatarOK(t *testing.T) {
	email := "alex@pagerduty.com"

	verifier := NewVerifier().EnableGravatarCheck()

	gravatar, err := verifier.CheckGravatar(context.Background(), email)
	assert.NoError(t, err)
	assert.True(t, gravatar.HasGravatar)
	assert.NotEmpty(t, gravatar.GravatarUrl)
}

func TestCheckGravatarFailed(t *testing.T) {
	email := "MyemailaddressHasNoGravatar@example.com"
	verifier := NewVerifier().EnableGravatarCheck()
	gravatar, err := verifier.CheckGravatar(context.Background(), email)
	assert.NoError(t, err)
	assert.False(t, gravatar.HasGravatar)
	assert.Empty(t, gravatar.GravatarUrl)
}
