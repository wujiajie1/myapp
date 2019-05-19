// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2015 The Go Authors.  All rights reserved.
// https://github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package jsonpb

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"
	"vendor"
)

var (
	marshaler = vendor.Marshaler{}

	marshalerAllOptions = vendor.Marshaler{
		Indent: "  ",
	}

	simpleObject = &vendor.Simple{
		OInt32:     vendor.Int32(-32),
		OInt32Str:  vendor.Int32(-32),
		OInt64:     vendor.Int64(-6400000000),
		OInt64Str:  vendor.Int64(-6400000000),
		OUint32:    vendor.Uint32(32),
		OUint32Str: vendor.Uint32(32),
		OUint64:    vendor.Uint64(6400000000),
		OUint64Str: vendor.Uint64(6400000000),
		OSint32:    vendor.Int32(-13),
		OSint32Str: vendor.Int32(-13),
		OSint64:    vendor.Int64(-2600000000),
		OSint64Str: vendor.Int64(-2600000000),
		OFloat:     vendor.Float32(3.14),
		OFloatStr:  vendor.Float32(3.14),
		ODouble:    vendor.Float64(6.02214179e23),
		ODoubleStr: vendor.Float64(6.02214179e23),
		OBool:      vendor.Bool(true),
		OString:    vendor.String("hello \"there\""),
		OBytes:     []byte("beep boop"),
	}

	simpleObjectInputJSON = `{` +
		`"oBool":true,` +
		`"oInt32":-32,` +
		`"oInt32Str":"-32",` +
		`"oInt64":-6400000000,` +
		`"oInt64Str":"-6400000000",` +
		`"oUint32":32,` +
		`"oUint32Str":"32",` +
		`"oUint64":6400000000,` +
		`"oUint64Str":"6400000000",` +
		`"oSint32":-13,` +
		`"oSint32Str":"-13",` +
		`"oSint64":-2600000000,` +
		`"oSint64Str":"-2600000000",` +
		`"oFloat":3.14,` +
		`"oFloatStr":"3.14",` +
		`"oDouble":6.02214179e+23,` +
		`"oDoubleStr":"6.02214179e+23",` +
		`"oString":"hello \"there\"",` +
		`"oBytes":"YmVlcCBib29w"` +
		`}`

	simpleObjectOutputJSON = `{` +
		`"oBool":true,` +
		`"oInt32":-32,` +
		`"oInt32Str":-32,` +
		`"oInt64":"-6400000000",` +
		`"oInt64Str":"-6400000000",` +
		`"oUint32":32,` +
		`"oUint32Str":32,` +
		`"oUint64":"6400000000",` +
		`"oUint64Str":"6400000000",` +
		`"oSint32":-13,` +
		`"oSint32Str":-13,` +
		`"oSint64":"-2600000000",` +
		`"oSint64Str":"-2600000000",` +
		`"oFloat":3.14,` +
		`"oFloatStr":3.14,` +
		`"oDouble":6.02214179e+23,` +
		`"oDoubleStr":6.02214179e+23,` +
		`"oString":"hello \"there\"",` +
		`"oBytes":"YmVlcCBib29w"` +
		`}`

	simpleObjectInputPrettyJSON = `{
  "oBool": true,
  "oInt32": -32,
  "oInt32Str": "-32",
  "oInt64": -6400000000,
  "oInt64Str": "-6400000000",
  "oUint32": 32,
  "oUint32Str": "32",
  "oUint64": 6400000000,
  "oUint64Str": "6400000000",
  "oSint32": -13,
  "oSint32Str": "-13",
  "oSint64": -2600000000,
  "oSint64Str": "-2600000000",
  "oFloat": 3.14,
  "oFloatStr": "3.14",
  "oDouble": 6.02214179e+23,
  "oDoubleStr": "6.02214179e+23",
  "oString": "hello \"there\"",
  "oBytes": "YmVlcCBib29w"
}`

	simpleObjectOutputPrettyJSON = `{
  "oBool": true,
  "oInt32": -32,
  "oInt32Str": -32,
  "oInt64": "-6400000000",
  "oInt64Str": "-6400000000",
  "oUint32": 32,
  "oUint32Str": 32,
  "oUint64": "6400000000",
  "oUint64Str": "6400000000",
  "oSint32": -13,
  "oSint32Str": -13,
  "oSint64": "-2600000000",
  "oSint64Str": "-2600000000",
  "oFloat": 3.14,
  "oFloatStr": 3.14,
  "oDouble": 6.02214179e+23,
  "oDoubleStr": 6.02214179e+23,
  "oString": "hello \"there\"",
  "oBytes": "YmVlcCBib29w"
}`

	repeatsObject = &vendor.Repeats{
		RBool:   []bool{true, false, true},
		RInt32:  []int32{-3, -4, -5},
		RInt64:  []int64{-123456789, -987654321},
		RUint32: []uint32{1, 2, 3},
		RUint64: []uint64{6789012345, 3456789012},
		RSint32: []int32{-1, -2, -3},
		RSint64: []int64{-6789012345, -3456789012},
		RFloat:  []float32{3.14, 6.28},
		RDouble: []float64{299792458 * 1e20, 6.62606957e-34},
		RString: []string{"happy", "days"},
		RBytes:  [][]byte{[]byte("skittles"), []byte("m&m's")},
	}

	repeatsObjectJSON = `{` +
		`"rBool":[true,false,true],` +
		`"rInt32":[-3,-4,-5],` +
		`"rInt64":["-123456789","-987654321"],` +
		`"rUint32":[1,2,3],` +
		`"rUint64":["6789012345","3456789012"],` +
		`"rSint32":[-1,-2,-3],` +
		`"rSint64":["-6789012345","-3456789012"],` +
		`"rFloat":[3.14,6.28],` +
		`"rDouble":[2.99792458e+28,6.62606957e-34],` +
		`"rString":["happy","days"],` +
		`"rBytes":["c2tpdHRsZXM=","bSZtJ3M="]` +
		`}`

	repeatsObjectPrettyJSON = `{
  "rBool": [
    true,
    false,
    true
  ],
  "rInt32": [
    -3,
    -4,
    -5
  ],
  "rInt64": [
    "-123456789",
    "-987654321"
  ],
  "rUint32": [
    1,
    2,
    3
  ],
  "rUint64": [
    "6789012345",
    "3456789012"
  ],
  "rSint32": [
    -1,
    -2,
    -3
  ],
  "rSint64": [
    "-6789012345",
    "-3456789012"
  ],
  "rFloat": [
    3.14,
    6.28
  ],
  "rDouble": [
    2.99792458e+28,
    6.62606957e-34
  ],
  "rString": [
    "happy",
    "days"
  ],
  "rBytes": [
    "c2tpdHRsZXM=",
    "bSZtJ3M="
  ]
}`

	innerSimple   = &vendor.Simple{OInt32: vendor.Int32(-32)}
	innerSimple2  = &vendor.Simple{OInt64: vendor.Int64(25)}
	innerRepeats  = &vendor.Repeats{RString: []string{"roses", "red"}}
	innerRepeats2 = &vendor.Repeats{RString: []string{"violets", "blue"}}
	complexObject = &vendor.Widget{
		Color:    vendor.Widget_GREEN.Enum(),
		RColor:   []vendor.Widget_Color{vendor.Widget_RED, vendor.Widget_GREEN, vendor.Widget_BLUE},
		Simple:   innerSimple,
		RSimple:  []*vendor.Simple{innerSimple, innerSimple2},
		Repeats:  innerRepeats,
		RRepeats: []*vendor.Repeats{innerRepeats, innerRepeats2},
	}

	complexObjectJSON = `{"color":"GREEN",` +
		`"rColor":["RED","GREEN","BLUE"],` +
		`"simple":{"oInt32":-32},` +
		`"rSimple":[{"oInt32":-32},{"oInt64":"25"}],` +
		`"repeats":{"rString":["roses","red"]},` +
		`"rRepeats":[{"rString":["roses","red"]},{"rString":["violets","blue"]}]` +
		`}`

	complexObjectPrettyJSON = `{
  "color": "GREEN",
  "rColor": [
    "RED",
    "GREEN",
    "BLUE"
  ],
  "simple": {
    "oInt32": -32
  },
  "rSimple": [
    {
      "oInt32": -32
    },
    {
      "oInt64": "25"
    }
  ],
  "repeats": {
    "rString": [
      "roses",
      "red"
    ]
  },
  "rRepeats": [
    {
      "rString": [
        "roses",
        "red"
      ]
    },
    {
      "rString": [
        "violets",
        "blue"
      ]
    }
  ]
}`

	colorPrettyJSON = `{
 "color": 2
}`

	colorListPrettyJSON = `{
  "color": 1000,
  "rColor": [
    "RED"
  ]
}`

	nummyPrettyJSON = `{
  "nummy": {
    "1": 2,
    "3": 4
  }
}`

	objjyPrettyJSON = `{
  "objjy": {
    "1": {
      "dub": 1
    }
  }
}`
	realNumber     = &vendor.Real{Value: vendor.Float64(3.14159265359)}
	realNumberName = "Pi"
	complexNumber  = &vendor.Complex{Imaginary: vendor.Float64(0.5772156649)}
	realNumberJSON = `{` +
		`"value":3.14159265359,` +
		`"[jsonpb.Complex.real_extension]":{"imaginary":0.5772156649},` +
		`"[jsonpb.name]":"Pi"` +
		`}`

	anySimple = &vendor.KnownTypes{
		An: &vendor.Any{
			TypeUrl: "something.example.com/jsonpb.Simple",
			Value: []byte{
				// &pb.Simple{OBool:true}
				1 << 3, 1,
			},
		},
	}
	anySimpleJSON       = `{"an":{"@type":"something.example.com/jsonpb.Simple","oBool":true}}`
	anySimplePrettyJSON = `{
  "an": {
    "@type": "something.example.com/jsonpb.Simple",
    "oBool": true
  }
}`

	anyWellKnown = &vendor.KnownTypes{
		An: &vendor.Any{
			TypeUrl: "type.googleapis.com/google.protobuf.Duration",
			Value: []byte{
				// &durpb.Duration{Seconds: 1, Nanos: 212000000 }
				1 << 3, 1, // seconds
				2 << 3, 0x80, 0xba, 0x8b, 0x65, // nanos
			},
		},
	}
	anyWellKnownJSON       = `{"an":{"@type":"type.googleapis.com/google.protobuf.Duration","value":"1.212s"}}`
	anyWellKnownPrettyJSON = `{
  "an": {
    "@type": "type.googleapis.com/google.protobuf.Duration",
    "value": "1.212s"
  }
}`

	nonFinites = &vendor.NonFinites{
		FNan:  vendor.Float32(float32(math.NaN())),
		FPinf: vendor.Float32(float32(math.Inf(1))),
		FNinf: vendor.Float32(float32(math.Inf(-1))),
		DNan:  vendor.Float64(float64(math.NaN())),
		DPinf: vendor.Float64(float64(math.Inf(1))),
		DNinf: vendor.Float64(float64(math.Inf(-1))),
	}
	nonFinitesJSON = `{` +
		`"fNan":"NaN",` +
		`"fPinf":"Infinity",` +
		`"fNinf":"-Infinity",` +
		`"dNan":"NaN",` +
		`"dPinf":"Infinity",` +
		`"dNinf":"-Infinity"` +
		`}`
)

