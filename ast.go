package mox

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

type Node fmt.Stringer
type NodeWhitespace string
type NodeComment string
type NodeNumber string
type NodeStringDouble string
type NodeStringGrave string
type NodeIdentifier string
type NodeCharacter string
type NodeOperator string
type NodeBlock []Node

func (self NodeWhitespace) String() string   { return string(self) }
func (self NodeComment) String() string      { return `((` + string(self) + `))` }
func (self NodeNumber) String() string       { return string(self) }
func (self NodeStringDouble) String() string { return strconv.Quote(string(self)) }
func (self NodeStringGrave) String() string  { return "`" + string(self) + "`" }
func (self NodeIdentifier) String() string   { return string(self) }
func (self NodeCharacter) String() string    { return `'` + string(self) + `'` }
func (self NodeOperator) String() string     { return string(self) }
func (self NodeBlock) String() string        { return `{` + Format([]Node(self)) + `}` }

func Parse(input string) ([]Node, error) {
	parser := Parser{Source: input}
	return parser.PopNodes()
}

type Parser struct {
	Source string
	Cursor int
}

func (self *Parser) PopNodes() ([]Node, error) {
	var nodes []Node

	for self.More() {
		start := self.Cursor
		node, err := self.PopNode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes, node)
		self.mustHaveAdvanced(start)
	}

	return nodes, nil
}

func (self *Parser) PopNode() (_ Node, err error) {
	if !self.More() {
		return nil, self.Error(io.EOF)
	}

	defer func(start int) {
		if err == nil {
			self.mustHaveAdvanced(start)
		}
	}(self.Cursor)

	switch {
	case self.NextCharIn(charMapWhitespace):
		return self.PopWhitespace()
	case self.NextPair('(', '('):
		return self.PopComment()
	case self.NextCharIn(charMapDigitsDecimal):
		return self.PopNumber()
	case self.NextCharIn(charMapIdentifierStart):
		return self.PopIdentifier()
	case self.NextChar('\''):
		return self.PopCharacter()
	case self.NextChar('"'):
		return self.PopStringDouble()
	case self.NextChar('`'):
		return self.PopStringGrave()
	case self.NextChar('{'):
		return self.PopBlock()
	default:
		return nil, self.Error(fmt.Errorf(`unexpected %q`, self.Preview()))
	}
}

func (self *Parser) PopWhitespace() (NodeWhitespace, error) {
	start := self.Cursor
	for self.NextCharIn(charMapWhitespace) {
		self.Advance()
	}

	node := NodeWhitespace(self.From(start))
	if len(node) == 0 {
		return "", self.Error(fmt.Errorf(`expected whitespace, found %q`, self.Preview()))
	}
	return node, nil
}

func (self *Parser) PopComment() (NodeComment, error) {
	if !self.NextPair('(', '(') {
		return "", self.Error(fmt.Errorf(`expected opening "((", found %q`, self.Preview()))
	}

	self.AdvancePair('(', '(')
	start := self.Cursor
	levels := 1

	for self.More() {
		if self.NextPair('(', '(') {
			levels++
			self.AdvancePair('(', '(')
			continue
		}

		if self.NextPair(')', ')') {
			levels--
			if levels == 0 {
				node := NodeComment(self.From(start))
				self.AdvancePair(')', ')')
				return node, nil
			}
		}

		self.Advance()
	}

	return "", self.Error(fmt.Errorf(`expected closing "))", found unexpected EOF`))
}

func (self *Parser) PopNumber() (NodeNumber, error) {
	switch {
	case self.NextPair('0', 'b'):
		return self.PopNumberBinary()
	case self.NextPair('0', 'o'):
		return self.PopNumberOctal()
	case self.NextPair('0', 'x'):
		return self.PopNumberHexadecimal()
	default:
		return self.PopNumberDecimal()
	}
}

func (self *Parser) PopNumberBinary() (NodeNumber, error) {
	return self.popIntegerWithBase('0', 'b', charMapDigitsBinary)
}

func (self *Parser) PopNumberOctal() (NodeNumber, error) {
	return self.popIntegerWithBase('0', 'o', charMapDigitsOctal)
}

