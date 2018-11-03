package module

// SortByPath implements sort.Interface to sort a slice of Module by path.
type SortByPath []Module

func (s SortByPath) Len() int           { return len(s) }
func (s SortByPath) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortByPath) Less(i, j int) bool { return s[i].Path < s[j].Path }
