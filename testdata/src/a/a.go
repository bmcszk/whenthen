package a

import "testing"

func TestGivenWhenThen(t *testing.T) {
	// given a decoded blob
	x := decode()
	// when it is decoded again
	x++
	// then the length matches
	_ = x
}

func TestWhenThenOnly(t *testing.T) {
	// when the blob is decoded
	x := decode()
	// then the length matches
	_ = x
}

func TestAliases(t *testing.T) {
	// arrange a blob
	x := decode()
	// act on it
	x++
	// assert the result
	_ = x
}

func TestFluentChain(t *testing.T) {
	_, when, then := newParts(t)

	when.
		do()

	then.
		do().and().
		do()
}

func TestFluentMissingThen(t *testing.T) {
	_, when, _ := newParts(t)

	when.
		do()
} // want `TestFluentMissingThen: missing then phase\(s\)`

func TestEnumParityWhenOnly(t *testing.T) {
	// when
	// reading constant values to verify enum parity
	// then
	if got := decode(); got != 1 {
		t.Fatalf("theme count = %d, want 1", got)
	}
} // want `TestEnumParityWhenOnly: when phase has no statements`

func TestEmptyThen(t *testing.T) {
	// when the blob is decoded
	decode()
	// then the length matches
} // want `TestEmptyThen: then phase has no statements`

func TestMissingWhen(t *testing.T) {
	// given a blob
	_ = 1
	// then it decodes
} // want `TestMissingWhen: missing when phase\(s\)`

func TestMissingThen(t *testing.T) {
	// when it decodes
	_ = 1
} // want `TestMissingThen: missing then phase\(s\)`

func TestMissingBoth(t *testing.T) {
	_ = 1
} // want `TestMissingBoth: missing when, then phase\(s\)`

func TestSameLine(t *testing.T) {
	// when it decodes, then the length matches
	_ = 1
} // want `TestSameLine: when and then on the same line`

func TestThenFirst(t *testing.T) {
	// then the length matches
	_ = 1
	// when it decodes
} // want `TestThenFirst: then before when`

func TestCodeIdentifiers(t *testing.T) {
	given, when, then := 1, 2, 3
	_, _, _ = given, when, then
} // want `TestCodeIdentifiers: missing when, then phase\(s\)`

func helper(t *testing.T) {
	_ = t
}

func TestNoTestingParam() {
	_ = 1
}

func decode() int {
	return 1
}

type part struct{ t *testing.T }

func newParts(t *testing.T) (part, part, part) {
	parts := part{t: t}
	return parts, parts, parts
}

func (p part) do() part { return p }

func (p part) and() part { return p }

type phases struct{}

func (phases) TestMethod(t *testing.T) {
	_ = t
}
