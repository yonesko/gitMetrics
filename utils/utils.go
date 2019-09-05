package utils

import "sort"

func SortByValue(m map[string]uint64) pairList {
	pairList := make(pairList, 0, len(m))

	for k, v := range m {
		pairList = append(pairList, Pair{k, v})
	}

	sort.Sort(sort.Reverse(pairList))

	return pairList
}

type Pair struct {
	Key string
	Val uint64
}
type pairList []Pair

func (p pairList) Len() int {
	return len(p)
}

func (p pairList) Less(i, j int) bool { return p[i].Val < p[j].Val }
func (p pairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
