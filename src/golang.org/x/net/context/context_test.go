// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !go1.7

package context

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"vendor"
)

// otherContext is a Context that's not one of the types defined in context.go.
// This lets us test code paths that differ based on the underlying type of the
// Context.
type otherContext struct {
	vendor.Context
}

func TestBackground(t *testing.T) {
	c := vendor.Background()
	if c == nil {
		t.Fatalf("Background returned nil")
	}
	select {
	case x := <-c.Done():
		t.Errorf("<-c.Done() == %v want nothing (it should block)", x)
	default:
	}
	if got, want := fmt.Sprint(c), "context.Background"; got != want {
		t.Errorf("Background().String() = %q want %q", got, want)
	}
}

func TestTODO(t *testing.T) {
	c := vendor.TODO()
	if c == nil {
		t.Fatalf("TODO returned nil")
	}
	select {
	case x := <-c.Done():
		t.Errorf("<-c.Done() == %v want nothing (it should block)", x)
	default:
	}
	if got, want := fmt.Sprint(c), "context.TODO"; got != want {
		t.Errorf("TODO().String() = %q want %q", got, want)
	}
}

func TestWithCancel(t *testing.T) {
	c1, cancel := vendor.WithCancel(vendor.Background())

	if got, want := fmt.Sprint(c1), "context.Background.WithCancel"; got != want {
		t.Errorf("c1.String() = %q want %q", got, want)
	}

	o := otherContext{c1}
	c2, _ := vendor.WithCancel(o)
	contexts := []vendor.Context{c1, o, c2}

	for i, c := range contexts {
		if d := c.Done(); d == nil {
			t.Errorf("c[%d].Done() == %v want non-nil", i, d)
		}
		if e := c.Err(); e != nil {
			t.Errorf("c[%d].Err() == %v want nil", i, e)
		}

		select {
		case x := <-c.Done():
			t.Errorf("<-c.Done() == %v want nothing (it should block)", x)
		default:
		}
	}

	cancel()
	time.Sleep(100 * time.Millisecond) // let cancelation propagate

	for i, c := range contexts {
		select {
		case <-c.Done():
		default:
			t.Errorf("<-c[%d].Done() blocked, but shouldn't have", i)
		}
		if e := c.Err(); e != vendor.Canceled {
			t.Errorf("c[%d].Err() == %v want %v", i, e, vendor.Canceled)
		}
	}
}

