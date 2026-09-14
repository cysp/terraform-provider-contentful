package provider

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

var errSigningSecretTestDetail = errors.New("rejected old+secret and new/secret; encoded old%2Bsecret; fragment old; unknown remote")

func TestSigningSecretErrorDetailRedactionPolicy(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		values []types.String
		want   string
	}{
		"known values": {
			values: []types.String{types.StringValue("old+secret"), types.StringValue("new/secret")},
			want:   "rejected *** and ***; encoded old%2Bsecret; fragment old; unknown remote",
		},
		"unavailable values": {
			values: []types.String{types.StringNull(), types.StringUnknown(), types.StringValue("")},
			want:   "rejected old+secret and new/secret; encoded old%2Bsecret; fragment old; unknown remote",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, signingSecretErrorDetail(nil, errSigningSecretTestDetail, test.values...))
		})
	}
}
