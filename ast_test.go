package mox

import (
	"reflect"
	"testing"
)

func TestFormat(t *testing.T) {
	ast := []Node{
		NodeComment(`# comment`),
		NodeWhitespace(` `),
		NodeBlock{
			NodeNumber(`123.456`),
			NodeWhitespace(` `),
			NodeStringDouble(`hello world`),
		},
	}

	expected := `[# comment] (123.456 "hello world")`
	actual := Format(ast)

	if expected != actual {
		t.Fatalf("expected formatted AST to be:\n%s\ngot:\n%s\n", expected, actual)
	}
}

func TestFormatWithNil(t *testing.T) {
	ast := []Node{nil, NodeIdentifier(`one`), nil, NodeIdentifier(`two`), nil}

	expected := `one two`
	actual := Format(ast)

	if expected != actual {
		t.Fatalf("expected formatted AST to be:\n%s\ngot:\n%s\n", expected, actual)
	}
}

func TestParseAndFormat(t *testing.T) {
	source := `[# comment [nested]] (123.456 ++ ident "hello world")`

	ast, err := Parse(source)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expectedAst := []Node{
		NodeComment(`# comment [nested]`),
		NodeWhitespace(` `),
		NodeBlock{
			NodeNumber(`123.456`),
			NodeWhitespace(` `),
			NodeOperator(`++`),
			NodeWhitespace(` `),
			NodeIdentifier(`ident`),
			NodeWhitespace(` `),
			NodeStringDouble(`hello world`),
		},
	}

	if !reflect.DeepEqual(expectedAst, ast) {
		t.Fatalf("expected parsed AST to be:\n%#v\ngot:\n%#v\n", expectedAst, ast)
	}

	formatted := Format(ast)
	if source != formatted {
		t.Fatalf("expected formatted AST to be:\n%s\ngot:\n%s\n", source, formatted)
	}
}

func TestParseAndFormatWithMinimalWhitespace(t *testing.T) {
	source := `[comment](123"one"++two"three"(four|456)789)five 012`

	ast, err := Parse(source)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expectedAst := []Node{
		NodeComment(`comment`),
		NodeBlock{
			NodeNumber(`123`),
			NodeStringDouble(`one`),
			NodeOperator(`++`),
			NodeIdentifier(`two`),
			NodeStringDouble(`three`),
			NodeBlock{
				NodeIdentifier(`four`),
				NodeOperator(`|`),
				NodeNumber(`456`),
			},
			NodeNumber(`789`),
		},
		NodeIdentifier(`five`),
		NodeWhitespace(` `),
		NodeNumber(`012`),
	}

	if !reflect.DeepEqual(expectedAst, ast) {
		t.Fatalf("expected parsed AST to be:\n%#v\ngot:\n%#v\n", expectedAst, ast)
	}

	formatted := Format(ast)
	if source != formatted {
		t.Fatalf("expected formatted AST to be:\n%s\ngot:\n%s\n", source, formatted)
	}
}

func TestFormatAddingWhitespace(t *testing.T) {
	ast := []Node{
		NodeIdentifier(`one`),
		NodeIdentifier(`two`),
		NodeStringDouble(`three`),
		NodeIdentifier(`four`),
		NodeNumber(`123`),
		NodeNumber(`456`),
		NodeIdentifier(`five`),
		NodeOperator(`++`),
		NodeIdentifier(`six`),
	}

	expected := `one two"three"four 123 456 five++six`
	formatted := Format(ast)
	if expected != formatted {
		t.Fatalf("expected formatted AST to be:\n%s\ngot:\n%s\n", expected, formatted)
	}
}

func TestParseIncomplete(t *testing.T) {
	t.Run("incomplete_block", func(t *testing.T) {
		source := `(`
		ast, err := Parse(source)
		if err == nil {
			t.Fatalf("expected parse error for %q, got parsed AST: %#v", source, ast)
		}
	})

	t.Run("incomplete_string_double", func(t *testing.T) {
		source := `"`
		ast, err := Parse(source)
		if err == nil {
			t.Fatalf("expected parse error for %q, got parsed AST: %#v", source, ast)
		}
	})

	t.Run("incomplete_string_grave", func(t *testing.T) {
		source := "`"
		ast, err := Parse(source)
		if err == nil {
			t.Fatalf("expected parse error for %q, got parsed AST: %#v", source, ast)
		}
	})
}