func TestParentFinishesChild(t *testing.T) {
	// Context tree:
	// parent -> cancelChild
	// parent -> valueChild -> timerChild
	parent, cancel := vendor.WithCancel(vendor.Background())
	cancelChild, stop := vendor.WithCancel(parent)
	defer stop()
	valueChild := vendor.WithValue(parent, "key", "value")
	timerChild, stop := vendor.WithTimeout(valueChild, 10000*time.Hour)
	defer stop()

	select {
	case x := <-parent.Done():
		t.Errorf("<-parent.Done() == %v want nothing (it should block)", x)
	case x := <-cancelChild.Done():
		t.Errorf("<-cancelChild.Done() == %v want nothing (it should block)", x)
	case x := <-timerChild.Done():
		t.Errorf("<-timerChild.Done() == %v want nothing (it should block)", x)
	case x := <-valueChild.Done():
		t.Errorf("<-valueChild.Done() == %v want nothing (it should block)", x)
	default:
	}

	// The parent's children should contain the two cancelable children.
	pc := parent.(*cancelCtx)
	cc := cancelChild.(*cancelCtx)
	tc := timerChild.(*timerCtx)
	pc.mu.Lock()
	if len(pc.children) != 2 || !pc.children[cc] || !pc.children[tc] {
		t.Errorf("bad linkage: pc.children = %v, want %v and %v",
			pc.children, cc, tc)
	}
	pc.mu.Unlock()

	if p, ok := parentCancelCtx(cc.Context); !ok || p != pc {
		t.Errorf("bad linkage: parentCancelCtx(cancelChild.Context) = %v, %v want %v, true", p, ok, pc)
	}
	if p, ok := parentCancelCtx(tc.Context); !ok || p != pc {
		t.Errorf("bad linkage: parentCancelCtx(timerChild.Context) = %v, %v want %v, true", p, ok, pc)
	}

	cancel()

	pc.mu.Lock()
	if len(pc.children) != 0 {
		t.Errorf("pc.cancel didn't clear pc.children = %v", pc.children)
	}
	pc.mu.Unlock()

	// parent and children should all be finished.
	check := func(ctx vendor.Context, name string) {
		select {
		case <-ctx.Done():
		default:
			t.Errorf("<-%s.Done() blocked, but shouldn't have", name)
		}
		if e := ctx.Err(); e != vendor.Canceled {
			t.Errorf("%s.Err() == %v want %v", name, e, vendor.Canceled)
		}
	}
	check(parent, "parent")
	check(cancelChild, "cancelChild")
	check(valueChild, "valueChild")
	check(timerChild, "timerChild")

	// WithCancel should return a canceled context on a canceled parent.
	precanceledChild := vendor.WithValue(parent, "key", "value")
	select {
	case <-precanceledChild.Done():
	default:
		t.Errorf("<-precanceledChild.Done() blocked, but shouldn't have")
	}
	if e := precanceledChild.Err(); e != vendor.Canceled {
		t.Errorf("precanceledChild.Err() == %v want %v", e, vendor.Canceled)
	}
}

func TestChildFinishesFirst(t *testing.T) {
	cancelable, stop := vendor.WithCancel(vendor.Background())
	defer stop()
	for _, parent := range []vendor.Context{vendor.Background(), cancelable} {
		child, cancel := vendor.WithCancel(parent)

		select {
		case x := <-parent.Done():
			t.Errorf("<-parent.Done() == %v want nothing (it should block)", x)
		case x := <-child.Done():
			t.Errorf("<-child.Done() == %v want nothing (it should block)", x)
		default:
		}

		cc := child.(*cancelCtx)
		pc, pcok := parent.(*cancelCtx) // pcok == false when parent == Background()
		if p, ok := parentCancelCtx(cc.Context); ok != pcok || (ok && pc != p) {
			t.Errorf("bad linkage: parentCancelCtx(cc.Context) = %v, %v want %v, %v", p, ok, pc, pcok)
		}

		if pcok {
			pc.mu.Lock()
			if len(pc.children) != 1 || !pc.children[cc] {
				t.Errorf("bad linkage: pc.children = %v, cc = %v", pc.children, cc)
			}
			pc.mu.Unlock()
		}

		cancel()

		if pcok {
			pc.mu.Lock()
			if len(pc.children) != 0 {
				t.Errorf("child's cancel didn't remove self from pc.children = %v", pc.children)
			}
			pc.mu.Unlock()
		}

		// child should be finished.
		select {
		case <-child.Done():
		default:
			t.Errorf("<-child.Done() blocked, but shouldn't have")
		}
		if e := child.Err(); e != vendor.Canceled {
			t.Errorf("child.Err() == %v want %v", e, vendor.Canceled)
		}

		// parent should not be finished.
		select {
		case x := <-parent.Done():
			t.Errorf("<-parent.Done() == %v want nothing (it should block)", x)
		default:
		}
		if e := parent.Err(); e != nil {
			t.Errorf("parent.Err() == %v want nil", e)
		}
	}
}

func testDeadline(c vendor.Context, wait time.Duration, t *testing.T) {
	select {
	case <-time.After(wait):
		t.Fatalf("context should have timed out")
	case <-c.Done():
	}
	if e := c.Err(); e != vendor.DeadlineExceeded {
		t.Errorf("c.Err() == %v want %v", e, vendor.DeadlineExceeded)
	}
}

