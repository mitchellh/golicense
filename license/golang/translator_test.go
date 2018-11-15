package golang

import (
	"context"
	"testing"

	"github.com/mitchellh/golicense/module"
	"github.com/stretchr/testify/require"
)

func TestTranslator(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"github.com/foo/bar",
			"",
		},

		{
			"golang.org/x/crypto",
			"github.com/golang/crypto",
		},
	}

	for _, tt := range cases {
		t.Run(tt.Input, func(t *testing.T) {
			var tr Translator
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