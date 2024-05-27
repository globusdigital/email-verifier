package emailverifier

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckEmailOK_SMTPHostNotExists(t *testing.T) {
	const (
		// trueVal  = true
		username = "email_username"
		domain   = "domainnotexists.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(context.Background(), email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: false,
		Disposable:   false,
		RoleAccount:  false,
		Reachable:    reachableNo,
		Free:         false,
		SMTP:         nil,
		TLDExists:    true,
	}
	assert.ErrorContains(t, err, ErrNoSuchHost)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmailOK_SMTPHostExists_NotCatchAll(t *testing.T) {
	const (
		// trueVal  = true
		username = "email_username"
		domain   = "github.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(context.Background(), email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Reachable:    reachableUnknown,
		Disposable:   false,
		RoleAccount:  false,
		Free:         false,
		TLDExists:    true,
		SMTP: &SMTP{
			HostExists:  true,
			FullInbox:   false,
			CatchAll:    true,
			Deliverable: false,
			Disabled:    false,
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmailOK_SMTPHostExists_FreeDomain(t *testing.T) {
	const (
		// trueVal  = true
		username = "email_username"
		domain   = "gmail.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(context.Background(), email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Reachable:    reachableNo,
		Disposable:   false,
		RoleAccount:  false,
		Free:         true,
		SMTP: &SMTP{
			HostExists:  true,
			FullInbox:   false,
			CatchAll:    false,
			Deliverable: false,
			Disabled:    false,
		},
		TLDExists: true,
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_ErrorSyntax(t *testing.T) {
	const (
		// trueVal  = true
		username = ""
		domain   = "yahoo.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(context.Background(), email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   "",
			Valid:    false,
		},
		HasMxRecords: false,
		Reachable:    reachableUnknown,
		Disposable:   false,
		RoleAccount:  false,
		Free:         false,
		SMTP:         nil,
		TLDExists:    false,
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_Disposable(t *testing.T) {
	const (
		// trueVal  = true
		username = "exampleuser"
		domain   = "zzjbfwqi.shop"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(context.Background(), email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: false,
		Reachable:    reachableUnknown,
		Disposable:   true,
		RoleAccount:  false,
		Free:         false,
		SMTP:         nil,
		TLDExists:    true,
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_Disposable_override(t *testing.T) {
	const (
		username = "exampleuser"
		domain   = "iamdisposableemail.de"
		address  = username + "@" + domain
		email    = address
	)

	verifier := NewVerifier().EnableSMTPCheck().AddDisposableDomains([]string{"iamdisposableemail.de"})
	ret, err := verifier.Verify(context.Background(), email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: false,
		Reachable:    reachableUnknown,
		Disposable:   true,
		RoleAccount:  false,
		Free:         false,
		SMTP:         nil,
		TLDExists:    true,
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_TLD_NotExists(t *testing.T) {
	const (
		username = "exampleuser"
		domain   = "iamdisposableemail.testing"
		address  = username + "@" + domain
		email    = address
	)

	verifier := NewVerifier().DisableMXCheck().DisableSMTPCheck()
	ret, err := verifier.Verify(context.Background(), email)
	assert.Nil(t, ret)
	assert.EqualError(t, err, "TLD domain \"iamdisposableemail.testing\" does not exist")
}

func TestCheckEmail_Concurrency(t *testing.T) {
	const (
		username = "exampleuser"
		domain   = "microsoft.com"
		address  = username + "@" + domain
		email    = address
	)

	verifier := NewVerifier().EnableGravatarCheck().EnableMXCheck().EnableSMTPCheck()

	ret, err := verifier.Verify(context.Background(), email)
	assert.NoError(t, err)
	assert.NotEmpty(t, ret)
}

func TestCheckEmail_RoleAccount(t *testing.T) {
	const (
		// trueVal  = true
		username = "admin"
		domain   = "github.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(context.Background(), email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Reachable:    reachableUnknown,
		Disposable:   false,
		RoleAccount:  true,
		Free:         false,
		SMTP: &SMTP{
			HostExists:  true,
			FullInbox:   false,
			CatchAll:    true,
			Deliverable: false,
			Disabled:    false,
		},
		TLDExists: true,
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_DisabledSMTPCheck(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "randomain.com"
		address  = username + "@" + domain
		email    = address
	)

	verifier.DisableSMTPCheck()
	ret, err := verifier.Verify(context.Background(), email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Disposable:   false,
		RoleAccount:  false,
		Reachable:    reachableUnknown,
		Free:         false,
		SMTP:         nil,
		TLDExists:    true,
	}
	verifier.EnableSMTPCheck()
	assert.NoError(t, err)
	assert.Equal(t, &expected, ret)
}

func TestNewVerifierOK_AutoUpdateDisposable(t *testing.T) {
	verifier.EnableAutoUpdateDisposable()
}

func TestNewVerifierOK_EnableAutoUpdateDisposable(t *testing.T) {
	verifier.EnableAutoUpdateDisposable()
}

func TestNewVerifierOK_AutoUpdateDisposableDuplicate(t *testing.T) {
	verifier.DisableAutoUpdateDisposable()

	verifier.EnableAutoUpdateDisposable()
	verifier.DisableAutoUpdateDisposable()

	verifier.EnableAutoUpdateDisposable()
	verifier.DisableAutoUpdateDisposable()
	verifier.EnableAutoUpdateDisposable()
}

func TestStopCurrentSchedule_ScheduleIsNil(t *testing.T) {
	verifier.schedule = nil
	verifier.stopCurrentSchedule()
}

func TestStopCurrentScheduleOK(t *testing.T) {
	verifier.EnableAutoUpdateDisposable()
	verifier.stopCurrentSchedule()
}

func TestCheckEmail_EnableDomainSuggest(t *testing.T) {
	const (
		// trueVal  = true
		username = "email_username"
		domain   = "hotmail.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, _ := verifier.Verify(context.Background(), email)

	assert.Empty(t, ret.Suggestion)
}

func TestCheckEmail_EnableDomainSuggest_Gmail(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "gmai.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, _ := verifier.EnableDomainSuggest().Verify(email)

	assert.Equal(t, "gmail.com", ret.Suggestion)
}
