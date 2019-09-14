package main

import (
	"sort"
)

type email struct {
	size    int64
	gmailID string
	date    string // retrieved from message header
	snippet string
}

type messageSorter struct {
	msg  []email
	less func(i, j email) bool
}

func sortBySize(msg []email) {
	sort.Sort(messageSorter{msg, func(i, j email) bool {
		return i.size > j.size
	}})
}

func (s messageSorter) Len() int {
	return len(s.msg)
}

func (s messageSorter) Swap(i, j int) {
	s.msg[i], s.msg[j] = s.msg[j], s.msg[i]
}

func (s messageSorter) Less(i, j int) bool {
	return s.less(s.msg[i], s.msg[j])
}