func init() {
	if err := vendor.SetExtension(realNumber, vendor.E_Name, &realNumberName); err != nil {
		panic(err)
	}
	if err := vendor.SetExtension(realNumber, vendor.E_Complex_RealExtension, complexNumber); err != nil {
		panic(err)
	}
}

var marshalingTests = []struct {
	desc      string
	marshaler vendor.Marshaler
	pb        vendor.Message
	json      string
}{
	{"simple flat object", marshaler, simpleObject, simpleObjectOutputJSON},
	{"simple pretty object", marshalerAllOptions, simpleObject, simpleObjectOutputPrettyJSON},
	{"non-finite floats fields object", marshaler, nonFinites, nonFinitesJSON},
	{"repeated fields flat object", marshaler, repeatsObject, repeatsObjectJSON},
	{"repeated fields pretty object", marshalerAllOptions, repeatsObject, repeatsObjectPrettyJSON},
	{"nested message/enum flat object", marshaler, complexObject, complexObjectJSON},
	{"nested message/enum pretty object", marshalerAllOptions, complexObject, complexObjectPrettyJSON},
	{"enum-string flat object", vendor.Marshaler{},
		&vendor.Widget{Color: vendor.Widget_BLUE.Enum()}, `{"color":"BLUE"}`},
	{"enum-value pretty object", vendor.Marshaler{EnumsAsInts: true, Indent: " "},
		&vendor.Widget{Color: vendor.Widget_BLUE.Enum()}, colorPrettyJSON},
	{"unknown enum value object", marshalerAllOptions,
		&vendor.Widget{Color: vendor.Widget_Color(1000).Enum(), RColor: []vendor.Widget_Color{vendor.Widget_RED}}, colorListPrettyJSON},
	{"repeated proto3 enum", vendor.Marshaler{},
		&vendor.Message{RFunny: []vendor.Message_Humour{
			vendor.Message_PUNS,
			vendor.Message_SLAPSTICK,
		}},
		`{"rFunny":["PUNS","SLAPSTICK"]}`},
	{"repeated proto3 enum as int", vendor.Marshaler{EnumsAsInts: true},
		&vendor.Message{RFunny: []vendor.Message_Humour{
			vendor.Message_PUNS,
			vendor.Message_SLAPSTICK,
		}},
		`{"rFunny":[1,2]}`},
	{"empty value", marshaler, &vendor.Simple3{}, `{}`},
	{"empty value emitted", vendor.Marshaler{EmitDefaults: true}, &vendor.Simple3{}, `{"dub":0}`},
	{"empty repeated emitted", vendor.Marshaler{EmitDefaults: true}, &vendor.SimpleSlice3{}, `{"slices":[]}`},
	{"empty map emitted", vendor.Marshaler{EmitDefaults: true}, &vendor.SimpleMap3{}, `{"stringy":{}}`},
	{"nested struct null", vendor.Marshaler{EmitDefaults: true}, &vendor.SimpleNull3{}, `{"simple":null}`},
	{"map<int64, int32>", marshaler, &vendor.Mappy{Nummy: map[int64]int32{1: 2, 3: 4}}, `{"nummy":{"1":2,"3":4}}`},
	{"map<int64, int32>", marshalerAllOptions, &vendor.Mappy{Nummy: map[int64]int32{1: 2, 3: 4}}, nummyPrettyJSON},
	{"map<string, string>", marshaler,
		&vendor.Mappy{Strry: map[string]string{`"one"`: "two", "three": "four"}},
		`{"strry":{"\"one\"":"two","three":"four"}}`},
	{"map<int32, Object>", marshaler,
		&vendor.Mappy{Objjy: map[int32]*vendor.Simple3{1: {Dub: 1}}}, `{"objjy":{"1":{"dub":1}}}`},
	{"map<int32, Object>", marshalerAllOptions,
		&vendor.Mappy{Objjy: map[int32]*vendor.Simple3{1: {Dub: 1}}}, objjyPrettyJSON},
	{"map<int64, string>", marshaler, &vendor.Mappy{Buggy: map[int64]string{1234: "yup"}},
		`{"buggy":{"1234":"yup"}}`},
	{"map<bool, bool>", marshaler, &vendor.Mappy{Booly: map[bool]bool{false: true}}, `{"booly":{"false":true}}`},
	{"map<string, enum>", marshaler, &vendor.Mappy{Enumy: map[string]vendor.Numeral{"XIV": vendor.Numeral_ROMAN}}, `{"enumy":{"XIV":"ROMAN"}}`},
	{"map<string, enum as int>", vendor.Marshaler{EnumsAsInts: true}, &vendor.Mappy{Enumy: map[string]vendor.Numeral{"XIV": vendor.Numeral_ROMAN}}, `{"enumy":{"XIV":2}}`},
	{"map<int32, bool>", marshaler, &vendor.Mappy{S32Booly: map[int32]bool{1: true, 3: false, 10: true, 12: false}}, `{"s32booly":{"1":true,"3":false,"10":true,"12":false}}`},
	{"map<int64, bool>", marshaler, &vendor.Mappy{S64Booly: map[int64]bool{1: true, 3: false, 10: true, 12: false}}, `{"s64booly":{"1":true,"3":false,"10":true,"12":false}}`},
	{"map<uint32, bool>", marshaler, &vendor.Mappy{U32Booly: map[uint32]bool{1: true, 3: false, 10: true, 12: false}}, `{"u32booly":{"1":true,"3":false,"10":true,"12":false}}`},
	{"map<uint64, bool>", marshaler, &vendor.Mappy{U64Booly: map[uint64]bool{1: true, 3: false, 10: true, 12: false}}, `{"u64booly":{"1":true,"3":false,"10":true,"12":false}}`},
	{"proto2 map<int64, string>", marshaler, &vendor.Maps{MInt64Str: map[int64]string{213: "cat"}},
		`{"mInt64Str":{"213":"cat"}}`},
	{"proto2 map<bool, Object>", marshaler,
		&vendor.Maps{MBoolSimple: map[bool]*vendor.Simple{true: {OInt32: vendor.Int32(1)}}},
		`{"mBoolSimple":{"true":{"oInt32":1}}}`},
	{"oneof, not set", marshaler, &vendor.MsgWithOneof{}, `{}`},
	{"oneof, set", marshaler, &vendor.MsgWithOneof{Union: &vendor.MsgWithOneof_Title{"Grand Poobah"}}, `{"title":"Grand Poobah"}`},
	{"force orig_name", vendor.Marshaler{OrigName: true}, &vendor.Simple{OInt32: vendor.Int32(4)},
		`{"o_int32":4}`},
	{"proto2 extension", marshaler, realNumber, realNumberJSON},
	{"Any with message", marshaler, anySimple, anySimpleJSON},
	{"Any with message and indent", marshalerAllOptions, anySimple, anySimplePrettyJSON},
	{"Any with WKT", marshaler, anyWellKnown, anyWellKnownJSON},
	{"Any with WKT and indent", marshalerAllOptions, anyWellKnown, anyWellKnownPrettyJSON},
	{"Duration", marshaler, &vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 3}}, `{"dur":"3s"}`},
	{"Duration", marshaler, &vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 3, Nanos: 1e6}}, `{"dur":"3.001s"}`},
	{"Duration beyond float64 precision", marshaler, &vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 100000000, Nanos: 1}}, `{"dur":"100000000.000000001s"}`},
	{"negative Duration", marshaler, &vendor.KnownTypes{Dur: &vendor.Duration{Seconds: -123, Nanos: -456}}, `{"dur":"-123.000000456s"}`},
	{"Struct", marshaler, &vendor.KnownTypes{St: &vendor.Struct{
		Fields: map[string]*vendor.Value{
			"one": {Kind: &vendor.Value_StringValue{"loneliest number"}},
			"two": {Kind: &vendor.Value_NullValue{vendor.NullValue_NULL_VALUE}},
		},
	}}, `{"st":{"one":"loneliest number","two":null}}`},
	{"empty ListValue", marshaler, &vendor.KnownTypes{Lv: &vendor.ListValue{}}, `{"lv":[]}`},
	{"basic ListValue", marshaler, &vendor.KnownTypes{Lv: &vendor.ListValue{Values: []*vendor.Value{
		{Kind: &vendor.Value_StringValue{"x"}},
		{Kind: &vendor.Value_NullValue{}},
		{Kind: &vendor.Value_NumberValue{3}},
		{Kind: &vendor.Value_BoolValue{true}},
	}}}, `{"lv":["x",null,3,true]}`},
	{"Timestamp", marshaler, &vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: 14e8, Nanos: 21e6}}, `{"ts":"2014-05-13T16:53:20.021Z"}`},
	{"Timestamp", marshaler, &vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: 14e8, Nanos: 0}}, `{"ts":"2014-05-13T16:53:20Z"}`},
	{"number Value", marshaler, &vendor.KnownTypes{Val: &vendor.Value{Kind: &vendor.Value_NumberValue{1}}}, `{"val":1}`},
	{"null Value", marshaler, &vendor.KnownTypes{Val: &vendor.Value{Kind: &vendor.Value_NullValue{vendor.NullValue_NULL_VALUE}}}, `{"val":null}`},
	{"string number value", marshaler, &vendor.KnownTypes{Val: &vendor.Value{Kind: &vendor.Value_StringValue{"9223372036854775807"}}}, `{"val":"9223372036854775807"}`},
	{"list of lists Value", marshaler, &vendor.KnownTypes{Val: &vendor.Value{
		Kind: &vendor.Value_ListValue{&vendor.ListValue{
			Values: []*vendor.Value{
				{Kind: &vendor.Value_StringValue{"x"}},
				{Kind: &vendor.Value_ListValue{&vendor.ListValue{
					Values: []*vendor.Value{
						{Kind: &vendor.Value_ListValue{&vendor.ListValue{
							Values: []*vendor.Value{{Kind: &vendor.Value_StringValue{"y"}}},
						}}},
						{Kind: &vendor.Value_StringValue{"z"}},
					},
				}}},
			},
		}},
	}}, `{"val":["x",[["y"],"z"]]}`},

	{"DoubleValue", marshaler, &vendor.KnownTypes{Dbl: &vendor.DoubleValue{Value: 1.2}}, `{"dbl":1.2}`},
	{"FloatValue", marshaler, &vendor.KnownTypes{Flt: &vendor.FloatValue{Value: 1.2}}, `{"flt":1.2}`},
	{"Int64Value", marshaler, &vendor.KnownTypes{I64: &vendor.Int64Value{Value: -3}}, `{"i64":"-3"}`},
	{"UInt64Value", marshaler, &vendor.KnownTypes{U64: &vendor.UInt64Value{Value: 3}}, `{"u64":"3"}`},
	{"Int32Value", marshaler, &vendor.KnownTypes{I32: &vendor.Int32Value{Value: -4}}, `{"i32":-4}`},
	{"UInt32Value", marshaler, &vendor.KnownTypes{U32: &vendor.UInt32Value{Value: 4}}, `{"u32":4}`},
	{"BoolValue", marshaler, &vendor.KnownTypes{Bool: &vendor.BoolValue{Value: true}}, `{"bool":true}`},
	{"StringValue", marshaler, &vendor.KnownTypes{Str: &vendor.StringValue{Value: "plush"}}, `{"str":"plush"}`},
	{"BytesValue", marshaler, &vendor.KnownTypes{Bytes: &vendor.BytesValue{Value: []byte("wow")}}, `{"bytes":"d293"}`},

	{"required", marshaler, &vendor.MsgWithRequired{Str: vendor.String("hello")}, `{"str":"hello"}`},
	{"required bytes", marshaler, &vendor.MsgWithRequiredBytes{Byts: []byte{}}, `{"byts":""}`},
}

