package emailverifier

import (
	"strings"
	"sync"
)

var (
	disposableSyncDomains sync.Map // concurrent safe map to store disposable domains data

	// additional list of disposable domains set via users of this library.
	// Concurrent safe for the same reason as the map above: it is written by
	// AddDisposableDomains, which callers reach from whatever goroutine they
	// like, and read by updateDisposableDomains on the schedule goroutine that
	// EnableAutoUpdateDisposable starts.
	additionalDisposableDomains sync.Map

	blockedExchangeDomains sync.Map // concurrent safe map to store blocked mail exchanger/mx target domains data
)

// IsRoleAccount checks if username is a role-based account
func (v *Verifier) IsRoleAccount(username string) bool {
	return roleAccounts[strings.ToLower(username)]
}

// IsFreeDomain checks if domain is a free domain
func (v *Verifier) IsFreeDomain(domain string) bool {
	return freeDomains[domain]
}

// IsDisposable checks if domain is a disposable domain
func (v *Verifier) IsDisposable(domain string) bool {
	domain = domainToASCII(domain)
	_, found := disposableSyncDomains.Load(domain)
	return found
}

// IsBlockedExchangeDomain checks if domain in the ANSWER SECTION of a MX query is blocked.
func (v *Verifier) IsBlockedExchangeDomain(domain string) bool {
	domain = domainToASCII(domain)
	_, found := blockedExchangeDomains.Load(domain)
	return found
}
