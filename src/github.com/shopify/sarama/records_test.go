package sarama

import (
	"bytes"
	"reflect"
	"testing"
	"vendor"
)

func TestLegacyRecords(t *testing.T) {
	set := &vendor.MessageSet{
		Messages: []*vendor.MessageBlock{
			{
				Msg: &vendor.Message{
					Version: 1,
				},
			},
		},
	}
	r := vendor.newLegacyRecords(set)

	exp, err := vendor.encode(set, nil)
	if err != nil {
		t.Fatal(err)
	}
	buf, err := vendor.encode(&r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, exp) {
		t.Errorf("Wrong encoding for legacy records, wanted %v, got %v", exp, buf)
	}

	set = &vendor.MessageSet{}
	r = vendor.Records{}

	err = vendor.decode(exp, set)
	if err != nil {
		t.Fatal(err)
	}
	err = vendor.decode(buf, &r)
	if err != nil {
		t.Fatal(err)
	}

	if r.recordsType != vendor.legacyRecords {
		t.Fatalf("Wrong records type %v, expected %v", r.recordsType, vendor.legacyRecords)
	}
	if !reflect.DeepEqual(set, r.MsgSet) {
		t.Errorf("Wrong decoding for legacy records, wanted %#+v, got %#+v", set, r.MsgSet)
	}

	n, err := r.numRecords()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("Wrong number of records, wanted 1, got %d", n)
	}

	p, err := r.isPartial()
	if err != nil {
		t.Fatal(err)
	}
	if p {
		t.Errorf("MessageSet shouldn't have a partial trailing message")
	}

	c, err := r.isControl()
	if err != nil {
		t.Fatal(err)
	}
	if c {
		t.Errorf("MessageSet can't be a control batch")
	}
}

func TestDefaultRecords(t *testing.T) {
	batch := &vendor.RecordBatch{
		IsTransactional: true,
		Version:         2,
		Records: []*vendor.Record{
			{
				Value: []byte{1},
			},
		},
	}

	r := vendor.newDefaultRecords(batch)

	exp, err := vendor.encode(batch, nil)
	if err != nil {
		t.Fatal(err)
	}
	buf, err := vendor.encode(&r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, exp) {
		t.Errorf("Wrong encoding for default records, wanted %v, got %v", exp, buf)
	}

	batch = &vendor.RecordBatch{}
	r = vendor.Records{}

	err = vendor.decode(exp, batch)
	if err != nil {
		t.Fatal(err)
	}
	err = vendor.decode(buf, &r)
	if err != nil {
		t.Fatal(err)
	}

	if r.recordsType != vendor.defaultRecords {
		t.Fatalf("Wrong records type %v, expected %v", r.recordsType, vendor.defaultRecords)
	}
	if !reflect.DeepEqual(batch, r.RecordBatch) {
		t.Errorf("Wrong decoding for default records, wanted %#+v, got %#+v", batch, r.RecordBatch)
	}

	n, err := r.numRecords()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("Wrong number of records, wanted 1, got %d", n)
	}

	p, err := r.isPartial()
	if err != nil {
		t.Fatal(err)
	}
	if p {
		t.Errorf("RecordBatch shouldn't have a partial trailing record")
	}

	c, err := r.isControl()
	if err != nil {
		t.Fatal(err)
	}
	if c {
		t.Errorf("RecordBatch shouldn't be a control batch")
	}
}