func TestMarshaling(t *testing.T) {
	for _, tt := range marshalingTests {
		json, err := tt.marshaler.MarshalToString(tt.pb)
		if err != nil {
			t.Errorf("%s: marshaling error: %v", tt.desc, err)
		} else if tt.json != json {
			t.Errorf("%s: got [%v] want [%v]", tt.desc, json, tt.json)
		}
	}
}

func TestMarshalingNil(t *testing.T) {
	var msg *vendor.Simple
	m := &vendor.Marshaler{}
	if _, err := m.MarshalToString(msg); err == nil {
		t.Errorf("mashaling nil returned no error")
	}
}

func TestMarshalIllegalTime(t *testing.T) {
	tests := []struct {
		pb   vendor.Message
		fail bool
	}{
		{&vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 1, Nanos: 0}}, false},
		{&vendor.KnownTypes{Dur: &vendor.Duration{Seconds: -1, Nanos: 0}}, false},
		{&vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 1, Nanos: -1}}, true},
		{&vendor.KnownTypes{Dur: &vendor.Duration{Seconds: -1, Nanos: 1}}, true},
		{&vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 1, Nanos: 1000000000}}, true},
		{&vendor.KnownTypes{Dur: &vendor.Duration{Seconds: -1, Nanos: -1000000000}}, true},
		{&vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: 1, Nanos: 1}}, false},
		{&vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: 1, Nanos: -1}}, true},
		{&vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: 1, Nanos: 1000000000}}, true},
	}
	for _, tt := range tests {
		_, err := marshaler.MarshalToString(tt.pb)
		if err == nil && tt.fail {
			t.Errorf("marshaler.MarshalToString(%v) = _, <nil>; want _, <non-nil>", tt.pb)
		}
		if err != nil && !tt.fail {
			t.Errorf("marshaler.MarshalToString(%v) = _, %v; want _, <nil>", tt.pb, err)
		}
	}
}

