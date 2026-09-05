// Command ansifilter converts text containing ANSI terminal escape codes into
// text, HTML, Pango markup, LaTeX, plain TeX, RTF, BBCode or SVG.
//
// It is a Go port of ansifilter 2.23 by André Simon and accepts the same
// options.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	af "github.com/0magnet/ansifilter-go/ansifilter"
)

const (
	appEmail   = "a.simon@mailbox.org"
	appWebsite = "http://andre-simon.de/" //nolint:gosec // a URL, not a credential
)

// appVersion tracks the library version so -ldflags pins both at once.
var appVersion = af.Version

// argKind describes whether an option takes an argument.
type argKind int

const (
	argNo argKind = iota
	argYes
	argMaybe
)

// optSpec describes one command line option.
type optSpec struct {
	code byte
	name string
	kind argKind
}

// optionTable mirrors ansifilter's Arg_parser option table.
var optionTable = []optSpec{
	{'a', "anchors", argMaybe},
	{'b', "append", argNo},
	{'n', "tee", argNo},
	{'d', "doc-title", argYes},
	{'e', "encoding", argYes},
	{'f', "fragment", argNo},
	{'F', "font", argYes},
	{'h', "help", argNo},
	{'H', "html", argNo},
	{'M', "pango", argNo},
	{'i', "input", argYes},
	{'l', "line-numbers", argNo},
	{'L', "latex", argNo},
	{'P', "tex", argNo},
	{'B', "bbcode", argNo},
	{'o', "output", argYes},
	{'O', "outdir", argYes},
	{'p', "plain", argNo},
	{'r', "style-ref", argYes},
	{'R', "rtf", argNo},
	{'s', "font-size", argYes},
	{'t', "tail", argNo},
	{'T', "text", argNo},
	{'w', "wrap", argYes},
	{'v', "version", argNo},
	{'V', "version", argNo},
	{'W', "wrap-no-numbers", argNo},
	{'X', "art-cp437", argNo},
	{'U', "art-bin", argNo},
	{'D', "art-tundra", argNo},
	{'Y', "art-width", argYes},
	{'Z', "art-height", argYes},
	{'m', "map", argYes},
	{'N', "no-trailing-nl", argNo},
	{'C', "no-version-info", argNo},
	{'k', "ignore-clear", argMaybe},
	{'c', "ignore-csi", argNo},
	{'y', "derived-styles", argNo},
	{'S', "svg", argNo},
	{'Q', "width", argYes},
	{'E', "height", argYes},
	{'x', "max-size", argYes},
	{'g', "no-default-fg", argNo},
	{'A', "line-append", argYes},
}

// parsedArg is one parsed command line entry. A zero code marks an operand.
type parsedArg struct {
	code byte
	arg  string
}

