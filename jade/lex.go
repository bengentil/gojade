// Copyright 2011 The Go Authors.
// 2013 Benjamin Gentil
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jade

import (
	"fmt"
	"strings"
	//	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemText                  // plain text
	itemEndl
	itemTag // html tag
	itemAttr
	itemIdentSpace
	itemIdentTab
	itemDoctype
	itemComment
	itemBlank
	itemId
	itemClass
	itemEOF // End Of File
)

// Make the types prettyprint.
var itemName = map[itemType]string{
	itemError:      "error",
	itemText:       "text",
	itemEndl:       "endl",
	itemTag:        "tag",
	itemAttr:       "attr",
	itemIdentSpace: "identSpace",
	itemIdentTab:   "identTab",
	itemDoctype:    "doctype",
	itemComment:    "comment",
	itemBlank:      "blank",
	itemId:         "id",
	itemClass:      "class",
	itemEOF:        "EOF",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

/*
var key = map[string]itemType{
	".":        itemDot,
	"define":   itemDefine,
	"else":     itemElse,
	"end":      itemEnd,
	"if":       itemIf,
	"range":    itemRange,
	"template": itemTemplate,
	"with":     itemWith,
}*/

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports.
	input      string    // the string being scanned.
	leftDelim  string    // start of action.
	rightDelim string    // end of action.
	state      stateFn   // the next lexing function to enter.
	pos        int       // current position in the input.
	start      int       // start position of this item.
	width      int       // width of last rune read from input.
	items      chan item // channel of scanned items.
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

// error returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
	panic("not reached")
}

// lex creates a new scanner for the input string.
func lex(name, input, left, right string) *lexer {
	if left == "" {
		left = leftDelim
	}
	if right == "" {
		right = rightDelim
	}
	l := &lexer{
		name:       name,
		input:      input,
		leftDelim:  left,
		rightDelim: right,
		state:      lexTag,
		items:      make(chan item, 2), // Two items of buffering is sufficient for all state functions
	}
	return l
}

// state functions

const (
	leftDelim    = "{{"
	rightDelim   = "}}"
	Comment      = "//"
	docTypeShort = "!!!"
	docTypeLong  = "doctype"
)

// lexTag scans html tag name.
func lexTag(l *lexer) stateFn {

	switch l.input[l.start] {
	case '.':
		l.emit(itemTag)
		l.next()
		return lexClass
	case '#':
		l.emit(itemTag)
		l.next()
		return lexId
	case ' ', '\t':
		return lexIndent
	case '|':
		l.ignore()
		return lexText
	case '\r', '\n':
		return lexNewLine
	}

	if strings.HasPrefix(l.input[l.start:], Comment) {
		return lexComment
	}

	if strings.HasPrefix(l.input[l.start:], docTypeShort) || strings.HasPrefix(l.input[l.start:], docTypeLong) {
		return lexDocType
	}

	switch r := l.peek(); {
	case r == eof:
		if l.pos > l.start {
			l.emit(itemTag)
		}
		return lexEOF
	case r == ' ':
		l.emit(itemTag)
		l.next()
		l.ignore()
		return lexText
	case r == '(':
		l.emit(itemTag)
		l.next()
		l.ignore()
		return lexAttr
	case r == '.':
		l.emit(itemTag)
		l.next()
		l.ignore()
		return lexClass
	case r == '#':
		l.emit(itemTag)
		l.next()
		l.ignore()
		return lexId
	case r == '\n' || r == '\r':
		l.emit(itemTag)
		l.next()
		return lexNewLine
	}

	l.next()

	return lexTag
}

func lexEOF(l *lexer) stateFn {
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

func lexDocType(l *lexer) stateFn {
	pre := len("doctype ")
	if strings.HasPrefix(l.input[l.pos:], "!!!") {
		pre = len("!!! ")
	}
	i := strings.Index(l.input[l.pos:], "\n")
	if i < 0 {
		return l.errorf("unclosed doctype")
	}
	l.start += pre
	l.pos += i
	l.emit(itemDoctype)
	l.next()
	return lexNewLine
}

func lexId(l *lexer) stateFn {
	switch r := l.peek(); {
	case r == '\r' || r == '\n':
		l.emit(itemId)
		l.next()
		return lexNewLine
	case r == '(':
		l.emit(itemId)
		l.next()
		return lexAttr
	case r == ' ':
		l.emit(itemId)
		l.next()
		return lexText
	case r == '.':
		l.emit(itemId)
		l.next()
		return lexClass
	case r == eof:
		l.emit(itemId)
		return lexEOF
	default:
		l.next()
	}
	return lexId
}

func lexClass(l *lexer) stateFn {
	switch r := l.peek(); {
	case r == '\r' || r == '\n':
		l.emit(itemClass)
		l.next()
		return lexNewLine
	case r == '(':
		l.emit(itemClass)
		l.next()
		return lexAttr
	case r == ' ':
		l.emit(itemClass)
		l.next()
		return lexText
	case r == eof:
		l.emit(itemClass)
		return lexEOF
	default:
		l.next()
	}
	return lexClass
}

func lexText(l *lexer) stateFn {
	if l.unexpected('\n') || l.unexpected('\r') {
		return lexNewLine
	}

	switch r := l.peek(); {
	case r == '\r' || r == '\n':
		l.emit(itemText)
		l.next()
		return lexNewLine
	case r == eof:
		l.emit(itemText)
		return lexEOF
	default:
		l.next()
	}
	return lexText
}

func lexComment(l *lexer) stateFn {
	switch r := l.peek(); {
	case r == '\r' || r == '\n':
		l.emit(itemComment)
		l.next()
		return lexNewLine
	case r == eof:
		l.emit(itemComment)
		return lexEOF
	default:
		l.next()
	}
	return lexComment
}

func lexAttr(l *lexer) stateFn {
Loop:
	for {
		switch r := l.peek(); {
		case r == ')':
			break Loop
		case r == ',':
			l.emit(itemAttr)
			l.next()
			l.ignore()
		default:
			l.next()
		}
	}
	l.emit(itemAttr)
	l.next()
	l.ignore()
	l.next()
	return lexText
}

func lexIndent(l *lexer) stateFn {
	if l.input[l.start] == '\t' {
		l.emit(itemIdentTab)
		l.next()
		return lexIndent
	}
	if l.input[l.start] == ' ' {
		l.emit(itemIdentSpace)
		l.next()
		return lexIndent
	}

	return lexTag
}

func lexNewLine(l *lexer) stateFn {
	if l.input[l.start] == '\r' || l.input[l.start] == '\n' {
		l.emit(itemEndl)
		l.next()
		return lexNewLine
	}

	return lexIndent
}

func (l *lexer) unexpected(r uint8) bool {
	if l.input[l.start] == r {
		return true
	}
	return false
}