func TestMarshalJSONPBMarshaler(t *testing.T) {
	rawJson := `{ "foo": "bar", "baz": [0, 1, 2, 3] }`
	msg := dynamicMessage{RawJson: rawJson}
	str, err := new(vendor.Marshaler).MarshalToString(&msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshalling JSONPBMarshaler: %v", err)
	}
	if str != rawJson {
		t.Errorf("marshalling JSON produced incorrect output: got %s, wanted %s", str, rawJson)
	}
}

func TestMarshalAnyJSONPBMarshaler(t *testing.T) {
	msg := dynamicMessage{RawJson: `{ "foo": "bar", "baz": [0, 1, 2, 3] }`}
	a, err := vendor.MarshalAny(&msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshalling to Any: %v", err)
	}
	str, err := new(vendor.Marshaler).MarshalToString(a)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshalling Any to JSON: %v", err)
	}
	// after custom marshaling, it's round-tripped through JSON decoding/encoding already,
	// so the keys are sorted, whitespace is compacted, and "@type" key has been added
	expected := `{"@type":"type.googleapis.com/` + dynamicMessageName + `","baz":[0,1,2,3],"foo":"bar"}`
	if str != expected {
		t.Errorf("marshalling JSON produced incorrect output: got %s, wanted %s", str, expected)
	}

	// Do it again, but this time with indentation:

	marshaler := vendor.Marshaler{Indent: "  "}
	str, err = marshaler.MarshalToString(a)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshalling Any to JSON: %v", err)
	}
	// same as expected above, but pretty-printed w/ indentation
	expected =
`{
  "@type": "type.googleapis.com/` + dynamicMessageName + `",
  "baz": [
    0,
    1,
    2,
    3
  ],
  "foo": "bar"
}`
	if str != expected {
		t.Errorf("marshalling JSON produced incorrect output: got %s, wanted %s", str, expected)
	}
}


func TestMarshalWithCustomValidation(t *testing.T) {
	msg := dynamicMessage{RawJson: `{ "foo": "bar", "baz": [0, 1, 2, 3] }`, Dummy: &dynamicMessage{}}

	js, err := new(vendor.Marshaler).MarshalToString(&msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshalling to json: %v", err)
	}
	err = vendor.Unmarshal(strings.NewReader(js), &msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when unmarshalling from json: %v", err)
	}
}

// Test marshaling message containing unset required fields should produce error.
func TestMarshalUnsetRequiredFields(t *testing.T) {
	msgExt := &vendor.Real{}
	vendor.SetExtension(msgExt, vendor.E_Extm, &vendor.MsgWithRequired{})

	tests := []struct {
		desc      string
		marshaler *vendor.Marshaler
		pb        vendor.Message
	}{
		{
			desc:      "direct required field",
			marshaler: &vendor.Marshaler{},
			pb:        &vendor.MsgWithRequired{},
		},
		{
			desc:      "direct required field + emit defaults",
			marshaler: &vendor.Marshaler{EmitDefaults: true},
			pb:        &vendor.MsgWithRequired{},
		},
		{
			desc:      "indirect required field",
			marshaler: &vendor.Marshaler{},
			pb:        &vendor.MsgWithIndirectRequired{Subm: &vendor.MsgWithRequired{}},
		},
		{
			desc:      "indirect required field + emit defaults",
			marshaler: &vendor.Marshaler{EmitDefaults: true},
			pb:        &vendor.MsgWithIndirectRequired{Subm: &vendor.MsgWithRequired{}},
		},
		{
			desc:      "direct required wkt field",
			marshaler: &vendor.Marshaler{},
			pb:        &vendor.MsgWithRequiredWKT{},
		},
		{
			desc:      "direct required wkt field + emit defaults",
			marshaler: &vendor.Marshaler{EmitDefaults: true},
			pb:        &vendor.MsgWithRequiredWKT{},
		},
		{
			desc:      "direct required bytes field",
			marshaler: &vendor.Marshaler{},
			pb:        &vendor.MsgWithRequiredBytes{},
		},
		{
			desc:      "required in map value",
			marshaler: &vendor.Marshaler{},
			pb: &vendor.MsgWithIndirectRequired{
				MapField: map[string]*vendor.MsgWithRequired{
					"key": {},
				},
			},
		},
		{
			desc:      "required in repeated item",
			marshaler: &vendor.Marshaler{},
			pb: &vendor.MsgWithIndirectRequired{
				SliceField: []*vendor.MsgWithRequired{
					{Str: vendor.String("hello")},
					{},
				},
			},
		},
		{
			desc:      "required inside oneof",
			marshaler: &vendor.Marshaler{},
			pb: &vendor.MsgWithOneof{
				Union: &vendor.MsgWithOneof_MsgWithRequired{&vendor.MsgWithRequired{}},
			},
		},
		{
			desc:      "required inside extension",
			marshaler: &vendor.Marshaler{},
			pb:        msgExt,
		},
	}

	for _, tc := range tests {
		if _, err := tc.marshaler.MarshalToString(tc.pb); err == nil {
			t.Errorf("%s: expecting error in marshaling with unset required fields %+v", tc.desc, tc.pb)
		}
	}
}

