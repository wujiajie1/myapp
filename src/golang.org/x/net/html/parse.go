// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"vendor"
)

// A parser implements the HTML5 parsing algorithm:
// https://html.spec.whatwg.org/multipage/syntax.html#tree-construction
type parser struct {
	// tokenizer provides the tokens for the parser.
	tokenizer *vendor.Tokenizer
	// tok is the most recently read token.
	tok vendor.Token
	// Self-closing tags like <hr/> are treated as start tags, except that
	// hasSelfClosingToken is set while they are being processed.
	hasSelfClosingToken bool
	// doc is the document root element.
	doc *vendor.Node
	// The stack of open elements (section 12.2.4.2) and active formatting
	// elements (section 12.2.4.3).
	oe, afe vendor.nodeStack
	// Element pointers (section 12.2.4.4).
	head, form *vendor.Node
	// Other parsing state flags (section 12.2.4.5).
	scripting, framesetOK bool
	// The stack of template insertion modes
	templateStack vendor.insertionModeStack
	// im is the current insertion mode.
	im insertionMode
	// originalIM is the insertion mode to go back to after completing a text
	// or inTableText insertion mode.
	originalIM insertionMode
	// fosterParenting is whether new elements should be inserted according to
	// the foster parenting rules (section 12.2.6.1).
	fosterParenting bool
	// quirks is whether the parser is operating in "quirks mode."
	quirks bool
	// fragment is whether the parser is parsing an HTML fragment.
	fragment bool
	// context is the context element when parsing an HTML fragment
	// (section 12.4).
	context *vendor.Node
}

func (p *parser) top() *vendor.Node {
	if n := p.oe.top(); n != nil {
		return n
	}
	return p.doc
}

// Stop tags for use in popUntil. These come from section 12.2.4.2.
var (
	defaultScopeStopTags = map[string][]vendor.Atom{
		"":     {vendor.Applet, vendor.Caption, vendor.Html, vendor.Table, vendor.Td, vendor.Th, vendor.Marquee, vendor.Object, vendor.Template},
		"math": {vendor.AnnotationXml, vendor.Mi, vendor.Mn, vendor.Mo, vendor.Ms, vendor.Mtext},
		"svg":  {vendor.Desc, vendor.ForeignObject, vendor.Title},
	}
)

type scope int

const (
	defaultScope scope = iota
	listItemScope
	buttonScope
	tableScope
	tableRowScope
	tableBodyScope
	selectScope
)

// popUntil pops the stack of open elements at the highest element whose tag
// is in matchTags, provided there is no higher element in the scope's stop
// tags (as defined in section 12.2.4.2). It returns whether or not there was
// such an element. If there was not, popUntil leaves the stack unchanged.
//
// For example, the set of stop tags for table scope is: "html", "table". If
// the stack was:
// ["html", "body", "font", "table", "b", "i", "u"]
// then popUntil(tableScope, "font") would return false, but
// popUntil(tableScope, "i") would return true and the stack would become:
// ["html", "body", "font", "table", "b"]
//
// If an element's tag is in both the stop tags and matchTags, then the stack
// will be popped and the function returns true (provided, of course, there was
// no higher element in the stack that was also in the stop tags). For example,
// popUntil(tableScope, "table") returns true and leaves:
// ["html", "body", "font"]
func (p *parser) popUntil(s scope, matchTags ...vendor.Atom) bool {
	if i := p.indexOfElementInScope(s, matchTags...); i != -1 {
		p.oe = p.oe[:i]
		return true
	}
	return false
}

// indexOfElementInScope returns the index in p.oe of the highest element whose
// tag is in matchTags that is in scope. If no matching element is in scope, it
// returns -1.
func (p *parser) indexOfElementInScope(s scope, matchTags ...vendor.Atom) int {
	for i := len(p.oe) - 1; i >= 0; i-- {
		tagAtom := p.oe[i].DataAtom
		if p.oe[i].Namespace == "" {
			for _, t := range matchTags {
				if t == tagAtom {
					return i
				}
			}
			switch s {
			case defaultScope:
				// No-op.
			case listItemScope:
				if tagAtom == vendor.Ol || tagAtom == vendor.Ul {
					return -1
				}
			case buttonScope:
				if tagAtom == vendor.Button {
					return -1
				}
			case tableScope:
				if tagAtom == vendor.Html || tagAtom == vendor.Table || tagAtom == vendor.Template {
					return -1
				}
			case selectScope:
				if tagAtom != vendor.Optgroup && tagAtom != vendor.Option {
					return -1
				}
			default:
				panic("unreachable")
			}
		}
		switch s {
		case defaultScope, listItemScope, buttonScope:
			for _, t := range defaultScopeStopTags[p.oe[i].Namespace] {
				if t == tagAtom {
					return -1
				}
			}
		}
	}
	return -1
}

// elementInScope is like popUntil, except that it doesn't modify the stack of
// open elements.
func (p *parser) elementInScope(s scope, matchTags ...vendor.Atom) bool {
	return p.indexOfElementInScope(s, matchTags...) != -1
}

// clearStackToContext pops elements off the stack of open elements until a
// scope-defined element is found.
func (p *parser) clearStackToContext(s scope) {
	for i := len(p.oe) - 1; i >= 0; i-- {
		tagAtom := p.oe[i].DataAtom
		switch s {
		case tableScope:
			if tagAtom == vendor.Html || tagAtom == vendor.Table || tagAtom == vendor.Template {
				p.oe = p.oe[:i+1]
				return
			}
		case tableRowScope:
			if tagAtom == vendor.Html || tagAtom == vendor.Tr || tagAtom == vendor.Template {
				p.oe = p.oe[:i+1]
				return
			}
		case tableBodyScope:
			if tagAtom == vendor.Html || tagAtom == vendor.Tbody || tagAtom == vendor.Tfoot || tagAtom == vendor.Thead || tagAtom == vendor.Template {
				p.oe = p.oe[:i+1]
				return
			}
		default:
			panic("unreachable")
		}
	}
}

// generateImpliedEndTags pops nodes off the stack of open elements as long as
// the top node has a tag name of dd, dt, li, optgroup, option, p, rb, rp, rt or rtc.
// If exceptions are specified, nodes with that name will not be popped off.
func (p *parser) generateImpliedEndTags(exceptions ...string) {
	var i int
loop:
	for i = len(p.oe) - 1; i >= 0; i-- {
		n := p.oe[i]
		if n.Type == vendor.ElementNode {
			switch n.DataAtom {
			case vendor.Dd, vendor.Dt, vendor.Li, vendor.Optgroup, vendor.Option, vendor.P, vendor.Rb, vendor.Rp, vendor.Rt, vendor.Rtc:
				for _, except := range exceptions {
					if n.Data == except {
						break loop
					}
				}
				continue
			}
		}
		break
	}

	p.oe = p.oe[:i+1]
}

// addChild adds a child node n to the top element, and pushes n onto the stack
// of open elements if it is an element node.
func (p *parser) addChild(n *vendor.Node) {
	if p.shouldFosterParent() {
		p.fosterParent(n)
	} else {
		p.top().AppendChild(n)
	}

	if n.Type == vendor.ElementNode {
		p.oe = append(p.oe, n)
	}
}

// shouldFosterParent returns whether the next node to be added should be
// foster parented.
func (p *parser) shouldFosterParent() bool {
	if p.fosterParenting {
		switch p.top().DataAtom {
		case vendor.Table, vendor.Tbody, vendor.Tfoot, vendor.Thead, vendor.Tr:
			return true
		}
	}
	return false
}

// fosterParent adds a child node according to the foster parenting rules.
// Section 12.2.6.1, "foster parenting".
func (p *parser) fosterParent(n *vendor.Node) {
	var table, parent, prev, template *vendor.Node
	var i int
	for i = len(p.oe) - 1; i >= 0; i-- {
		if p.oe[i].DataAtom == vendor.Table {
			table = p.oe[i]
			break
		}
	}

	var j int
	for j = len(p.oe) - 1; j >= 0; j-- {
		if p.oe[j].DataAtom == vendor.Template {
			template = p.oe[j]
			break
		}
	}

	if template != nil && (table == nil || j > i) {
		template.AppendChild(n)
		return
	}

	if table == nil {
		// The foster parent is the html element.
		parent = p.oe[0]
	} else {
		parent = table.Parent
	}
	if parent == nil {
		parent = p.oe[i-1]
	}

	if table != nil {
		prev = table.PrevSibling
	} else {
		prev = parent.LastChild
	}
	if prev != nil && prev.Type == vendor.TextNode && n.Type == vendor.TextNode {
		prev.Data += n.Data
		return
	}

	parent.InsertBefore(n, table)
}

