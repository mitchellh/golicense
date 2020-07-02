package mapper

import (
	"context"
	"testing"

	"github.com/cloudentity/golicense/module"
	"github.com/stretchr/testify/require"
)

func TestTranslator(t *testing.T) {
	cases := []struct {
		Map    map[string]string
		Input  string
		Output string
	}{
		{
			nil,
			"github.com/foo/bar",
			"",
		},

		{
			map[string]string{
				"gopkg.in/pkg.v3": "github.com/go-pkg/pkg",
			},
			"gopkg.in/pkg.v3",
			"github.com/go-pkg/pkg",
		},

		{
			map[string]string{
				`/^gopkg\.in/([^/]+)/([^/]+)\./`: `github.com/\1/\2`,
			},
			"gopkg.in/cloudentity/foo.v22",
			"github.com/cloudentity/foo",
		},
	}

	for _, tt := range cases {
		t.Run(tt.Input, func(t *testing.T) {
			tr := &Translator{Map: tt.Map}
			actual, ok := tr.Translate(context.Background(), module.Module{
				Path: tt.Input,
			})

			if tt.Output == "" {
				require.False(t, ok)
				return
			}

			require.True(t, ok)
			require.Equal(t, tt.Output, actual.Path)
		})
	}
}