var unmarshalingTests = []struct {
	desc        string
	unmarshaler vendor.Unmarshaler
	json        string
	pb          vendor.Message
}{
	{"simple flat object", vendor.Unmarshaler{}, simpleObjectInputJSON, simpleObject},
	{"simple pretty object", vendor.Unmarshaler{}, simpleObjectInputPrettyJSON, simpleObject},
	{"repeated fields flat object", vendor.Unmarshaler{}, repeatsObjectJSON, repeatsObject},
	{"repeated fields pretty object", vendor.Unmarshaler{}, repeatsObjectPrettyJSON, repeatsObject},
	{"nested message/enum flat object", vendor.Unmarshaler{}, complexObjectJSON, complexObject},
	{"nested message/enum pretty object", vendor.Unmarshaler{}, complexObjectPrettyJSON, complexObject},
	{"enum-string object", vendor.Unmarshaler{}, `{"color":"BLUE"}`, &vendor.Widget{Color: vendor.Widget_BLUE.Enum()}},
	{"enum-value object", vendor.Unmarshaler{}, "{\n \"color\": 2\n}", &vendor.Widget{Color: vendor.Widget_BLUE.Enum()}},
	{"unknown field with allowed option", vendor.Unmarshaler{AllowUnknownFields: true}, `{"unknown": "foo"}`, new(vendor.Simple)},
	{"proto3 enum string", vendor.Unmarshaler{}, `{"hilarity":"PUNS"}`, &vendor.Message{Hilarity: vendor.Message_PUNS}},
	{"proto3 enum value", vendor.Unmarshaler{}, `{"hilarity":1}`, &vendor.Message{Hilarity: vendor.Message_PUNS}},
	{"unknown enum value object",
		vendor.Unmarshaler{},
		"{\n  \"color\": 1000,\n  \"r_color\": [\n    \"RED\"\n  ]\n}",
		&vendor.Widget{Color: vendor.Widget_Color(1000).Enum(), RColor: []vendor.Widget_Color{vendor.Widget_RED}}},
	{"repeated proto3 enum", vendor.Unmarshaler{}, `{"rFunny":["PUNS","SLAPSTICK"]}`,
		&vendor.Message{RFunny: []vendor.Message_Humour{
			vendor.Message_PUNS,
			vendor.Message_SLAPSTICK,
		}}},
	{"repeated proto3 enum as int", vendor.Unmarshaler{}, `{"rFunny":[1,2]}`,
		&vendor.Message{RFunny: []vendor.Message_Humour{
			vendor.Message_PUNS,
			vendor.Message_SLAPSTICK,
		}}},
	{"repeated proto3 enum as mix of strings and ints", vendor.Unmarshaler{}, `{"rFunny":["PUNS",2]}`,
		&vendor.Message{RFunny: []vendor.Message_Humour{
			vendor.Message_PUNS,
			vendor.Message_SLAPSTICK,
		}}},
	{"unquoted int64 object", vendor.Unmarshaler{}, `{"oInt64":-314}`, &vendor.Simple{OInt64: vendor.Int64(-314)}},
	{"unquoted uint64 object", vendor.Unmarshaler{}, `{"oUint64":123}`, &vendor.Simple{OUint64: vendor.Uint64(123)}},
	{"NaN", vendor.Unmarshaler{}, `{"oDouble":"NaN"}`, &vendor.Simple{ODouble: vendor.Float64(math.NaN())}},
	{"Inf", vendor.Unmarshaler{}, `{"oFloat":"Infinity"}`, &vendor.Simple{OFloat: vendor.Float32(float32(math.Inf(1)))}},
	{"-Inf", vendor.Unmarshaler{}, `{"oDouble":"-Infinity"}`, &vendor.Simple{ODouble: vendor.Float64(math.Inf(-1))}},
	{"map<int64, int32>", vendor.Unmarshaler{}, `{"nummy":{"1":2,"3":4}}`, &vendor.Mappy{Nummy: map[int64]int32{1: 2, 3: 4}}},
	{"map<string, string>", vendor.Unmarshaler{}, `{"strry":{"\"one\"":"two","three":"four"}}`, &vendor.Mappy{Strry: map[string]string{`"one"`: "two", "three": "four"}}},
	{"map<int32, Object>", vendor.Unmarshaler{}, `{"objjy":{"1":{"dub":1}}}`, &vendor.Mappy{Objjy: map[int32]*vendor.Simple3{1: {Dub: 1}}}},
	{"proto2 extension", vendor.Unmarshaler{}, realNumberJSON, realNumber},
	{"Any with message", vendor.Unmarshaler{}, anySimpleJSON, anySimple},
	{"Any with message and indent", vendor.Unmarshaler{}, anySimplePrettyJSON, anySimple},
	{"Any with WKT", vendor.Unmarshaler{}, anyWellKnownJSON, anyWellKnown},
	{"Any with WKT and indent", vendor.Unmarshaler{}, anyWellKnownPrettyJSON, anyWellKnown},
	{"map<string, enum>", vendor.Unmarshaler{}, `{"enumy":{"XIV":"ROMAN"}}`, &vendor.Mappy{Enumy: map[string]vendor.Numeral{"XIV": vendor.Numeral_ROMAN}}},
	{"map<string, enum as int>", vendor.Unmarshaler{}, `{"enumy":{"XIV":2}}`, &vendor.Mappy{Enumy: map[string]vendor.Numeral{"XIV": vendor.Numeral_ROMAN}}},
	{"oneof", vendor.Unmarshaler{}, `{"salary":31000}`, &vendor.MsgWithOneof{Union: &vendor.MsgWithOneof_Salary{31000}}},
	{"oneof spec name", vendor.Unmarshaler{}, `{"Country":"Australia"}`, &vendor.MsgWithOneof{Union: &vendor.MsgWithOneof_Country{"Australia"}}},
	{"oneof orig_name", vendor.Unmarshaler{}, `{"Country":"Australia"}`, &vendor.MsgWithOneof{Union: &vendor.MsgWithOneof_Country{"Australia"}}},
	{"oneof spec name2", vendor.Unmarshaler{}, `{"homeAddress":"Australia"}`, &vendor.MsgWithOneof{Union: &vendor.MsgWithOneof_HomeAddress{"Australia"}}},
	{"oneof orig_name2", vendor.Unmarshaler{}, `{"home_address":"Australia"}`, &vendor.MsgWithOneof{Union: &vendor.MsgWithOneof_HomeAddress{"Australia"}}},
	{"orig_name input", vendor.Unmarshaler{}, `{"o_bool":true}`, &vendor.Simple{OBool: vendor.Bool(true)}},
	{"camelName input", vendor.Unmarshaler{}, `{"oBool":true}`, &vendor.Simple{OBool: vendor.Bool(true)}},

	{"Duration", vendor.Unmarshaler{}, `{"dur":"3.000s"}`, &vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 3}}},
	{"Duration", vendor.Unmarshaler{}, `{"dur":"4s"}`, &vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 4}}},
	{"Duration with unicode", vendor.Unmarshaler{}, `{"dur": "3\u0073"}`, &vendor.KnownTypes{Dur: &vendor.Duration{Seconds: 3}}},
	{"null Duration", vendor.Unmarshaler{}, `{"dur":null}`, &vendor.KnownTypes{Dur: nil}},
	{"Timestamp", vendor.Unmarshaler{}, `{"ts":"2014-05-13T16:53:20.021Z"}`, &vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: 14e8, Nanos: 21e6}}},
	{"Timestamp", vendor.Unmarshaler{}, `{"ts":"2014-05-13T16:53:20Z"}`, &vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: 14e8, Nanos: 0}}},
	{"Timestamp with unicode", vendor.Unmarshaler{}, `{"ts": "2014-05-13T16:53:20\u005a"}`, &vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: 14e8, Nanos: 0}}},
	{"PreEpochTimestamp", vendor.Unmarshaler{}, `{"ts":"1969-12-31T23:59:58.999999995Z"}`, &vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: -2, Nanos: 999999995}}},
	{"ZeroTimeTimestamp", vendor.Unmarshaler{}, `{"ts":"0001-01-01T00:00:00Z"}`, &vendor.KnownTypes{Ts: &vendor.Timestamp{Seconds: -62135596800, Nanos: 0}}},
	{"null Timestamp", vendor.Unmarshaler{}, `{"ts":null}`, &vendor.KnownTypes{Ts: nil}},
	{"null Struct", vendor.Unmarshaler{}, `{"st": null}`, &vendor.KnownTypes{St: nil}},
	{"empty Struct", vendor.Unmarshaler{}, `{"st": {}}`, &vendor.KnownTypes{St: &vendor.Struct{}}},
	{"basic Struct", vendor.Unmarshaler{}, `{"st": {"a": "x", "b": null, "c": 3, "d": true}}`, &vendor.KnownTypes{St: &vendor.Struct{Fields: map[string]*vendor.Value{
		"a": {Kind: &vendor.Value_StringValue{"x"}},
		"b": {Kind: &vendor.Value_NullValue{}},
		"c": {Kind: &vendor.Value_NumberValue{3}},
		"d": {Kind: &vendor.Value_BoolValue{true}},
	}}}},
	{"nested Struct", vendor.Unmarshaler{}, `{"st": {"a": {"b": 1, "c": [{"d": true}, "f"]}}}`, &vendor.KnownTypes{St: &vendor.Struct{Fields: map[string]*vendor.Value{
		"a": {Kind: &vendor.Value_StructValue{&vendor.Struct{Fields: map[string]*vendor.Value{
			"b": {Kind: &vendor.Value_NumberValue{1}},
			"c": {Kind: &vendor.Value_ListValue{&vendor.ListValue{Values: []*vendor.Value{
				{Kind: &vendor.Value_StructValue{&vendor.Struct{Fields: map[string]*vendor.Value{"d": {Kind: &vendor.Value_BoolValue{true}}}}}},
				{Kind: &vendor.Value_StringValue{"f"}},
			}}}},
		}}}},
	}}}},
	{"null ListValue", vendor.Unmarshaler{}, `{"lv": null}`, &vendor.KnownTypes{Lv: nil}},
	{"empty ListValue", vendor.Unmarshaler{}, `{"lv": []}`, &vendor.KnownTypes{Lv: &vendor.ListValue{}}},
	{"basic ListValue", vendor.Unmarshaler{}, `{"lv": ["x", null, 3, true]}`, &vendor.KnownTypes{Lv: &vendor.ListValue{Values: []*vendor.Value{
		{Kind: &vendor.Value_StringValue{"x"}},
		{Kind: &vendor.Value_NullValue{}},
		{Kind: &vendor.Value_NumberValue{3}},
		{Kind: &vendor.Value_BoolValue{true}},
	}}}},
	{"number Value", vendor.Unmarshaler{}, `{"val":1}`, &vendor.KnownTypes{Val: &vendor.Value{Kind: &vendor.Value_NumberValue{1}}}},
	{"null Value", vendor.Unmarshaler{}, `{"val":null}`, &vendor.KnownTypes{Val: &vendor.Value{Kind: &vendor.Value_NullValue{vendor.NullValue_NULL_VALUE}}}},
	{"bool Value", vendor.Unmarshaler{}, `{"val":true}`, &vendor.KnownTypes{Val: &vendor.Value{Kind: &vendor.Value_BoolValue{true}}}},
	{"string Value", vendor.Unmarshaler{}, `{"val":"x"}`, &vendor.KnownTypes{Val: &vendor.Value{Kind: &vendor.Value_StringValue{"x"}}}},
	{"string number value", vendor.Unmarshaler{}, `{"val":"9223372036854775807"}`, &vendor.KnownTypes{Val: &vendor.Value{Kind: &vendor.Value_StringValue{"9223372036854775807"}}}},
	{"list of lists Value", vendor.Unmarshaler{}, `{"val":["x", [["y"], "z"]]}`, &vendor.KnownTypes{Val: &vendor.Value{
		Kind: &vendor.Value_ListValue{&vendor.ListValue{
			Values: []*vendor.Value{
				{Kind: &vendor.Value_StringValue{"x"}},
				{Kind: &vendor.Value_ListValue{&vendor.ListValue{
					Values: []*vendor.Value{
						{Kind: &vendor.Value_ListValue{&vendor.ListValue{
							Values: []*vendor.Value{{Kind: &vendor.Value_StringValue{"y"}}},
						}}},
						{Kind: &vendor.Value_StringValue{"z"}},
					},
				}}},
			},
		}}}}},

	{"DoubleValue", vendor.Unmarshaler{}, `{"dbl":1.2}`, &vendor.KnownTypes{Dbl: &vendor.DoubleValue{Value: 1.2}}},
	{"FloatValue", vendor.Unmarshaler{}, `{"flt":1.2}`, &vendor.KnownTypes{Flt: &vendor.FloatValue{Value: 1.2}}},
	{"Int64Value", vendor.Unmarshaler{}, `{"i64":"-3"}`, &vendor.KnownTypes{I64: &vendor.Int64Value{Value: -3}}},
	{"UInt64Value", vendor.Unmarshaler{}, `{"u64":"3"}`, &vendor.KnownTypes{U64: &vendor.UInt64Value{Value: 3}}},
	{"Int32Value", vendor.Unmarshaler{}, `{"i32":-4}`, &vendor.KnownTypes{I32: &vendor.Int32Value{Value: -4}}},
	{"UInt32Value", vendor.Unmarshaler{}, `{"u32":4}`, &vendor.KnownTypes{U32: &vendor.UInt32Value{Value: 4}}},
	{"BoolValue", vendor.Unmarshaler{}, `{"bool":true}`, &vendor.KnownTypes{Bool: &vendor.BoolValue{Value: true}}},
	{"StringValue", vendor.Unmarshaler{}, `{"str":"plush"}`, &vendor.KnownTypes{Str: &vendor.StringValue{Value: "plush"}}},
	{"StringValue containing escaped character", vendor.Unmarshaler{}, `{"str":"a\/b"}`, &vendor.KnownTypes{Str: &vendor.StringValue{Value: "a/b"}}},
	{"StructValue containing StringValue's", vendor.Unmarshaler{}, `{"escaped": "a\/b", "unicode": "\u00004E16\u0000754C"}`,
		&vendor.Struct{
			Fields: map[string]*vendor.Value{
				"escaped": {Kind: &vendor.Value_StringValue{"a/b"}},
				"unicode": {Kind: &vendor.Value_StringValue{"\u00004E16\u0000754C"}},
			},
		}},
	{"BytesValue", vendor.Unmarshaler{}, `{"bytes":"d293"}`, &vendor.KnownTypes{Bytes: &vendor.BytesValue{Value: []byte("wow")}}},

	// Ensure that `null` as a value ends up with a nil pointer instead of a [type]Value struct.
	{"null DoubleValue", vendor.Unmarshaler{}, `{"dbl":null}`, &vendor.KnownTypes{Dbl: nil}},
	{"null FloatValue", vendor.Unmarshaler{}, `{"flt":null}`, &vendor.KnownTypes{Flt: nil}},
	{"null Int64Value", vendor.Unmarshaler{}, `{"i64":null}`, &vendor.KnownTypes{I64: nil}},
	{"null UInt64Value", vendor.Unmarshaler{}, `{"u64":null}`, &vendor.KnownTypes{U64: nil}},
	{"null Int32Value", vendor.Unmarshaler{}, `{"i32":null}`, &vendor.KnownTypes{I32: nil}},
	{"null UInt32Value", vendor.Unmarshaler{}, `{"u32":null}`, &vendor.KnownTypes{U32: nil}},
	{"null BoolValue", vendor.Unmarshaler{}, `{"bool":null}`, &vendor.KnownTypes{Bool: nil}},
	{"null StringValue", vendor.Unmarshaler{}, `{"str":null}`, &vendor.KnownTypes{Str: nil}},
	{"null BytesValue", vendor.Unmarshaler{}, `{"bytes":null}`, &vendor.KnownTypes{Bytes: nil}},

	{"required", vendor.Unmarshaler{}, `{"str":"hello"}`, &vendor.MsgWithRequired{Str: vendor.String("hello")}},
	{"required bytes", vendor.Unmarshaler{}, `{"byts": []}`, &vendor.MsgWithRequiredBytes{Byts: []byte{}}},
}