// parseArgs parses argv the way Arg_parser does: short options may be
// clustered, long options accept "=value" and unique abbreviations, and
// operands are recorded with a zero code in the order they appear.
func parseArgs(argv []string) ([]parsedArg, error) {
	var out []parsedArg
	i := 0
	for i < len(argv) {
		a := argv[i]
		switch {
		case a == "--":
			for i++; i < len(argv); i++ {
				out = append(out, parsedArg{0, argv[i]})
			}
			return out, nil

		case strings.HasPrefix(a, "--"):
			body := a[2:]
			name, val := body, ""
			hasVal := false
			if eq := strings.IndexByte(body, '='); eq >= 0 {
				name, val, hasVal = body[:eq], body[eq+1:], true
			}
			spec, err := lookupLong(name)
			if err != nil {
				return nil, err
			}
			switch spec.kind {
			case argNo:
				if hasVal {
					return nil, fmt.Errorf("option '--%s' doesn't allow an argument", name)
				}
				out = append(out, parsedArg{spec.code, ""})
			case argYes:
				if !hasVal {
					i++
					if i >= len(argv) {
						return nil, fmt.Errorf("option '--%s' requires an argument", name)
					}
					val = argv[i]
				}
				out = append(out, parsedArg{spec.code, val})
			case argMaybe:
				out = append(out, parsedArg{spec.code, val})
			}
			i++

		case len(a) > 1 && a[0] == '-':
			j := 1
			for j < len(a) {
				c := a[j]
				spec := lookupShort(c)
				if spec == nil {
					return nil, fmt.Errorf("invalid option '-%c'", c)
				}
				switch spec.kind {
				case argNo:
					out = append(out, parsedArg{spec.code, ""})
					j++
				case argYes:
					if j+1 < len(a) {
						out = append(out, parsedArg{spec.code, a[j+1:]})
					} else {
						i++
						if i >= len(argv) {
							return nil, fmt.Errorf("option '-%c' requires an argument", c)
						}
						out = append(out, parsedArg{spec.code, argv[i]})
					}
					j = len(a)
				case argMaybe:
					val := ""
					if j+1 < len(a) {
						val = a[j+1:]
					}
					out = append(out, parsedArg{spec.code, val})
					j = len(a)
				}
			}
			i++

		default:
			out = append(out, parsedArg{0, a})
			i++
		}
	}
	return out, nil
}

// lookupLong resolves a long option name, allowing unique abbreviations.
func lookupLong(name string) (*optSpec, error) {
	var match *optSpec
	for idx := range optionTable {
		if optionTable[idx].name == name {
			return &optionTable[idx], nil
		}
	}
	count := 0
	for idx := range optionTable {
		if strings.HasPrefix(optionTable[idx].name, name) {
			// The two "version" entries are the same option under two codes.
			if match != nil && match.name == optionTable[idx].name {
				continue
			}
			match = &optionTable[idx]
			count++
		}
	}
	if count == 1 {
		return match, nil
	}
	if count > 1 {
		return nil, fmt.Errorf("option '--%s' is ambiguous", name)
	}
	return nil, fmt.Errorf("unrecognized option '--%s'", name)
}

// lookupShort resolves a short option code.
func lookupShort(c byte) *optSpec {
	for idx := range optionTable {
		if optionTable[idx].code == c {
			return &optionTable[idx]
		}
	}
	return nil
}

// options holds the resolved command line configuration.
type options struct {
	outputType af.OutputType

	help, version                bool
	fragment, plain, ignoreEOF   bool
	linenum, wrapNoNum           bool
	anchors, funnyAnchors        bool
	cp437, asciiBin, asciiTundra bool
	teeMode, appendMode          bool
	omitTrailingCR               bool
	omitVersionInfo              bool
	ignoreClear                  bool
	ignoreCSI                    bool
	applyDynStyles               bool
	omitDefaultFgColor           bool

	encodingName   string
	font, fontSize string
	docTitle       string
	styleSheetPath string
	colorMapPath   string
	outFilename    string
	outDirectory   string
	lineAppendage  string
	width, height  string

	wrapLineLen    int
	asciiArtWidth  int
	asciiArtHeight int
	maxFileSize    int64

	inputFileNames []string
}

// newOptions returns the defaults used by ansifilter's CmdLineOptions.
func newOptions() *options {
	return &options{
		outputType:     af.TEXT,
		ignoreClear:    true,
		encodingName:   "ISO-8859-1",
		font:           "Courier New",
		fontSize:       "10pt",
		asciiArtWidth:  80,
		asciiArtHeight: 100,
		maxFileSize:    268435456,
	}
}

