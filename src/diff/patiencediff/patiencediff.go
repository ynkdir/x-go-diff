// Patience Diff Advantages
// http://bramcohen.livejournal.com/73318.html
//
// Patience Diff, a brief summary
// http://alfedenzo.livejournal.com/170301.html
//
// 1. Match the first lines of both if they're identical, then match the
//    second, third, etc. until a pair doesn't match.
// 2. Match the last lines of both if they're identical, then match the next to
//    last, second to last, etc. until a pair doesn't match.
// 3. Find all lines which occur exactly once on both sides, then do longest
//    common subsequence on those lines, matching them up.
// 4. Do steps 1-2 on each section between matched lines

package patiencediff

import (
	"github.com/hattya/go.diff"
	"sort"
)

type Record struct {
	aline  int
	acount int
	bline  int
	bcount int
	prev   *Record
}

type ByAline []*Record

func (a ByAline) Len() int           { return len(a) }
func (a ByAline) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAline) Less(i, j int) bool { return a[i].aline < a[j].aline }

func Strings(al []string, bl []string) []diff.Change {
	return patience_diff(al, 0, len(al), bl, 0, len(bl))
}

func patience_diff(al []string, astart int, aend int, bl []string, bstart int, bend int) []diff.Change {
	for astart < aend && bstart < bend && al[astart] == bl[bstart] {
		astart++
		bstart++
	}
	for astart < aend && bstart < bend && al[aend-1] == bl[bend-1] {
		aend--
		bend--
	}
	if astart == aend && bstart == bend {
		return []diff.Change{}
	} else if astart == aend || bstart == bend {
		return []diff.Change{diff.Change{A: astart, B: bstart, Del: aend - astart, Ins: bend - bstart}}
	}
	ul := find_all_unique_common_lines(al, astart, aend, bl, bstart, bend)
	if len(ul) == 0 {
		return fallback_diff(al, astart, aend, bl, bstart, bend)
	}
	lcs := find_longest_common_subsequence(ul)
	cl := []diff.Change{}
	for _, r := range lcs {
		subcl := patience_diff(al, astart, r.aline, bl, bstart, r.bline)
		cl = append(cl, subcl...)
		astart = r.aline + 1
		bstart = r.bline + 1
	}
	if astart < aend || bstart < bend {
		subcl := patience_diff(al, astart, aend, bl, bstart, bend)
		cl = append(cl, subcl...)
	}
	return cl
}

func fallback_diff(al []string, astart int, aend int, bl []string, bstart int, bend int) []diff.Change {
	cl := diff.Strings(al[astart:aend], bl[bstart:bend])
	for i, c := range cl {
		c.A += astart
		c.B += bstart
		cl[i] = c
	}
	return cl
}

func find_all_unique_common_lines(al []string, astart int, aend int, bl []string, bstart int, bend int) []*Record {
	rm := map[string]*Record{}
	for a := astart; a < aend; a++ {
		if _, ok := rm[al[a]]; ok {
			rm[al[a]].acount++
		} else {
			rm[al[a]] = &Record{aline: a, acount: 1}
		}
	}
	for b := bstart; b < bend; b++ {
		if _, ok := rm[bl[b]]; ok && rm[bl[b]].acount == 1 {
			rm[bl[b]].bline = b
			rm[bl[b]].bcount++
		}
	}
	rl := []*Record{}
	for _, r := range rm {
		if r.acount == 1 && r.bcount == 1 {
			rl = append(rl, r)
		}
	}
	return rl
}

func find_longest_common_subsequence(rl []*Record) []*Record {
	// Either is fine.
	// sort.Sort(ByBline(rl))
	sort.Sort(ByAline(rl))
	piles := []*Record{}
	for _, r := range rl {
		i := sort.Search(len(piles), func(i int) bool {
			// return piles[i].aline > r.aline
			return piles[i].bline > r.bline
		})
		if i != 0 {
			r.prev = piles[i-1]
		}
		if i < len(piles) {
			piles[i] = r
		} else {
			piles = append(piles, r)
		}
	}
	lcs := make([]*Record, len(piles))
	i := len(piles) - 1
	for r := piles[i]; r != nil; r = r.prev {
		lcs[i] = r
		i--
	}
	return lcs
}