// addText adds text to the preceding node if it is a text node, or else it
// calls addChild with a new text node.
func (p *parser) addText(text string) {
	if text == "" {
		return
	}

	if p.shouldFosterParent() {
		p.fosterParent(&vendor.Node{
			Type: vendor.TextNode,
			Data: text,
		})
		return
	}

	t := p.top()
	if n := t.LastChild; n != nil && n.Type == vendor.TextNode {
		n.Data += text
		return
	}
	p.addChild(&vendor.Node{
		Type: vendor.TextNode,
		Data: text,
	})
}

// addElement adds a child element based on the current token.
func (p *parser) addElement() {
	p.addChild(&vendor.Node{
		Type:     vendor.ElementNode,
		DataAtom: p.tok.DataAtom,
		Data:     p.tok.Data,
		Attr:     p.tok.Attr,
	})
}

// Section 12.2.4.3.
func (p *parser) addFormattingElement() {
	tagAtom, attr := p.tok.DataAtom, p.tok.Attr
	p.addElement()

	// Implement the Noah's Ark clause, but with three per family instead of two.
	identicalElements := 0
findIdenticalElements:
	for i := len(p.afe) - 1; i >= 0; i-- {
		n := p.afe[i]
		if n.Type == vendor.scopeMarkerNode {
			break
		}
		if n.Type != vendor.ElementNode {
			continue
		}
		if n.Namespace != "" {
			continue
		}
		if n.DataAtom != tagAtom {
			continue
		}
		if len(n.Attr) != len(attr) {
			continue
		}
	compareAttributes:
		for _, t0 := range n.Attr {
			for _, t1 := range attr {
				if t0.Key == t1.Key && t0.Namespace == t1.Namespace && t0.Val == t1.Val {
					// Found a match for this attribute, continue with the next attribute.
					continue compareAttributes
				}
			}
			// If we get here, there is no attribute that matches a.
			// Therefore the element is not identical to the new one.
			continue findIdenticalElements
		}

		identicalElements++
		if identicalElements >= 3 {
			p.afe.remove(n)
		}
	}

	p.afe = append(p.afe, p.top())
}

// Section 12.2.4.3.
func (p *parser) clearActiveFormattingElements() {
	for {
		n := p.afe.pop()
		if len(p.afe) == 0 || n.Type == vendor.scopeMarkerNode {
			return
		}
	}
}

// Section 12.2.4.3.
func (p *parser) reconstructActiveFormattingElements() {
	n := p.afe.top()
	if n == nil {
		return
	}
	if n.Type == vendor.scopeMarkerNode || p.oe.index(n) != -1 {
		return
	}
	i := len(p.afe) - 1
	for n.Type != vendor.scopeMarkerNode && p.oe.index(n) == -1 {
		if i == 0 {
			i = -1
			break
		}
		i--
		n = p.afe[i]
	}
	for {
		i++
		clone := p.afe[i].clone()
		p.addChild(clone)
		p.afe[i] = clone
		if i == len(p.afe)-1 {
			break
		}
	}
}

// Section 12.2.5.
func (p *parser) acknowledgeSelfClosingTag() {
	p.hasSelfClosingToken = false
}

// An insertion mode (section 12.2.4.1) is the state transition function from
// a particular state in the HTML5 parser's state machine. It updates the
// parser's fields depending on parser.tok (where ErrorToken means EOF).
// It returns whether the token was consumed.
type insertionMode func(*parser) bool

// setOriginalIM sets the insertion mode to return to after completing a text or
// inTableText insertion mode.
// Section 12.2.4.1, "using the rules for".
func (p *parser) setOriginalIM() {
	if p.originalIM != nil {
		panic("html: bad parser state: originalIM was set twice")
	}
	p.originalIM = p.im
}

// Section 12.2.4.1, "reset the insertion mode".
func (p *parser) resetInsertionMode() {
	for i := len(p.oe) - 1; i >= 0; i-- {
		n := p.oe[i]
		last := i == 0
		if last && p.context != nil {
			n = p.context
		}

		switch n.DataAtom {
		case vendor.Select:
			if !last {
				for ancestor, first := n, p.oe[0]; ancestor != first; {
					ancestor = p.oe[p.oe.index(ancestor)-1]
					switch ancestor.DataAtom {
					case vendor.Template:
						p.im = inSelectIM
						return
					case vendor.Table:
						p.im = inSelectInTableIM
						return
					}
				}
			}
			p.im = inSelectIM
		case vendor.Td, vendor.Th:
			// TODO: remove this divergence from the HTML5 spec.
			//
			// See https://bugs.chromium.org/p/chromium/issues/detail?id=829668
			p.im = inCellIM
		case vendor.Tr:
			p.im = inRowIM
		case vendor.Tbody, vendor.Thead, vendor.Tfoot:
			p.im = inTableBodyIM
		case vendor.Caption:
			p.im = inCaptionIM
		case vendor.Colgroup:
			p.im = inColumnGroupIM
		case vendor.Table:
			p.im = inTableIM
		case vendor.Template:
			// TODO: remove this divergence from the HTML5 spec.
			if n.Namespace != "" {
				continue
			}
			p.im = p.templateStack.top()
		case vendor.Head:
			// TODO: remove this divergence from the HTML5 spec.
			//
			// See https://bugs.chromium.org/p/chromium/issues/detail?id=829668
			p.im = inHeadIM
		case vendor.Body:
			p.im = inBodyIM
		case vendor.Frameset:
			p.im = inFramesetIM
		case vendor.Html:
			if p.head == nil {
				p.im = beforeHeadIM
			} else {
				p.im = afterHeadIM
			}
		default:
			if last {
				p.im = inBodyIM
				return
			}
			continue
		}
		return
	}
}

const whitespace = " \t\r\n\f"

// Section 12.2.6.4.1.
func initialIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		p.tok.Data = strings.TrimLeft(p.tok.Data, whitespace)
		if len(p.tok.Data) == 0 {
			// It was all whitespace, so ignore it.
			return true
		}
	case vendor.CommentToken:
		p.doc.AppendChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	case vendor.DoctypeToken:
		n, quirks := vendor.parseDoctype(p.tok.Data)
		p.doc.AppendChild(n)
		p.quirks = quirks
		p.im = beforeHTMLIM
		return true
	}
	p.quirks = true
	p.im = beforeHTMLIM
	return false
}

// Section 12.2.6.4.2.
func beforeHTMLIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.DoctypeToken:
		// Ignore the token.
		return true
	case vendor.TextToken:
		p.tok.Data = strings.TrimLeft(p.tok.Data, whitespace)
		if len(p.tok.Data) == 0 {
			// It was all whitespace, so ignore it.
			return true
		}
	case vendor.StartTagToken:
		if p.tok.DataAtom == vendor.Html {
			p.addElement()
			p.im = beforeHeadIM
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Head, vendor.Body, vendor.Html, vendor.Br:
			p.parseImpliedToken(vendor.StartTagToken, vendor.Html, vendor.Html.String())
			return false
		default:
			// Ignore the token.
			return true
		}
	case vendor.CommentToken:
		p.doc.AppendChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	}
	p.parseImpliedToken(vendor.StartTagToken, vendor.Html, vendor.Html.String())
	return false
}

// Section 12.2.6.4.3.
func beforeHeadIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		p.tok.Data = strings.TrimLeft(p.tok.Data, whitespace)
		if len(p.tok.Data) == 0 {
			// It was all whitespace, so ignore it.
			return true
		}
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Head:
			p.addElement()
			p.head = p.top()
			p.im = inHeadIM
			return true
		case vendor.Html:
			return inBodyIM(p)
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Head, vendor.Body, vendor.Html, vendor.Br:
			p.parseImpliedToken(vendor.StartTagToken, vendor.Head, vendor.Head.String())
			return false
		default:
			// Ignore the token.
			return true
		}
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	case vendor.DoctypeToken:
		// Ignore the token.
		return true
	}

	p.parseImpliedToken(vendor.StartTagToken, vendor.Head, vendor.Head.String())
	return false
}

