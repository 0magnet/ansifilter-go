// The demo is its own module on purpose.
//
// It composes the whole desk — winbox windows, a websh shell, the netscrape
// browser — to show ansifilter's two halves in the renderers they were
// written for. That is a large dependency graph, and none of it belongs to a
// port whose point is being byte-exact with no dependencies. Keeping it here
// leaves the library's own go.mod, and the graph its README prints, as they
// were.
module github.com/0magnet/ansifilter-go/demo

go 1.26.6

require (
	github.com/0magnet/ansifilter-go v0.0.1-0.20261008155320-d133b6e470ff
	github.com/0magnet/desk v0.0.3-0.20261012010524-59791fc58bd4
	github.com/0magnet/desk/panes v0.0.1-0.20261012010524-59791fc58bd4
	github.com/0magnet/netscrape v0.0.1-0.20261012010338-59d0d3bbd354
)

require (
	github.com/0magnet/afero v1.15.1-0.20261003211811-482680d00992 // indirect
	github.com/0magnet/bottle v0.0.1-0.20261010202827-a5ff39f4742b // indirect
	github.com/0magnet/sh/v3 v3.13.2-0.20261004194540-aa2d6e4a31a5 // indirect
	github.com/0magnet/u-root v0.16.1-0.20261003214924-44e47b732754 // indirect
	github.com/0magnet/websh v0.0.1-0.20261012010421-aa9f8595d2d0 // indirect
	github.com/0magnet/winbox-go v0.0.1-0.20261012010320-88f1679695ed // indirect
	github.com/0magnet/xterm-go v0.0.1-0.20261011000522-9adb4a350db6 // indirect
	github.com/benhoyt/goawk v1.32.0 // indirect
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/itchyny/gojq v0.12.19 // indirect
	github.com/itchyny/timefmt-go v0.1.9 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)