// parseRuntimeOptions applies parsed arguments. Parsing stops at the first
// operand; the remainder become input file names when -i was not used.
func (o *options) parseRuntimeOptions(argv []string, readInputFilenames bool) error {
	args, err := parseArgs(argv)
	if err != nil {
		return err
	}

	argind := 0
	for ; argind < len(args); argind++ {
		code := args[argind].code
		arg := args[argind].arg
		if code == 0 {
			break
		}
		switch code {
		case 'a':
			o.anchors = true
			if arg == "self" {
				o.funnyAnchors = true
			}
		case 'b':
			o.appendMode = true
		case 'n':
			o.teeMode = true
		case 'B':
			o.outputType = af.BBCODE
		case 'd':
			o.docTitle = arg
		case 'e':
			o.encodingName = arg
		case 'f':
			o.fragment = true
		case 'F':
			o.font = arg
		case 'h':
			o.help = true
		case 'H':
			o.outputType = af.HTML
		case 'i':
			o.inputFileNames = append(o.inputFileNames, arg)
		case 'l':
			o.linenum = true
		case 'L':
			o.outputType = af.LATEX
		case 'm':
			o.colorMapPath = arg
		case 'M':
			o.outputType = af.PANGO
		case 'P':
			o.outputType = af.TEX
		case 'o':
			o.outFilename = arg
		case 'p':
			o.plain = true
		case 'r':
			o.styleSheetPath = arg
		case 'R':
			o.outputType = af.RTF
		case 'S':
			o.outputType = af.SVG
		case 's':
			o.fontSize = arg
		case 't':
			o.ignoreEOF = true
		case 'T':
			o.outputType = af.TEXT
		case 'v', 'V':
			o.version = true
		case 'O':
			o.outDirectory = validateDirPath(arg)
		case 'w':
			o.wrapLineLen = atoiPrefix(arg)
		case 'W':
			o.wrapNoNum = true
		case 'X':
			o.cp437 = true
		case 'U':
			o.asciiBin = true
		case 'D':
			o.asciiTundra = true
		case 'Y':
			o.asciiArtWidth = atoiPrefix(arg)
		case 'Z':
			o.asciiArtHeight = atoiPrefix(arg)
		case 'N':
			o.omitTrailingCR = true
		case 'C':
			o.omitVersionInfo = true
		case 'k':
			if arg != "" {
				o.ignoreClear = arg == "true" || arg == "1"
			}
		case 'c':
			o.ignoreCSI = true
		case 'y':
			o.applyDynStyles = true
		case 'g':
			o.omitDefaultFgColor = true
		case 'Q':
			o.width = arg
		case 'E':
			o.height = arg
		case 'A':
			o.lineAppendage = arg
		case 'x':
			o.maxFileSize = int64(atoiPrefix(arg))
			if arg != "" {
				// The C++ switch falls through, so G scales by 1024^3.
				switch arg[len(arg)-1] {
				case 'G':
					o.maxFileSize *= 1024 * 1024 * 1024
				case 'M':
					o.maxFileSize *= 1024 * 1024
				case 'K':
					o.maxFileSize *= 1024
				}
			}
		default:
			fmt.Fprintln(os.Stderr, "ansifilter: option parsing failed")
		}
	}

	if readInputFilenames {
		if argind < len(args) {
			if len(o.inputFileNames) == 0 {
				for ; argind < len(args); argind++ {
					o.inputFileNames = append(o.inputFileNames, args[argind].arg)
				}
			}
		} else if len(o.inputFileNames) == 0 {
			o.inputFileNames = append(o.inputFileNames, "")
		}
	}
	return nil
}

// atoiPrefix parses the leading integer of s, yielding 0 when there is none.
func atoiPrefix(s string) int {
	s = strings.TrimSpace(s)
	i := 0
	if i < len(s) && (s[i] == '-' || s[i] == '+') {
		i++
	}
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == start {
		return 0
	}
	v, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0
	}
	return v
}

// validateDirPath ensures the directory path ends with a separator.
func validateDirPath(path string) string {
	if path == "" || path[len(path)-1] != '/' {
		return path + "/"
	}
	return path
}

