// The demo is its own module on purpose.
//
// It composes the whole desk — winbox windows, a websh shell, the netscrape
// browser — to show ansifilter's two halves in the renderers they were
// written for. That is a large dependency graph, and none of it belongs to a
// port whose point is being byte-exact with no dependencies. Keeping it here
// leaves the library's own go.mod, and the graph its README prints, as they
// were.
module github.com/0magnet/ansifilter-go/demo

go 1.26

require (
	github.com/0magnet/ansifilter-go v0.0.0-20260904173921-5aa9613eeacc
	github.com/0magnet/desk v0.0.0-20260904164415-376edc2698e1
	github.com/0magnet/desk/panes v0.0.0-20260904164415-376edc2698e1
	github.com/0magnet/netscrape v0.0.0-20260904175241-ee6ac0cf35ed
)

require (
	github.com/0magnet/afero v1.15.1-0.20260816202415-9f9d46a34dcd // indirect
	github.com/0magnet/sh/v3 v3.13.2-0.20260818190530-13d0024da85c // indirect
	github.com/0magnet/u-root v0.16.1-0.20260814161052-156e0b67262b // indirect
	github.com/0magnet/websh v0.0.0-20260821231944-8cefc6a09852 // indirect
	github.com/0magnet/winbox-go v0.0.0-20260903022448-a6277067b114 // indirect
	github.com/0magnet/xterm-go v0.0.0-20260821223040-7fc35994fbca // indirect
	github.com/benhoyt/goawk v1.31.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/itchyny/gojq v0.12.19 // indirect
	github.com/itchyny/timefmt-go v0.1.8 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