func TestUnmarshaling(t *testing.T) {
	for _, tt := range unmarshalingTests {
		// Make a new instance of the type of our expected object.
		p := reflect.New(reflect.TypeOf(tt.pb).Elem()).Interface().(vendor.Message)

		err := tt.unmarshaler.Unmarshal(strings.NewReader(tt.json), p)
		if err != nil {
			t.Errorf("unmarshalling %s: %v", tt.desc, err)
			continue
		}

		// For easier diffs, compare text strings of the protos.
		exp := vendor.MarshalTextString(tt.pb)
		act := vendor.MarshalTextString(p)
		if string(exp) != string(act) {
			t.Errorf("%s: got [%s] want [%s]", tt.desc, act, exp)
		}
	}
}

func TestUnmarshalNullArray(t *testing.T) {
	var repeats vendor.Repeats
	if err := vendor.UnmarshalString(`{"rBool":null}`, &repeats); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repeats, vendor.Repeats{}) {
		t.Errorf("got non-nil fields in [%#v]", repeats)
	}
}

func TestUnmarshalNullObject(t *testing.T) {
	var maps vendor.Maps
	if err := vendor.UnmarshalString(`{"mInt64Str":null}`, &maps); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(maps, vendor.Maps{}) {
		t.Errorf("got non-nil fields in [%#v]", maps)
	}
}

