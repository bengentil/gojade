// Copyright 2011 The Go Authors.
// 2013 Benjamin Gentil
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jade

import (
	//	"reflect"
	"fmt"
	"testing"
)

var jadestring = `// test.jade
doctype 5
html(lang="en")
  head
    title pageTitle
    script(type='text/javascript')
      | if (foo) {
      |    bar()
      | }
  body
    h1 Jade - node template engine
    #container.class
      !--
      if youAreUsingJade
        p You are amazing
      else
          p Get on it!
    .test
    h3.r1
    h6#r4
    h1.dd#fd`

func TestJade(t *testing.T) {
	l := lex("name", jadestring, "", "")
	for {
		item := l.nextItem()
		fmt.Printf("%s: %s\n", item.typ, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
}