// Section 12.2.6.4.4.
func inHeadIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		s := strings.TrimLeft(p.tok.Data, whitespace)
		if len(s) < len(p.tok.Data) {
			// Add the initial whitespace to the current node.
			p.addText(p.tok.Data[:len(p.tok.Data)-len(s)])
			if s == "" {
				return true
			}
			p.tok.Data = s
		}
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			return inBodyIM(p)
		case vendor.Base, vendor.Basefont, vendor.Bgsound, vendor.Command, vendor.Link, vendor.Meta:
			p.addElement()
			p.oe.pop()
			p.acknowledgeSelfClosingTag()
			return true
		case vendor.Noscript:
			p.addElement()
			if p.scripting {
				p.setOriginalIM()
				p.im = textIM
			} else {
				p.im = inHeadNoscriptIM
			}
			return true
		case vendor.Script, vendor.Title, vendor.Noframes, vendor.Style:
			p.addElement()
			p.setOriginalIM()
			p.im = textIM
			return true
		case vendor.Head:
			// Ignore the token.
			return true
		case vendor.Template:
			p.addElement()
			p.afe = append(p.afe, &vendor.scopeMarker)
			p.framesetOK = false
			p.im = inTemplateIM
			p.templateStack = append(p.templateStack, inTemplateIM)
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Head:
			p.oe.pop()
			p.im = afterHeadIM
			return true
		case vendor.Body, vendor.Html, vendor.Br:
			p.parseImpliedToken(vendor.EndTagToken, vendor.Head, vendor.Head.String())
			return false
		case vendor.Template:
			if !p.oe.contains(vendor.Template) {
				return true
			}
			// TODO: remove this divergence from the HTML5 spec.
			//
			// See https://bugs.chromium.org/p/chromium/issues/detail?id=829668
			p.generateImpliedEndTags()
			for i := len(p.oe) - 1; i >= 0; i-- {
				if n := p.oe[i]; n.Namespace == "" && n.DataAtom == vendor.Template {
					p.oe = p.oe[:i]
					break
				}
			}
			p.clearActiveFormattingElements()
			p.templateStack.pop()
			p.resetInsertionMode()
			return true
		default:
			// Ignore the token.
			return true
		}
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	case vendor.DoctypeToken:
		// Ignore the token.
		return true
	}

	p.parseImpliedToken(vendor.EndTagToken, vendor.Head, vendor.Head.String())
	return false
}

// 12.2.6.4.5.
func inHeadNoscriptIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.DoctypeToken:
		// Ignore the token.
		return true
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			return inBodyIM(p)
		case vendor.Basefont, vendor.Bgsound, vendor.Link, vendor.Meta, vendor.Noframes, vendor.Style:
			return inHeadIM(p)
		case vendor.Head, vendor.Noscript:
			// Ignore the token.
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Noscript, vendor.Br:
		default:
			// Ignore the token.
			return true
		}
	case vendor.TextToken:
		s := strings.TrimLeft(p.tok.Data, whitespace)
		if len(s) == 0 {
			// It was all whitespace.
			return inHeadIM(p)
		}
	case vendor.CommentToken:
		return inHeadIM(p)
	}
	p.oe.pop()
	if p.top().DataAtom != vendor.Head {
		panic("html: the new current node will be a head element.")
	}
	p.im = inHeadIM
	if p.tok.DataAtom == vendor.Noscript {
		return true
	}
	return false
}

// Section 12.2.6.4.6.
func afterHeadIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		s := strings.TrimLeft(p.tok.Data, whitespace)
		if len(s) < len(p.tok.Data) {
			// Add the initial whitespace to the current node.
			p.addText(p.tok.Data[:len(p.tok.Data)-len(s)])
			if s == "" {
				return true
			}
			p.tok.Data = s
		}
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			return inBodyIM(p)
		case vendor.Body:
			p.addElement()
			p.framesetOK = false
			p.im = inBodyIM
			return true
		case vendor.Frameset:
			p.addElement()
			p.im = inFramesetIM
			return true
		case vendor.Base, vendor.Basefont, vendor.Bgsound, vendor.Link, vendor.Meta, vendor.Noframes, vendor.Script, vendor.Style, vendor.Template, vendor.Title:
			p.oe = append(p.oe, p.head)
			defer p.oe.remove(p.head)
			return inHeadIM(p)
		case vendor.Head:
			// Ignore the token.
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Body, vendor.Html, vendor.Br:
			// Drop down to creating an implied <body> tag.
		case vendor.Template:
			return inHeadIM(p)
		default:
			// Ignore the token.
			return true
		}
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	case vendor.DoctypeToken:
		// Ignore the token.
		return true
	}

	p.parseImpliedToken(vendor.StartTagToken, vendor.Body, vendor.Body.String())
	p.framesetOK = true
	return false
}

// copyAttributes copies attributes of src not found on dst to dst.
func copyAttributes(dst *vendor.Node, src vendor.Token) {
	if len(src.Attr) == 0 {
		return
	}
	attr := map[string]string{}
	for _, t := range dst.Attr {
		attr[t.Key] = t.Val
	}
	for _, t := range src.Attr {
		if _, ok := attr[t.Key]; !ok {
			dst.Attr = append(dst.Attr, t)
			attr[t.Key] = t.Val
		}
	}
}

