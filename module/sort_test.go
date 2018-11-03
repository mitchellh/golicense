package module

import (
	"sort"
	"testing"
)

func TestSort_interface(t *testing.T) {
	var _ sort.Interface = SortByPath(nil)
}
