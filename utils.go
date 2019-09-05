package main

import "sort"

func sortByValue(m map[string]uint64) PairList {
	pairList := make(PairList, 0, len(m))

	for k, v := range m {
		pairList = append(pairList, Pair{k, v})
	}

	sort.Sort(sort.Reverse(pairList))

	return pairList
}

type Pair struct {
	key string
	val uint64
}
type PairList []Pair

func (p PairList) Len() int {
	return len(p)
}

func (p PairList) Less(i, j int) bool { return p[i].val < p[j].val }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