// Section 12.2.6.4.7.
func inBodyIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		d := p.tok.Data
		switch n := p.oe.top(); n.DataAtom {
		case vendor.Pre, vendor.Listing:
			if n.FirstChild == nil {
				// Ignore a newline at the start of a <pre> block.
				if d != "" && d[0] == '\r' {
					d = d[1:]
				}
				if d != "" && d[0] == '\n' {
					d = d[1:]
				}
			}
		}
		d = strings.Replace(d, "\x00", "", -1)
		if d == "" {
			return true
		}
		p.reconstructActiveFormattingElements()
		p.addText(d)
		if p.framesetOK && strings.TrimLeft(d, whitespace) != "" {
			// There were non-whitespace characters inserted.
			p.framesetOK = false
		}
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			if p.oe.contains(vendor.Template) {
				return true
			}
			copyAttributes(p.oe[0], p.tok)
		case vendor.Base, vendor.Basefont, vendor.Bgsound, vendor.Command, vendor.Link, vendor.Meta, vendor.Noframes, vendor.Script, vendor.Style, vendor.Template, vendor.Title:
			return inHeadIM(p)
		case vendor.Body:
			if p.oe.contains(vendor.Template) {
				return true
			}
			if len(p.oe) >= 2 {
				body := p.oe[1]
				if body.Type == vendor.ElementNode && body.DataAtom == vendor.Body {
					p.framesetOK = false
					copyAttributes(body, p.tok)
				}
			}
		case vendor.Frameset:
			if !p.framesetOK || len(p.oe) < 2 || p.oe[1].DataAtom != vendor.Body {
				// Ignore the token.
				return true
			}
			body := p.oe[1]
			if body.Parent != nil {
				body.Parent.RemoveChild(body)
			}
			p.oe = p.oe[:1]
			p.addElement()
			p.im = inFramesetIM
			return true
		case vendor.Address, vendor.Article, vendor.Aside, vendor.Blockquote, vendor.Center, vendor.Details, vendor.Dir, vendor.Div, vendor.Dl, vendor.Fieldset, vendor.Figcaption, vendor.Figure, vendor.Footer, vendor.Header, vendor.Hgroup, vendor.Menu, vendor.Nav, vendor.Ol, vendor.P, vendor.Section, vendor.Summary, vendor.Ul:
			p.popUntil(buttonScope, vendor.P)
			p.addElement()
		case vendor.H1, vendor.H2, vendor.H3, vendor.H4, vendor.H5, vendor.H6:
			p.popUntil(buttonScope, vendor.P)
			switch n := p.top(); n.DataAtom {
			case vendor.H1, vendor.H2, vendor.H3, vendor.H4, vendor.H5, vendor.H6:
				p.oe.pop()
			}
			p.addElement()
		case vendor.Pre, vendor.Listing:
			p.popUntil(buttonScope, vendor.P)
			p.addElement()
			// The newline, if any, will be dealt with by the TextToken case.
			p.framesetOK = false
		case vendor.Form:
			if p.form != nil && !p.oe.contains(vendor.Template) {
				// Ignore the token
				return true
			}
			p.popUntil(buttonScope, vendor.P)
			p.addElement()
			if !p.oe.contains(vendor.Template) {
				p.form = p.top()
			}
		case vendor.Li:
			p.framesetOK = false
			for i := len(p.oe) - 1; i >= 0; i-- {
				node := p.oe[i]
				switch node.DataAtom {
				case vendor.Li:
					p.oe = p.oe[:i]
				case vendor.Address, vendor.Div, vendor.P:
					continue
				default:
					if !vendor.isSpecialElement(node) {
						continue
					}
				}
				break
			}
			p.popUntil(buttonScope, vendor.P)
			p.addElement()
		case vendor.Dd, vendor.Dt:
			p.framesetOK = false
			for i := len(p.oe) - 1; i >= 0; i-- {
				node := p.oe[i]
				switch node.DataAtom {
				case vendor.Dd, vendor.Dt:
					p.oe = p.oe[:i]
				case vendor.Address, vendor.Div, vendor.P:
					continue
				default:
					if !vendor.isSpecialElement(node) {
						continue
					}
				}
				break
			}
			p.popUntil(buttonScope, vendor.P)
			p.addElement()
		case vendor.Plaintext:
			p.popUntil(buttonScope, vendor.P)
			p.addElement()
		case vendor.Button:
			p.popUntil(defaultScope, vendor.Button)
			p.reconstructActiveFormattingElements()
			p.addElement()
			p.framesetOK = false
		case vendor.A:
			for i := len(p.afe) - 1; i >= 0 && p.afe[i].Type != vendor.scopeMarkerNode; i-- {
				if n := p.afe[i]; n.Type == vendor.ElementNode && n.DataAtom == vendor.A {
					p.inBodyEndTagFormatting(vendor.A, "a")
					p.oe.remove(n)
					p.afe.remove(n)
					break
				}
			}
			p.reconstructActiveFormattingElements()
			p.addFormattingElement()
		case vendor.B, vendor.Big, vendor.Code, vendor.Em, vendor.Font, vendor.I, vendor.S, vendor.Small, vendor.Strike, vendor.Strong, vendor.Tt, vendor.U:
			p.reconstructActiveFormattingElements()
			p.addFormattingElement()
		case vendor.Nobr:
			p.reconstructActiveFormattingElements()
			if p.elementInScope(defaultScope, vendor.Nobr) {
				p.inBodyEndTagFormatting(vendor.Nobr, "nobr")
				p.reconstructActiveFormattingElements()
			}
			p.addFormattingElement()
		case vendor.Applet, vendor.Marquee, vendor.Object:
			p.reconstructActiveFormattingElements()
			p.addElement()
			p.afe = append(p.afe, &vendor.scopeMarker)
			p.framesetOK = false
		case vendor.Table:
			if !p.quirks {
				p.popUntil(buttonScope, vendor.P)
			}
			p.addElement()
			p.framesetOK = false
			p.im = inTableIM
			return true
		case vendor.Area, vendor.Br, vendor.Embed, vendor.Img, vendor.Input, vendor.Keygen, vendor.Wbr:
			p.reconstructActiveFormattingElements()
			p.addElement()
			p.oe.pop()
			p.acknowledgeSelfClosingTag()
			if p.tok.DataAtom == vendor.Input {
				for _, t := range p.tok.Attr {
					if t.Key == "type" {
						if strings.ToLower(t.Val) == "hidden" {
							// Skip setting framesetOK = false
							return true
						}
					}
				}
			}
			p.framesetOK = false
		case vendor.Param, vendor.Source, vendor.Track:
			p.addElement()
			p.oe.pop()
			p.acknowledgeSelfClosingTag()
		case vendor.Hr:
			p.popUntil(buttonScope, vendor.P)
			p.addElement()
			p.oe.pop()
			p.acknowledgeSelfClosingTag()
			p.framesetOK = false
		case vendor.Image:
			p.tok.DataAtom = vendor.Img
			p.tok.Data = vendor.Img.String()
			return false
		case vendor.Isindex:
			if p.form != nil {
				// Ignore the token.
				return true
			}
			action := ""
			prompt := "This is a searchable index. Enter search keywords: "
			attr := []vendor.Attribute{{Key: "name", Val: "isindex"}}
			for _, t := range p.tok.Attr {
				switch t.Key {
				case "action":
					action = t.Val
				case "name":
					// Ignore the attribute.
				case "prompt":
					prompt = t.Val
				default:
					attr = append(attr, t)
				}
			}
			p.acknowledgeSelfClosingTag()
			p.popUntil(buttonScope, vendor.P)
			p.parseImpliedToken(vendor.StartTagToken, vendor.Form, vendor.Form.String())
			if p.form == nil {
				// NOTE: The 'isindex' element has been removed,
				// and the 'template' element has not been designed to be
				// collaborative with the index element.
				//
				// Ignore the token.
				return true
			}
			if action != "" {
				p.form.Attr = []vendor.Attribute{{Key: "action", Val: action}}
			}
			p.parseImpliedToken(vendor.StartTagToken, vendor.Hr, vendor.Hr.String())
			p.parseImpliedToken(vendor.StartTagToken, vendor.Label, vendor.Label.String())
			p.addText(prompt)
			p.addChild(&vendor.Node{
				Type:     vendor.ElementNode,
				DataAtom: vendor.Input,
				Data:     vendor.Input.String(),
				Attr:     attr,
			})
			p.oe.pop()
			p.parseImpliedToken(vendor.EndTagToken, vendor.Label, vendor.Label.String())
			p.parseImpliedToken(vendor.StartTagToken, vendor.Hr, vendor.Hr.String())
			p.parseImpliedToken(vendor.EndTagToken, vendor.Form, vendor.Form.String())
		case vendor.Textarea:
			p.addElement()
			p.setOriginalIM()
			p.framesetOK = false
			p.im = textIM
		case vendor.Xmp:
			p.popUntil(buttonScope, vendor.P)
			p.reconstructActiveFormattingElements()
			p.framesetOK = false
			p.addElement()
			p.setOriginalIM()
			p.im = textIM
		case vendor.Iframe:
			p.framesetOK = false
			p.addElement()
			p.setOriginalIM()
			p.im = textIM
		case vendor.Noembed, vendor.Noscript:
			p.addElement()
			p.setOriginalIM()
			p.im = textIM
		case vendor.Select:
			p.reconstructActiveFormattingElements()
			p.addElement()
			p.framesetOK = false
			p.im = inSelectIM
			return true
		case vendor.Optgroup, vendor.Option:
			if p.top().DataAtom == vendor.Option {
				p.oe.pop()
			}
			p.reconstructActiveFormattingElements()
			p.addElement()
		case vendor.Rb, vendor.Rtc:
			if p.elementInScope(defaultScope, vendor.Ruby) {
				p.generateImpliedEndTags()
			}
			p.addElement()
		case vendor.Rp, vendor.Rt:
			if p.elementInScope(defaultScope, vendor.Ruby) {
				p.generateImpliedEndTags("rtc")
			}
			p.addElement()
		case vendor.Math, vendor.Svg:
			p.reconstructActiveFormattingElements()
			if p.tok.DataAtom == vendor.Math {
				vendor.adjustAttributeNames(p.tok.Attr, vendor.mathMLAttributeAdjustments)
			} else {
				vendor.adjustAttributeNames(p.tok.Attr, vendor.svgAttributeAdjustments)
			}
			vendor.adjustForeignAttributes(p.tok.Attr)
			p.addElement()
			p.top().Namespace = p.tok.Data
			if p.hasSelfClosingToken {
				p.oe.pop()
				p.acknowledgeSelfClosingTag()
			}
			return true
		case vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Frame, vendor.Head, vendor.Tbody, vendor.Td, vendor.Tfoot, vendor.Th, vendor.Thead, vendor.Tr:
			// Ignore the token.
		default:
			p.reconstructActiveFormattingElements()
			p.addElement()
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Body:
			if p.elementInScope(defaultScope, vendor.Body) {
				p.im = afterBodyIM
			}
		case vendor.Html:
			if p.elementInScope(defaultScope, vendor.Body) {
				p.parseImpliedToken(vendor.EndTagToken, vendor.Body, vendor.Body.String())
				return false
			}
			return true
		case vendor.Address, vendor.Article, vendor.Aside, vendor.Blockquote, vendor.Button, vendor.Center, vendor.Details, vendor.Dir, vendor.Div, vendor.Dl, vendor.Fieldset, vendor.Figcaption, vendor.Figure, vendor.Footer, vendor.Header, vendor.Hgroup, vendor.Listing, vendor.Menu, vendor.Nav, vendor.Ol, vendor.Pre, vendor.Section, vendor.Summary, vendor.Ul:
			p.popUntil(defaultScope, p.tok.DataAtom)
		case vendor.Form:
			if p.oe.contains(vendor.Template) {
				i := p.indexOfElementInScope(defaultScope, vendor.Form)
				if i == -1 {
					// Ignore the token.
					return true
				}
				p.generateImpliedEndTags()
				if p.oe[i].DataAtom != vendor.Form {
					// Ignore the token.
					return true
				}
				p.popUntil(defaultScope, vendor.Form)
			} else {
				node := p.form
				p.form = nil
				i := p.indexOfElementInScope(defaultScope, vendor.Form)
				if node == nil || i == -1 || p.oe[i] != node {
					// Ignore the token.
					return true
				}
				p.generateImpliedEndTags()
				p.oe.remove(node)
			}
		case vendor.P:
			if !p.elementInScope(buttonScope, vendor.P) {
				p.parseImpliedToken(vendor.StartTagToken, vendor.P, vendor.P.String())
			}
			p.popUntil(buttonScope, vendor.P)
		case vendor.Li:
			p.popUntil(listItemScope, vendor.Li)
		case vendor.Dd, vendor.Dt:
			p.popUntil(defaultScope, p.tok.DataAtom)
		case vendor.H1, vendor.H2, vendor.H3, vendor.H4, vendor.H5, vendor.H6:
			p.popUntil(defaultScope, vendor.H1, vendor.H2, vendor.H3, vendor.H4, vendor.H5, vendor.H6)
		case vendor.A, vendor.B, vendor.Big, vendor.Code, vendor.Em, vendor.Font, vendor.I, vendor.Nobr, vendor.S, vendor.Small, vendor.Strike, vendor.Strong, vendor.Tt, vendor.U:
			p.inBodyEndTagFormatting(p.tok.DataAtom, p.tok.Data)
		case vendor.Applet, vendor.Marquee, vendor.Object:
			if p.popUntil(defaultScope, p.tok.DataAtom) {
				p.clearActiveFormattingElements()
			}
		case vendor.Br:
			p.tok.Type = vendor.StartTagToken
			return false
		case vendor.Template:
			return inHeadIM(p)
		default:
			p.inBodyEndTagOther(p.tok.DataAtom, p.tok.Data)
		}
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
	case vendor.ErrorToken:
		// TODO: remove this divergence from the HTML5 spec.
		if len(p.templateStack) > 0 {
			p.im = inTemplateIM
			return false
		} else {
			for _, e := range p.oe {
				switch e.DataAtom {
				case vendor.Dd, vendor.Dt, vendor.Li, vendor.Optgroup, vendor.Option, vendor.P, vendor.Rb, vendor.Rp, vendor.Rt, vendor.Rtc, vendor.Tbody, vendor.Td, vendor.Tfoot, vendor.Th,
					vendor.Thead, vendor.Tr, vendor.Body, vendor.Html:
				default:
					return true
				}
			}
		}
	}

	return true
}

