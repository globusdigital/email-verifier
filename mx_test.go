package emailverifier

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckMxOK(t *testing.T) {
	domain := "github.com"

	mx, err := verifier.CheckMX(context.Background(), domain)
	require.NoError(t, err)
	assert.True(t, mx.HasMXRecord)
}

func TestCheckNoMxOK(t *testing.T) {
	domain := "githubexists.com"

	mx, err := verifier.CheckMX(context.Background(), domain)
	assert.Nil(t, mx)

	// CheckMX surfaces the resolver error as-is; it is Verify that maps it to
	// ErrNoSuchHost. Assert the DNS semantics rather than the message text,
	// which carries the resolver address and so varies by environment.
	var dnsErr *net.DNSError
	require.ErrorAs(t, err, &dnsErr)
	assert.True(t, dnsErr.IsNotFound, "expected NXDOMAIN, got %v", err)
}

func TestCheckMx_WithCustomResolver(t *testing.T) {
	domain := "github.com"
	wantErr := errors.New("custom resolver dial invoked")

	customResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, wantErr
		},
	}

	v := NewVerifier().Resolver(customResolver)
	mx, err := v.CheckMX(context.Background(), domain)
	assert.Nil(t, mx)
	assert.ErrorContains(t, err, wantErr.Error())
}

// fakeMXCache is a minimal in-memory MXCacher for tests.
type fakeMXCache struct {
	mu      sync.Mutex
	entries map[string]MXCacheResult
	ttls    map[string]time.Duration
}

func newFakeMXCache() *fakeMXCache {
	return &fakeMXCache{entries: map[string]MXCacheResult{}, ttls: map[string]time.Duration{}}
}

func (c *fakeMXCache) Get(key string, _ ...any) *MXCacheResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r, ok := c.entries[key]; ok {
		return &r
	}
	return nil
}

func (c *fakeMXCache) Set(key string, value MXCacheResult, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = value
	c.ttls[key] = ttl
}

func (c *fakeMXCache) GetOrSet(key string, value MXCacheResult, ttl time.Duration) (*MXCacheResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r, ok := c.entries[key]; ok {
		return &r, true
	}
	c.entries[key] = value
	c.ttls[key] = ttl
	return &value, false
}

// failingResolver fails every DNS query, proving a result came from the cache.
func failingResolver(t *testing.T) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			t.Error("resolver must not be called")
			return nil, errors.New("resolver must not be called")
		},
	}
}

func TestCheckMx_CacheHit(t *testing.T) {
	cache := newFakeMXCache()
	cache.Set("example.com", MXCacheResult{MX: []*net.MX{{Host: "mail.evilc0rp.org.", Pref: 10}}}, time.Hour)

	v := NewVerifier().Resolver(failingResolver(t)).WithMXCache(cache, 0).
		AddBlockedExchangeDomains([]string{"evilc0rp.org"})
	t.Cleanup(func() { blockedExchangeDomains.Delete("evilc0rp.org") })

	mx, err := v.CheckMX(context.Background(), "Example.COM.")
	require.NoError(t, err)
	assert.True(t, mx.HasMXRecord)
	require.Len(t, mx.Records, 1)
	assert.Equal(t, "mail.evilc0rp.org.", mx.Records[0].Host)
	assert.Equal(t, []string{"evilc0rp.org"}, mx.BlockedMXTargets)

	// The caller gets a copy: mutating it must not alter the cached entry.
	mx.Records[0].Host = "changed."
	assert.Equal(t, "mail.evilc0rp.org.", cache.Get("example.com").MX[0].Host)
}

func TestCheckMx_CacheHitNegative(t *testing.T) {
	cache := newFakeMXCache()
	notFound := &net.DNSError{Err: "no such host", Name: "example.com", IsNotFound: true}
	cache.Set("example.com", MXCacheResult{LookupError: notFound}, time.Hour)

	v := NewVerifier().Resolver(failingResolver(t)).WithMXCache(cache, 0)

	mx, err := v.CheckMX(context.Background(), "example.com")
	assert.Nil(t, mx)
	assert.ErrorIs(t, err, notFound)
}

func TestCheckMx_TransientErrorNotCached(t *testing.T) {
	cache := newFakeMXCache()
	v := NewVerifier().Resolver(&net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, errors.New("dial failed")
		},
	}).WithMXCache(cache, 0)

	_, err := v.CheckMX(context.Background(), "github.com")
	require.Error(t, err)
	assert.Nil(t, cache.Get("github.com"), "transient errors must not be cached")
}

func TestCheckMx_CacheStoresLookups(t *testing.T) {
	cache := newFakeMXCache()
	v := NewVerifier().WithMXCache(cache, 5*time.Minute)

	mx, err := v.CheckMX(context.Background(), "github.com")
	require.NoError(t, err)
	require.True(t, mx.HasMXRecord)
	r := cache.Get("github.com")
	require.NotNil(t, r)
	assert.NoError(t, r.LookupError)
	assert.Len(t, r.MX, len(mx.Records))
	assert.Equal(t, 5*time.Minute, cache.ttls["github.com"])

	_, err = v.CheckMX(context.Background(), "githubexists.com")
	require.Error(t, err)
	r = cache.Get("githubexists.com")
	require.NotNil(t, r, "NXDOMAIN must be cached")
	assert.Error(t, r.LookupError)
	assert.Equal(t, 5*time.Minute, cache.ttls["githubexists.com"])
}