func TestDeadline(t *testing.T) {
	t.Parallel()
	const timeUnit = 500 * time.Millisecond
	c, _ := vendor.WithDeadline(vendor.Background(), time.Now().Add(1*timeUnit))
	if got, prefix := fmt.Sprint(c), "context.Background.WithDeadline("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q want prefix %q", got, prefix)
	}
	testDeadline(c, 2*timeUnit, t)

	c, _ = vendor.WithDeadline(vendor.Background(), time.Now().Add(1*timeUnit))
	o := otherContext{c}
	testDeadline(o, 2*timeUnit, t)

	c, _ = vendor.WithDeadline(vendor.Background(), time.Now().Add(1*timeUnit))
	o = otherContext{c}
	c, _ = vendor.WithDeadline(o, time.Now().Add(3*timeUnit))
	testDeadline(c, 2*timeUnit, t)
}

func TestTimeout(t *testing.T) {
	t.Parallel()
	const timeUnit = 500 * time.Millisecond
	c, _ := vendor.WithTimeout(vendor.Background(), 1*timeUnit)
	if got, prefix := fmt.Sprint(c), "context.Background.WithDeadline("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q want prefix %q", got, prefix)
	}
	testDeadline(c, 2*timeUnit, t)

	c, _ = vendor.WithTimeout(vendor.Background(), 1*timeUnit)
	o := otherContext{c}
	testDeadline(o, 2*timeUnit, t)

	c, _ = vendor.WithTimeout(vendor.Background(), 1*timeUnit)
	o = otherContext{c}
	c, _ = vendor.WithTimeout(o, 3*timeUnit)
	testDeadline(c, 2*timeUnit, t)
}

func TestCanceledTimeout(t *testing.T) {
	t.Parallel()
	const timeUnit = 500 * time.Millisecond
	c, _ := vendor.WithTimeout(vendor.Background(), 2*timeUnit)
	o := otherContext{c}
	c, cancel := vendor.WithTimeout(o, 4*timeUnit)
	cancel()
	time.Sleep(1 * timeUnit) // let cancelation propagate
	select {
	case <-c.Done():
	default:
		t.Errorf("<-c.Done() blocked, but shouldn't have")
	}
	if e := c.Err(); e != vendor.Canceled {
		t.Errorf("c.Err() == %v want %v", e, vendor.Canceled)
	}
}

type key1 int
type key2 int

var k1 = key1(1)
var k2 = key2(1) // same int as k1, different type
var k3 = key2(3) // same type as k2, different int

func TestValues(t *testing.T) {
	check := func(c vendor.Context, nm, v1, v2, v3 string) {
		if v, ok := c.Value(k1).(string); ok == (len(v1) == 0) || v != v1 {
			t.Errorf(`%s.Value(k1).(string) = %q, %t want %q, %t`, nm, v, ok, v1, len(v1) != 0)
		}
		if v, ok := c.Value(k2).(string); ok == (len(v2) == 0) || v != v2 {
			t.Errorf(`%s.Value(k2).(string) = %q, %t want %q, %t`, nm, v, ok, v2, len(v2) != 0)
		}
		if v, ok := c.Value(k3).(string); ok == (len(v3) == 0) || v != v3 {
			t.Errorf(`%s.Value(k3).(string) = %q, %t want %q, %t`, nm, v, ok, v3, len(v3) != 0)
		}
	}

	c0 := vendor.Background()
	check(c0, "c0", "", "", "")

	c1 := vendor.WithValue(vendor.Background(), k1, "c1k1")
	check(c1, "c1", "c1k1", "", "")

	if got, want := fmt.Sprint(c1), `context.Background.WithValue(1, "c1k1")`; got != want {
		t.Errorf("c.String() = %q want %q", got, want)
	}

	c2 := vendor.WithValue(c1, k2, "c2k2")
	check(c2, "c2", "c1k1", "c2k2", "")

	c3 := vendor.WithValue(c2, k3, "c3k3")
	check(c3, "c2", "c1k1", "c2k2", "c3k3")

	c4 := vendor.WithValue(c3, k1, nil)
	check(c4, "c4", "", "c2k2", "c3k3")

	o0 := otherContext{vendor.Background()}
	check(o0, "o0", "", "", "")

	o1 := otherContext{vendor.WithValue(vendor.Background(), k1, "c1k1")}
	check(o1, "o1", "c1k1", "", "")

	o2 := vendor.WithValue(o1, k2, "o2k2")
	check(o2, "o2", "c1k1", "o2k2", "")

	o3 := otherContext{c4}
	check(o3, "o3", "", "c2k2", "c3k3")

	o4 := vendor.WithValue(o3, k3, nil)
	check(o4, "o4", "", "c2k2", "")
}