func (p *parser) inBodyEndTagFormatting(tagAtom vendor.Atom, tagName string) {
	// This is the "adoption agency" algorithm, described at
	// https://html.spec.whatwg.org/multipage/syntax.html#adoptionAgency

	// TODO: this is a fairly literal line-by-line translation of that algorithm.
	// Once the code successfully parses the comprehensive test suite, we should
	// refactor this code to be more idiomatic.

	// Steps 1-4. The outer loop.
	for i := 0; i < 8; i++ {
		// Step 5. Find the formatting element.
		var formattingElement *vendor.Node
		for j := len(p.afe) - 1; j >= 0; j-- {
			if p.afe[j].Type == vendor.scopeMarkerNode {
				break
			}
			if p.afe[j].DataAtom == tagAtom {
				formattingElement = p.afe[j]
				break
			}
		}
		if formattingElement == nil {
			p.inBodyEndTagOther(tagAtom, tagName)
			return
		}
		feIndex := p.oe.index(formattingElement)
		if feIndex == -1 {
			p.afe.remove(formattingElement)
			return
		}
		if !p.elementInScope(defaultScope, tagAtom) {
			// Ignore the tag.
			return
		}

		// Steps 9-10. Find the furthest block.
		var furthestBlock *vendor.Node
		for _, e := range p.oe[feIndex:] {
			if vendor.isSpecialElement(e) {
				furthestBlock = e
				break
			}
		}
		if furthestBlock == nil {
			e := p.oe.pop()
			for e != formattingElement {
				e = p.oe.pop()
			}
			p.afe.remove(e)
			return
		}

		// Steps 11-12. Find the common ancestor and bookmark node.
		commonAncestor := p.oe[feIndex-1]
		bookmark := p.afe.index(formattingElement)

		// Step 13. The inner loop. Find the lastNode to reparent.
		lastNode := furthestBlock
		node := furthestBlock
		x := p.oe.index(node)
		// Steps 13.1-13.2
		for j := 0; j < 3; j++ {
			// Step 13.3.
			x--
			node = p.oe[x]
			// Step 13.4 - 13.5.
			if p.afe.index(node) == -1 {
				p.oe.remove(node)
				continue
			}
			// Step 13.6.
			if node == formattingElement {
				break
			}
			// Step 13.7.
			clone := node.clone()
			p.afe[p.afe.index(node)] = clone
			p.oe[p.oe.index(node)] = clone
			node = clone
			// Step 13.8.
			if lastNode == furthestBlock {
				bookmark = p.afe.index(node) + 1
			}
			// Step 13.9.
			if lastNode.Parent != nil {
				lastNode.Parent.RemoveChild(lastNode)
			}
			node.AppendChild(lastNode)
			// Step 13.10.
			lastNode = node
		}

		// Step 14. Reparent lastNode to the common ancestor,
		// or for misnested table nodes, to the foster parent.
		if lastNode.Parent != nil {
			lastNode.Parent.RemoveChild(lastNode)
		}
		switch commonAncestor.DataAtom {
		case vendor.Table, vendor.Tbody, vendor.Tfoot, vendor.Thead, vendor.Tr:
			p.fosterParent(lastNode)
		default:
			commonAncestor.AppendChild(lastNode)
		}

		// Steps 15-17. Reparent nodes from the furthest block's children
		// to a clone of the formatting element.
		clone := formattingElement.clone()
		vendor.reparentChildren(clone, furthestBlock)
		furthestBlock.AppendChild(clone)

		// Step 18. Fix up the list of active formatting elements.
		if oldLoc := p.afe.index(formattingElement); oldLoc != -1 && oldLoc < bookmark {
			// Move the bookmark with the rest of the list.
			bookmark--
		}
		p.afe.remove(formattingElement)
		p.afe.insert(bookmark, clone)

		// Step 19. Fix up the stack of open elements.
		p.oe.remove(formattingElement)
		p.oe.insert(p.oe.index(furthestBlock)+1, clone)
	}
}

// inBodyEndTagOther performs the "any other end tag" algorithm for inBodyIM.
// "Any other end tag" handling from 12.2.6.5 The rules for parsing tokens in foreign content
// https://html.spec.whatwg.org/multipage/syntax.html#parsing-main-inforeign
func (p *parser) inBodyEndTagOther(tagAtom vendor.Atom, tagName string) {
	for i := len(p.oe) - 1; i >= 0; i-- {
		// Two element nodes have the same tag if they have the same Data (a
		// string-typed field). As an optimization, for common HTML tags, each
		// Data string is assigned a unique, non-zero DataAtom (a uint32-typed
		// field), since integer comparison is faster than string comparison.
		// Uncommon (custom) tags get a zero DataAtom.
		//
		// The if condition here is equivalent to (p.oe[i].Data == tagName).
		if (p.oe[i].DataAtom == tagAtom) &&
			((tagAtom != 0) || (p.oe[i].Data == tagName)) {
			p.oe = p.oe[:i]
			break
		}
		if vendor.isSpecialElement(p.oe[i]) {
			break
		}
	}
}

// Section 12.2.6.4.8.
func textIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.ErrorToken:
		p.oe.pop()
	case vendor.TextToken:
		d := p.tok.Data
		if n := p.oe.top(); n.DataAtom == vendor.Textarea && n.FirstChild == nil {
			// Ignore a newline at the start of a <textarea> block.
			if d != "" && d[0] == '\r' {
				d = d[1:]
			}
			if d != "" && d[0] == '\n' {
				d = d[1:]
			}
		}
		if d == "" {
			return true
		}
		p.addText(d)
		return true
	case vendor.EndTagToken:
		p.oe.pop()
	}
	p.im = p.originalIM
	p.originalIM = nil
	return p.tok.Type == vendor.EndTagToken
}

