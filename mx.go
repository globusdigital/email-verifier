package emailverifier

import (
	"context"
	"net"
)

// Mx is detail about the Mx host
type Mx struct {
	HasMXRecord bool      // whether has 1 or more MX record
	Records     []*net.MX // represent DNS MX records
}

// CheckMX will return the DNS MX records for the given domain name sorted by preference.
func (v *Verifier) CheckMX(ctx context.Context, domain string) (*Mx, error) {
	if !v.mxCheckEnabled {
		return &Mx{}, nil
	}

	domain = domainToASCII(domain)
	mx, err := net.DefaultResolver.LookupMX(ctx, domain)
	if err != nil && len(mx) == 0 {
		return nil, err
	}
	return &Mx{
		HasMXRecord: len(mx) > 0,
		Records:     mx,
	}, nil
}
