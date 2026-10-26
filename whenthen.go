// Package whenthen provides an analyzer that requires the given, when and then
// phases in every test function, one phase per line.
package whenthen

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
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

// marker is one phase marker found in a scope: a comment line or a fluent chain
// receiver such as when.act().
type marker struct {
	line int
	text string
}

// scope is one function body that owns its markers: the test body, or a closure
// such as a t.Run subtest. Markers never cross a func literal boundary, so a
// subtest is judged on its own statements instead of the test body's.
type scope struct {
	body    *ast.BlockStmt
	label   string // empty for the test body, "subtest at line N" for a closure
	markers []marker
	spans   [][2]int
	end     int
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
		checkTest(pass, file, fn)
	}
}

// checkTest judges the test body and every subtest closure that carries markers.
// A test that marks nothing itself but marks its subtests is fine: the phases
// belong to the closure that owns them.
func checkTest(pass *analysis.Pass, file *ast.File, fn *ast.FuncDecl) {
	root := newScope(pass, file, fn.Body, "")
	children := childScopes(pass, file, fn.Body)
	if participates(root) || len(children) == 0 {
		report(pass, fn, root, fn.Body.Rbrace)
	}
	for _, child := range children {
		report(pass, fn, child, child.body.Rbrace)
	}
}

// participates reports whether the scope names a phase at all. A bare t.Run call
// is not a marker, so a test that only fans out to subtests does not participate
// and leaves the phases to them.
func participates(s scope) bool {
	return len(phaseMarkers(s.markers)) > 0
}

// phaseMarkers drops the selector receivers that name nothing, such as t in
// t.Run(...).
func phaseMarkers(markers []marker) []marker {
	var kept []marker
	for _, m := range markers {
		for _, phase := range Phases {
			if matchesAny(m.text, phase.Aliases) {
				kept = append(kept, m)
				break
			}
		}
	}
	return kept
}

// report prints the first broken rule of one scope, nothing when it is clean.
func report(pass *analysis.Pass, fn *ast.FuncDecl, s scope, at token.Pos) {
	msg := scopeViolation(s)
	if msg == "" {
		return
	}
	if s.label == "" {
		pass.Reportf(at, "%s: %s%s", fn.Name.Name, msg, RuleHint)
		return
	}
	pass.Reportf(at, "%s: %s: %s%s", fn.Name.Name, s.label, msg, RuleHint)
}

// scopeViolation returns the first broken rule for s, empty when s is clean.
func scopeViolation(s scope) string {
	lineOf, present := locatePhases(s.markers)
	if missing := missingPhases(present); len(missing) > 0 {
		return "missing " + strings.Join(missing, ", ") + " phase(s)"
	}
	if msg := sameLine(present, lineOf); msg != "" {
		return msg
	}
	if msg := outOfOrder(present, lineOf); msg != "" {
		return msg
	}
	return emptyPhase(present, lineOf, s.spans, s.end)
}

func newScope(pass *analysis.Pass, file *ast.File, body *ast.BlockStmt, label string) scope {
	return scope{
		body:    body,
		label:   label,
		markers: scopeMarkers(pass, file, body),
		spans:   stmtSpans(pass, body),
		end:     pass.Fset.Position(body.Rbrace).Line,
	}
}

// childScopes returns every function literal in body that carries its own
// markers. A closure without markers is a step of the test, not a test of its
// own, so it is left alone.
func childScopes(pass *analysis.Pass, file *ast.File, body *ast.BlockStmt) []scope {
	var scopes []scope
	ast.Inspect(body, func(node ast.Node) bool {
		lit, ok := node.(*ast.FuncLit)
		if !ok {
			return true
		}
		label := "subtest at line " + strconv.Itoa(litLine(pass, lit))
		if s := newScope(pass, file, lit.Body, label); len(phaseMarkers(s.markers)) > 0 {
			scopes = append(scopes, s)
		}
		return true
	})
	return scopes
}