func TestUnmarshalNext(t *testing.T) {
	// We only need to check against a few, not all of them.
	tests := unmarshalingTests[:5]

	// Create a buffer with many concatenated JSON objects.
	var b bytes.Buffer
	for _, tt := range tests {
		b.WriteString(tt.json)
	}

	dec := json.NewDecoder(&b)
	for _, tt := range tests {
		// Make a new instance of the type of our expected object.
		p := reflect.New(reflect.TypeOf(tt.pb).Elem()).Interface().(vendor.Message)

		err := tt.unmarshaler.UnmarshalNext(dec, p)
		if err != nil {
			t.Errorf("%s: %v", tt.desc, err)
			continue
		}

		// For easier diffs, compare text strings of the protos.
		exp := vendor.MarshalTextString(tt.pb)
		act := vendor.MarshalTextString(p)
		if string(exp) != string(act) {
			t.Errorf("%s: got [%s] want [%s]", tt.desc, act, exp)
		}
	}

	p := &vendor.Simple{}
	err := new(vendor.Unmarshaler).UnmarshalNext(dec, p)
	if err != io.EOF {
		t.Errorf("eof: got %v, expected io.EOF", err)
	}
}

var unmarshalingShouldError = []struct {
	desc string
	in   string
	pb   vendor.Message
}{
	{"a value", "666", new(vendor.Simple)},
	{"gibberish", "{adskja123;l23=-=", new(vendor.Simple)},
	{"unknown field", `{"unknown": "foo"}`, new(vendor.Simple)},
	{"unknown enum name", `{"hilarity":"DAVE"}`, new(vendor.Message)},
	{"Duration containing invalid character", `{"dur": "3\U0073"}`, &vendor.KnownTypes{}},
	{"Timestamp containing invalid character", `{"ts": "2014-05-13T16:53:20\U005a"}`, &vendor.KnownTypes{}},
	{"StringValue containing invalid character", `{"str": "\U00004E16\U0000754C"}`, &vendor.KnownTypes{}},
	{"StructValue containing invalid character", `{"str": "\U00004E16\U0000754C"}`, &vendor.Struct{}},
	{"repeated proto3 enum with non array input", `{"rFunny":"PUNS"}`, &vendor.Message{RFunny: []vendor.Message_Humour{}}},
}

func TestUnmarshalingBadInput(t *testing.T) {
	for _, tt := range unmarshalingShouldError {
		err := vendor.UnmarshalString(tt.in, tt.pb)
		if err == nil {
			t.Errorf("an error was expected when parsing %q instead of an object", tt.desc)
		}
	}
}

type funcResolver func(turl string) (vendor.Message, error)

func (fn funcResolver) Resolve(turl string) (vendor.Message, error) {
	return fn(turl)
}

func TestAnyWithCustomResolver(t *testing.T) {
	var resolvedTypeUrls []string
	resolver := funcResolver(func(turl string) (vendor.Message, error) {
		resolvedTypeUrls = append(resolvedTypeUrls, turl)
		return new(vendor.Simple), nil
	})
	msg := &vendor.Simple{
		OBytes:  []byte{1, 2, 3, 4},
		OBool:   vendor.Bool(true),
		OString: vendor.String("foobar"),
		OInt64:  vendor.Int64(1020304),
	}
	msgBytes, err := vendor.Marshal(msg)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshaling message: %v", err)
	}
	// make an Any with a type URL that won't resolve w/out custom resolver
	any := &vendor.Any{
		TypeUrl: "https://foobar.com/some.random.MessageKind",
		Value:   msgBytes,
	}

	m := vendor.Marshaler{AnyResolver: resolver}
	js, err := m.MarshalToString(any)
	if err != nil {
		t.Errorf("an unexpected error occurred when marshaling any to JSON: %v", err)
	}
	if len(resolvedTypeUrls) != 1 {
		t.Errorf("custom resolver was not invoked during marshaling")
	} else if resolvedTypeUrls[0] != "https://foobar.com/some.random.MessageKind" {
		t.Errorf("custom resolver was invoked with wrong URL: got %q, wanted %q", resolvedTypeUrls[0], "https://foobar.com/some.random.MessageKind")
	}
	wanted := `{"@type":"https://foobar.com/some.random.MessageKind","oBool":true,"oInt64":"1020304","oString":"foobar","oBytes":"AQIDBA=="}`
	if js != wanted {
		t.Errorf("marshalling JSON produced incorrect output: got %s, wanted %s", js, wanted)
	}

	u := vendor.Unmarshaler{AnyResolver: resolver}
	roundTrip := &vendor.Any{}
	err = u.Unmarshal(bytes.NewReader([]byte(js)), roundTrip)
	if err != nil {
		t.Errorf("an unexpected error occurred when unmarshaling any from JSON: %v", err)
	}
	if len(resolvedTypeUrls) != 2 {
		t.Errorf("custom resolver was not invoked during marshaling")
	} else if resolvedTypeUrls[1] != "https://foobar.com/some.random.MessageKind" {
		t.Errorf("custom resolver was invoked with wrong URL: got %q, wanted %q", resolvedTypeUrls[1], "https://foobar.com/some.random.MessageKind")
	}
	if !vendor.Equal(any, roundTrip) {
		t.Errorf("message contents not set correctly after unmarshalling JSON: got %s, wanted %s", roundTrip, any)
	}
}

