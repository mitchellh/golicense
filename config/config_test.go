package config

import (
	"testing"

	"github.com/cloudentity/golicense/license"
	"github.com/stretchr/testify/require"
)

func TestConfigAllowed(t *testing.T) {
	cases := []struct {
		Name   string
		Config *Config
		Lic    *license.License
		Result AllowState
	}{
		{
			"empty lists",
			&Config{},
			&license.License{Name: "FOO"},
			StateUnknown,
		},

		{
			"name allowed",
			&Config{
				Allow: []string{"FOO"},
			},
			&license.License{Name: "FOO"},
			StateAllowed,
		},

		{
			"name allowed and denied",
			&Config{
				Allow: []string{"FOO"},
				Deny:  []string{"FOO"},
			},
			&license.License{Name: "FOO"},
			StateDenied,
		},

		{
			"spdx allowed",
			&Config{
				Allow: []string{"FOO"},
			},
			&license.License{SPDX: "FOO"},
			StateAllowed,
		},

		{
			"spdx allowed and denied",
			&Config{
				Allow: []string{"FOO"},
				Deny:  []string{"FOO"},
			},
			&license.License{SPDX: "FOO"},
			StateDenied,
		},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			actual := tt.Config.Allowed(tt.Lic)
			require.Equal(t, actual, tt.Result)
		})
	}
}
