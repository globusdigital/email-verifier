package emailverifier

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckMxOK(t *testing.T) {
	domain := "github.com"

	mx, err := verifier.CheckMX(context.Background(), domain)
	assert.NoError(t, err)
	assert.True(t, mx.HasMXRecord)
}

func TestCheckNoMxOK(t *testing.T) {
	domain := "githubexists.com"

	mx, err := verifier.CheckMX(context.Background(), domain)
	assert.Nil(t, mx)
	assert.Error(t, err, ErrNoSuchHost)
}
