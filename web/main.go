//go:build js && wasm

// Command web is the ansifilter-go demo: ANSI in, every output format out.
//
// The point of the tool is that the same escape sequences become HTML, or
// SVG, or LaTeX, or BBCode, byte for byte as the C original produces them.
// A page can show that directly: one input, a format picker, the generated
// markup, and — for HTML and SVG — the markup actually rendered, so you can
// see that it says what it looks like.
package main

import (
	"strings"
	"syscall/js"

	"github.com/0magnet/ansifilter-go/ansifilter"
)

type format struct {
	name     string
	typ      ansifilter.OutputType
	fragment bool // ask for a fragment, so the preview is not a whole document
	preview  bool // can be shown rendered as well as as source
}

// The order is the one the README lists, so the page and the docs agree.
var formats = []format{
	{"HTML", ansifilter.HTML, true, true},
	{"SVG", ansifilter.SVG, false, true},
	{"Text", ansifilter.TEXT, false, false},
	{"Pango", ansifilter.PANGO, true, false},
	{"LaTeX", ansifilter.LATEX, true, false},
	{"TeX", ansifilter.TEX, true, false},
	{"RTF", ansifilter.RTF, true, false},
	{"BBCode", ansifilter.BBCODE, true, false},
}

func main() {
	doc := js.Global().Get("document")
	in := doc.Call("getElementById", "in")
	sel := doc.Call("getElementById", "fmt")
	src := doc.Call("getElementById", "src")
	term := doc.Call("getElementById", "term")
	prev := doc.Call("getElementById", "prev")
	prevWrap := doc.Call("getElementById", "prevwrap")
	prevLabel := doc.Call("getElementById", "prevlabel")

	for i, f := range formats {
		o := doc.Call("createElement", "option")
		o.Set("value", i)
		o.Set("textContent", f.name)
		sel.Call("appendChild", o)
	}

	render := func() {
		// The escape sequences are typed as text, so \x1b arrives as a
		// backslash and an e rather than as one byte. Accept both.
		input := in.Get("value").String()
		input = strings.NewReplacer(`\e`, "\x1b", `\033`, "\x1b", `\x1b`, "\x1b").Replace(input)

		// What a terminal would show, whatever format is selected. This is
		// the same HTML generator, on a dark ground: ansifilter emits colour
		// only where the text sets one, so default text inherits the pane's
		// and looks the way it would on a terminal.
		tg := ansifilter.New(ansifilter.HTML)
		tg.SetFragmentCode(true)
		term.Set("innerHTML", tg.GenerateString(input))

		f := formats[sel.Get("selectedIndex").Int()]
		g := ansifilter.New(f.typ)
		if f.fragment {
			g.SetFragmentCode(true)
		}
		out := g.GenerateString(input)
		src.Set("textContent", out)

		if f.preview {
			prevWrap.Get("style").Set("display", "")
			prevLabel.Set("textContent", "as the exported "+f.name+" renders")
			prev.Set("innerHTML", out)
		} else {
			prevWrap.Get("style").Set("display", "none")
		}
	}

	cb := js.FuncOf(func(js.Value, []js.Value) any { render(); return nil })
	in.Call("addEventListener", "input", cb)
	sel.Call("addEventListener", "change", cb)
	render()

	select {}
}
