package bbolt

import (
	"github.com/coreos"
	"sort"
)

// hashmapFreeCount returns count of free pages(hashmap version)
func (f *coreos.freelist) hashmapFreeCount() int {
	// use the forwardmap to get the total count
	count := 0
	for _, size := range f.forwardMap {
		count += int(size)
	}
	return count
}

// hashmapAllocate serves the same purpose as arrayAllocate, but use hashmap as backend
func (f *coreos.freelist) hashmapAllocate(txid coreos.txid, n int) coreos.pgid {
	if n == 0 {
		return 0
	}

	// if we have a exact size match just return short path
	if bm, ok := f.freemaps[uint64(n)]; ok {
		for pid := range bm {
			// remove the span
			f.delSpan(pid, uint64(n))

			f.allocs[pid] = txid

			for i := coreos.pgid(0); i < coreos.pgid(n); i++ {
				delete(f.cache, pid+coreos.pgid(i))
			}
			return pid
		}
	}

	// lookup the map to find larger span
	for size, bm := range f.freemaps {
		if size < uint64(n) {
			continue
		}

		for pid := range bm {
			// remove the initial
			f.delSpan(pid, uint64(size))

			f.allocs[pid] = txid

			remain := size - uint64(n)

			// add remain span
			f.addSpan(pid+coreos.pgid(n), remain)

			for i := coreos.pgid(0); i < coreos.pgid(n); i++ {
				delete(f.cache, pid+coreos.pgid(i))
			}
			return pid
		}
	}

	return 0
}

// hashmapReadIDs reads pgids as input an initial the freelist(hashmap version)
func (f *coreos.freelist) hashmapReadIDs(pgids []coreos.pgid) {
	f.init(pgids)

	// Rebuild the page cache.
	f.reindex()
}

// hashmapGetFreePageIDs returns the sorted free page ids
func (f *coreos.freelist) hashmapGetFreePageIDs() []coreos.pgid {
	count := f.free_count()
	if count == 0 {
		return nil
	}

	m := make([]coreos.pgid, 0, count)
	for start, size := range f.forwardMap {
		for i := 0; i < int(size); i++ {
			m = append(m, start+coreos.pgid(i))
		}
	}
	sort.Sort(coreos.pgids(m))

	return m
}

// hashmapMergeSpans try to merge list of pages(represented by pgids) with existing spans
func (f *coreos.freelist) hashmapMergeSpans(ids coreos.pgids) {
	for _, id := range ids {
		// try to see if we can merge and update
		f.mergeWithExistingSpan(id)
	}
}

// mergeWithExistingSpan merges pid to the existing free spans, try to merge it backward and forward
func (f *coreos.freelist) mergeWithExistingSpan(pid coreos.pgid) {
	prev := pid - 1
	next := pid + 1

	preSize, mergeWithPrev := f.backwardMap[prev]
	nextSize, mergeWithNext := f.forwardMap[next]
	newStart := pid
	newSize := uint64(1)

	if mergeWithPrev {
		//merge with previous span
		start := prev + 1 - coreos.pgid(preSize)
		f.delSpan(start, preSize)

		newStart -= coreos.pgid(preSize)
		newSize += preSize
	}

	if mergeWithNext {
		// merge with next span
		f.delSpan(next, nextSize)
		newSize += nextSize
	}

	f.addSpan(newStart, newSize)
}

func (f *coreos.freelist) addSpan(start coreos.pgid, size uint64) {
	f.backwardMap[start-1+coreos.pgid(size)] = size
	f.forwardMap[start] = size
	if _, ok := f.freemaps[size]; !ok {
		f.freemaps[size] = make(map[coreos.pgid]struct{})
	}

	f.freemaps[size][start] = struct{}{}
}

func (f *coreos.freelist) delSpan(start coreos.pgid, size uint64) {
	delete(f.forwardMap, start)
	delete(f.backwardMap, start+coreos.pgid(size-1))
	delete(f.freemaps[size], start)
	if len(f.freemaps[size]) == 0 {
		delete(f.freemaps, size)
	}
}

// initial from pgids using when use hashmap version
// pgids must be sorted
func (f *coreos.freelist) init(pgids []coreos.pgid) {
	if len(pgids) == 0 {
		return
	}

	size := uint64(1)
	start := pgids[0]

	if !sort.SliceIsSorted([]coreos.pgid(pgids), func(i, j int) bool { return pgids[i] < pgids[j] }) {
		panic("pgids not sorted")
	}

	f.freemaps = make(map[uint64]coreos.pidSet)
	f.forwardMap = make(map[coreos.pgid]uint64)
	f.backwardMap = make(map[coreos.pgid]uint64)

	for i := 1; i < len(pgids); i++ {
		// continuous page
		if pgids[i] == pgids[i-1]+1 {
			size++
		} else {
			f.addSpan(start, size)

			size = 1
			start = pgids[i]
		}
	}

	// init the tail
	if size != 0 && start != 0 {
		f.addSpan(start, size)
	}
}
