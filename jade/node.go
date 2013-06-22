// Copyright 2011 The Go Authors.
// 2013 Benjamin Gentil
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parse nodes.

package jade

import (
	"bytes"
	"fmt"

//	"strconv"
//	"strings"
)

var doctypes = map[string]string{
	"5":            `<!DOCTYPE html>`,
	"default":      `<!DOCTYPE html>`,
	"xml":          `<?xml version="1.0" encoding="utf-8" ?>`,
	"transitional": `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">`,
	"strict":       `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">`,
	"frameset":     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Frameset//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd">`,
	"1.1":          `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">`,
	"basic":        `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML Basic 1.1//EN" "http://www.w3.org/TR/xhtml-basic/xhtml-basic11.dtd">`,
	"mobile":       `<!DOCTYPE html PUBLIC "-//WAPFORUM//DTD XHTML Mobile 1.2//EN" "http://www.openmobilealliance.org/tech/DTD/xhtml-mobile12.dtd">`,
}

// A node is an element in the parse tree. The interface is trivial.
type Node interface {
	Type() NodeType
	String() string
	HTMLString() string
	// Copy does a deep copy of the Node and all its components.
	// To avoid type assertions, some XxxNodes also have specialized
	// CopyXxx methods that return *XxxNode.
	Copy() Node
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeText NodeType = iota
	NodeTag
	NodeAttr
	NodeDoctype
	NodeId
	NodeClass
	NodeList
)

var nodeName = map[NodeType]string{
	NodeText:    "text",
	NodeTag:     "tag",
	NodeAttr:    "attr",
	NodeDoctype: "doctype",
	NodeId:      "id",
	NodeClass:   "class",
	NodeList:    "List",
}

func (i NodeType) String() string {
	s := nodeName[i]
	if s == "" {
		return fmt.Sprintf("node%d", int(i))
	}
	return s
}

// Nodes.

// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Nodes []Node // The element nodes in lexical order.
}

func newList() *ListNode {
	return &ListNode{NodeType: NodeList}
}

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

func (l *ListNode) String() string {
	b := new(bytes.Buffer)
	for _, n := range l.Nodes {
		fmt.Fprint(b, n)
	}
	return b.String()
}

func (l *ListNode) HTMLString() string {
	b := new(bytes.Buffer)
	for _, n := range l.Nodes {
		fmt.Fprint(b, n.HTMLString())
	}
	return b.String()
}

func (l *ListNode) CopyList() *ListNode {
	if l == nil {
		return l
	}
	n := newList()
	for _, elem := range l.Nodes {
		n.append(elem.Copy())
	}
	return n
}

func (l *ListNode) Copy() Node {
	return l.CopyList()
}

// TagNode
type TagNode struct {
	NodeType
	Tag   []byte
	Nodes []Node // The element nodes in lexical order.
}

func newTag(tag string) *TagNode {
	return &TagNode{NodeType: NodeTag, Tag: []byte(tag)}
}

func (l *TagNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

func (l *TagNode) String() string {
	return fmt.Sprintf("%s", l.Tag)
}

func (l *TagNode) HTMLString() string {
	b := new(bytes.Buffer)
	n_classes := 0
	classes := new(bytes.Buffer)
	fmt.Fprint(b, fmt.Sprintf("<%s", l.Tag))
	for _, n := range l.Nodes {
		if n.Type() == NodeClass {
			if n_classes > 0 {
				fmt.Fprint(classes, " ")
			}
			fmt.Fprintf(classes, "%s", n.HTMLString())
			n_classes++
		}
		if n.Type() == NodeAttr {
			fmt.Fprintf(b, " %s", n.HTMLString())
		}
		if n.Type() == NodeId {
			fmt.Fprintf(b, " id=\"%s\"", n.HTMLString())
		}
	}

	if len(classes.String()) > 0 {
		fmt.Fprintf(b, " class=\"%s\"", classes.String())
	}

	fmt.Fprint(b, ">")

	for _, n := range l.Nodes {
		if n.Type() == NodeText || n.Type() == NodeTag {
			fmt.Fprint(b, n.HTMLString())
		}
	}

	fmt.Fprintf(b, "</%s>", l.Tag)
	return b.String()

	return fmt.Sprintf("%s", l.Tag)
}

func (l *TagNode) CopyTag() *TagNode {
	if l == nil {
		return l
	}
	n := newTag(string(l.Tag))
	for _, elem := range l.Nodes {
		n.append(elem.Copy())
	}
	return n
}

func (l *TagNode) Copy() Node {
	return l.CopyTag()
}

// TextNode
type TextNode struct {
	NodeType
	Text []byte
}

func newText(text string) *TextNode {
	return &TextNode{NodeType: NodeText, Text: []byte(text)}
}

func (t *TextNode) String() string {
	return fmt.Sprintf("%s", t.Text)
}

func (t *TextNode) HTMLString() string {
	return t.String()
}

func (t *TextNode) Copy() Node {
	return &TextNode{NodeType: NodeText, Text: append([]byte{}, t.Text...)}
}

// DoctypeNode
type DoctypeNode struct {
	NodeType
	Doctype []byte
}

func newDoctype(doctype string) *DoctypeNode {
	return &DoctypeNode{NodeType: NodeDoctype, Doctype: []byte(doctype)}
}

func (t *DoctypeNode) String() string {
	return fmt.Sprintf("%s", t.Doctype)
}

func (t *DoctypeNode) HTMLString() string {
	if defined := doctypes[string(t.Doctype)]; len(defined) != 0 {
		return defined
	}

	return fmt.Sprintf("<!DOCTYPE %s >", t.Doctype)
}

func (t *DoctypeNode) Copy() Node {
	return &DoctypeNode{NodeType: NodeDoctype, Doctype: append([]byte{}, t.Doctype...)}
}

// AttrNode
type AttrNode struct {
	NodeType
	Attr []byte
}

func newAttr(attr string) *AttrNode {
	return &AttrNode{NodeType: NodeAttr, Attr: []byte(attr)}
}

func (t *AttrNode) String() string {
	return fmt.Sprintf("%s", t.Attr)
}

func (t *AttrNode) HTMLString() string {
	return t.String()
}

func (t *AttrNode) Copy() Node {
	return &AttrNode{NodeType: NodeAttr, Attr: append([]byte{}, t.Attr...)}
}

// IdNode
type IdNode struct {
	NodeType
	Id []byte
}

func newId(id string) *IdNode {
	return &IdNode{NodeType: NodeId, Id: []byte(id)}
}

func (t *IdNode) String() string {
	return fmt.Sprintf("%s", t.Id)
}

func (t *IdNode) HTMLString() string {
	return t.String()
}

func (t *IdNode) Copy() Node {
	return &IdNode{NodeType: NodeId, Id: append([]byte{}, t.Id...)}
}

// ClassNode
type ClassNode struct {
	NodeType
	Class []byte
}

func newClass(class string) *ClassNode {
	return &ClassNode{NodeType: NodeClass, Class: []byte(class)}
}

func (t *ClassNode) String() string {
	return fmt.Sprintf("%s", t.Class)
}

func (t *ClassNode) HTMLString() string {
	return t.String()
}

func (t *ClassNode) Copy() Node {
	return &ClassNode{NodeType: NodeClass, Class: append([]byte{}, t.Class...)}
}
