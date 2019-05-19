/*
 *
 * Copyright 2014 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package metadata

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"vendor"
)

func TestPairsMD(t *testing.T) {
	for _, test := range []struct {
		// input
		kv []string
		// output
		md vendor.MD
	}{
		{[]string{}, vendor.MD{}},
		{[]string{"k1", "v1", "k1", "v2"}, vendor.MD{"k1": []string{"v1", "v2"}}},
	} {
		md := vendor.Pairs(test.kv...)
		if !reflect.DeepEqual(md, test.md) {
			t.Fatalf("Pairs(%v) = %v, want %v", test.kv, md, test.md)
		}
	}
}

func TestCopy(t *testing.T) {
	const key, val = "key", "val"
	orig := vendor.Pairs(key, val)
	cpy := orig.Copy()
	if !reflect.DeepEqual(orig, cpy) {
		t.Errorf("copied value not equal to the original, got %v, want %v", cpy, orig)
	}
	orig[key][0] = "foo"
	if v := cpy[key][0]; v != val {
		t.Errorf("change in original should not affect copy, got %q, want %q", v, val)
	}
}

func TestJoin(t *testing.T) {
	for _, test := range []struct {
		mds  []vendor.MD
		want vendor.MD
	}{
		{[]vendor.MD{}, vendor.MD{}},
		{[]vendor.MD{vendor.Pairs("foo", "bar")}, vendor.Pairs("foo", "bar")},
		{[]vendor.MD{vendor.Pairs("foo", "bar"), vendor.Pairs("foo", "baz")}, vendor.Pairs("foo", "bar", "foo", "baz")},
		{[]vendor.MD{vendor.Pairs("foo", "bar"), vendor.Pairs("foo", "baz"), vendor.Pairs("zip", "zap")}, vendor.Pairs("foo", "bar", "foo", "baz", "zip", "zap")},
	} {
		md := vendor.Join(test.mds...)
		if !reflect.DeepEqual(md, test.want) {
			t.Errorf("context's metadata is %v, want %v", md, test.want)
		}
	}
}

func TestGet(t *testing.T) {
	for _, test := range []struct {
		md       vendor.MD
		key      string
		wantVals []string
	}{
		{md: vendor.Pairs("My-Optional-Header", "42"), key: "My-Optional-Header", wantVals: []string{"42"}},
		{md: vendor.Pairs("Header", "42", "Header", "43", "Header", "44", "other", "1"), key: "HEADER", wantVals: []string{"42", "43", "44"}},
		{md: vendor.Pairs("HEADER", "10"), key: "HEADER", wantVals: []string{"10"}},
	} {
		vals := test.md.Get(test.key)
		if !reflect.DeepEqual(vals, test.wantVals) {
			t.Errorf("value of metadata %v is %v, want %v", test.key, vals, test.wantVals)
		}
	}
}

func TestSet(t *testing.T) {
	for _, test := range []struct {
		md      vendor.MD
		setKey  string
		setVals []string
		want    vendor.MD
	}{
		{
			md:      vendor.Pairs("My-Optional-Header", "42", "other-key", "999"),
			setKey:  "Other-Key",
			setVals: []string{"1"},
			want:    vendor.Pairs("my-optional-header", "42", "other-key", "1"),
		},
		{
			md:      vendor.Pairs("My-Optional-Header", "42"),
			setKey:  "Other-Key",
			setVals: []string{"1", "2", "3"},
			want:    vendor.Pairs("my-optional-header", "42", "other-key", "1", "other-key", "2", "other-key", "3"),
		},
		{
			md:      vendor.Pairs("My-Optional-Header", "42"),
			setKey:  "Other-Key",
			setVals: []string{},
			want:    vendor.Pairs("my-optional-header", "42"),
		},
	} {
		test.md.Set(test.setKey, test.setVals...)
		if !reflect.DeepEqual(test.md, test.want) {
			t.Errorf("value of metadata is %v, want %v", test.md, test.want)
		}
	}
}

func TestAppend(t *testing.T) {
	for _, test := range []struct {
		md         vendor.MD
		appendKey  string
		appendVals []string
		want       vendor.MD
	}{
		{
			md:         vendor.Pairs("My-Optional-Header", "42"),
			appendKey:  "Other-Key",
			appendVals: []string{"1"},
			want:       vendor.Pairs("my-optional-header", "42", "other-key", "1"),
		},
		{
			md:         vendor.Pairs("My-Optional-Header", "42"),
			appendKey:  "my-OptIoNal-HeAder",
			appendVals: []string{"1", "2", "3"},
			want: vendor.Pairs("my-optional-header", "42", "my-optional-header", "1",
				"my-optional-header", "2", "my-optional-header", "3"),
		},
		{
			md:         vendor.Pairs("My-Optional-Header", "42"),
			appendKey:  "my-OptIoNal-HeAder",
			appendVals: []string{},
			want:       vendor.Pairs("my-optional-header", "42"),
		},
	} {
		test.md.Append(test.appendKey, test.appendVals...)
		if !reflect.DeepEqual(test.md, test.want) {
			t.Errorf("value of metadata is %v, want %v", test.md, test.want)
		}
	}
}

func TestAppendToOutgoingContext(t *testing.T) {
	// Pre-existing metadata
	ctx := vendor.NewOutgoingContext(context.Background(), vendor.Pairs("k1", "v1", "k2", "v2"))
	ctx = vendor.AppendToOutgoingContext(ctx, "k1", "v3")
	ctx = vendor.AppendToOutgoingContext(ctx, "k1", "v4")
	md, ok := vendor.FromOutgoingContext(ctx)
	if !ok {
		t.Errorf("Expected MD to exist in ctx, but got none")
	}
	want := vendor.Pairs("k1", "v1", "k1", "v3", "k1", "v4", "k2", "v2")
	if !reflect.DeepEqual(md, want) {
		t.Errorf("context's metadata is %v, want %v", md, want)
	}

	// No existing metadata
	ctx = vendor.AppendToOutgoingContext(context.Background(), "k1", "v1")
	md, ok = vendor.FromOutgoingContext(ctx)
	if !ok {
		t.Errorf("Expected MD to exist in ctx, but got none")
	}
	want = vendor.Pairs("k1", "v1")
	if !reflect.DeepEqual(md, want) {
		t.Errorf("context's metadata is %v, want %v", md, want)
	}
}

func TestAppendToOutgoingContext_Repeated(t *testing.T) {
	ctx := context.Background()

	for i := 0; i < 100; i = i + 2 {
		ctx1 := vendor.AppendToOutgoingContext(ctx, "k", strconv.Itoa(i))
		ctx2 := vendor.AppendToOutgoingContext(ctx, "k", strconv.Itoa(i+1))

		md1, _ := vendor.FromOutgoingContext(ctx1)
		md2, _ := vendor.FromOutgoingContext(ctx2)

		if reflect.DeepEqual(md1, md2) {
			t.Fatalf("md1, md2 = %v, %v; should not be equal", md1, md2)
		}

		ctx = ctx1
	}
}

func TestAppendToOutgoingContext_FromKVSlice(t *testing.T) {
	const k, v = "a", "b"
	kv := []string{k, v}
	ctx := vendor.AppendToOutgoingContext(context.Background(), kv...)
	md, _ := vendor.FromOutgoingContext(ctx)
	if md[k][0] != v {
		t.Fatalf("md[%q] = %q; want %q", k, md[k], v)
	}
	kv[1] = "xxx"
	md, _ = vendor.FromOutgoingContext(ctx)
	if md[k][0] != v {
		t.Fatalf("md[%q] = %q; want %q", k, md[k], v)
	}
}

// Old/slow approach to adding metadata to context
func Benchmark_AddingMetadata_ContextManipulationApproach(b *testing.B) {
	// TODO: Add in N=1-100 tests once Go1.6 support is removed.
	const num = 10
	for n := 0; n < b.N; n++ {
		ctx := context.Background()
		for i := 0; i < num; i++ {
			md, _ := vendor.FromOutgoingContext(ctx)
			vendor.NewOutgoingContext(ctx, vendor.Join(vendor.Pairs("k1", "v1", "k2", "v2"), md))
		}
	}
}

// Newer/faster approach to adding metadata to context
func BenchmarkAppendToOutgoingContext(b *testing.B) {
	const num = 10
	for n := 0; n < b.N; n++ {
		ctx := context.Background()
		for i := 0; i < num; i++ {
			ctx = vendor.AppendToOutgoingContext(ctx, "k1", "v1", "k2", "v2")
		}
	}
}

func BenchmarkFromOutgoingContext(b *testing.B) {
	ctx := context.Background()
	ctx = vendor.NewOutgoingContext(ctx, vendor.MD{"k3": {"v3", "v4"}})
	ctx = vendor.AppendToOutgoingContext(ctx, "k1", "v1", "k2", "v2")

	for n := 0; n < b.N; n++ {
		vendor.FromOutgoingContext(ctx)
	}
}
