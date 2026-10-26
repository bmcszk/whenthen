package whenthen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	// given the fixtures in testdata/src/a, each carrying its own want comment
	fixtures := analysistest.TestData()

	// when the analyzer runs over them
	results := analysistest.Run(t, fixtures, Analyzer, "a")

	// then the package was checked and every diagnostic matched its want
	if len(results) != 1 {
		t.Fatalf("checked %d packages, want 1", len(results))
	}
}

func TestPhases(t *testing.T) {
	t.Parallel()

	// given the declared phase table
	names := make([]string, 0, len(Phases))
	for _, phase := range Phases {
		names = append(names, phase.Name)
	}

	// when the canonical alias of every group is read
	canonical := make([]string, 0, len(Phases))
	for _, phase := range Phases {
		canonical = append(canonical, phase.Aliases[0])
	}

	// then the order, canonical aliases and optionality match the coding rule
	want := []string{"given", "when", "then"}
	if !slices.Equal(names, want) {
		t.Fatalf("phase names = %v, want %v", names, want)
	}
	if !slices.Equal(canonical, want) {
		t.Errorf("canonical aliases = %v, want %v", canonical, want)
	}
	if !Phases[0].Optional {
		t.Error("given must be optional")
	}
	for i, phase := range Phases[1:] {
		if phase.Optional {
			t.Errorf("%s must be mandatory", names[i+1])
		}
	}
}

func TestIsTest(t *testing.T) {
	t.Parallel()

	// given declarations that look like tests to varying degrees
	const filePrefix = "package p\n\ntype R struct{}\n\n"
	cases := []struct {
		name string
		decl string
		want bool
	}{
		{name: "empty body", decl: "func TestX(t *testing.T) {}", want: true},
		{name: "body statements", decl: "func TestX(t *testing.T) { _ = t }", want: true},
		{name: "value receiver", decl: "func (R) TestX(t *testing.T) {}", want: false},
		{name: "pointer receiver", decl: "func (r *R) TestX(t *testing.T) {}", want: false},
		{name: "no body", decl: "func TestX(t *testing.T)", want: false},
		{name: "helper name", decl: "func helper(t *testing.T) {}", want: false},
		{name: "no param", decl: "func TestNoParam() {}", want: false},
		{name: "benchmark param", decl: "func TestX(t *testing.B) {}", want: false},
		{name: "two params", decl: "func TestX(t, u *testing.T) {}", want: false},
		{name: "value testing.T", decl: "func TestX(t testing.T) {}", want: false},
		{name: "foreign T", decl: "func TestX(t *foo.T) {}", want: false},
	}

	// when each declaration is parsed and inspected
	got := make([]bool, len(cases))
	for i, tc := range cases {
		got[i] = parseIsTest(t, filePrefix+tc.decl)
	}

	// then only func TestX(t *testing.T) with a body counts as a test
	for i, tc := range cases {
		if got[i] != tc.want {
			t.Errorf("%s: isTest(%s) = %v, want %v", tc.name, tc.decl, got[i], tc.want)
		}
	}
}

func parseIsTest(t *testing.T, src string) bool {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "x.go", src, 0)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			return isTest(fn)
		}
	}
	t.Fatalf("no func decl in %q", src)
	return false
}

func TestHasWord(t *testing.T) {
	t.Parallel()

	// given comment texts and the words to look for
	cases := []struct {
		text string
		word string
		want bool
	}{
		{text: "// when it decodes", word: "when", want: true},
		{text: "// WHEN it decodes", word: "when", want: true},
		{text: "// given a blob, then it decodes", word: "then", want: true},
		{text: "whenever", word: "when", want: false},
		{text: "assertion", word: "assert", want: false},
		{text: "when2", word: "when", want: false},
		{text: "_when", word: "when", want: false},
	}

	// when hasWord compares each text with its word
	got := make([]bool, len(cases))
	for i, tc := range cases {
		got[i] = hasWord(tc.text, tc.word)
	}

	// then only whole words match, case-insensitively
	for i, tc := range cases {
		if got[i] != tc.want {
			t.Errorf("hasWord(%q, %q) = %v, want %v", tc.text, tc.word, got[i], tc.want)
		}
	}
}

