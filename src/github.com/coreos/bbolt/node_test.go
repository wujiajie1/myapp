package bbolt

import (
	"github.com/coreos"
	"testing"
	"unsafe"
)

// Ensure that a node can insert a key/value.
func TestNode_put(t *testing.T) {
	n := &coreos.node{inodes: make(coreos.inodes, 0), bucket: &coreos.Bucket{tx: &coreos.Tx{meta: &coreos.meta{pgid: 1}}}}
	n.put([]byte("baz"), []byte("baz"), []byte("2"), 0, 0)
	n.put([]byte("foo"), []byte("foo"), []byte("0"), 0, 0)
	n.put([]byte("bar"), []byte("bar"), []byte("1"), 0, 0)
	n.put([]byte("foo"), []byte("foo"), []byte("3"), 0, coreos.leafPageFlag)

	if len(n.inodes) != 3 {
		t.Fatalf("exp=3; got=%d", len(n.inodes))
	}
	if k, v := n.inodes[0].key, n.inodes[0].value; string(k) != "bar" || string(v) != "1" {
		t.Fatalf("exp=<bar,1>; got=<%s,%s>", k, v)
	}
	if k, v := n.inodes[1].key, n.inodes[1].value; string(k) != "baz" || string(v) != "2" {
		t.Fatalf("exp=<baz,2>; got=<%s,%s>", k, v)
	}
	if k, v := n.inodes[2].key, n.inodes[2].value; string(k) != "foo" || string(v) != "3" {
		t.Fatalf("exp=<foo,3>; got=<%s,%s>", k, v)
	}
	if n.inodes[2].flags != uint32(coreos.leafPageFlag) {
		t.Fatalf("not a leaf: %d", n.inodes[2].flags)
	}
}

// Ensure that a node can deserialize from a leaf page.
func TestNode_read_LeafPage(t *testing.T) {
	// Create a page.
	var buf [4096]byte
	page := (*coreos.page)(unsafe.Pointer(&buf[0]))
	page.flags = coreos.leafPageFlag
	page.count = 2

	// Insert 2 elements at the beginning. sizeof(leafPageElement) == 16
	nodes := (*[3]coreos.leafPageElement)(unsafe.Pointer(&page.ptr))
	nodes[0] = coreos.leafPageElement{flags: 0, pos: 32, ksize: 3, vsize: 4}  // pos = sizeof(leafPageElement) * 2
	nodes[1] = coreos.leafPageElement{flags: 0, pos: 23, ksize: 10, vsize: 3} // pos = sizeof(leafPageElement) + 3 + 4

	// Write data for the nodes at the end.
	data := (*[4096]byte)(unsafe.Pointer(&nodes[2]))
	copy(data[:], "barfooz")
	copy(data[7:], "helloworldbye")

	// Deserialize page into a leaf.
	n := &coreos.node{}
	n.read(page)

	// Check that there are two inodes with correct data.
	if !n.isLeaf {
		t.Fatal("expected leaf")
	}
	if len(n.inodes) != 2 {
		t.Fatalf("exp=2; got=%d", len(n.inodes))
	}
	if k, v := n.inodes[0].key, n.inodes[0].value; string(k) != "bar" || string(v) != "fooz" {
		t.Fatalf("exp=<bar,fooz>; got=<%s,%s>", k, v)
	}
	if k, v := n.inodes[1].key, n.inodes[1].value; string(k) != "helloworld" || string(v) != "bye" {
		t.Fatalf("exp=<helloworld,bye>; got=<%s,%s>", k, v)
	}
}

// Ensure that a node can serialize into a leaf page.
func TestNode_write_LeafPage(t *testing.T) {
	// Create a node.
	n := &coreos.node{isLeaf: true, inodes: make(coreos.inodes, 0), bucket: &coreos.Bucket{tx: &coreos.Tx{db: &coreos.DB{}, meta: &coreos.meta{pgid: 1}}}}
	n.put([]byte("susy"), []byte("susy"), []byte("que"), 0, 0)
	n.put([]byte("ricki"), []byte("ricki"), []byte("lake"), 0, 0)
	n.put([]byte("john"), []byte("john"), []byte("johnson"), 0, 0)

	// Write it to a page.
	var buf [4096]byte
	p := (*coreos.page)(unsafe.Pointer(&buf[0]))
	n.write(p)

	// Read the page back in.
	n2 := &coreos.node{}
	n2.read(p)

	// Check that the two pages are the same.
	if len(n2.inodes) != 3 {
		t.Fatalf("exp=3; got=%d", len(n2.inodes))
	}
	if k, v := n2.inodes[0].key, n2.inodes[0].value; string(k) != "john" || string(v) != "johnson" {
		t.Fatalf("exp=<john,johnson>; got=<%s,%s>", k, v)
	}
	if k, v := n2.inodes[1].key, n2.inodes[1].value; string(k) != "ricki" || string(v) != "lake" {
		t.Fatalf("exp=<ricki,lake>; got=<%s,%s>", k, v)
	}
	if k, v := n2.inodes[2].key, n2.inodes[2].value; string(k) != "susy" || string(v) != "que" {
		t.Fatalf("exp=<susy,que>; got=<%s,%s>", k, v)
	}
}

// Ensure that a node can split into appropriate subgroups.
func TestNode_split(t *testing.T) {
	// Create a node.
	n := &coreos.node{inodes: make(coreos.inodes, 0), bucket: &coreos.Bucket{tx: &coreos.Tx{db: &coreos.DB{}, meta: &coreos.meta{pgid: 1}}}}
	n.put([]byte("00000001"), []byte("00000001"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000002"), []byte("00000002"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000003"), []byte("00000003"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000004"), []byte("00000004"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000005"), []byte("00000005"), []byte("0123456701234567"), 0, 0)

	// Split between 2 & 3.
	n.split(100)

	var parent = n.parent
	if len(parent.children) != 2 {
		t.Fatalf("exp=2; got=%d", len(parent.children))
	}
	if len(parent.children[0].inodes) != 2 {
		t.Fatalf("exp=2; got=%d", len(parent.children[0].inodes))
	}
	if len(parent.children[1].inodes) != 3 {
		t.Fatalf("exp=3; got=%d", len(parent.children[1].inodes))
	}
}

// Ensure that a page with the minimum number of inodes just returns a single node.
func TestNode_split_MinKeys(t *testing.T) {
	// Create a node.
	n := &coreos.node{inodes: make(coreos.inodes, 0), bucket: &coreos.Bucket{tx: &coreos.Tx{db: &coreos.DB{}, meta: &coreos.meta{pgid: 1}}}}
	n.put([]byte("00000001"), []byte("00000001"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000002"), []byte("00000002"), []byte("0123456701234567"), 0, 0)

	// Split.
	n.split(20)
	if n.parent != nil {
		t.Fatalf("expected nil parent")
	}
}

// Ensure that a node that has keys that all fit on a page just returns one leaf.
func TestNode_split_SinglePage(t *testing.T) {
	// Create a node.
	n := &coreos.node{inodes: make(coreos.inodes, 0), bucket: &coreos.Bucket{tx: &coreos.Tx{db: &coreos.DB{}, meta: &coreos.meta{pgid: 1}}}}
	n.put([]byte("00000001"), []byte("00000001"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000002"), []byte("00000002"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000003"), []byte("00000003"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000004"), []byte("00000004"), []byte("0123456701234567"), 0, 0)
	n.put([]byte("00000005"), []byte("00000005"), []byte("0123456701234567"), 0, 0)

	// Split.
	n.split(4096)
	if n.parent != nil {
		t.Fatalf("expected nil parent")
	}
}
