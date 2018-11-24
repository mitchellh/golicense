package module

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseExeData(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected []Module
		Error    string
	}{
		{
			"typical (from golicense itself)",
			testExeData,
			[]Module{
				Module{
					Path:    "github.com/fatih/color",
					Version: "v1.7.0",
					Hash:    "h1:DkWD4oS2D8LGGgTQ6IvwJJXSL5Vp2ffcQg58nFV38Ys=",
				},
				Module{
					Path:    "github.com/mattn/go-colorable",
					Version: "v0.0.9",
					Hash:    "h1:UVL0vNpWh04HeJXV0KLcaT7r06gOH2l4OW6ddYRUIY4=",
				},
				Module{
					Path:    "github.com/mattn/go-isatty",
					Version: "v0.0.4",
					Hash:    "h1:bnP0vzxcAdeI1zdubAl5PjU6zsERjGZb7raWodagDYs=",
				},
				Module{
					Path:    "github.com/rsc/goversion",
					Version: "v1.2.0",
					Hash:    "h1:zVF4y5ciA/rw779S62bEAq4Yif1cBc/UwRkXJ2xZyT4=",
				},
				Module{
					Path:    "github.com/rsc/goversion",
					Version: "v12.0.0",
					Hash:    "h1:zVF4y5ciA/rw779S62bEAq4Yif1cBc/UwRkXJ2xZyT4=",
				},
			},
			"",
		},

		{
			"replacement syntax",
			strings.TrimSpace(replacement),
			[]Module{
				Module{
					Path:    "github.com/markbates/inflect",
					Version: "v0.0.0-20171215194931-a12c3aec81a6",
					Hash:    "h1:LZhVjIISSbj8qLf2qDPP0D8z0uvOWAW5C85ly5mJW6c=",
				},
			},
			"",
		},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			require := require.New(t)
			actual, err := ParseExeData(tt.Input)
			if tt.Error != "" {
				require.Error(err)
				require.Contains(err.Error(), tt.Error)
				return
			}
			require.NoError(err)
			require.Equal(tt.Expected, actual)
		})
	}
}

const testExeData = "path\tgithub.com/mitchellh/golicense\nmod\tgithub.com/mitchellh/golicense\t(devel)\t\ndep\tgithub.com/fatih/color\tv1.7.0\th1:DkWD4oS2D8LGGgTQ6IvwJJXSL5Vp2ffcQg58nFV38Ys=\ndep\tgithub.com/mattn/go-colorable\tv0.0.9\th1:UVL0vNpWh04HeJXV0KLcaT7r06gOH2l4OW6ddYRUIY4=\ndep\tgithub.com/mattn/go-isatty\tv0.0.4\th1:bnP0vzxcAdeI1zdubAl5PjU6zsERjGZb7raWodagDYs=\ndep\tgithub.com/rsc/goversion\tv1.2.0\th1:zVF4y5ciA/rw779S62bEAq4Yif1cBc/UwRkXJ2xZyT4=\ndep\tgithub.com/rsc/goversion/v12\tv12.0.0\th1:zVF4y5ciA/rw779S62bEAq4Yif1cBc/UwRkXJ2xZyT4=\n"

const replacement = `
path	github.com/gohugoio/hugo
mod	github.com/gohugoio/hugo	(devel)
dep	github.com/markbates/inflect	v1.0.0
=>	github.com/markbates/inflect	v0.0.0-20171215194931-a12c3aec81a6	h1:LZhVjIISSbj8qLf2qDPP0D8z0uvOWAW5C85ly5mJW6c=
`