// outFileSuffix returns the extension used when writing multiple files.
func (o *options) outFileSuffix() string {
	switch o.outputType {
	case af.HTML:
		return ".html"
	case af.PANGO:
		return ".pango"
	case af.XHTML:
		return ".xhtml"
	case af.RTF:
		return ".rtf"
	case af.TEX, af.LATEX:
		return ".tex"
	case af.BBCODE:
		return ".bbcode"
	case af.SVG:
		return ".svg"
	default:
		return ".txt"
	}
}

// singleOutFilename resolves the output path for single file conversions.
func (o *options) singleOutFilename() string {
	if len(o.inputFileNames) != 0 && o.outDirectory != "" {
		if o.outFilename == "" {
			o.outFilename = o.outDirectory
			in := o.inputFileNames[0]
			delim := strings.LastIndexByte(in, '/') + 1
			o.outFilename += in[delim:] + o.outFileSuffix()
		}
	}
	return o.outFilename
}

// getOutDirectory derives the output directory from -o when no -O was given.
func (o *options) getOutDirectory() string {
	if o.outFilename != "" && len(o.inputFileNames) == 0 {
		o.outDirectory = dirName(o.outFilename)
	}
	return o.outDirectory
}

// dirName returns the directory portion of path, including the separator.
func dirName(path string) string {
	i := strings.LastIndexByte(path, '/')
	if i < 0 {
		return ""
	}
	return path[:i+1]
}

func printVersionInfo() {
	fmt.Print("\n ansifilter version " + appVersion +
		"\n Copyright (C) 2007-2026 Andre Simon <a dot simon at mailbox.org>" +
		"\n\n Argparser class" +
		"\n Copyright (C) 2006-2008 Antonio Diaz Diaz <ant_diaz at teleline.es>" +
		"\n\n This software is released under the terms of the GNU General " +
		"Public License." +
		"\n For more information about these matters, see the file named " +
		"COPYING.\n")
}

func printHelp() {
	fmt.Print(`Invocation: ansifilter [OPTION]... [FILE]...

ansifilter handles text files containing ANSI terminal escape codes.

File handling:
  -i, --input=<file>     Name of input file (default stdin)
  -o, --output=<file>    Name of output file (default stdout)
  -O, --outdir=<dir>     Name of output directory
  -n, --tee              Copy stdin to stdout and write filtered output to file
  -b, --append           Append to output file (use with --tee)
  -t, --tail             Continue reading after end-of-file (like tail -f)
  -x, --max-size=<size>  Set maximum input file size
                         (examples: 512M, 1G; default: 256M)

Output text formats:
  -T, --text (default)   Output text
  -H, --html             Output HTML
  -M, --pango            Output Pango Markup
  -L, --latex            Output LaTeX
  -P, --tex              Output Plain TeX
  -R, --rtf              Output RTF
  -S, --svg              Output SVG
  -B, --bbcode           Output BBCode

Format options:
  -a, --anchors(=self)   Add HTML line anchors (opt: self referencing, assumes -l)
  -d, --doc-title        Set HTML/LaTeX/SVG document title
  -e, --encoding=<enc>   Set HTML/RTF encoding (must match input file encoding);
                         omit encoding information if enc=NONE
  -f, --fragment         Omit HTML header and footer
  -F, --font=<font>      Set HTML/RTF/SVG font face
  -k, --ignore-clear(=0) Do not adhere to clear (ESC K) commands (default: true)
  -c, --ignore-csi       Do not adhere to CSI commands (useful for UTF-8 input)
  -A, --line-append=<s>  Output string after each output line
  -l, --line-numbers     Print line numbers in output file
  -m, --map=<path>       Read color mapping file (see README)
  -r, --style-ref=<rf>   Set HTML/TeX/LaTeX/SVG stylesheet path
  -s, --font-size=<fs>   Set HTML/RTF/SVG font size
  -p, --plain            Ignore ANSI formatting information
  -w, --wrap=<len>       Wrap long lines
  -g, --no-default-fg    Omit default foreground color
      --no-trailing-nl   Omit trailing newline
      --no-version-info  Omit version info comment
      --wrap-no-numbers  Omit line numbers of wrapped lines (assumes -l)
      --derived-styles   Output dynamic stylesheets (HTML/SVG)

ANSI art options:
      --art-cp437        Parse codepage 437 ANSI art (HTML and RTF output)
      --art-bin          Parse BIN/XBIN ANSI art (HTML output, no stdin)
      --art-tundra       Parse Tundra ANSI art (HTML output, no stdin)
      --art-width        Set ANSI art width (default 80)
      --art-height       Set ANSI art height (default 150)

SVG output options:
      --height           set image height (units allowed)
      --width            set image width (see --height)

Other options:
  -h, --help             Print help
  -v, --version          Print version and license info

Examples:
ansifilter -i input.ansi -o output.txt
ansifilter *.txt
tail -f server.log | ansifilter

Parsing XBIN files overrides --art-width, --art-height and --map options.
The ANSI art file formats BIN, XBIN and TND cannot be read from stdin.

Please report bugs to ` + appEmail + `
For updates see ` + appWebsite + "\n")
}

