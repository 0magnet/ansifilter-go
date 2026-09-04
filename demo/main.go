//go:build js && wasm

// Command demo is the ansifilter-go demo, as a desk.
//
// ansifilter has two halves: escape codes going in, markup coming out. Showing
// either one honestly means using the renderer it was written for — a terminal
// for the ANSI, a browser for the HTML — and a page that draws its own
// approximation of both is showing its approximation, not the tool. So the
// terminal is websh on xterm-go, the browser is netscrape, and each renders
// the same input its own way, in a window of its own.
//
// The input goes into the shell's filesystem as a file, which is why the
// terminal can `cat` it: nothing is being simulated for the demo's benefit.
package main

import (
	"strings"
	"syscall/js"

	"github.com/0magnet/ansifilter-go/ansifilter"
	"github.com/0magnet/desk"
	"github.com/0magnet/desk/panes/term"
	"github.com/0magnet/netscrape"
)

// The path the input is written to, in the shell's filesystem and in the URL
// the browser asks the transport for. Same name on both sides so the two
// windows are visibly looking at one thing.
const (
	ansiPath = "/home/user/sample.ansi"
	outURL   = "http://ansifilter.local/sample.html"
)

const sample = "" +
	"\x1b[1;36mansifilter-go\x1b[0m — \x1b[32mANSI\x1b[0m escape codes into markup\n" +
	"\n" +
	"\x1b[31mred\x1b[0m \x1b[32mgreen\x1b[0m \x1b[33myellow\x1b[0m \x1b[34mblue\x1b[0m \x1b[35mmagenta\x1b[0m \x1b[36mcyan\x1b[0m\n" +
	"\x1b[1mbold\x1b[0m \x1b[3mitalic\x1b[0m \x1b[4munderline\x1b[0m \x1b[7mreverse\x1b[0m\n" +
	"\x1b[38;5;208m256-colour\x1b[0m \x1b[38;2;120;200;255mtruecolour\x1b[0m\n"

const greeting = "" +
	"\x1b[1;36mansifilter-go\x1b[0m — the ANSI half, in the terminal it was written for\r\n" +
	"\x1b[2mthe shell is \x1b[0m\x1b[1mwebsh\x1b[0m\x1b[2m; the input is a file, so it can be \x1b[0m\x1b[1mcat\x1b[0m\x1b[2med\x1b[0m\r\n\r\n" +
	"\x1b[2medit it in the input window · the browser beside this shows the exported HTML\x1b[0m\r\n\r\n"

var (
	shell   *term.Pane
	current = sample
)

// html is the export the browser is served, regenerated whenever the input
// changes. A full document rather than a fragment: it is what the browser is
// being asked to render, and ansifilter writes the <pre> and the header.
func html() string {
	g := ansifilter.New(ansifilter.HTML)
	g.SetTitle("ansifilter-go — exported HTML")
	return g.GenerateString(current)
}

// setInput takes new ANSI, puts it where both windows look, and pokes each of
// them. The terminal re-cats the file; the browser re-fetches the URL.
func setInput(s string) {
	current = s
	if fs := term.FS(); fs != nil {
		afero := fs
		if f, err := afero.Create(ansiPath); err == nil {
			f.Write([]byte(s)) //nolint:errcheck // in-memory
			f.Close()          //nolint:errcheck
		}
	}
	if shell != nil {
		if sess := shell.Session(); sess != nil {
			sess.Submit("clear; cat " + ansiPath)
		}
	}
	netscrape.Navigate(outURL)
}

func main() {
	doc := js.Global().Get("document")
	if el := doc.Call("getElementById", "desktop"); el.Truthy() {
		desk.SetRoot(el)
	}

	// The browser's transport. It is asked for one URL and answers with the
	// current export; anything else is somebody typing in the address bar,
	// and a static page cannot fetch it.
	js.Global().Set("__netscrapeFetch", js.FuncOf(func(_ js.Value, args []js.Value) any {
		url := ""
		if len(args) > 0 {
			url = args[0].String()
		}
		body := html()
		if !strings.HasPrefix(url, "http://ansifilter.local/") {
			body = "<body style='font:15px/1.6 ui-sans-serif,system-ui,sans-serif;padding:2em'>" +
				"<h1 style='font-size:1.2em'>Nothing to fetch</h1>" +
				"<p>This browser is wired to one transport, which serves the exported HTML " +
				"at <code>" + outURL + "</code>. A page on static hosting cannot reach anywhere else.</p></body>"
		}
		// A promise, not a Response: the transport is awaited with .then, and
		// handing back a bare Response panics the runtime on the first fetch.
		resp := js.Global().Get("Response").New(body, map[string]any{
			"status":  200,
			"headers": map[string]any{"content-type": "text/html"},
		})
		return js.Global().Get("Promise").Call("resolve", resp)
	}))

	if fs := term.FS(); fs != nil {
		if f, err := fs.Create(ansiPath); err == nil {
			f.Write([]byte(sample)) //nolint:errcheck // in-memory
			f.Close()               //nolint:errcheck
		}
	}

	desk.Register(desk.App{
		Name: "input", Title: "input — ANSI", Help: "the escape codes both windows render",
		Width: 620, Height: 300,
		Open: func([]string) (desk.Pane, error) { return newInputPane(), nil },
	})
	desk.Register(desk.App{
		Name: "term", Title: "terminal — the ANSI", Help: "a websh shell; cat the input",
		Width: 620, Height: 420,
		Open: func([]string) (desk.Pane, error) {
			shell = term.New(greeting, "ansifilter").Run("cat " + ansiPath)
			return shell, nil
		},
	})
	desk.Register(desk.App{
		Name: "browser", Title: "browser — the exported HTML", Help: "netscrape, rendering the export",
		Width: 620, Height: 500,
		Open: func([]string) (desk.Pane, error) { return newBrowserPane(), nil },
	})

	// Tiled rather than cascaded: the point is seeing the three at once.
	w, h := deskSize()
	half := w / 2
	desk.LaunchOpts("input", desk.Options{X: 8, Y: 8, Width: half - 16, Height: h/2 - 16})    //nolint:errcheck
	desk.LaunchOpts("term", desk.Options{X: 8, Y: h / 2, Width: half - 16, Height: h/2 - 16}) //nolint:errcheck
	desk.LaunchOpts("browser", desk.Options{X: half, Y: 8, Width: half - 16, Height: h - 24}) //nolint:errcheck

	select {}
}

func deskSize() (float64, float64) {
	el := js.Global().Get("document").Call("getElementById", "desktop")
	if !el.Truthy() {
		return 1200, 800
	}
	w := el.Get("clientWidth").Float()
	h := el.Get("clientHeight").Float()
	if w < 600 {
		w = 1200
	}
	if h < 400 {
		h = 800
	}
	return w, h
}
