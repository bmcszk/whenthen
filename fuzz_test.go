package whenthen

import "testing"

func FuzzHasWord(f *testing.F) {
	// given seeds covering whole-word, substring and empty shapes
	f.Add("// when it decodes", "when")
	f.Add("whenever", "when")
	f.Add("// then", "then")
	f.Add("", "")
	f.Add("\x8e", "\x8e")

	// when hasWord runs on arbitrary comment text
	f.Fuzz(func(t *testing.T, text, word string) {
		// then it reports a match without panicking
		_ = hasWord(text, word)
	})
}

func FuzzEmptyPhase(f *testing.F) {
	// given marker lines and a body end
	f.Add(3, 9)
	f.Add(0, 1)
	f.Add(12, 13)

	// when the rule runs on arbitrary line numbers
	f.Fuzz(func(t *testing.T, marker, boundary int) {
		if marker < 0 || boundary <= marker {
			return
		}
		// then it reports without panicking on out-of-range input
		spans := [][2]int{{marker, marker + 1}}
		_ = emptyPhase([]bool{false, true, true}, []int{0, marker, boundary}, spans, boundary+1)
	})
}