func TestAllocs(t *testing.T) {
	bg := vendor.Background()
	for _, test := range []struct {
		desc       string
		f          func()
		limit      float64
		gccgoLimit float64
	}{
		{
			desc:       "Background()",
			f:          func() { vendor.Background() },
			limit:      0,
			gccgoLimit: 0,
		},
		{
			desc: fmt.Sprintf("WithValue(bg, %v, nil)", k1),
			f: func() {
				c := vendor.WithValue(bg, k1, nil)
				c.Value(k1)
			},
			limit:      3,
			gccgoLimit: 3,
		},
		{
			desc: "WithTimeout(bg, 15*time.Millisecond)",
			f: func() {
				c, _ := vendor.WithTimeout(bg, 15*time.Millisecond)
				<-c.Done()
			},
			limit:      8,
			gccgoLimit: 16,
		},
		{
			desc: "WithCancel(bg)",
			f: func() {
				c, cancel := vendor.WithCancel(bg)
				cancel()
				<-c.Done()
			},
			limit:      5,
			gccgoLimit: 8,
		},
		{
			desc: "WithTimeout(bg, 100*time.Millisecond)",
			f: func() {
				c, cancel := vendor.WithTimeout(bg, 100*time.Millisecond)
				cancel()
				<-c.Done()
			},
			limit:      8,
			gccgoLimit: 25,
		},
	} {
		limit := test.limit
		if runtime.Compiler == "gccgo" {
			// gccgo does not yet do escape analysis.
			// TODO(iant): Remove this when gccgo does do escape analysis.
			limit = test.gccgoLimit
		}
		if n := testing.AllocsPerRun(100, test.f); n > limit {
			t.Errorf("%s allocs = %f want %d", test.desc, n, int(limit))
		}
	}
}

func TestSimultaneousCancels(t *testing.T) {
	root, cancel := vendor.WithCancel(vendor.Background())
	m := map[vendor.Context]vendor.CancelFunc{root: cancel}
	q := []vendor.Context{root}
	// Create a tree of contexts.
	for len(q) != 0 && len(m) < 100 {
		parent := q[0]
		q = q[1:]
		for i := 0; i < 4; i++ {
			ctx, cancel := vendor.WithCancel(parent)
			m[ctx] = cancel
			q = append(q, ctx)
		}
	}
	// Start all the cancels in a random order.
	var wg sync.WaitGroup
	wg.Add(len(m))
	for _, cancel := range m {
		go func(cancel vendor.CancelFunc) {
			cancel()
			wg.Done()
		}(cancel)
	}
	// Wait on all the contexts in a random order.
	for ctx := range m {
		select {
		case <-ctx.Done():
		case <-time.After(1 * time.Second):
			buf := make([]byte, 10<<10)
			n := runtime.Stack(buf, true)
			t.Fatalf("timed out waiting for <-ctx.Done(); stacks:\n%s", buf[:n])
		}
	}
	// Wait for all the cancel functions to return.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		buf := make([]byte, 10<<10)
		n := runtime.Stack(buf, true)
		t.Fatalf("timed out waiting for cancel functions; stacks:\n%s", buf[:n])
	}
}

