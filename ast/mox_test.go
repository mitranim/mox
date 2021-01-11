package mox

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func init() {
	spew.Config.Indent = "  "
	spew.Config.ContinueOnMethod = true
}

func TestTokenize(t *testing.T) {
	const src = `
; one
10
"two"
three
; four
; five
[six]
`

	tok := NewTokenizer(src)
	var tokens []Token

	for {
		token, err := tok.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("unexpected tokenization error: %+v", err)
		}
		tokens = append(tokens, token)
	}

	expected := []Token{
		TokenSpace("\n"),
		TokenComment(" one\n"),
		TokenNumber("10"),
		TokenSpace("\n"),
		TokenStringDouble("two"),
		TokenSpace("\n"),
		TokenIdent("three"),
		TokenSpace("\n"),
		TokenComment(" four\n"),
		TokenComment(" five\n"),
		TokenBracketOpen{},
		TokenIdent("six"),
		TokenBracketClose{},
		TokenSpace("\n"),
	}

	if !reflect.DeepEqual(expected, tokens) {
		t.Fatalf("tokenization mismatch;\nexpected tokens:\n%v\nreceived tokens:\n%v",
			spew.Sdump(expected), spew.Sdump(tokens))
	}
}