func TestUnmarshalJSONPBUnmarshaler(t *testing.T) {
	rawJson := `{ "foo": "bar", "baz": [0, 1, 2, 3] }`
	var msg dynamicMessage
	if err := vendor.Unmarshal(strings.NewReader(rawJson), &msg); err != nil {
		t.Errorf("an unexpected error occurred when parsing into JSONPBUnmarshaler: %v", err)
	}
	if msg.RawJson != rawJson {
		t.Errorf("message contents not set correctly after unmarshalling JSON: got %s, wanted %s", msg.RawJson, rawJson)
	}
}

func TestUnmarshalNullWithJSONPBUnmarshaler(t *testing.T) {
	rawJson := `{"stringField":null}`
	var ptrFieldMsg ptrFieldMessage
	if err := vendor.Unmarshal(strings.NewReader(rawJson), &ptrFieldMsg); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

	want := ptrFieldMessage{StringField: &stringField{IsSet: true, StringValue: "null"}}
	if !vendor.Equal(&ptrFieldMsg, &want) {
		t.Errorf("unmarshal result StringField: got %v, want %v", ptrFieldMsg, want)
	}
}

func TestUnmarshalAnyJSONPBUnmarshaler(t *testing.T) {
	rawJson := `{ "@type": "blah.com/` + dynamicMessageName + `", "foo": "bar", "baz": [0, 1, 2, 3] }`
	var got vendor.Any
	if err := vendor.Unmarshal(strings.NewReader(rawJson), &got); err != nil {
		t.Errorf("an unexpected error occurred when parsing into JSONPBUnmarshaler: %v", err)
	}

	dm := &dynamicMessage{RawJson: `{"baz":[0,1,2,3],"foo":"bar"}`}
	var want vendor.Any
	if b, err := vendor.Marshal(dm); err != nil {
		t.Errorf("an unexpected error occurred when marshaling message: %v", err)
	} else {
		want.TypeUrl = "blah.com/" + dynamicMessageName
		want.Value = b
	}

	if !vendor.Equal(&got, &want) {
		t.Errorf("message contents not set correctly after unmarshalling JSON: got %v, wanted %v", got, want)
	}
}

const (
	dynamicMessageName = "google.protobuf.jsonpb.testing.dynamicMessage"
)

func init() {
	// we register the custom type below so that we can use it in Any types
	vendor.RegisterType((*dynamicMessage)(nil), dynamicMessageName)
}

type ptrFieldMessage struct {
	StringField *stringField `protobuf:"bytes,1,opt,name=stringField"`
}

func (m *ptrFieldMessage) Reset() {
}

func (m *ptrFieldMessage) String() string {
	return m.StringField.StringValue
}

func (m *ptrFieldMessage) ProtoMessage() {
}

type stringField struct {
	IsSet       bool   `protobuf:"varint,1,opt,name=isSet"`
	StringValue string `protobuf:"bytes,2,opt,name=stringValue"`
}

func (s *stringField) Reset() {
}

func (s *stringField) String() string {
	return s.StringValue
}

func (s *stringField) ProtoMessage() {
}

func (s *stringField) UnmarshalJSONPB(jum *vendor.Unmarshaler, js []byte) error {
	s.IsSet = true
	s.StringValue = string(js)
	return nil
}

// dynamicMessage implements protobuf.Message but is not a normal generated message type.
// It provides implementations of JSONPBMarshaler and JSONPBUnmarshaler for JSON support.
type dynamicMessage struct {
	RawJson string `protobuf:"bytes,1,opt,name=rawJson"`

	// an unexported nested message is present just to ensure that it
	// won't result in a panic (see issue #509)
	Dummy *dynamicMessage `protobuf:"bytes,2,opt,name=dummy"`
}

func (m *dynamicMessage) Reset() {
	m.RawJson = "{}"
}

func (m *dynamicMessage) String() string {
	return m.RawJson
}

func (m *dynamicMessage) ProtoMessage() {
}

func (m *dynamicMessage) MarshalJSONPB(jm *vendor.Marshaler) ([]byte, error) {
	return []byte(m.RawJson), nil
}

func (m *dynamicMessage) UnmarshalJSONPB(jum *vendor.Unmarshaler, js []byte) error {
	m.RawJson = string(js)
	return nil
}

// Test unmarshaling message containing unset required fields should produce error.
func TestUnmarshalUnsetRequiredFields(t *testing.T) {
	tests := []struct {
		desc string
		pb   vendor.Message
		json string
	}{
		{
			desc: "direct required field missing",
			pb:   &vendor.MsgWithRequired{},
			json: `{}`,
		},
		{
			desc: "direct required field set to null",
			pb:   &vendor.MsgWithRequired{},
			json: `{"str": null}`,
		},
		{
			desc: "indirect required field missing",
			pb:   &vendor.MsgWithIndirectRequired{},
			json: `{"subm": {}}`,
		},
		{
			desc: "indirect required field set to null",
			pb:   &vendor.MsgWithIndirectRequired{},
			json: `{"subm": {"str": null}}`,
		},
		{
			desc: "direct required bytes field missing",
			pb:   &vendor.MsgWithRequiredBytes{},
			json: `{}`,
		},
		{
			desc: "direct required bytes field set to null",
			pb:   &vendor.MsgWithRequiredBytes{},
			json: `{"byts": null}`,
		},
		{
			desc: "direct required wkt field missing",
			pb:   &vendor.MsgWithRequiredWKT{},
			json: `{}`,
		},
		{
			desc: "direct required wkt field set to null",
			pb:   &vendor.MsgWithRequiredWKT{},
			json: `{"str": null}`,
		},
		{
			desc: "any containing message with required field set to null",
			pb:   &vendor.KnownTypes{},
			json: `{"an": {"@type": "example.com/jsonpb.MsgWithRequired", "str": null}}`,
		},
		{
			desc: "any containing message with missing required field",
			pb:   &vendor.KnownTypes{},
			json: `{"an": {"@type": "example.com/jsonpb.MsgWithRequired"}}`,
		},
		{
			desc: "missing required in map value",
			pb:   &vendor.MsgWithIndirectRequired{},
			json: `{"map_field": {"a": {}, "b": {"str": "hi"}}}`,
		},
		{
			desc: "required in map value set to null",
			pb:   &vendor.MsgWithIndirectRequired{},
			json: `{"map_field": {"a": {"str": "hello"}, "b": {"str": null}}}`,
		},
		{
			desc: "missing required in slice item",
			pb:   &vendor.MsgWithIndirectRequired{},
			json: `{"slice_field": [{}, {"str": "hi"}]}`,
		},
		{
			desc: "required in slice item set to null",
			pb:   &vendor.MsgWithIndirectRequired{},
			json: `{"slice_field": [{"str": "hello"}, {"str": null}]}`,
		},
		{
			desc: "required inside oneof missing",
			pb:   &vendor.MsgWithOneof{},
			json: `{"msgWithRequired": {}}`,
		},
		{
			desc: "required inside oneof set to null",
			pb:   &vendor.MsgWithOneof{},
			json: `{"msgWithRequired": {"str": null}}`,
		},
		{
			desc: "required field in extension missing",
			pb:   &vendor.Real{},
			json: `{"[jsonpb.extm]":{}}`,
		},
		{
			desc: "required field in extension set to null",
			pb:   &vendor.Real{},
			json: `{"[jsonpb.extm]":{"str": null}}`,
		},
	}

	for _, tc := range tests {
		if err := vendor.UnmarshalString(tc.json, tc.pb); err == nil {
			t.Errorf("%s: expecting error in unmarshaling with unset required fields %s", tc.desc, tc.json)
		}
	}
}