// Section 12.2.6.4.9.
func inTableIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		p.tok.Data = strings.Replace(p.tok.Data, "\x00", "", -1)
		switch p.oe.top().DataAtom {
		case vendor.Table, vendor.Tbody, vendor.Tfoot, vendor.Thead, vendor.Tr:
			if strings.Trim(p.tok.Data, whitespace) == "" {
				p.addText(p.tok.Data)
				return true
			}
		}
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Caption:
			p.clearStackToContext(tableScope)
			p.afe = append(p.afe, &vendor.scopeMarker)
			p.addElement()
			p.im = inCaptionIM
			return true
		case vendor.Colgroup:
			p.clearStackToContext(tableScope)
			p.addElement()
			p.im = inColumnGroupIM
			return true
		case vendor.Col:
			p.parseImpliedToken(vendor.StartTagToken, vendor.Colgroup, vendor.Colgroup.String())
			return false
		case vendor.Tbody, vendor.Tfoot, vendor.Thead:
			p.clearStackToContext(tableScope)
			p.addElement()
			p.im = inTableBodyIM
			return true
		case vendor.Td, vendor.Th, vendor.Tr:
			p.parseImpliedToken(vendor.StartTagToken, vendor.Tbody, vendor.Tbody.String())
			return false
		case vendor.Table:
			if p.popUntil(tableScope, vendor.Table) {
				p.resetInsertionMode()
				return false
			}
			// Ignore the token.
			return true
		case vendor.Style, vendor.Script, vendor.Template:
			return inHeadIM(p)
		case vendor.Input:
			for _, t := range p.tok.Attr {
				if t.Key == "type" && strings.ToLower(t.Val) == "hidden" {
					p.addElement()
					p.oe.pop()
					return true
				}
			}
			// Otherwise drop down to the default action.
		case vendor.Form:
			if p.oe.contains(vendor.Template) || p.form != nil {
				// Ignore the token.
				return true
			}
			p.addElement()
			p.form = p.oe.pop()
		case vendor.Select:
			p.reconstructActiveFormattingElements()
			switch p.top().DataAtom {
			case vendor.Table, vendor.Tbody, vendor.Tfoot, vendor.Thead, vendor.Tr:
				p.fosterParenting = true
			}
			p.addElement()
			p.fosterParenting = false
			p.framesetOK = false
			p.im = inSelectInTableIM
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Table:
			if p.popUntil(tableScope, vendor.Table) {
				p.resetInsertionMode()
				return true
			}
			// Ignore the token.
			return true
		case vendor.Body, vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Html, vendor.Tbody, vendor.Td, vendor.Tfoot, vendor.Th, vendor.Thead, vendor.Tr:
			// Ignore the token.
			return true
		case vendor.Template:
			return inHeadIM(p)
		}
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	case vendor.DoctypeToken:
		// Ignore the token.
		return true
	case vendor.ErrorToken:
		return inBodyIM(p)
	}

	p.fosterParenting = true
	defer func() { p.fosterParenting = false }()

	return inBodyIM(p)
}

// Section 12.2.6.4.11.
func inCaptionIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Tbody, vendor.Td, vendor.Tfoot, vendor.Thead, vendor.Tr:
			if p.popUntil(tableScope, vendor.Caption) {
				p.clearActiveFormattingElements()
				p.im = inTableIM
				return false
			} else {
				// Ignore the token.
				return true
			}
		case vendor.Select:
			p.reconstructActiveFormattingElements()
			p.addElement()
			p.framesetOK = false
			p.im = inSelectInTableIM
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Caption:
			if p.popUntil(tableScope, vendor.Caption) {
				p.clearActiveFormattingElements()
				p.im = inTableIM
			}
			return true
		case vendor.Table:
			if p.popUntil(tableScope, vendor.Caption) {
				p.clearActiveFormattingElements()
				p.im = inTableIM
				return false
			} else {
				// Ignore the token.
				return true
			}
		case vendor.Body, vendor.Col, vendor.Colgroup, vendor.Html, vendor.Tbody, vendor.Td, vendor.Tfoot, vendor.Th, vendor.Thead, vendor.Tr:
			// Ignore the token.
			return true
		}
	}
	return inBodyIM(p)
}

// Section 12.2.6.4.12.
func inColumnGroupIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		s := strings.TrimLeft(p.tok.Data, whitespace)
		if len(s) < len(p.tok.Data) {
			// Add the initial whitespace to the current node.
			p.addText(p.tok.Data[:len(p.tok.Data)-len(s)])
			if s == "" {
				return true
			}
			p.tok.Data = s
		}
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	case vendor.DoctypeToken:
		// Ignore the token.
		return true
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			return inBodyIM(p)
		case vendor.Col:
			p.addElement()
			p.oe.pop()
			p.acknowledgeSelfClosingTag()
			return true
		case vendor.Template:
			return inHeadIM(p)
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Colgroup:
			if p.oe.top().DataAtom == vendor.Colgroup {
				p.oe.pop()
				p.im = inTableIM
			}
			return true
		case vendor.Col:
			// Ignore the token.
			return true
		case vendor.Template:
			return inHeadIM(p)
		}
	case vendor.ErrorToken:
		return inBodyIM(p)
	}
	if p.oe.top().DataAtom != vendor.Colgroup {
		return true
	}
	p.oe.pop()
	p.im = inTableIM
	return false
}

// Section 12.2.6.4.13.
func inTableBodyIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Tr:
			p.clearStackToContext(tableBodyScope)
			p.addElement()
			p.im = inRowIM
			return true
		case vendor.Td, vendor.Th:
			p.parseImpliedToken(vendor.StartTagToken, vendor.Tr, vendor.Tr.String())
			return false
		case vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Tbody, vendor.Tfoot, vendor.Thead:
			if p.popUntil(tableScope, vendor.Tbody, vendor.Thead, vendor.Tfoot) {
				p.im = inTableIM
				return false
			}
			// Ignore the token.
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Tbody, vendor.Tfoot, vendor.Thead:
			if p.elementInScope(tableScope, p.tok.DataAtom) {
				p.clearStackToContext(tableBodyScope)
				p.oe.pop()
				p.im = inTableIM
			}
			return true
		case vendor.Table:
			if p.popUntil(tableScope, vendor.Tbody, vendor.Thead, vendor.Tfoot) {
				p.im = inTableIM
				return false
			}
			// Ignore the token.
			return true
		case vendor.Body, vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Html, vendor.Td, vendor.Th, vendor.Tr:
			// Ignore the token.
			return true
		}
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	}

	return inTableIM(p)
}

// Section 12.2.6.4.14.
func inRowIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Td, vendor.Th:
			p.clearStackToContext(tableRowScope)
			p.addElement()
			p.afe = append(p.afe, &vendor.scopeMarker)
			p.im = inCellIM
			return true
		case vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Tbody, vendor.Tfoot, vendor.Thead, vendor.Tr:
			if p.popUntil(tableScope, vendor.Tr) {
				p.im = inTableBodyIM
				return false
			}
			// Ignore the token.
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Tr:
			if p.popUntil(tableScope, vendor.Tr) {
				p.im = inTableBodyIM
				return true
			}
			// Ignore the token.
			return true
		case vendor.Table:
			if p.popUntil(tableScope, vendor.Tr) {
				p.im = inTableBodyIM
				return false
			}
			// Ignore the token.
			return true
		case vendor.Tbody, vendor.Tfoot, vendor.Thead:
			if p.elementInScope(tableScope, p.tok.DataAtom) {
				p.parseImpliedToken(vendor.EndTagToken, vendor.Tr, vendor.Tr.String())
				return false
			}
			// Ignore the token.
			return true
		case vendor.Body, vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Html, vendor.Td, vendor.Th:
			// Ignore the token.
			return true
		}
	}

	return inTableIM(p)
}

// Section 12.2.6.4.15.
func inCellIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Tbody, vendor.Td, vendor.Tfoot, vendor.Th, vendor.Thead, vendor.Tr:
			if p.popUntil(tableScope, vendor.Td, vendor.Th) {
				// Close the cell and reprocess.
				p.clearActiveFormattingElements()
				p.im = inRowIM
				return false
			}
			// Ignore the token.
			return true
		case vendor.Select:
			p.reconstructActiveFormattingElements()
			p.addElement()
			p.framesetOK = false
			p.im = inSelectInTableIM
			return true
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Td, vendor.Th:
			if !p.popUntil(tableScope, p.tok.DataAtom) {
				// Ignore the token.
				return true
			}
			p.clearActiveFormattingElements()
			p.im = inRowIM
			return true
		case vendor.Body, vendor.Caption, vendor.Col, vendor.Colgroup, vendor.Html:
			// Ignore the token.
			return true
		case vendor.Table, vendor.Tbody, vendor.Tfoot, vendor.Thead, vendor.Tr:
			if !p.elementInScope(tableScope, p.tok.DataAtom) {
				// Ignore the token.
				return true
			}
			// Close the cell and reprocess.
			if p.popUntil(tableScope, vendor.Td, vendor.Th) {
				p.clearActiveFormattingElements()
			}
			p.im = inRowIM
			return false
		}
	}
	return inBodyIM(p)
}

