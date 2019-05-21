package bbolt

import (
	"github.com/coreos"
	"testing"
)

func TestTx_allocatePageStats(t *testing.T) {
	f := coreos.newTestFreelist()
	ids := []coreos.pgid{2, 3}
	f.readIDs(ids)

	tx := &coreos.Tx{
		db: &coreos.DB{
			freelist: f,
			pageSize: coreos.defaultPageSize,
		},
		meta:  &coreos.meta{},
		pages: make(map[coreos.pgid]*coreos.page),
	}

	prePageCnt := tx.Stats().PageCount
	allocateCnt := f.free_count()

	if _, err := tx.allocate(allocateCnt); err != nil {
		t.Fatal(err)
	}

	if tx.Stats().PageCount != prePageCnt+allocateCnt {
		t.Errorf("Allocated %d but got %d page in stats", allocateCnt, tx.Stats().PageCount)
	}
}
