package emailverifier

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type MXCacheResult struct {
	LookupError error
	MX          []*net.MX
}

type MXCacher interface {
	Get(key string, opts ...any) *MXCacheResult
	GetOrSet(key string, value MXCacheResult, ttl time.Duration) (*MXCacheResult, bool)
}

// Mx is detail about the Mx host
type Mx struct {
	HasMXRecord      bool      // whether has 1 or more MX record
	Records          []*net.MX // represent DNS MX records
	BlockedMXTargets []string  // contains a list of blocked MX target domains
}

// CheckMX will return the DNS MX records for the given domain name sorted by preference.
func (v *Verifier) CheckMX(ctx context.Context, domain string) (*Mx, error) {
	// A trailing dot marks a fully qualified name ("example.com.") and is
	// valid in DNS; drop it so the TLD check sees the real TLD.
	domain = strings.TrimSuffix(domainToASCII(domain), ".")

	if !TopLevelDomainExists(domain) {
		return nil, fmt.Errorf("TLD domain %q does not exist", domain)
	}

	if !v.mxCheckEnabled {
		return &Mx{}, nil
	}

	mxs, err := v.lookupMX(ctx, domain)
	if err != nil && len(mxs) == 0 {
		return nil, err
	}

	ret := &Mx{
		HasMXRecord: len(mxs) > 0,
		Records:     mxs,
	}

	for _, mx := range mxs {
		host := strings.TrimSuffix(strings.ToLower(mx.Host), ".")
		if d, ok := v.matchDomain(host); ok {
			ret.BlockedMXTargets = append(ret.BlockedMXTargets, d)
		}
	}

	return ret, nil
}

// lookupMX resolves the MX records of domain, consulting v.mxCache first when
// one is configured. Successful answers and authoritative "not found" answers
// are cached for v.mxCacheTTL; transient failures (timeouts, SERVFAIL, dial
// errors, context cancellation) are not, so the next call retries the lookup.
// Callers always receive their own copy of the records, so mutating them never
// affects the cached entry.
func (v *Verifier) lookupMX(ctx context.Context, domain string) ([]*net.MX, error) {
	if v.mxCache == nil {
		return v.dnsResolver().LookupMX(ctx, domain)
	}

	key := strings.TrimSuffix(strings.ToLower(domain), ".")
	if r := v.mxCache.Get(key); r != nil {
		return cloneMXs(r.MX), r.LookupError
	}

	mxs, err := v.dnsResolver().LookupMX(ctx, domain)
	if !isCacheableMXError(err) {
		return mxs, err
	}

	// GetOrSet keeps the entry of a concurrent caller that stored first, so
	// all callers racing on the same domain observe the same answer.
	r, _ := v.mxCache.GetOrSet(key, MXCacheResult{LookupError: err, MX: cloneMXs(mxs)}, v.mxCacheTTL)
	if r == nil {
		return mxs, err
	}
	return cloneMXs(r.MX), r.LookupError
}

// isCacheableMXError reports whether a LookupMX outcome is stable enough to be
// cached: a successful lookup, or a DNS answer stating the name or its MX
// records do not exist.
func isCacheableMXError(err error) bool {
	if err == nil {
		return true
	}
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr) && dnsErr.IsNotFound
}

func cloneMXs(mxs []*net.MX) []*net.MX {
	if mxs == nil {
		return nil
	}
	out := make([]*net.MX, len(mxs))
	for i, mx := range mxs {
		if mx != nil {
			c := *mx
			out[i] = &c
		}
	}
	return out
}

// matchDomain returns the entry of list that equals domain or one of its parent domains.
func (v *Verifier) matchDomain(domain string) (string, bool) {
	for d := domain; d != ""; {
		if v.IsBlockedExchangeDomain(d) {
			return d, true
		}
		dot := strings.IndexByte(d, '.')
		if dot < 0 {
			break
		}
		d = d[dot+1:]
	}
	return "", false
}