func TestPhaseLine(t *testing.T) {
	t.Parallel()

	// given markers collected from a test body
	markers := []marker{
		{line: 3, text: "// given a blob"},
		{line: 5, text: "// when decoded"},
		{line: 7, text: "// then the length matches"},
	}
	cases := []struct {
		name    string
		aliases []string
		want    int
		wantOK  bool
	}{
		{name: "canonical alias", aliases: []string{"given"}, want: 3, wantOK: true},
		{name: "later phase", aliases: []string{"then"}, want: 7, wantOK: true},
		{name: "alias group", aliases: []string{"arrange", "when"}, want: 5, wantOK: true},
		{name: "absent", aliases: []string{"setup"}, want: 0, wantOK: false},
	}

	// when phaseLine looks each alias group up
	type result struct {
		line int
		ok   bool
	}
	got := make([]result, len(cases))
	for i, tc := range cases {
		line, ok := phaseLine(markers, tc.aliases)
		got[i] = result{line: line, ok: ok}
	}

	// then the earliest matching marker line comes back
	for i, tc := range cases {
		if got[i].ok != tc.wantOK || got[i].line != tc.want {
			t.Errorf("%s: phaseLine(%v) = (%d, %v), want (%d, %v)",
				tc.name, tc.aliases, got[i].line, got[i].ok, tc.want, tc.wantOK)
		}
	}
}

func TestEmptyPhase(t *testing.T) {
	t.Parallel()

	// given a body whose statements start at lines 4 and 9
	const (
		whenEmpty  = "when phase has no statements, a marker must be followed by its code"
		thenEmpty  = "then phase has no statements, a marker must be followed by its code"
		givenEmpty = "given marker has no statements, given is optional: delete the marker or move the setup under it"
	)
	spans := [][2]int{{4, 5}, {9, 9}}

	// when each marker window is searched for a statement
	cases := []struct {
		name    string
		present []bool
		lineOf  []int
		want    string
	}{
		{name: "every marker has code", present: []bool{false, true, true}, lineOf: []int{0, 3, 9}},
		{name: "when glued to then", present: []bool{false, true, true}, lineOf: []int{0, 3, 4}, want: whenEmpty},
		{name: "dangling then", present: []bool{false, true, true}, lineOf: []int{0, 3, 11}, want: thenEmpty},
		{name: "dangling given", present: []bool{true, true, true}, lineOf: []int{2, 3, 9}, want: givenEmpty},
	}
	got := make([]string, len(cases))
	for i, tc := range cases {
		got[i] = emptyPhase(tc.present, tc.lineOf, spans, 20)
	}

	// then only a marker with no statements in its window is reported
	for i, tc := range cases {
		if got[i] != tc.want {
			t.Errorf("%s: emptyPhase = %q, want %q", tc.name, got[i], tc.want)
		}
	}
}

func TestHasStatement(t *testing.T) {
	t.Parallel()

	// given statement spans inside a body
	spans := [][2]int{{3, 4}, {9, 9}}

	// when a marker window is searched
	cases := []struct {
		name string
		from int
		to   int
		want bool
	}{
		{name: "statement starts after the marker", from: 3, to: 9, want: true},
		{name: "statement starts on the marker line", from: 9, to: 20, want: true},
		{name: "window before any statement", from: 1, to: 3, want: false},
		{name: "window between statements", from: 5, to: 9, want: false},
		{name: "window past the end", from: 10, to: 20, want: false},
	}
	got := make([]bool, len(cases))
	for i, tc := range cases {
		got[i] = hasStatement(spans, tc.from, tc.to)
	}

	// then only a statement starting inside the window counts
	for i, tc := range cases {
		if got[i] != tc.want {
			t.Errorf("%s: hasStatement(%d, %d) = %v, want %v",
				tc.name, tc.from, tc.to, got[i], tc.want)
		}
	}
}