// Section 12.2.6.4.16.
func inSelectIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		p.addText(strings.Replace(p.tok.Data, "\x00", "", -1))
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			return inBodyIM(p)
		case vendor.Option:
			if p.top().DataAtom == vendor.Option {
				p.oe.pop()
			}
			p.addElement()
		case vendor.Optgroup:
			if p.top().DataAtom == vendor.Option {
				p.oe.pop()
			}
			if p.top().DataAtom == vendor.Optgroup {
				p.oe.pop()
			}
			p.addElement()
		case vendor.Select:
			if p.popUntil(selectScope, vendor.Select) {
				p.resetInsertionMode()
			} else {
				// Ignore the token.
				return true
			}
		case vendor.Input, vendor.Keygen, vendor.Textarea:
			if p.elementInScope(selectScope, vendor.Select) {
				p.parseImpliedToken(vendor.EndTagToken, vendor.Select, vendor.Select.String())
				return false
			}
			// In order to properly ignore <textarea>, we need to change the tokenizer mode.
			p.tokenizer.NextIsNotRawText()
			// Ignore the token.
			return true
		case vendor.Script, vendor.Template:
			return inHeadIM(p)
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Option:
			if p.top().DataAtom == vendor.Option {
				p.oe.pop()
			}
		case vendor.Optgroup:
			i := len(p.oe) - 1
			if p.oe[i].DataAtom == vendor.Option {
				i--
			}
			if p.oe[i].DataAtom == vendor.Optgroup {
				p.oe = p.oe[:i]
			}
		case vendor.Select:
			if p.popUntil(selectScope, vendor.Select) {
				p.resetInsertionMode()
			} else {
				// Ignore the token.
				return true
			}
		case vendor.Template:
			return inHeadIM(p)
		}
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
	case vendor.DoctypeToken:
		// Ignore the token.
		return true
	case vendor.ErrorToken:
		return inBodyIM(p)
	}

	return true
}

// Section 12.2.6.4.17.
func inSelectInTableIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.StartTagToken, vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Caption, vendor.Table, vendor.Tbody, vendor.Tfoot, vendor.Thead, vendor.Tr, vendor.Td, vendor.Th:
			if p.tok.Type == vendor.EndTagToken && !p.elementInScope(tableScope, p.tok.DataAtom) {
				// Ignore the token.
				return true
			}
			// This is like p.popUntil(selectScope, a.Select), but it also
			// matches <math select>, not just <select>. Matching the MathML
			// tag is arguably incorrect (conceptually), but it mimics what
			// Chromium does.
			for i := len(p.oe) - 1; i >= 0; i-- {
				if n := p.oe[i]; n.DataAtom == vendor.Select {
					p.oe = p.oe[:i]
					break
				}
			}
			p.resetInsertionMode()
			return false
		}
	}
	return inSelectIM(p)
}

// Section 12.2.6.4.18.
func inTemplateIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken, vendor.CommentToken, vendor.DoctypeToken:
		return inBodyIM(p)
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Base, vendor.Basefont, vendor.Bgsound, vendor.Link, vendor.Meta, vendor.Noframes, vendor.Script, vendor.Style, vendor.Template, vendor.Title:
			return inHeadIM(p)
		case vendor.Caption, vendor.Colgroup, vendor.Tbody, vendor.Tfoot, vendor.Thead:
			p.templateStack.pop()
			p.templateStack = append(p.templateStack, inTableIM)
			p.im = inTableIM
			return false
		case vendor.Col:
			p.templateStack.pop()
			p.templateStack = append(p.templateStack, inColumnGroupIM)
			p.im = inColumnGroupIM
			return false
		case vendor.Tr:
			p.templateStack.pop()
			p.templateStack = append(p.templateStack, inTableBodyIM)
			p.im = inTableBodyIM
			return false
		case vendor.Td, vendor.Th:
			p.templateStack.pop()
			p.templateStack = append(p.templateStack, inRowIM)
			p.im = inRowIM
			return false
		default:
			p.templateStack.pop()
			p.templateStack = append(p.templateStack, inBodyIM)
			p.im = inBodyIM
			return false
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Template:
			return inHeadIM(p)
		default:
			// Ignore the token.
			return true
		}
	case vendor.ErrorToken:
		if !p.oe.contains(vendor.Template) {
			// Ignore the token.
			return true
		}
		// TODO: remove this divergence from the HTML5 spec.
		//
		// See https://bugs.chromium.org/p/chromium/issues/detail?id=829668
		p.generateImpliedEndTags()
		for i := len(p.oe) - 1; i >= 0; i-- {
			if n := p.oe[i]; n.Namespace == "" && n.DataAtom == vendor.Template {
				p.oe = p.oe[:i]
				break
			}
		}
		p.clearActiveFormattingElements()
		p.templateStack.pop()
		p.resetInsertionMode()
		return false
	}
	return false
}

// Section 12.2.6.4.19.
func afterBodyIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.ErrorToken:
		// Stop parsing.
		return true
	case vendor.TextToken:
		s := strings.TrimLeft(p.tok.Data, whitespace)
		if len(s) == 0 {
			// It was all whitespace.
			return inBodyIM(p)
		}
	case vendor.StartTagToken:
		if p.tok.DataAtom == vendor.Html {
			return inBodyIM(p)
		}
	case vendor.EndTagToken:
		if p.tok.DataAtom == vendor.Html {
			if !p.fragment {
				p.im = afterAfterBodyIM
			}
			return true
		}
	case vendor.CommentToken:
		// The comment is attached to the <html> element.
		if len(p.oe) < 1 || p.oe[0].DataAtom != vendor.Html {
			panic("html: bad parser state: <html> element not found, in the after-body insertion mode")
		}
		p.oe[0].AppendChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	}
	p.im = inBodyIM
	return false
}

// Section 12.2.6.4.20.
func inFramesetIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
	case vendor.TextToken:
		// Ignore all text but whitespace.
		s := strings.Map(func(c rune) rune {
			switch c {
			case ' ', '\t', '\n', '\f', '\r':
				return c
			}
			return -1
		}, p.tok.Data)
		if s != "" {
			p.addText(s)
		}
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			return inBodyIM(p)
		case vendor.Frameset:
			p.addElement()
		case vendor.Frame:
			p.addElement()
			p.oe.pop()
			p.acknowledgeSelfClosingTag()
		case vendor.Noframes:
			return inHeadIM(p)
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Frameset:
			if p.oe.top().DataAtom != vendor.Html {
				p.oe.pop()
				if p.oe.top().DataAtom != vendor.Frameset {
					p.im = afterFramesetIM
					return true
				}
			}
		}
	default:
		// Ignore the token.
	}
	return true
}

// Section 12.2.6.4.21.
func afterFramesetIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
	case vendor.TextToken:
		// Ignore all text but whitespace.
		s := strings.Map(func(c rune) rune {
			switch c {
			case ' ', '\t', '\n', '\f', '\r':
				return c
			}
			return -1
		}, p.tok.Data)
		if s != "" {
			p.addText(s)
		}
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			return inBodyIM(p)
		case vendor.Noframes:
			return inHeadIM(p)
		}
	case vendor.EndTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			p.im = afterAfterFramesetIM
			return true
		}
	default:
		// Ignore the token.
	}
	return true
}

// Section 12.2.6.4.22.
func afterAfterBodyIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.ErrorToken:
		// Stop parsing.
		return true
	case vendor.TextToken:
		s := strings.TrimLeft(p.tok.Data, whitespace)
		if len(s) == 0 {
			// It was all whitespace.
			return inBodyIM(p)
		}
	case vendor.StartTagToken:
		if p.tok.DataAtom == vendor.Html {
			return inBodyIM(p)
		}
	case vendor.CommentToken:
		p.doc.AppendChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
		return true
	case vendor.DoctypeToken:
		return inBodyIM(p)
	}
	p.im = inBodyIM
	return false
}

// Section 12.2.6.4.23.
func afterAfterFramesetIM(p *parser) bool {
	switch p.tok.Type {
	case vendor.CommentToken:
		p.doc.AppendChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
	case vendor.TextToken:
		// Ignore all text but whitespace.
		s := strings.Map(func(c rune) rune {
			switch c {
			case ' ', '\t', '\n', '\f', '\r':
				return c
			}
			return -1
		}, p.tok.Data)
		if s != "" {
			p.tok.Data = s
			return inBodyIM(p)
		}
	case vendor.StartTagToken:
		switch p.tok.DataAtom {
		case vendor.Html:
			return inBodyIM(p)
		case vendor.Noframes:
			return inHeadIM(p)
		}
	case vendor.DoctypeToken:
		return inBodyIM(p)
	default:
		// Ignore the token.
	}
	return true
}

const whitespaceOrNUL = whitespace + "\x00"