// litLine returns the line of the closure's first statement, or of its func
// keyword when the body is empty.
func litLine(pass *analysis.Pass, lit *ast.FuncLit) int {
	if len(lit.Body.List) > 0 {
		return pass.Fset.Position(lit.Body.List[0].Pos()).Line
	}
	return pass.Fset.Position(lit.Type.Func).Line
}

// scopeMarkers collects phase markers from two styles: phase comments
// (// given, // when, // then) and fluent chains whose receiver is the phase
// name (when.act(), then.assert()). Bare identifiers, string literals and
// anything inside a nested closure do not count.
func scopeMarkers(pass *analysis.Pass, file *ast.File, body *ast.BlockStmt) []marker {
	markers := commentMarkers(pass, file, body)
	ast.Inspect(body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.SelectorExpr:
			if id, ok := n.X.(*ast.Ident); ok {
				markers = append(markers, marker{
					line: pass.Fset.Position(id.Pos()).Line,
					text: id.Name,
				})
			}
		}
		return true
	})
	return markers
}

// commentMarkers returns the comments inside body, one marker per line, skipping
// the comments that belong to a nested closure.
func commentMarkers(pass *analysis.Pass, file *ast.File, body *ast.BlockStmt) []marker {
	inner := funcLitRanges(body)
	var markers []marker
	for _, group := range file.Comments {
		if group.Pos() < body.Lbrace || group.End() > body.Rbrace {
			continue
		}
		for _, comment := range group.List {
			if inRanges(inner, comment.Pos()) {
				continue
			}
			line := pass.Fset.Position(comment.Slash).Line
			for i, text := range strings.Split(comment.Text, "\n") {
				if !anyPhaseWord(firstWord(text)) {
					continue
				}
				markers = append(markers, marker{line: line + i, text: strings.TrimSpace(text)})
			}
		}
	}
	return markers
}

// firstWord returns the leading word of a comment line with the comment markers
// stripped. A phase marker has to open the line, so prose that merely mentions
// "then" mid-sentence does not count.
func firstWord(text string) string {
	trimmed := strings.TrimLeft(strings.TrimSpace(text), "/* 	")
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], "*/")
}

// anyPhaseWord reports whether word names one of the phases.
func anyPhaseWord(word string) bool {
	for _, phase := range Phases {
		if matchesAny(word, phase.Aliases) {
			return true
		}
	}
	return false
}

// funcLitRanges returns the source range of every function literal in body.
func funcLitRanges(body *ast.BlockStmt) [][2]token.Pos {
	var ranges [][2]token.Pos
	ast.Inspect(body, func(node ast.Node) bool {
		if lit, ok := node.(*ast.FuncLit); ok {
			ranges = append(ranges, [2]token.Pos{lit.Pos(), lit.End()})
		}
		return true
	})
	return ranges
}

func inRanges(ranges [][2]token.Pos, pos token.Pos) bool {
	for _, r := range ranges {
		if pos >= r[0] && pos <= r[1] {
			return true
		}
	}
	return false
}

// stmtSpans returns the line span of every statement in the scope. Statements
// inside a nested closure belong to that closure, so a marker above a loop still
// owns the loop's statements.
func stmtSpans(pass *analysis.Pass, body *ast.BlockStmt) [][2]int {
	var spans [][2]int
	ast.Inspect(body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			return false
		case ast.Stmt:
			spans = append(spans, [2]int{
				pass.Fset.Position(n.Pos()).Line,
				pass.Fset.Position(n.End()).Line,
			})
		}
		return true
	})
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
			return emptyPhaseMessage(Phases[i])
		}
	}
	return ""
}

// emptyPhaseMessage points at the fix that fits the phase: a mandatory phase
// needs code, an optional one is usually better deleted.
func emptyPhaseMessage(phase Phase) string {
	if phase.Optional {
		return phase.Name + " marker has no statements, " + phase.Name +
			" is optional: delete the marker or move the setup under it"
	}
	return phase.Name + " phase has no statements, a marker must be followed by its code"
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