// setGeneratorProperties transfers the command line options to the generator.
func setGeneratorProperties(g *af.Generator, o *options, title string) {
	g.SetTitle(title)
	g.SetEncoding(o.encodingName)
	g.SetFragmentCode(o.fragment)
	g.SetPlainOutput(o.plain)
	g.SetReadAfterEOF(o.ignoreEOF)
	g.SetFont(o.font)
	g.SetFontSize(o.fontSize)
	g.SetStyleSheet(o.styleSheetPath)
	g.SetPreformatting(o.wrapLineLen)
	g.SetShowLineNumbers(o.linenum)
	g.SetWrapNoNumbers(!o.wrapNoNum)
	g.SetAddAnchors(o.anchors)
	g.SetAddFunnyAnchors(o.funnyAnchors)
	g.SetParseCodePage437(o.cp437)
	g.SetParseAsciiBin(o.asciiBin)
	g.SetParseAsciiTundra(o.asciiTundra)
	g.SetIgnoreClearSeq(o.ignoreClear)
	g.SetIgnoreCSISeq(o.ignoreCSI)
	g.SetApplyDynStyles(o.applyDynStyles)
	g.SetAsciiArtSize(o.asciiArtWidth, o.asciiArtHeight)
	g.SetOmitTrailingNewline(o.omitTrailingCR)
	g.SetOmitVersionInfo(o.omitVersionInfo)
	g.SetSVGSize(o.width, o.height)
}

