package metrics

import (
	"runtime"
	"testing"
	"time"
	"vendor"
)

func BenchmarkRuntimeMemStats(b *testing.B) {
	r := vendor.NewRegistry()
	vendor.RegisterRuntimeMemStats(r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vendor.CaptureRuntimeMemStatsOnce(r)
	}
}

func TestRuntimeMemStats(t *testing.T) {
	r := vendor.NewRegistry()
	vendor.RegisterRuntimeMemStats(r)
	vendor.CaptureRuntimeMemStatsOnce(r)
	zero := vendor.runtimeMetrics.MemStats.PauseNs.Count() // Get a "zero" since GC may have run before these tests.
	runtime.GC()
	vendor.CaptureRuntimeMemStatsOnce(r)
	if count := vendor.runtimeMetrics.MemStats.PauseNs.Count(); 1 != count-zero {
		t.Fatal(count - zero)
	}
	runtime.GC()
	runtime.GC()
	vendor.CaptureRuntimeMemStatsOnce(r)
	if count := vendor.runtimeMetrics.MemStats.PauseNs.Count(); 3 != count-zero {
		t.Fatal(count - zero)
	}
	for i := 0; i < 256; i++ {
		runtime.GC()
	}
	vendor.CaptureRuntimeMemStatsOnce(r)
	if count := vendor.runtimeMetrics.MemStats.PauseNs.Count(); 259 != count-zero {
		t.Fatal(count - zero)
	}
	for i := 0; i < 257; i++ {
		runtime.GC()
	}
	vendor.CaptureRuntimeMemStatsOnce(r)
	if count := vendor.runtimeMetrics.MemStats.PauseNs.Count(); 515 != count-zero { // We lost one because there were too many GCs between captures.
		t.Fatal(count - zero)
	}
}

func TestRuntimeMemStatsNumThread(t *testing.T) {
	r := vendor.NewRegistry()
	vendor.RegisterRuntimeMemStats(r)
	vendor.CaptureRuntimeMemStatsOnce(r)

	if value := vendor.runtimeMetrics.NumThread.Value(); value < 1 {
		t.Fatalf("got NumThread: %d, wanted at least 1", value)
	}
}

func TestRuntimeMemStatsBlocking(t *testing.T) {
	if g := runtime.GOMAXPROCS(0); g < 2 {
		t.Skipf("skipping TestRuntimeMemStatsBlocking with GOMAXPROCS=%d\n", g)
	}
	ch := make(chan int)
	go testRuntimeMemStatsBlocking(ch)
	var memStats runtime.MemStats
	t0 := time.Now()
	runtime.ReadMemStats(&memStats)
	t1 := time.Now()
	t.Log("i++ during runtime.ReadMemStats:", <-ch)
	go testRuntimeMemStatsBlocking(ch)
	d := t1.Sub(t0)
	t.Log(d)
	time.Sleep(d)
	t.Log("i++ during time.Sleep:", <-ch)
}

func testRuntimeMemStatsBlocking(ch chan int) {
	i := 0
	for {
		select {
		case ch <- i:
			return
		default:
			i++
		}
	}
}
