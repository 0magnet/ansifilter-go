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
	github.com/0magnet/ansifilter-go v0.0.1-0.20261106131638-92935b5b9b4e
	github.com/0magnet/desk v0.0.3-0.20261107115715-e5ac49895780
	github.com/0magnet/desk/panes v0.0.1-0.20261107115330-c83823566fa2
	github.com/0magnet/netscrape v0.0.1-0.20261106130345-a8165604cf34
)

require (
	github.com/0magnet/afero v1.15.1-0.20261003211811-482680d00992 // indirect
	github.com/0magnet/bottle v0.0.1-0.20261028202309-2b922f28f74e // indirect
	github.com/0magnet/sh/v3 v3.13.2-0.20261024150911-145e01c63128 // indirect
	github.com/0magnet/u-root v0.16.1-0.20261003214924-44e47b732754 // indirect
	github.com/0magnet/websh v0.0.1-0.20261106131606-1973f8bee280 // indirect
	github.com/0magnet/winbox-go v0.0.1-0.20261106130421-357df562ff6c // indirect
	github.com/0magnet/xterm-go v0.0.1-0.20261102175652-6023ed24df89 // indirect
	github.com/benhoyt/goawk v1.32.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/itchyny/gojq v0.12.20 // indirect
	github.com/itchyny/timefmt-go v0.1.9 // indirect
	github.com/mattn/go-runewidth v0.0.31 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)
