package emailverifier

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckGravatarOK(t *testing.T) {
	email := "alex@pagerduty.com"

	verifier := NewVerifier().EnableGravatarCheck()

	gravatar, err := verifier.CheckGravatar(context.Background(), email)
	require.NoError(t, err)
	assert.True(t, gravatar.HasGravatar)
	assert.NotEmpty(t, gravatar.GravatarUrl)
}

func TestCheckGravatarFailed(t *testing.T) {
	email := "MyemailaddressHasNoGravatar@example.com"
	verifier := NewVerifier().EnableGravatarCheck()
	gravatar, err := verifier.CheckGravatar(context.Background(), email)
	require.NoError(t, err)
	assert.False(t, gravatar.HasGravatar)
	assert.Empty(t, gravatar.GravatarUrl)
}
