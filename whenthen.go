// Package whenthen provides an analyzer that requires the given, when and then
// phases in every test function, one phase per line.
package whenthen

import (
	"go/ast"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Phase is one step of the given/when/then structure.
type Phase struct {
	Name     string   // canonical name used in messages
	Aliases  []string // any of these words marks the phase, case-insensitive
	Optional bool     // given is setup, not every test needs it
}

// RuleHint ends every message so a failing gate points back at the rule.
const RuleHint = " [coding-rules: given/when/then]"

// Phases lists the phases in required order.
var Phases = []Phase{
	{Name: "given", Aliases: []string{"given", "arrange", "setup"}, Optional: true},
	{Name: "when", Aliases: []string{"when", "act"}},
	{Name: "then", Aliases: []string{"then", "assert"}},
}

var Analyzer = &analysis.Analyzer{
	Name: "whenthen",
	Doc:  "check that every Test function marks given, when and then phases on separate lines",
	Run:  run,
}

// marker is one phase marker found in a test body: a comment line or a fluent
// chain receiver such as when.act().
type marker struct {
	line int
	text string
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		checkFile(pass, file)
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, file *ast.File) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !isTest(fn) {
			continue
		}
		if msg := violation(pass, file, fn); msg != "" {
			pass.Reportf(fn.Body.Rbrace, "%s: %s%s", fn.Name.Name, msg, RuleHint)
		}
	}
}

// violation returns the first broken rule for fn, empty when the test is clean.
func violation(pass *analysis.Pass, file *ast.File, fn *ast.FuncDecl) string {
	lineOf, present := locatePhases(bodyMarkers(pass, file, fn))
	if missing := missingPhases(present); len(missing) > 0 {
		return "missing " + strings.Join(missing, ", ") + " phase(s)"
	}
	if msg := sameLine(present, lineOf); msg != "" {
		return msg
	}
	if msg := outOfOrder(present, lineOf); msg != "" {
		return msg
	}
	bodyEnd := pass.Fset.Position(fn.Body.Rbrace).Line
	return emptyPhase(present, lineOf, stmtSpans(pass, fn), bodyEnd)
}

// stmtSpans returns the line span of every top-level statement in the body.
func stmtSpans(pass *analysis.Pass, fn *ast.FuncDecl) [][2]int {
	spans := make([][2]int, 0, len(fn.Body.List))
	for _, stmt := range fn.Body.List {
		spans = append(spans, [2]int{
			pass.Fset.Position(stmt.Pos()).Line,
			pass.Fset.Position(stmt.End()).Line,
		})
	}
	return spans
}

// emptyPhase reports the first marker that introduces no statements. That is how
// a marker pair ends up glued together with only comments in between.
func emptyPhase(present []bool, lineOf []int, spans [][2]int, bodyEnd int) string {
	for i := range Phases {
		if !present[i] {
			continue
		}
		if !hasStatement(spans, lineOf[i], nextMarker(present, lineOf, i, bodyEnd)) {
			return Phases[i].Name + " phase has no statements, a marker must be followed by its code"
		}
	}
	return ""
}

// nextMarker returns the line of the closest marker after Phases[i], or one past
// the closing brace when it is the last phase.
func nextMarker(present []bool, lineOf []int, i, bodyEnd int) int {
	boundary := bodyEnd + 1
	for j := range Phases {
		if present[j] && lineOf[j] > lineOf[i] && lineOf[j] < boundary {
			boundary = lineOf[j]
		}
	}
	return boundary
}

// hasStatement reports whether a statement starts inside [from, to).
func hasStatement(spans [][2]int, from, to int) bool {
	for _, span := range spans {
		if span[0] >= from && span[0] < to {
			return true
		}
	}
	return false
}

// bodyMarkers collects phase markers from two styles: phase comments
// (// given, // when, // then) and fluent chains whose receiver is the phase
// name (when.act(), then.assert()). Bare identifiers and string literals do not
// count, so a table field named want or a slice of phase names stays silent.
func bodyMarkers(pass *analysis.Pass, file *ast.File, fn *ast.FuncDecl) []marker {
	markers := commentMarkers(pass, file, fn)
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		sel, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		receiver, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		markers = append(markers, marker{
			line: pass.Fset.Position(receiver.Pos()).Line,
			text: receiver.Name,
		})
		return true
	})
	return markers
}

// commentMarkers returns the comments inside fn's body, one marker per line.
func commentMarkers(pass *analysis.Pass, file *ast.File, fn *ast.FuncDecl) []marker {
	var markers []marker
	start, end := fn.Body.Lbrace, fn.Body.Rbrace
	for _, group := range file.Comments {
		if group.Pos() < start || group.End() > end {
			continue
		}
		for _, comment := range group.List {
			line := pass.Fset.Position(comment.Slash).Line
			for i, text := range strings.Split(comment.Text, "\n") {
				markers = append(markers, marker{line: line + i, text: text})
			}
		}
	}
	return markers
}

func locatePhases(markers []marker) ([]int, []bool) {
	lineOf := make([]int, len(Phases))
	present := make([]bool, len(Phases))
	for i, phase := range Phases {
		lineOf[i], present[i] = phaseLine(markers, phase.Aliases)
	}
	return lineOf, present
}

// phaseLine returns the line of the earliest marker that names the phase.
func phaseLine(markers []marker, aliases []string) (int, bool) {
	line, found := 0, false
	for _, m := range markers {
		if !matchesAny(m.text, aliases) {
			continue
		}
		if !found || m.line < line {
			line, found = m.line, true
		}
	}
	return line, found
}

func matchesAny(text string, aliases []string) bool {
	for _, alias := range aliases {
		if hasWord(text, alias) {
			return true
		}
	}
	return false
}

func missingPhases(present []bool) []string {
	var missing []string
	for i, phase := range Phases {
		if !phase.Optional && !present[i] {
			missing = append(missing, phase.Name)
		}
	}
	return missing
}

func sameLine(present []bool, lineOf []int) string {
	for i := range Phases {
		for j := i + 1; j < len(Phases); j++ {
			if present[i] && present[j] && lineOf[i] == lineOf[j] {
				return Phases[i].Name + " and " + Phases[j].Name +
					" on the same line, one phase per line"
			}
		}
	}
	return ""
}

func outOfOrder(present []bool, lineOf []int) string {
	for i := range Phases {
		for j := i + 1; j < len(Phases); j++ {
			if present[i] && present[j] && lineOf[j] < lineOf[i] {
				return Phases[j].Name + " before " + Phases[i].Name +
					", want given, when, then"
			}
		}
	}
	return ""
}

func isTest(fn *ast.FuncDecl) bool {
	if fn.Recv != nil || fn.Body == nil || !strings.HasPrefix(fn.Name.Name, "Test") {
		return false
	}
	params := fn.Type.Params.List
	if len(params) != 1 || len(params[0].Names) != 1 {
		return false
	}
	star, ok := params[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "testing" && sel.Sel.Name == "T"
}

func hasWord(text, word string) bool {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`).MatchString(text)
}
