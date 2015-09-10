// https://github.com/eclipse/jgit/blob/master/org.eclipse.jgit/src/org/eclipse/jgit/diff/HistogramDiff.java
// https://github.com/git/git/blob/master/xdiff/xhistogram.c

/*
 * Copyright (C) 2010, Google Inc.
 * and other copyright owners as documented in the project's IP log.
 *
 * This program and the accompanying materials are made available
 * under the terms of the Eclipse Distribution License v1.0 which
 * accompanies this distribution, is reproduced below, and is
 * available at http://www.eclipse.org/org/documents/edl-v10.php
 *
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or
 * without modification, are permitted provided that the following
 * conditions are met:
 *
 * - Redistributions of source code must retain the above copyright
 *   notice, this list of conditions and the following disclaimer.
 *
 * - Redistributions in binary form must reproduce the above
 *   copyright notice, this list of conditions and the following
 *   disclaimer in the documentation and/or other materials provided
 *   with the distribution.
 *
 * - Neither the name of the Eclipse Foundation, Inc. nor the
 *   names of its contributors may be used to endorse or promote
 *   products derived from this software without specific prior
 *   written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
 * CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
 * INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
 * OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR
 * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
 * NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
 * STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
 * ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

/**
 * An extended form of Bram Cohen's patience diff algorithm.
 * <p>
 * This implementation was derived by using the 4 rules that are outlined in
 * Bram Cohen's <a href="http://bramcohen.livejournal.com/73318.html">blog</a>,
 * and then was further extended to support low-occurrence common elements.
 * <p>
 * The basic idea of the algorithm is to create a histogram of occurrences for
 * each element of sequence A. Each element of sequence B is then considered in
 * turn. If the element also exists in sequence A, and has a lower occurrence
 * count, the positions are considered as a candidate for the longest common
 * subsequence (LCS). After scanning of B is complete the LCS that has the
 * lowest number of occurrences is chosen as a split point. The region is split
 * around the LCS, and the algorithm is recursively applied to the sections
 * before and after the LCS.
 * <p>
 * By always selecting a LCS position with the lowest occurrence count, this
 * algorithm behaves exactly like Bram Cohen's patience diff whenever there is a
 * unique common element available between the two sequences. When no unique
 * elements exist, the lowest occurrence element is chosen instead. This offers
 * more readable diffs than simply falling back on the standard Myers' O(ND)
 * algorithm would produce.
 * <p>
 * To prevent the algorithm from having an O(N^2) running time, an upper limit
 * on the number of unique elements in a histogram bucket is configured by
 * {@link #setMaxChainLength(int)}. If sequence A has more than this many
 * elements that hash into the same hash bucket, the algorithm passes the region
 * to {@link #setFallbackAlgorithm(DiffAlgorithm)}. If no fallback algorithm is
 * configured, the region is emitted as a replace edit.
 * <p>
 * During scanning of sequence B, any element of A that occurs more than
 * {@link #setMaxChainLength(int)} times is never considered for an LCS match
 * position, even if it is common between the two sequences. This limits the
 * number of locations in sequence A that must be considered to find the LCS,
 * and helps maintain a lower running time bound.
 * <p>
 * So long as {@link #setMaxChainLength(int)} is a small constant (such as 64),
 * the algorithm runs in O(N * D) time, where N is the sum of the input lengths
 * and D is the number of edits in the resulting EditList. If the supplied
 * {@link SequenceComparator} has a good hash function, this implementation
 * typically out-performs {@link MyersDiff}, even though its theoretical running
 * time is the same.
 * <p>
 * This implementation has an internal limitation that prevents it from handling
 * sequences with more than 268,435,456 (2^28) elements.
 */

package histogramdiff

import (
	"github.com/hattya/go.diff"
)

const MAX_OCCURRENCE = 64

type HistIndex struct {
	rm         map[string]*Record
	count      int
	lcs        Region
	has_common bool
	has_lcs    bool
}

type Record struct {
	lines []int
}

type Region struct {
	astart int
	aend   int
	bstart int
	bend   int
}

func Strings(al []string, bl []string) []diff.Change {
	return histogram_diff(al, 0, len(al), bl, 0, len(bl))
}

func histogram_diff(al []string, astart int, aend int, bl []string, bstart int, bend int) []diff.Change {
	if astart == aend && bstart == bend {
		return []diff.Change{}
	} else if astart == aend || bstart == bend {
		return []diff.Change{diff.Change{A: astart, B: bstart, Del: aend - astart, Ins: bend - bstart}}
	}
	index := HistIndex{
		rm:    map[string]*Record{},
		count: MAX_OCCURRENCE,
	}
	find_lcs(&index, al, astart, aend, bl, bstart, bend)
	if !index.has_lcs {
		if index.has_common {
			return fallback_diff(al, astart, aend, bl, bstart, bend)
		} else {
			return []diff.Change{diff.Change{A: astart, B: bstart, Del: aend - astart, Ins: bend - bstart}}
		}
	}
	cl := []diff.Change{}
	subcl := histogram_diff(al, astart, index.lcs.astart, bl, bstart, index.lcs.bstart)
	cl = append(cl, subcl...)
	subcl = histogram_diff(al, index.lcs.aend, aend, bl, index.lcs.bend, bend)
	cl = append(cl, subcl...)
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

func find_lcs(index *HistIndex, al []string, astart int, aend int, bl []string, bstart int, bend int) {
	scanA(index, al, astart, aend)
	for b := bstart; b < bend; {
		b = try_lcs(index, b, al, astart, aend, bl, bstart, bend)
	}
}

func scanA(index *HistIndex, al []string, astart int, aend int) {
	for a := astart; a < aend; a++ {
		if _, ok := index.rm[al[a]]; ok {
			index.rm[al[a]].lines = append(index.rm[al[a]].lines, a)
		} else {
			index.rm[al[a]] = &Record{lines: []int{a}}
		}
	}
}

func try_lcs(index *HistIndex, b int, al []string, astart int, aend int, bl []string, bstart int, bend int) int {
	b_next := b + 1
	r, ok := index.rm[bl[b]]
	if !ok {
		return b_next
	}
	index.has_common = true
	if len(r.lines) > index.count {
		return b_next
	}
	prev_ae := 0
	for _, a := range r.lines {
		if a < prev_ae {
			continue
		}
		as := a
		ae := a + 1
		bs := b
		be := b + 1
		rc := len(r.lines)
		for astart < as && bstart < bs && al[as-1] == bl[bs-1] {
			as--
			bs--
			if len(index.rm[al[as]].lines) < rc {
				rc = len(index.rm[al[as]].lines)
			}
		}
		for ae < aend && be < bend && al[ae] == bl[be] {
			ae++
			be++
			if len(index.rm[al[ae-1]].lines) < rc {
				rc = len(index.rm[al[ae-1]].lines)
			}
		}
		if b_next < be {
			b_next = be
		}
		if index.lcs.aend-index.lcs.astart < ae-as || rc < index.count {
			index.lcs.astart = as
			index.lcs.aend = ae
			index.lcs.bstart = bs
			index.lcs.bend = be
			index.count = rc
			index.has_lcs = true
		}
		prev_ae = ae
	}
	return b_next
}