func (self *Parser) PopNumberHexadecimal() (NodeNumber, error) {
	return self.popIntegerWithBase('0', 'x', charMapDigitsHexadecimal)
}

func (self *Parser) popIntegerWithBase(a rune, b rune, charMap []bool) (NodeNumber, error) {
	if !self.NextPair(a, b) {
		return "", self.Error(fmt.Errorf(`expected opening "%c%c", found %q`, a, b, self.Preview()))
	}

	start := self.Cursor
	self.AdvancePair(a, b)
	for self.NextCharIn(charMap) {
		self.Advance()
	}

	node := NodeNumber(self.From(start))
	if !(len(node) > 2) {
		return node, self.Error(fmt.Errorf(`expected at least one digit, found %q`, self.Preview()))
	}
	return node, nil
}

// Placeholder implementation without exponents.
func (self *Parser) PopNumberDecimal() (NodeNumber, error) {
	start := self.Cursor

	if !self.NextCharIn(charMapDigitsDecimal) {
		return "", self.Error(fmt.Errorf(`expected digit, found %q`, self.Preview()))
	}

	for self.More() {
		if self.NextCharIn(charMapDigitsDecimal) {
			self.Advance()
			continue
		}

		if self.NextChar('.') {
			self.Advance()
			goto fraction
		}

		goto end
	}
	goto end

fraction:
	if !self.NextCharIn(charMapDigitsDecimal) {
		return "", self.Error(fmt.Errorf(`expected digit, found %q`, self.Preview()))
	}
	for self.NextCharIn(charMapDigitsDecimal) {
		self.Advance()
	}

end:
	node := NodeNumber(self.From(start))
	// Internal sanity check.
	if !(len(node) > 0) {
		return node, self.Error(fmt.Errorf(`expected at least one digit, found %q`, self.Preview()))
	}
	return node, nil
}

func (self *Parser) PopIdentifier() (NodeIdentifier, error) {
	if !self.NextCharIn(charMapIdentifierStart) {
		return "", self.Error(fmt.Errorf(`expected beginning of identifier, found %q`, self.Preview()))
	}

	start := self.Cursor
	self.Advance()
	for self.NextCharIn(charMapIdentifier) {
		self.Advance()
	}

	return NodeIdentifier(self.From(start)), nil
}

// Placeholder implementation without escapes.
func (self *Parser) PopCharacter() (NodeCharacter, error) {
	str, err := self.popStringBetween('\'', '\'')
	return NodeCharacter(str), err
}

// Placeholder implementation without escapes.
func (self *Parser) PopStringDouble() (NodeStringDouble, error) {
	str, err := self.popStringBetween('"', '"')
	return NodeStringDouble(str), err
}

func (self *Parser) PopStringGrave() (NodeStringGrave, error) {
	str, err := self.popStringBetween('`', '`')
	return NodeStringGrave(str), err
}

func (self *Parser) popStringBetween(prefix rune, suffix rune) (string, error) {
	if !self.NextChar(prefix) {
		return "", self.Error(fmt.Errorf(`expected opening %q, found %q`, prefix, self.Preview()))
	}

	self.Advance()
	start := self.Cursor

	for self.More() {
		if self.NextChar(suffix) {
			if self.Cursor == start {
				return "", self.Error(fmt.Errorf(`expected character, found unexpected closing %q`, suffix))
			}

			str := self.From(start)
			self.Advance()
			return str, nil
		}

		self.Advance()
	}

	return "", self.Error(fmt.Errorf(`expected character or closing %q, found EOF`, suffix))
}

