// Copyright 2013 Benjamin Gentil
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"gojade/jade"
)

var debug = flag.Bool("debug", false, "Enable debug output")

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

func indent(n int) {
	for i := 0; i < n; i++ {
		print(" ")
	}
}

func displayNode(n jade.Node, ind int) {
	indent(ind)
	fmt.Printf("%s:%s\n", n.Type(), n.String())

	switch n := n.(type) {
	case *jade.TagNode:
		for _, node := range n.Nodes {
			displayNode(node, ind+2)
		}
	case *jade.ListNode:
		for _, node := range n.Nodes {
			displayNode(node, ind+2)
		}
	}
	/*default:
	  println("!! unexpected !!")
	}*/
}

func main() {
	flag.Parse()

	if *debug {
		jade.EnableDebug()
	}

	tmpl, err := jade.New("name").Parse(jadestring, "", "", make(map[string]*jade.Tree), nil)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
		return
	}

	if *debug {
		fmt.Printf("\ndisplayNode:\n")
		displayNode(tmpl.Root, 0)
	}

	print(tmpl.Root.HTMLString())
}
