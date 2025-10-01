package mox

/*
EXTREMELY incomplete sketch for a Mox fmter.

The current implementation prioritizes simplicity over efficiency. There are
many obvious optimizations, but they must wait.

Pending / TODO:

* Different rules for different syntax styles (data notation, various langs).

* Optional: enforce consistent inline vs multiline alignment of expressions
within blocks. Requires expression grouping, which requires knowledge of infix
operators. Requires knowledge of the call convention; the first token within a
block may be enforced on the same line.

* Enforce consistent spacing for operators. Optionally, use special knowledge of
prefix, infix, postfix to determine when to avoid spacing.
*/

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

var FmtConfDefault = FmtConf{Indent: "  "}

type FmtConf struct {
	Indent string `json:"indent"`
}

func Fmt(out io.Writer, src string, conf FmtConf) error {
	fmter := fmter{
		out:  out,
		tok:  Tokenizer{Source: src},
		conf: conf,
	}
	return fmter.top()
}

const (
	separator = ' '
	newline   = '\n'
)

type fmter struct {
	out     io.Writer
	tok     Tokenizer
	conf    FmtConf
	prev    Token
	delims  int // Balance of opened/closed delims on current line.
	indent  int // Indentation level (amount of "tabs").
	shifted int // Current line indent change. Possible values: -1, 0, +1.
	inline  bool
	bufByte [1]byte
	bufRune [utf8.UTFMax]byte
}

func (self *fmter) top() (err error) {
	defer allowErr(&err, io.EOF)
	defer rec(&err)
	for {
		self.token(self.nextToken())
	}
	return
}

func (self *fmter) token(token Token) {
	switch token := token.(type) {
	case TokenSpace:
		self.space(token)
	// case TokenNewlines:
	// self.newlines(token)
	case TokenStringDouble, TokenStringGrave:
		self.string(token)
	case TokenParenOpen, TokenBracketOpen, TokenBraceOpen:
		self.delimOpen(token)
	case TokenParenClose, TokenBracketClose, TokenBraceClose:
		self.delimClose(token)
	default:
		self.atom(token)
	}
}

func (self *fmter) space(token TokenSpace) {
	lines := countNewlines(string(token))
	if lines > 2 {
		lines = 2
	}
	if self.prev != nil && !isDelimOpen(self.prev) {
		self.writeNewlines(lines)
		self.prev = token
	}
}

// func (self *fmter) newlines(token TokenNewlines) {
// 	if token > 2 {
// 		token = 2
// 	}
// 	if self.prev != nil && !isDelimOpen(self.prev) {
// 		self.writeToken(token)
// 		self.prev = token
// 	}
// }

func (self *fmter) atom(token Token) {
	self.before(token)
	self.writeToken(token)
	self.prev = token
}

func (self *fmter) string(token Token) {
	self.before(token)
	self.delimInc()
	self.writeToken(token)
	self.delimDec()
	self.prev = token
}

func (self *fmter) delimOpen(token Token) {
	self.before(token)
	self.writeToken(token)
	self.delimInc()
	self.prev = token
}

func (self *fmter) delimClose(token Token) {
	self.writeNewline()
	self.delimDec()
	self.before(token)
	self.writeToken(token)
	self.prev = token
}

func (self *fmter) nextToken() Token {
	token, err := self.tok.Token()
	if err != nil {
		panic(err)
	}
	return token
}

func (self *fmter) before(next Token) {
	if !self.inline {
		self.writeIndent()
	} else if !isDelim(self.prev) && !isDelim(next) {
		self.writeSeparator()
	}
}

func (self *fmter) delimInc() {
	self.delims++
}

func (self *fmter) delimDec() {
	self.delims--
	if self.delims < 0 && self.shifted == 0 {
		self.indent--
		self.shifted = -1
	}
}

func (self *fmter) newlineReset() {
	if self.delims > 0 {
		self.indent++
	}
	self.delims = 0
	self.shifted = 0
	self.inline = false
}

func (self *fmter) writeSeparator() {
	self.writeByte(separator)
}

func (self *fmter) writeNewline() {
	self.writeByte(newline)
}

func (self *fmter) writeNewlines(lines int) {
	for lines > 0 {
		self.writeNewline()
		lines--
	}
}

func (self *fmter) writeIndent() {
	for i := 0; i < self.indent; i++ {
		self.writeString(self.conf.Indent)
	}
}

func (self *fmter) writeToken(token Token) {
	self.writeString(token.String())
}

func (self *fmter) writeByte(char byte) {
	self.bufByte[0] = char
	self.writeBytes(self.bufByte[:])
}

func (self *fmter) writeRune(char rune) {
	size := utf8.EncodeRune(self.bufRune[:], char)
	self.writeBytes(self.bufRune[:size])
}

func (self *fmter) writeString(str string) {
	self.writeBytes(stringToBytesAlloc(str))
}

func (self *fmter) writeBytes(input []byte) {
	_, err := self.out.Write(input)
	if err != nil {
		panic(err)
	}
	if hasNewline(input) {
		self.newlineReset()
	}
	if hasNonWhitespaceSuffix(input) {
		self.inline = true
	}
}

// TODO consider supporting `\r`.
func hasNewline(input []byte) bool {
	return bytes.ContainsRune(input, newline)
}

// Slightly misleading name: checks if there's a non-whitespace suffix following
// the last newline, if any, ignoring suffix whitespace that isn't newlines.
// TODO consider supporting `\r`.
func hasNonWhitespaceSuffix(input []byte) bool {
	index := bytes.IndexRune(input, newline)
	if index >= 0 {
		input = input[index+1:]
	}
	return len(bytes.TrimSpace(input)) > 0
}

func isDelim(token Token) bool {
	return isDelimOpen(token) || isDelimClose(token)
}

func isDelimOpen(token Token) bool {
	switch token.(type) {
	case TokenParenOpen, TokenBracketOpen, TokenBraceOpen:
		return true
	default:
		return false
	}
}

func isDelimClose(token Token) bool {
	switch token.(type) {
	case TokenParenClose, TokenBracketClose, TokenBraceClose:
		return true
	default:
		return false
	}
}

func rec(ptr *error) {
	if *ptr != nil {
		return
	}

	val := recover()
	if val == nil {
		return
	}

	err, _ := val.(error)
	if err != nil {
		*ptr = err
		return
	}

	panic(val)
}

func allowErr(ptr *error, predicate error) {
	if errors.Is(*ptr, predicate) {
		*ptr = nil
	}
}

func countNewlines(str string) int {
	var count int

	for i, char := range str {
		if char == '\r' && i < len(str) && str[i] == '\n' {
			continue
		}
		if char == '\r' || char == '\n' {
			count++
		}
	}

	return count
}
