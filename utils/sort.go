package utils

import "sort"

func SortByValue(m map[string]uint64) pairList {
	pairList := make(pairList, 0, len(m))

	for k, v := range m {
		pairList = append(pairList, pair{k, v})
	}

	sort.Sort(sort.Reverse(pairList))

	return pairList
}

type pair struct {
	Key string
	Val uint64
}
type pairList []pair

func (p pairList) Len() int {
	return len(p)
}

func (p pairList) Less(i, j int) bool { return p[i].Val < p[j].Val }
func (p pairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