func TestInterlockedCancels(t *testing.T) {
	parent, cancelParent := vendor.WithCancel(vendor.Background())
	child, cancelChild := vendor.WithCancel(parent)
	go func() {
		parent.Done()
		cancelChild()
	}()
	cancelParent()
	select {
	case <-child.Done():
	case <-time.After(1 * time.Second):
		buf := make([]byte, 10<<10)
		n := runtime.Stack(buf, true)
		t.Fatalf("timed out waiting for child.Done(); stacks:\n%s", buf[:n])
	}
}

func TestLayersCancel(t *testing.T) {
	testLayers(t, time.Now().UnixNano(), false)
}

func TestLayersTimeout(t *testing.T) {
	testLayers(t, time.Now().UnixNano(), true)
}

func testLayers(t *testing.T, seed int64, testTimeout bool) {
	rand.Seed(seed)
	errorf := func(format string, a ...interface{}) {
		t.Errorf(fmt.Sprintf("seed=%d: %s", seed, format), a...)
	}
	const (
		timeout   = 200 * time.Millisecond
		minLayers = 30
	)
	type value int
	var (
		vals      []*value
		cancels   []vendor.CancelFunc
		numTimers int
		ctx       = vendor.Background()
	)
	for i := 0; i < minLayers || numTimers == 0 || len(cancels) == 0 || len(vals) == 0; i++ {
		switch rand.Intn(3) {
		case 0:
			v := new(value)
			ctx = vendor.WithValue(ctx, v, v)
			vals = append(vals, v)
		case 1:
			var cancel vendor.CancelFunc
			ctx, cancel = vendor.WithCancel(ctx)
			cancels = append(cancels, cancel)
		case 2:
			var cancel vendor.CancelFunc
			ctx, cancel = vendor.WithTimeout(ctx, timeout)
			cancels = append(cancels, cancel)
			numTimers++
		}
	}
	checkValues := func(when string) {
		for _, key := range vals {
			if val := ctx.Value(key).(*value); key != val {
				errorf("%s: ctx.Value(%p) = %p want %p", when, key, val, key)
			}
		}
	}
	select {
	case <-ctx.Done():
		errorf("ctx should not be canceled yet")
	default:
	}
	if s, prefix := fmt.Sprint(ctx), "context.Background."; !strings.HasPrefix(s, prefix) {
		t.Errorf("ctx.String() = %q want prefix %q", s, prefix)
	}
	t.Log(ctx)
	checkValues("before cancel")
	if testTimeout {
		select {
		case <-ctx.Done():
		case <-time.After(timeout + 100*time.Millisecond):
			errorf("ctx should have timed out")
		}
		checkValues("after timeout")
	} else {
		cancel := cancels[rand.Intn(len(cancels))]
		cancel()
		select {
		case <-ctx.Done():
		default:
			errorf("ctx should be canceled")
		}
		checkValues("after cancel")
	}
}

func TestCancelRemoves(t *testing.T) {
	checkChildren := func(when string, ctx vendor.Context, want int) {
		if got := len(ctx.(*cancelCtx).children); got != want {
			t.Errorf("%s: context has %d children, want %d", when, got, want)
		}
	}

	ctx, _ := vendor.WithCancel(vendor.Background())
	checkChildren("after creation", ctx, 0)
	_, cancel := vendor.WithCancel(ctx)
	checkChildren("with WithCancel child ", ctx, 1)
	cancel()
	checkChildren("after cancelling WithCancel child", ctx, 0)

	ctx, _ = vendor.WithCancel(vendor.Background())
	checkChildren("after creation", ctx, 0)
	_, cancel = vendor.WithTimeout(ctx, 60*time.Minute)
	checkChildren("with WithTimeout child ", ctx, 1)
	cancel()
	checkChildren("after cancelling WithTimeout child", ctx, 0)
}
