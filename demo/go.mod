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
	github.com/0magnet/ansifilter-go v0.0.0
	github.com/0magnet/desk v0.0.1-0.20261004200012-7ac6a12489c1
	github.com/0magnet/desk/panes v0.0.1-0.20261004200012-7ac6a12489c1
	github.com/0magnet/netscrape v0.0.1-0.20261004014115-1bb4ba7bb673
)

require (
	github.com/0magnet/afero v1.15.1-0.20261003211811-482680d00992 // indirect
	github.com/0magnet/sh/v3 v3.13.2-0.20261004194540-aa2d6e4a31a5 // indirect
	github.com/0magnet/u-root v0.16.1-0.20261003214924-44e47b732754 // indirect
	github.com/0magnet/websh v0.0.1-0.20261004200853-adb76a8fa4c2 // indirect
	github.com/0magnet/winbox-go v0.0.0 // indirect
	github.com/0magnet/xterm-go v0.0.1-0.20261004020305-36b45f096b30 // indirect
	github.com/benhoyt/goawk v1.32.0 // indirect
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/itchyny/gojq v0.12.19 // indirect
	github.com/itchyny/timefmt-go v0.1.9 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)
