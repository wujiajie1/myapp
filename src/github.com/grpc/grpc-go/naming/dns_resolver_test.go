/*
 *
 * Copyright 2017 gRPC authors.
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

package naming

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
	"vendor"
)

func newUpdateWithMD(op vendor.Operation, addr, lb string) *vendor.Update {
	return &vendor.Update{
		Op:       op,
		Addr:     addr,
		Metadata: vendor.AddrMetadataGRPCLB{AddrType: vendor.GRPCLB, ServerName: lb},
	}
}

func toMap(u []*vendor.Update) map[string]*vendor.Update {
	m := make(map[string]*vendor.Update)
	for _, v := range u {
		m[v.Addr] = v
	}
	return m
}

func TestCompileUpdate(t *testing.T) {
	tests := []struct {
		oldAddrs []string
		newAddrs []string
		want     []*vendor.Update
	}{
		{
			[]string{},
			[]string{"1.0.0.1"},
			[]*vendor.Update{{Op: vendor.Add, Addr: "1.0.0.1"}},
		},
		{
			[]string{"1.0.0.1"},
			[]string{"1.0.0.1"},
			[]*vendor.Update{},
		},
		{
			[]string{"1.0.0.0"},
			[]string{"1.0.0.1"},
			[]*vendor.Update{{Op: vendor.Delete, Addr: "1.0.0.0"}, {Op: vendor.Add, Addr: "1.0.0.1"}},
		},
		{
			[]string{"1.0.0.1"},
			[]string{"1.0.0.0"},
			[]*vendor.Update{{Op: vendor.Add, Addr: "1.0.0.0"}, {Op: vendor.Delete, Addr: "1.0.0.1"}},
		},
		{
			[]string{"1.0.0.1"},
			[]string{"1.0.0.1", "1.0.0.2", "1.0.0.3"},
			[]*vendor.Update{{Op: vendor.Add, Addr: "1.0.0.2"}, {Op: vendor.Add, Addr: "1.0.0.3"}},
		},
		{
			[]string{"1.0.0.1", "1.0.0.2", "1.0.0.3"},
			[]string{"1.0.0.0"},
			[]*vendor.Update{{Op: vendor.Add, Addr: "1.0.0.0"}, {Op: vendor.Delete, Addr: "1.0.0.1"}, {Op: vendor.Delete, Addr: "1.0.0.2"}, {Op: vendor.Delete, Addr: "1.0.0.3"}},
		},
		{
			[]string{"1.0.0.1", "1.0.0.3", "1.0.0.5"},
			[]string{"1.0.0.2", "1.0.0.3", "1.0.0.6"},
			[]*vendor.Update{{Op: vendor.Delete, Addr: "1.0.0.1"}, {Op: vendor.Add, Addr: "1.0.0.2"}, {Op: vendor.Delete, Addr: "1.0.0.5"}, {Op: vendor.Add, Addr: "1.0.0.6"}},
		},
		{
			[]string{"1.0.0.1", "1.0.0.1", "1.0.0.2"},
			[]string{"1.0.0.1"},
			[]*vendor.Update{{Op: vendor.Delete, Addr: "1.0.0.2"}},
		},
	}

	var w vendor.dnsWatcher
	for _, c := range tests {
		w.curAddrs = make(map[string]*vendor.Update)
		newUpdates := make(map[string]*vendor.Update)
		for _, a := range c.oldAddrs {
			w.curAddrs[a] = &vendor.Update{Addr: a}
		}
		for _, a := range c.newAddrs {
			newUpdates[a] = &vendor.Update{Addr: a}
		}
		r := w.compileUpdate(newUpdates)
		if !reflect.DeepEqual(toMap(c.want), toMap(r)) {
			t.Errorf("w(%+v).compileUpdate(%+v) = %+v, want %+v", c.oldAddrs, c.newAddrs, updatesToSlice(r), updatesToSlice(c.want))
		}
	}
}

func TestResolveFunc(t *testing.T) {
	tests := []struct {
		addr string
		want error
	}{
		// TODO(yuxuanli): More false cases?
		{"www.google.com", nil},
		{"foo.bar:12345", nil},
		{"127.0.0.1", nil},
		{"127.0.0.1:12345", nil},
		{"[::1]:80", nil},
		{"[2001:db8:a0b:12f0::1]:21", nil},
		{":80", nil},
		{"127.0.0...1:12345", nil},
		{"[fe80::1%lo0]:80", nil},
		{"golang.org:http", nil},
		{"[2001:db8::1]:http", nil},
		{":", nil},
		{"", vendor.errMissingAddr},
		{"[2001:db8:a0b:12f0::1", fmt.Errorf("invalid target address %v", "[2001:db8:a0b:12f0::1")},
	}

	r, err := vendor.NewDNSResolver()
	if err != nil {
		t.Errorf("%v", err)
	}
	for _, v := range tests {
		_, err := r.Resolve(v.addr)
		if !reflect.DeepEqual(err, v.want) {
			t.Errorf("Resolve(%q) = %v, want %v", v.addr, err, v.want)
		}
	}
}

var hostLookupTbl = map[string][]string{
	"foo.bar.com":      {"1.2.3.4", "5.6.7.8"},
	"ipv4.single.fake": {"1.2.3.4"},
	"ipv4.multi.fake":  {"1.2.3.4", "5.6.7.8", "9.10.11.12"},
	"ipv6.single.fake": {"2607:f8b0:400a:801::1001"},
	"ipv6.multi.fake":  {"2607:f8b0:400a:801::1001", "2607:f8b0:400a:801::1002", "2607:f8b0:400a:801::1003"},
}

func hostLookup(host string) ([]string, error) {
	if addrs, ok := hostLookupTbl[host]; ok {
		return addrs, nil
	}
	return nil, fmt.Errorf("failed to lookup host:%s resolution in hostLookupTbl", host)
}

var srvLookupTbl = map[string][]*net.SRV{
	"_grpclb._tcp.srv.ipv4.single.fake": {&net.SRV{Target: "ipv4.single.fake", Port: 1234}},
	"_grpclb._tcp.srv.ipv4.multi.fake":  {&net.SRV{Target: "ipv4.multi.fake", Port: 1234}},
	"_grpclb._tcp.srv.ipv6.single.fake": {&net.SRV{Target: "ipv6.single.fake", Port: 1234}},
	"_grpclb._tcp.srv.ipv6.multi.fake":  {&net.SRV{Target: "ipv6.multi.fake", Port: 1234}},
}

func srvLookup(service, proto, name string) (string, []*net.SRV, error) {
	cname := "_" + service + "._" + proto + "." + name
	if srvs, ok := srvLookupTbl[cname]; ok {
		return cname, srvs, nil
	}
	return "", nil, fmt.Errorf("failed to lookup srv record for %s in srvLookupTbl", cname)
}

func updatesToSlice(updates []*vendor.Update) []vendor.Update {
	res := make([]vendor.Update, len(updates))
	for i, u := range updates {
		res[i] = *u
	}
	return res
}

func testResolver(t *testing.T, freq time.Duration, slp time.Duration) {
	tests := []struct {
		target string
		want   []*vendor.Update
	}{
		{
			"foo.bar.com",
			[]*vendor.Update{{Op: vendor.Add, Addr: "1.2.3.4" + colonDefaultPort}, {Op: vendor.Add, Addr: "5.6.7.8" + colonDefaultPort}},
		},
		{
			"foo.bar.com:1234",
			[]*vendor.Update{{Op: vendor.Add, Addr: "1.2.3.4:1234"}, {Op: vendor.Add, Addr: "5.6.7.8:1234"}},
		},
		{
			"srv.ipv4.single.fake",
			[]*vendor.Update{newUpdateWithMD(vendor.Add, "1.2.3.4:1234", "ipv4.single.fake")},
		},
		{
			"srv.ipv4.multi.fake",
			[]*vendor.Update{
				newUpdateWithMD(vendor.Add, "1.2.3.4:1234", "ipv4.multi.fake"),
				newUpdateWithMD(vendor.Add, "5.6.7.8:1234", "ipv4.multi.fake"),
				newUpdateWithMD(vendor.Add, "9.10.11.12:1234", "ipv4.multi.fake")},
		},
		{
			"srv.ipv6.single.fake",
			[]*vendor.Update{newUpdateWithMD(vendor.Add, "[2607:f8b0:400a:801::1001]:1234", "ipv6.single.fake")},
		},
		{
			"srv.ipv6.multi.fake",
			[]*vendor.Update{
				newUpdateWithMD(vendor.Add, "[2607:f8b0:400a:801::1001]:1234", "ipv6.multi.fake"),
				newUpdateWithMD(vendor.Add, "[2607:f8b0:400a:801::1002]:1234", "ipv6.multi.fake"),
				newUpdateWithMD(vendor.Add, "[2607:f8b0:400a:801::1003]:1234", "ipv6.multi.fake"),
			},
		},
	}

	for _, a := range tests {
		r, err := vendor.NewDNSResolverWithFreq(freq)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		w, err := r.Resolve(a.target)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		updates, err := w.Next()
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		if !reflect.DeepEqual(toMap(a.want), toMap(updates)) {
			t.Errorf("Resolve(%q) = %+v, want %+v\n", a.target, updatesToSlice(updates), updatesToSlice(a.want))
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				_, err := w.Next()
				if err != nil {
					return
				}
				t.Error("Execution shouldn't reach here, since w.Next() should be blocked until close happen.")
			}
		}()
		// Sleep for sometime to let watcher do more than one lookup
		time.Sleep(slp)
		w.Close()
		wg.Wait()
	}
}

func replaceNetFunc() func() {
	oldLookupHost := vendor.lookupHost
	oldLookupSRV := vendor.lookupSRV
	vendor.lookupHost = func(ctx context.Context, host string) ([]string, error) {
		return hostLookup(host)
	}
	vendor.lookupSRV = func(ctx context.Context, service, proto, name string) (string, []*net.SRV, error) {
		return srvLookup(service, proto, name)
	}
	return func() {
		vendor.lookupHost = oldLookupHost
		vendor.lookupSRV = oldLookupSRV
	}
}

func TestResolve(t *testing.T) {
	defer replaceNetFunc()()
	testResolver(t, time.Millisecond*5, time.Millisecond*10)
}

const colonDefaultPort = ":" + vendor.defaultPort

func TestIPWatcher(t *testing.T) {
	tests := []struct {
		target string
		want   []*vendor.Update
	}{
		{"127.0.0.1", []*vendor.Update{{Op: vendor.Add, Addr: "127.0.0.1" + colonDefaultPort}}},
		{"127.0.0.1:12345", []*vendor.Update{{Op: vendor.Add, Addr: "127.0.0.1:12345"}}},
		{"::1", []*vendor.Update{{Op: vendor.Add, Addr: "[::1]" + colonDefaultPort}}},
		{"[::1]:12345", []*vendor.Update{{Op: vendor.Add, Addr: "[::1]:12345"}}},
		{"[::1]:", []*vendor.Update{{Op: vendor.Add, Addr: "[::1]:443"}}},
		{"2001:db8:85a3::8a2e:370:7334", []*vendor.Update{{Op: vendor.Add, Addr: "[2001:db8:85a3::8a2e:370:7334]" + colonDefaultPort}}},
		{"[2001:db8:85a3::8a2e:370:7334]", []*vendor.Update{{Op: vendor.Add, Addr: "[2001:db8:85a3::8a2e:370:7334]" + colonDefaultPort}}},
		{"[2001:db8:85a3::8a2e:370:7334]:12345", []*vendor.Update{{Op: vendor.Add, Addr: "[2001:db8:85a3::8a2e:370:7334]:12345"}}},
		{"[2001:db8::1]:http", []*vendor.Update{{Op: vendor.Add, Addr: "[2001:db8::1]:http"}}},
		// TODO(yuxuanli): zone support?
	}

	for _, v := range tests {
		r, err := vendor.NewDNSResolverWithFreq(time.Millisecond * 5)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		w, err := r.Resolve(v.target)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		var updates []*vendor.Update
		var wg sync.WaitGroup
		wg.Add(1)
		count := 0
		go func() {
			defer wg.Done()
			for {
				u, err := w.Next()
				if err != nil {
					return
				}
				updates = u
				count++
			}
		}()
		// Sleep for sometime to let watcher do more than one lookup
		time.Sleep(time.Millisecond * 10)
		w.Close()
		wg.Wait()
		if !reflect.DeepEqual(v.want, updates) {
			t.Errorf("Resolve(%q) = %v, want %+v\n", v.target, updatesToSlice(updates), updatesToSlice(v.want))
		}
		if count != 1 {
			t.Errorf("IPWatcher Next() should return only once, not %d times\n", count)
		}
	}
}