// Section 12.2.6.5
func parseForeignContent(p *parser) bool {
	switch p.tok.Type {
	case vendor.TextToken:
		if p.framesetOK {
			p.framesetOK = strings.TrimLeft(p.tok.Data, whitespaceOrNUL) == ""
		}
		p.tok.Data = strings.Replace(p.tok.Data, "\x00", "\ufffd", -1)
		p.addText(p.tok.Data)
	case vendor.CommentToken:
		p.addChild(&vendor.Node{
			Type: vendor.CommentNode,
			Data: p.tok.Data,
		})
	case vendor.StartTagToken:
		b := vendor.breakout[p.tok.Data]
		if p.tok.DataAtom == vendor.Font {
		loop:
			for _, attr := range p.tok.Attr {
				switch attr.Key {
				case "color", "face", "size":
					b = true
					break loop
				}
			}
		}
		if b {
			for i := len(p.oe) - 1; i >= 0; i-- {
				n := p.oe[i]
				if n.Namespace == "" || vendor.htmlIntegrationPoint(n) || vendor.mathMLTextIntegrationPoint(n) {
					p.oe = p.oe[:i+1]
					break
				}
			}
			return false
		}
		switch p.top().Namespace {
		case "math":
			vendor.adjustAttributeNames(p.tok.Attr, vendor.mathMLAttributeAdjustments)
		case "svg":
			// Adjust SVG tag names. The tokenizer lower-cases tag names, but
			// SVG wants e.g. "foreignObject" with a capital second "O".
			if x := vendor.svgTagNameAdjustments[p.tok.Data]; x != "" {
				p.tok.DataAtom = vendor.Lookup([]byte(x))
				p.tok.Data = x
			}
			vendor.adjustAttributeNames(p.tok.Attr, vendor.svgAttributeAdjustments)
		default:
			panic("html: bad parser state: unexpected namespace")
		}
		vendor.adjustForeignAttributes(p.tok.Attr)
		namespace := p.top().Namespace
		p.addElement()
		p.top().Namespace = namespace
		if namespace != "" {
			// Don't let the tokenizer go into raw text mode in foreign content
			// (e.g. in an SVG <title> tag).
			p.tokenizer.NextIsNotRawText()
		}
		if p.hasSelfClosingToken {
			p.oe.pop()
			p.acknowledgeSelfClosingTag()
		}
	case vendor.EndTagToken:
		for i := len(p.oe) - 1; i >= 0; i-- {
			if p.oe[i].Namespace == "" {
				return p.im(p)
			}
			if strings.EqualFold(p.oe[i].Data, p.tok.Data) {
				p.oe = p.oe[:i]
				break
			}
		}
		return true
	default:
		// Ignore the token.
	}
	return true
}

// Section 12.2.6.
func (p *parser) inForeignContent() bool {
	if len(p.oe) == 0 {
		return false
	}
	n := p.oe[len(p.oe)-1]
	if n.Namespace == "" {
		return false
	}
	if vendor.mathMLTextIntegrationPoint(n) {
		if p.tok.Type == vendor.StartTagToken && p.tok.DataAtom != vendor.Mglyph && p.tok.DataAtom != vendor.Malignmark {
			return false
		}
		if p.tok.Type == vendor.TextToken {
			return false
		}
	}
	if n.Namespace == "math" && n.DataAtom == vendor.AnnotationXml && p.tok.Type == vendor.StartTagToken && p.tok.DataAtom == vendor.Svg {
		return false
	}
	if vendor.htmlIntegrationPoint(n) && (p.tok.Type == vendor.StartTagToken || p.tok.Type == vendor.TextToken) {
		return false
	}
	if p.tok.Type == vendor.ErrorToken {
		return false
	}
	return true
}

// parseImpliedToken parses a token as though it had appeared in the parser's
// input.
func (p *parser) parseImpliedToken(t vendor.TokenType, dataAtom vendor.Atom, data string) {
	realToken, selfClosing := p.tok, p.hasSelfClosingToken
	p.tok = vendor.Token{
		Type:     t,
		DataAtom: dataAtom,
		Data:     data,
	}
	p.hasSelfClosingToken = false
	p.parseCurrentToken()
	p.tok, p.hasSelfClosingToken = realToken, selfClosing
}

// parseCurrentToken runs the current token through the parsing routines
// until it is consumed.
func (p *parser) parseCurrentToken() {
	if p.tok.Type == vendor.SelfClosingTagToken {
		p.hasSelfClosingToken = true
		p.tok.Type = vendor.StartTagToken
	}

	consumed := false
	for !consumed {
		if p.inForeignContent() {
			consumed = parseForeignContent(p)
		} else {
			consumed = p.im(p)
		}
	}

	if p.hasSelfClosingToken {
		// This is a parse error, but ignore it.
		p.hasSelfClosingToken = false
	}
}

func (p *parser) parse() error {
	// Iterate until EOF. Any other error will cause an early return.
	var err error
	for err != io.EOF {
		// CDATA sections are allowed only in foreign content.
		n := p.oe.top()
		p.tokenizer.AllowCDATA(n != nil && n.Namespace != "")
		// Read and parse the next token.
		p.tokenizer.Next()
		p.tok = p.tokenizer.Token()
		if p.tok.Type == vendor.ErrorToken {
			err = p.tokenizer.Err()
			if err != nil && err != io.EOF {
				return err
			}
		}
		p.parseCurrentToken()
	}
	return nil
}

// Parse returns the parse tree for the HTML from the given Reader.
//
// It implements the HTML5 parsing algorithm
// (https://html.spec.whatwg.org/multipage/syntax.html#tree-construction),
// which is very complicated. The resultant tree can contain implicitly created
// nodes that have no explicit <tag> listed in r's data, and nodes' parents can
// differ from the nesting implied by a naive processing of start and end
// <tag>s. Conversely, explicit <tag>s in r's data can be silently dropped,
// with no corresponding node in the resulting tree.
//
// The input is assumed to be UTF-8 encoded.
func Parse(r io.Reader) (*vendor.Node, error) {
	return ParseWithOptions(r)
}

// ParseFragment parses a fragment of HTML and returns the nodes that were
// found. If the fragment is the InnerHTML for an existing element, pass that
// element in context.
//
// It has the same intricacies as Parse.
func ParseFragment(r io.Reader, context *vendor.Node) ([]*vendor.Node, error) {
	return ParseFragmentWithOptions(r, context)
}

// ParseOption configures a parser.
type ParseOption func(p *parser)

// ParseOptionEnableScripting configures the scripting flag.
// https://html.spec.whatwg.org/multipage/webappapis.html#enabling-and-disabling-scripting
//
// By default, scripting is enabled.
func ParseOptionEnableScripting(enable bool) ParseOption {
	return func(p *parser) {
		p.scripting = enable
	}
}

// ParseWithOptions is like Parse, with options.
func ParseWithOptions(r io.Reader, opts ...ParseOption) (*vendor.Node, error) {
	p := &parser{
		tokenizer: vendor.NewTokenizer(r),
		doc: &vendor.Node{
			Type: vendor.DocumentNode,
		},
		scripting:  true,
		framesetOK: true,
		im:         initialIM,
	}

	for _, f := range opts {
		f(p)
	}

	err := p.parse()
	if err != nil {
		return nil, err
	}
	return p.doc, nil
}

// ParseFragmentWithOptions is like ParseFragment, with options.
func ParseFragmentWithOptions(r io.Reader, context *vendor.Node, opts ...ParseOption) ([]*vendor.Node, error) {
	contextTag := ""
	if context != nil {
		if context.Type != vendor.ElementNode {
			return nil, errors.New("html: ParseFragment of non-element Node")
		}
		// The next check isn't just context.DataAtom.String() == context.Data because
		// it is valid to pass an element whose tag isn't a known atom. For example,
		// DataAtom == 0 and Data = "tagfromthefuture" is perfectly consistent.
		if context.DataAtom != vendor.Lookup([]byte(context.Data)) {
			return nil, fmt.Errorf("html: inconsistent Node: DataAtom=%q, Data=%q", context.DataAtom, context.Data)
		}
		contextTag = context.DataAtom.String()
	}
	p := &parser{
		tokenizer: vendor.NewTokenizerFragment(r, contextTag),
		doc: &vendor.Node{
			Type: vendor.DocumentNode,
		},
		scripting: true,
		fragment:  true,
		context:   context,
	}

	for _, f := range opts {
		f(p)
	}

	root := &vendor.Node{
		Type:     vendor.ElementNode,
		DataAtom: vendor.Html,
		Data:     vendor.Html.String(),
	}
	p.doc.AppendChild(root)
	p.oe = vendor.nodeStack{root}
	if context != nil && context.DataAtom == vendor.Template {
		p.templateStack = append(p.templateStack, inTemplateIM)
	}
	p.resetInsertionMode()

	for n := context; n != nil; n = n.Parent {
		if n.Type == vendor.ElementNode && n.DataAtom == vendor.Form {
			p.form = n
			break
		}
	}

	err := p.parse()
	if err != nil {
		return nil, err
	}

	parent := p.doc
	if context != nil {
		parent = root
	}

	var result []*vendor.Node
	for c := parent.FirstChild; c != nil; {
		next := c.NextSibling
		parent.RemoveChild(c)
		result = append(result, c)
		c = next
	}
	return result, nil
}
