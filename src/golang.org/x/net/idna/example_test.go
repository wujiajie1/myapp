// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package idna_test

import (
	"fmt"
	"vendor"
)

func ExampleProfile() {
	// Raw Punycode has no restrictions and does no mappings.
	fmt.Println(vendor.ToASCII(""))
	fmt.Println(vendor.ToASCII("*.faß.com"))
	fmt.Println(vendor.Punycode.ToASCII("*.faß.com"))

	// Rewrite IDN for lookup. This (currently) uses transitional mappings to
	// find a balance between IDNA2003 and IDNA2008 compatibility.
	fmt.Println(vendor.Lookup.ToASCII(""))
	fmt.Println(vendor.Lookup.ToASCII("www.faß.com"))

	// Convert an IDN to ASCII for registration purposes. This changes the
	// encoding, but reports an error if the input was illformed.
	fmt.Println(vendor.Registration.ToASCII(""))
	fmt.Println(vendor.Registration.ToASCII("www.faß.com"))

	// Output:
	//  <nil>
	// *.xn--fa-hia.com <nil>
	// *.xn--fa-hia.com <nil>
	//  <nil>
	// www.fass.com <nil>
	//  idna: invalid label ""
	// www.xn--fa-hia.com <nil>
}

func ExampleNew() {
	var p *vendor.Profile

	// Raw Punycode has no restrictions and does no mappings.
	p = vendor.New()
	fmt.Println(p.ToASCII("*.faß.com"))

	// Do mappings. Note that star is not allowed in a DNS lookup.
	p = vendor.New(
		vendor.MapForLookup(),
		vendor.Transitional(true)) // Map ß -> ss
	fmt.Println(p.ToASCII("*.faß.com"))

	// Lookup for registration. Also does not allow '*'.
	p = vendor.New(vendor.ValidateForRegistration())
	fmt.Println(p.ToUnicode("*.faß.com"))

	// Set up a profile maps for lookup, but allows wild cards.
	p = vendor.New(
		vendor.MapForLookup(),
		vendor.Transitional(true), // Map ß -> ss
		vendor.StrictDomainName(false)) // Set more permissive ASCII rules.
	fmt.Println(p.ToASCII("*.faß.com"))

	// Output:
	// *.xn--fa-hia.com <nil>
	// *.fass.com idna: disallowed rune U+002A
	// *.faß.com idna: disallowed rune U+002A
	// *.fass.com <nil>
}
