package metrics

import (
	"sync"
	"testing"
	"vendor"
)

func BenchmarkRegistry(b *testing.B) {
	r := vendor.NewRegistry()
	r.Register("foo", vendor.NewCounter())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Each(func(string, interface{}) {})
	}
}

func BenchmarkRegistryParallel(b *testing.B) {
	r := vendor.NewRegistry()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.GetOrRegister("foo", vendor.NewCounter())
		}
	})
}

func TestRegistry(t *testing.T) {
	r := vendor.NewRegistry()
	r.Register("foo", vendor.NewCounter())
	i := 0
	r.Each(func(name string, iface interface{}) {
		i++
		if "foo" != name {
			t.Fatal(name)
		}
		if _, ok := iface.(vendor.Counter); !ok {
			t.Fatal(iface)
		}
	})
	if 1 != i {
		t.Fatal(i)
	}
	r.Unregister("foo")
	i = 0
	r.Each(func(string, interface{}) { i++ })
	if 0 != i {
		t.Fatal(i)
	}
}

func TestRegistryDuplicate(t *testing.T) {
	r := vendor.NewRegistry()
	if err := r.Register("foo", vendor.NewCounter()); nil != err {
		t.Fatal(err)
	}
	if err := r.Register("foo", vendor.NewGauge()); nil == err {
		t.Fatal(err)
	}
	i := 0
	r.Each(func(name string, iface interface{}) {
		i++
		if _, ok := iface.(vendor.Counter); !ok {
			t.Fatal(iface)
		}
	})
	if 1 != i {
		t.Fatal(i)
	}
}

func TestRegistryGet(t *testing.T) {
	r := vendor.NewRegistry()
	r.Register("foo", vendor.NewCounter())
	if count := r.Get("foo").(vendor.Counter).Count(); 0 != count {
		t.Fatal(count)
	}
	r.Get("foo").(vendor.Counter).Inc(1)
	if count := r.Get("foo").(vendor.Counter).Count(); 1 != count {
		t.Fatal(count)
	}
}