func (self *Parser) PopBlock() (Node, error) {
	if !self.NextChar('{') {
		return nil, self.Error(fmt.Errorf(`expected opening "{", found %q`, self.Preview()))
	}

	self.Advance()
	var nodes NodeBlock

	for self.More() {
		if self.NextChar('}') {
			self.Advance()
			break
		}

		start := self.Cursor
		node, err := self.PopNode()
		if err != nil {
			return nodes, err
		}
		self.mustHaveAdvanced(start)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (self Parser) More() bool {
	return self.Left() > 0
}

func (self Parser) Left() int {
	return len(self.Source) - self.Cursor
}

func (self Parser) NextChar(char rune) bool {
	return self.Left() > 0 && self.Head() == char
}

func (self Parser) NextCharIn(chars []bool) bool {
	return self.Left() > 0 && isCharIn(chars, self.Head())
}

func (self Parser) NextPair(a rune, b rune) bool {
	for i, char := range self.Rest() {
		if i == 0 {
			if char != a {
				return false
			}
			continue
		}
		return char == b
	}
	return false
}

func (self Parser) Head() rune {
	char, _ := utf8.DecodeRuneInString(self.Rest())
	return char
}

func (self Parser) From(start int) string {
	if start < 0 {
		start = 0
	}
	if start < self.Cursor {
		return self.Source[start:self.Cursor]
	}
	return ""
}

func (self Parser) Rest() string {
	if self.More() {
		return self.Source[self.Cursor:]
	}
	return ""
}

func (self Parser) Preview() string {
	const limit = 32
	if self.Left() > limit {
		return self.Source[self.Cursor:self.Cursor+limit] + " ..."
	}
	return self.Source[self.Cursor:]
}

func (self *Parser) Advance() {
	_, size := utf8.DecodeRuneInString(self.Rest())
	self.Cursor += size
}

func (self *Parser) AdvancePair(a rune, b rune) {
	self.Cursor += utf8.RuneLen(a)
	self.Cursor += utf8.RuneLen(b)
}

func (self Parser) mustHaveAdvanced(start int) {
	if !(self.Cursor > start) {
		panic(self.Error(fmt.Errorf(`internal error: failed to advance cursor`)))
	}
}

func (self Parser) Error(cause error) error {
	return Error{Cause: cause, Parser: self}
}

type Error struct {
	Cause  error
	Parser Parser
}

// TODO: include parser info.
func (self Error) Error() string {
	if self.Cause != nil {
		return self.Cause.Error()
	}
	return ""
}

// Implement a hidden interface in "errors".
func (self Error) Unwrap() error {
	return self.Cause
}

// TODO: include parser info.
func (self Error) Format(fms fmt.State, verb rune) {
	switch verb {
	case 'v':
		if fms.Flag('#') || fms.Flag('+') {
			fmt.Fprintf(fms, "position %v: ", self.Parser.Cursor)
			if self.Cause != nil {
				fmt.Fprintf(fms, "%+v", self.Cause)
			}
			return
		}
		fms.Write(stringToBytesAlloc(self.Error()))
	default:
		fms.Write(stringToBytesAlloc(self.Error()))
	}
}

func Format(nodes []Node) string {
	var out string
	var prev Node

	for _, node := range nodes {
		/**
		Omitting nil can be convenient for AST editing. This allows to "remove" a
		node by replacing it with nil, instead of using `NodeWhitespace("")` or
		shifting the other nodes.
		*/
		if node == nil {
			continue
		}

		if requiresWhitespaceInfix(prev, node) {
			out += ` `
		}

		out += node.String()
		prev = node
	}

	return out
}

func requiresWhitespaceInfix(left Node, right Node) bool {
	return (isIdentifier(left) || isNumber(left)) &&
		(isIdentifier(right) || isNumber(right))
}

func isIdentifier(node Node) bool {
	_, ok := node.(NodeIdentifier)
	return ok
}

func isNumber(node Node) bool {
	_, ok := node.(NodeNumber)
	return ok
}

func stringToBytesAlloc(input string) []byte { return []byte(input) }

var charMapWhitespace = charMap(" \n\r\t\v")
var charMapOperator = charMap(`~!@#$%^&*:<>.?/\=+-`)
var charMapIdentifierStart = charMap(`ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_`)
var charMapIdentifier = charMap(`ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_0123456789`)
var charMapDigitsBinary = charMap(`01`)
var charMapDigitsOctal = charMap(`01234567`)
var charMapDigitsDecimal = charMap(`0123456789`)
var charMapDigitsHexadecimal = charMap(`0123456789abcdef`)

func isCharIn(chars []bool, char rune) bool {
	index := int(char)
	return index < len(chars) && chars[index]
}

func charMap(str string) []bool {
	var max int
	for _, char := range str {
		if int(char) > max {
			max = int(char)
		}
	}

	charMap := make([]bool, max+1)
	for _, char := range str {
		charMap[int(char)] = true
	}
	return charMap
}
