package metrics

import (
	"bytes"
	"encoding/json"
	"testing"
	"vendor"
)

func TestRegistryMarshallJSON(t *testing.T) {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	r := vendor.NewRegistry()
	r.Register("counter", vendor.NewCounter())
	enc.Encode(r)
	if s := b.String(); "{\"counter\":{\"count\":0}}\n" != s {
		t.Fatalf(s)
	}
}

func TestRegistryWriteJSONOnce(t *testing.T) {
	r := vendor.NewRegistry()
	r.Register("counter", vendor.NewCounter())
	b := &bytes.Buffer{}
	vendor.WriteJSONOnce(r, b)
	if s := b.String(); s != "{\"counter\":{\"count\":0}}\n" {
		t.Fail()
	}
}