func run(argv []string) int {
	o := newOptions()

	// ANSIFILTER_OPTIONS is applied before the command line.
	if env := os.Getenv("ANSIFILTER_OPTIONS"); env != "" {
		if err := o.parseRuntimeOptions(strings.Fields(env), false); err != nil {
			fmt.Fprintf(os.Stderr, "ansifilter: %v\n", err)
			fmt.Fprintln(os.Stderr, "Try 'ansifilter --help' for more information.")
			return 1
		}
	}
	if err := o.parseRuntimeOptions(argv, true); err != nil {
		fmt.Fprintf(os.Stderr, "ansifilter: %v\n", err)
		fmt.Fprintln(os.Stderr, "Try 'ansifilter --help' for more information.")
		return 1
	}

	if o.version {
		printVersionInfo()
		return 0
	}
	if o.help {
		printHelp()
		return 0
	}

	generator := af.New(o.outputType)
	if generator == nil {
		fmt.Fprintln(os.Stderr, "ansifilter: unsupported output type")
		return 1
	}

	if o.teeMode {
		return runTee(generator, o)
	}

	inFileList := o.inputFileNames
	outDirectory := o.getOutDirectory()

	if !o.omitDefaultFgColor {
		generator.SetDefaultForegroundColor()
	}
	if err := generator.SetColorMap(o.colorMapPath); err != nil {
		fmt.Fprintf(os.Stderr, "could not read map file: %s\n", sanitize(o.colorMapPath))
		return 1
	}

	failure := false
	for i := 0; i < len(inFileList) && !failure; i++ {
		var outFilePath string
		if len(inFileList) > 1 {
			inFileName := filepath.Base(inFileList[i])
			outFilePath = outDirectory + inFileName + o.outFileSuffix()
		} else {
			outFilePath = o.singleOutFilename()
		}

		if inFileList[i] != "" {
			if st, err := os.Stat(inFileList[i]); err == nil && st.Size() > o.maxFileSize {
				fmt.Fprintf(os.Stderr, "file exceeds max size (see --max-size): %s\n",
					sanitize(inFileList[i]))
				return 1
			}
		}

		title := o.docTitle
		if title == "" {
			title = inFileList[i]
		}
		setGeneratorProperties(generator, o, title)
		generator.SetLineAppendage(o.lineAppendage)

		err := generator.GenerateFile(inFileList[i], outFilePath)
		switch {
		case errors.Is(err, af.ErrBadInput):
			fmt.Fprintf(os.Stderr, "could not read input: %s\n", sanitize(inFileList[i]))
			failure = true
		case errors.Is(err, af.ErrBadOutput):
			fmt.Fprintf(os.Stderr, "could not write output: %s\n", sanitize(outFilePath))
			failure = true
		}
	}

	if o.applyDynStyles && !failure {
		// Reported like the other output failures above: without this a
		// stylesheet that could not be written vanished silently, and the
		// documents referencing it came out unstyled with a zero exit.
		if err := generator.PrintDynamicStyleFile(outDirectory + "derived_styles.css"); err != nil {
			fmt.Fprintf(os.Stderr, "could not write output: %s\n", sanitize(outDirectory+"derived_styles.css"))
			failure = true
		}
	}

	if failure {
		return 1
	}
	return 0
}

// runTee copies stdin to stdout while writing filtered output to a file.
func runTee(generator *af.Generator, o *options) int {
	if o.outputType != af.TEXT {
		fmt.Fprintln(os.Stderr, "ansifilter: --tee mode allows only text output")
		return 1
	}
	outFilePath := o.singleOutFilename()
	if outFilePath == "" {
		fmt.Fprintln(os.Stderr, "ansifilter: --tee mode requires an output file specified with -o")
		return 1
	}
	if !o.omitDefaultFgColor {
		generator.SetDefaultForegroundColor()
	}
	if err := generator.SetColorMap(o.colorMapPath); err != nil {
		fmt.Fprintf(os.Stderr, "could not read map file: %s\n", sanitize(o.colorMapPath))
		return 1
	}
	setGeneratorProperties(generator, o, o.docTitle)

	flags := os.O_CREATE | os.O_WRONLY
	if o.appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	outFile, err := os.OpenFile(outFilePath, flags, 0o644) //nolint:gosec // a document meant to be readable
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not write output: %s\n", sanitize(outFilePath))
		return 1
	}
	defer outFile.Close() //nolint:errcheck // the write errors below are what carry a failure out

	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for in.Scan() {
		buffer := in.Text()
		fmt.Println(buffer)
		if _, err := outFile.WriteString(generator.GenerateString(buffer)); err != nil {
			// Without this a full disk truncated the output silently and still
			// exited 0, so the copy looked complete.
			fmt.Fprintf(os.Stderr, "could not write output: %s\n", sanitize(outFilePath))
			return 1
		}
	}
	return 0
}

// sanitize strips control characters from a path before echoing it, so a
// crafted file name cannot inject escape sequences into the terminal.
func sanitize(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] == 0x7f {
			b.WriteByte('?')
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func main() {
	os.Exit(run(os.Args[1:]))
}
