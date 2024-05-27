package emailverifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopLevelDomainExists(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{
			name: "ch exists",
			args: "example.ch",
			want: true,
		},
		{
			name: "con not exists",
			args: "example.con",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, TopLevelDomainExists(tt.args), "TopLevelDomainExists(%v)", tt.args)
		})
	}
}
