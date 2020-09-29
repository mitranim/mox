package mox

import (
	"reflect"
	"testing"
)

func TestFormat(t *testing.T) {
	ast := []Node{
		NodeComment(` # Comment `),
		NodeWhitespace(` `),
		NodeBlock{
			NodeNumber("123.456"),
			NodeStringDouble(`hello world`),
		},
	}

	expected := `(( # Comment )) {123.456 "hello world"}`
	actual := FormatNodes(ast)

	if expected != actual {
		t.Fatalf("expected formatted AST to be:\n%s\ngot:\n%s\n", expected, actual)
	}
}

func TestParseAndFormat(t *testing.T) {
	source := `(( # Comment (( nested )) )) {123.456 ident "hello world" {'a'}}`

	parser := ParserFromUtf8String(source)

	ast, err := parser.PopNodes()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	expectedAst := []Node{
		NodeComment(` # Comment (( nested )) `),
		NodeWhitespace(` `),
		NodeBlock{
			NodeNumber(`123.456`),
			NodeWhitespace(` `),
			NodeIdentifier(`ident`),
			NodeWhitespace(` `),
			NodeStringDouble(`hello world`),
			NodeWhitespace(` `),
			NodeBlock{
				NodeCharacter(`a`),
			},
		},
	}

	if !reflect.DeepEqual(expectedAst, ast) {
		t.Fatalf("expected parsed AST to be:\n%#v\ngot:\n%#v\n", expectedAst, ast)
	}

	formatted := FormatNodes(ast)
	if source != formatted {
		t.Fatalf("expected formatted AST to be:\n%s\ngot:\n%s\n", source, formatted)
	}
}
