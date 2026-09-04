//go:build js && wasm

package main

import (
	"strings"
	"syscall/js"

	"github.com/0magnet/netscrape"
)

// inputPane is a textarea holding the ANSI both other windows render.
//
// Escape bytes cannot be typed into a textarea, so \e and \033 are accepted
// as they are written in a shell — the same convenience the command line's
// own examples rely on.
type inputPane struct {
	area js.Value
	cb   js.Func
}

func newInputPane() *inputPane { return &inputPane{} }

func (p *inputPane) Mount(el js.Value) error {
	doc := js.Global().Get("document")
	// The window owns the box; the pane only fills it. Positioning the pane
	// itself takes it out of that box and it lands over the title bar.
	el.Get("style").Set("background", "#0b0e14")

	p.area = doc.Call("createElement", "textarea")
	p.area.Set("spellcheck", false)
	p.area.Get("style").Set("cssText",
		"display:block;width:100%;height:100%;box-sizing:border-box;"+
			"background:#0b0e14;color:#d3d7cf;border:0;outline:none;resize:none;padding:10px 12px;"+
			"font:13px/1.5 ui-monospace,SFMono-Regular,Menlo,monospace")
	p.area.Set("value", escapesOut(current))
	el.Call("appendChild", p.area)

	p.cb = js.FuncOf(func(js.Value, []js.Value) any {
		setInput(escapesIn(p.area.Get("value").String()))
		return nil
	})
	p.area.Call("addEventListener", "change", p.cb)
	return nil
}

func (p *inputPane) Close() {
	if p.cb.Truthy() {
		p.cb.Release()
	}
}

// escapesIn turns what can be typed into the byte it stands for; escapesOut
// goes back, so the textarea shows something a person can edit.
func escapesIn(s string) string {
	return strings.NewReplacer(`\e`, "\x1b", `\033`, "\x1b", `\x1b`, "\x1b").Replace(s)
}

func escapesOut(s string) string { return strings.ReplaceAll(s, "\x1b", `\e`) }

// browserPane is netscrape in a window, pointed at the export.
type browserPane struct{ el js.Value }

func newBrowserPane() *browserPane { return &browserPane{} }

func (p *browserPane) Mount(el js.Value) error {
	// netscrape fills its mount element absolutely — right for a page, wrong
	// for a window: handed the window's own body it escapes the box and lands
	// over the title bar. It gets a child to fill instead, inside a wrapper
	// that provides the containing block.
	//
	// The wrapper is needed because the body must not be touched. Making the
	// body position:relative collapses it: its only child is then absolutely
	// positioned, contributes no height, and the window measures 0 tall.
	doc := js.Global().Get("document")
	box := doc.Call("createElement", "div")
	box.Get("style").Set("cssText", "position:relative;width:100%;height:100%;overflow:hidden")
	inner := doc.Call("createElement", "div")
	box.Call("appendChild", inner)
	el.Call("appendChild", box)
	p.el = box

	// Say where a new tab starts before opening, rather than navigating after:
	// Open finishes by opening a tab itself, so anything set afterwards in the
	// same turn is replaced a line later.
	js.Global().Set("__netscrapeStart", outURL)
	netscrape.Open(inner)
	return nil
}

func (p *browserPane) Close() {
	if p.el.Truthy() {
		p.el.Set("innerHTML", "")
	}
}