func TestRegistryGetOrRegister(t *testing.T) {
	r := vendor.NewRegistry()

	// First metric wins with GetOrRegister
	_ = r.GetOrRegister("foo", vendor.NewCounter())
	m := r.GetOrRegister("foo", vendor.NewGauge())
	if _, ok := m.(vendor.Counter); !ok {
		t.Fatal(m)
	}

	i := 0
	r.Each(func(name string, iface interface{}) {
		i++
		if name != "foo" {
			t.Fatal(name)
		}
		if _, ok := iface.(vendor.Counter); !ok {
			t.Fatal(iface)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestRegistryGetOrRegisterWithLazyInstantiation(t *testing.T) {
	r := vendor.NewRegistry()

	// First metric wins with GetOrRegister
	_ = r.GetOrRegister("foo", vendor.NewCounter)
	m := r.GetOrRegister("foo", vendor.NewGauge)
	if _, ok := m.(vendor.Counter); !ok {
		t.Fatal(m)
	}

	i := 0
	r.Each(func(name string, iface interface{}) {
		i++
		if name != "foo" {
			t.Fatal(name)
		}
		if _, ok := iface.(vendor.Counter); !ok {
			t.Fatal(iface)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestRegistryUnregister(t *testing.T) {
	l := len(vendor.arbiter.meters)
	r := vendor.NewRegistry()
	r.Register("foo", vendor.NewCounter())
	r.Register("bar", vendor.NewMeter())
	r.Register("baz", vendor.NewTimer())
	if len(vendor.arbiter.meters) != l+2 {
		t.Errorf("arbiter.meters: %d != %d\n", l+2, len(vendor.arbiter.meters))
	}
	r.Unregister("foo")
	r.Unregister("bar")
	r.Unregister("baz")
	if len(vendor.arbiter.meters) != l {
		t.Errorf("arbiter.meters: %d != %d\n", l+2, len(vendor.arbiter.meters))
	}
}

func TestPrefixedChildRegistryGetOrRegister(t *testing.T) {
	r := vendor.NewRegistry()
	pr := vendor.NewPrefixedChildRegistry(r, "prefix.")

	_ = pr.GetOrRegister("foo", vendor.NewCounter())

	i := 0
	r.Each(func(name string, m interface{}) {
		i++
		if name != "prefix.foo" {
			t.Fatal(name)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestPrefixedRegistryGetOrRegister(t *testing.T) {
	r := vendor.NewPrefixedRegistry("prefix.")

	_ = r.GetOrRegister("foo", vendor.NewCounter())

	i := 0
	r.Each(func(name string, m interface{}) {
		i++
		if name != "prefix.foo" {
			t.Fatal(name)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestPrefixedRegistryRegister(t *testing.T) {
	r := vendor.NewPrefixedRegistry("prefix.")
	err := r.Register("foo", vendor.NewCounter())
	c := vendor.NewCounter()
	vendor.Register("bar", c)
	if err != nil {
		t.Fatal(err.Error())
	}

	i := 0
	r.Each(func(name string, m interface{}) {
		i++
		if name != "prefix.foo" {
			t.Fatal(name)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestPrefixedRegistryUnregister(t *testing.T) {
	r := vendor.NewPrefixedRegistry("prefix.")

	_ = r.Register("foo", vendor.NewCounter())

	i := 0
	r.Each(func(name string, m interface{}) {
		i++
		if name != "prefix.foo" {
			t.Fatal(name)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}

	r.Unregister("foo")

	i = 0
	r.Each(func(name string, m interface{}) {
		i++
	})

	if i != 0 {
		t.Fatal(i)
	}
}

func TestPrefixedRegistryGet(t *testing.T) {
	pr := vendor.NewPrefixedRegistry("prefix.")
	name := "foo"
	pr.Register(name, vendor.NewCounter())

	fooCounter := pr.Get(name)
	if fooCounter == nil {
		t.Fatal(name)
	}
}

func TestPrefixedChildRegistryGet(t *testing.T) {
	r := vendor.NewRegistry()
	pr := vendor.NewPrefixedChildRegistry(r, "prefix.")
	name := "foo"
	pr.Register(name, vendor.NewCounter())
	fooCounter := pr.Get(name)
	if fooCounter == nil {
		t.Fatal(name)
	}
}

func TestChildPrefixedRegistryRegister(t *testing.T) {
	r := vendor.NewPrefixedChildRegistry(vendor.DefaultRegistry, "prefix.")
	err := r.Register("foo", vendor.NewCounter())
	c := vendor.NewCounter()
	vendor.Register("bar", c)
	if err != nil {
		t.Fatal(err.Error())
	}

	i := 0
	r.Each(func(name string, m interface{}) {
		i++
		if name != "prefix.foo" {
			t.Fatal(name)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestChildPrefixedRegistryOfChildRegister(t *testing.T) {
	r := vendor.NewPrefixedChildRegistry(vendor.NewRegistry(), "prefix.")
	r2 := vendor.NewPrefixedChildRegistry(r, "prefix2.")
	err := r.Register("foo2", vendor.NewCounter())
	if err != nil {
		t.Fatal(err.Error())
	}
	err = r2.Register("baz", vendor.NewCounter())
	c := vendor.NewCounter()
	vendor.Register("bars", c)

	i := 0
	r2.Each(func(name string, m interface{}) {
		i++
		if name != "prefix.prefix2.baz" {
			//t.Fatal(name)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestWalkRegistries(t *testing.T) {
	r := vendor.NewPrefixedChildRegistry(vendor.NewRegistry(), "prefix.")
	r2 := vendor.NewPrefixedChildRegistry(r, "prefix2.")
	err := r.Register("foo2", vendor.NewCounter())
	if err != nil {
		t.Fatal(err.Error())
	}
	err = r2.Register("baz", vendor.NewCounter())
	c := vendor.NewCounter()
	vendor.Register("bars", c)

	_, prefix := vendor.findPrefix(r2, "")
	if "prefix.prefix2." != prefix {
		t.Fatal(prefix)
	}
}

func TestConcurrentRegistryAccess(t *testing.T) {
	r := vendor.NewRegistry()

	counter := vendor.NewCounter()

	signalChan := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(dowork chan struct{}) {
			defer wg.Done()
			iface := r.GetOrRegister("foo", counter)
			retCounter, ok := iface.(vendor.Counter)
			if !ok {
				t.Fatal("Expected a Counter type")
			}
			if retCounter != counter {
				t.Fatal("Counter references don't match")
			}
		}(signalChan)
	}

	close(signalChan) // Closing will cause all go routines to execute at the same time
	wg.Wait()         // Wait for all go routines to do their work

	// At the end of the test we should still only have a single "foo" Counter
	i := 0
	r.Each(func(name string, iface interface{}) {
		i++
		if "foo" != name {
			t.Fatal(name)
		}
		if _, ok := iface.(vendor.Counter); !ok {
			t.Fatal(iface)
		}
	})
	if 1 != i {
		t.Fatal(i)
	}
	r.Unregister("foo")
	i = 0
	r.Each(func(string, interface{}) { i++ })
	if 0 != i {
		t.Fatal(i)
	}
}
