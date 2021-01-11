package mox

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const CommentStart = `;`
const CharEscape = '\\'

type Token fmt.Stringer
type TokenSpace string

// type TokenNewlines int
type TokenComment string
type TokenNumber string
type TokenStringDouble string
type TokenStringGrave string
type TokenIdent string
type TokenOperator string
type TokenParenOpen struct{}
type TokenParenClose struct{}
type TokenBracketOpen struct{}
type TokenBracketClose struct{}
type TokenBraceOpen struct{}
type TokenBraceClose struct{}

func (self TokenSpace) String() string { return string(self) }

// func (self TokenNewlines) String() string     { return strings.Repeat("\n", int(self)) }
func (self TokenComment) String() string      { return CommentStart + string(self) }
func (self TokenNumber) String() string       { return string(self) }
func (self TokenStringDouble) String() string { return `"` + string(self) + `"` }
func (self TokenStringGrave) String() string  { return "`" + string(self) + "`" }
func (self TokenIdent) String() string        { return string(self) }
func (self TokenOperator) String() string     { return string(self) }
func (self TokenParenOpen) String() string    { return `(` }
func (self TokenParenClose) String() string   { return `)` }
func (self TokenBracketOpen) String() string  { return `[` }
func (self TokenBracketClose) String() string { return `]` }
func (self TokenBraceOpen) String() string    { return `{` }
func (self TokenBraceClose) String() string   { return `}` }

func NewTokenizer(src string) *Tokenizer {
	return &Tokenizer{Source: src}
}

type Tokenizer struct {
	Source  string
	cursor  int
	closers []byte
}

func (self *Tokenizer) Token() (tok Token, err error) {
	// Allows to use panics for deep returns, simplifying the code.
	defer func() {
		if tok != nil || err != nil {
			return
		}
		val := recover()
		tok, _ = val.(Token)
		err, _ = val.(error)
		if tok == nil && err == nil && val != nil {
			panic(val)
		}
	}()

	self.maybeSpace()
	// self.maybeWhitespace()
	self.maybeComment()
	self.maybeNumber()
	self.maybeString()
	self.maybeOperator()
	self.maybeIdent()
	self.maybeParens()
	self.maybeBrackets()
	self.maybeBraces()

	if self.more() {
		return nil, self.error(fmt.Errorf(`failed to parse %q`, self.preview()))
	}

	return nil, io.EOF
}

func (self *Tokenizer) maybeSpace() {
	start := self.cursor
	for self.isNextWhitespace() {
		self.skipByte()
	}

	tok := TokenSpace(self.from(start))
	if len(tok) > 0 {
		panic(tok)
	}
}

// func (self *Tokenizer) maybeWhitespace() {
// 	var lines TokenNewlines

// 	for {
// 		if self.skippedNewline() {
// 			lines++
// 			continue
// 		}

// 		if self.isNextWhitespace() {
// 			self.skipByte()
// 			continue
// 		}

// 		break
// 	}

// 	if lines > 0 {
// 		panic(lines)
// 	}
// }

func (self *Tokenizer) maybeComment() {
	if !self.skippedString(CommentStart) {
		return
	}

	start := self.cursor
	for self.more() {
		if self.skippedNewline() {
			break
		}
		self.skipChar()
	}

	panic(TokenComment(self.from(start)))
}

func (self *Tokenizer) maybeNumber() {
	self.maybeNumberBin()
	self.maybeNumberOct()
	self.maybeNumberHex()
	self.maybeNumberDec()
}

func (self *Tokenizer) maybeNumberBin() {
	self.maybeNumberWithPrefix(`0b`, byteMapDigitsBin)
}

func (self *Tokenizer) maybeNumberOct() {
	self.maybeNumberWithPrefix(`0o`, byteMapDigitsOct)
}

func (self *Tokenizer) maybeNumberHex() {
	self.maybeNumberWithPrefix(`0h`, byteMapDigitsHex)
}

func (self *Tokenizer) maybeNumberDec() {
	if self.isNextByteIn(byteMapDigitsDec) {
		self.numberWith(byteMapDigitsDec)
	}
}

func (self *Tokenizer) maybeNumberWithPrefix(prefix string, byteMap []bool) {
	if self.skippedString(prefix) {
		self.numberWith(byteMap)
	}
}

func (self *Tokenizer) numberWith(byteMap []bool) {
	start := self.cursor

	if !self.isNextByteIn(byteMap) {
		panic(self.error(fmt.Errorf(`expected one of %q, found %q`, byteMapString(byteMap), self.preview())))
	}

	for self.more() {
		if self.isNextByteIn(byteMap) {
			self.skipByte()
			continue
		}
		if self.isNextByteIn(byteMapIdent) {
			panic(self.error(fmt.Errorf(`expected one of %q, found %q`, byteMapString(byteMap), self.preview())))
		}

		if self.skippedByte('.') {
			goto fraction
		}
		goto end
	}

fraction:
	if !self.isNextByteIn(byteMap) {
		panic(self.error(fmt.Errorf(`expected one of %q, found %q`, byteMapString(byteMap), self.preview())))
	}
	for self.isNextByteIn(byteMap) {
		self.skipByte()
	}
	if self.isNextByteIn(byteMapIdent) {
		panic(self.error(fmt.Errorf(`expected one of %q, found %q`, byteMapString(byteMap), self.preview())))
	}

end:
	tok := TokenNumber(self.from(start))
	// Internal sanity check.
	if !(len(tok) > 0) {
		panic(self.error(fmt.Errorf(`expected one of %q, found %q`, byteMapString(byteMap), self.preview())))
	}
	panic(tok)
}

func (self *Tokenizer) maybeString() {
	self.maybeStringDouble()
	self.maybeStringGrave()
}

func (self *Tokenizer) maybeStringDouble() {
	str, ok := self.maybeStringBetween('"', '"', true)
	if ok {
		panic(TokenStringDouble(str))
	}
}

func (self *Tokenizer) maybeStringGrave() {
	str, ok := self.maybeStringBetween('`', '`', false)
	if ok {
		panic(TokenStringGrave(str))
	}
}

func (self *Tokenizer) maybeStringBetween(opener byte, closer byte, escapes bool) (string, bool) {
	if !self.isNextByte(opener) {
		return "", false
	}

	self.skipByte()
	start := self.cursor

	for self.more() {
		if self.isNextByte(closer) {
			str := self.from(start)
			self.skipByte()
			return str, true
		}

		if escapes && self.isNextByte(CharEscape) {
			self.skipByte()
			if !self.more() {
				panic(self.error(fmt.Errorf(`expected escaped character after %q, got EOF`, CharEscape)))
			}
			self.skipChar()
			continue
		}

		self.skipChar()
	}

	panic(self.error(fmt.Errorf(`expected closing %q, got EOF`, '`')))
}

func (self *Tokenizer) maybeOperator() {
	start := self.cursor
	for self.isNextByteIn(byteMapOperator) {
		self.skipByte()
	}

	tok := TokenOperator(self.from(start))

	if len(tok) > 0 {
		// Mandated by spec to avoid ambiguities.
		if stringLastChar(string(tok)) == '.' && self.isNextByteIn(byteMapDigitsDec) {
			panic(self.error(fmt.Errorf(`unexpected digit after dot: %q`, self.previewFrom(start))))
		}
		panic(tok)
	}
}

func (self *Tokenizer) maybeIdent() {
	start := self.cursor

	if self.isNextByteIn(byteMapIdentStart) {
		self.skipByte()
		for self.isNextByteIn(byteMapIdent) {
			self.skipByte()
		}
		panic(TokenIdent(self.from(start)))
	}
}

func (self *Tokenizer) maybeParens() {
	self.maybeDelims('(', ')', TokenParenOpen{}, TokenParenClose{})
}

func (self *Tokenizer) maybeBrackets() {
	self.maybeDelims('[', ']', TokenBracketOpen{}, TokenBracketClose{})
}

func (self *Tokenizer) maybeBraces() {
	self.maybeDelims('{', '}', TokenBraceOpen{}, TokenBraceClose{})
}

// Note: delim tokens are zero-sized. Their interface values should be
// allocation-free in modern Go. Therefore it's okay to instantiate them before
// passing.
func (self *Tokenizer) maybeDelims(opener byte, closer byte, open Token, close Token) {
	if self.skippedByte(opener) {
		self.pushCloser(closer)
		panic(open)
	}

	if self.isNextByte(closer) {
		if self.popCloser(closer) {
			self.skipByte()
			panic(close)
		}
		panic(self.error(fmt.Errorf(`unexpected closing %q`, closer)))
	}
}

func (self *Tokenizer) pushCloser(char byte) {
	self.closers = append(self.closers, char)
}

func (self *Tokenizer) popCloser(char byte) bool {
	index := len(self.closers) - 1
	if index >= 0 && self.closers[index] == char {
		self.closers = self.closers[:index]
		return true
	}
	return false
}

func (self *Tokenizer) skippedNewline() bool {
	return self.skippedString("\r\n") || self.skippedByte('\n') || self.skippedByte('\r')
}

func (self *Tokenizer) more() bool {
	return self.left() > 0
}

func (self *Tokenizer) left() int {
	return len(self.Source) - self.cursor
}

func (self *Tokenizer) headByte() byte {
	if self.cursor < len(self.Source) {
		return self.Source[self.cursor]
	}
	return 0
}

func (self *Tokenizer) from(start int) string {
	if start < 0 {
		start = 0
	}
	if start < self.cursor {
		return self.Source[start:self.cursor]
	}
	return ""
}

func (self *Tokenizer) rest() string {
	if self.more() {
		return self.Source[self.cursor:]
	}
	return ""
}

func (self *Tokenizer) preview() string {
	return self.previewFrom(self.cursor)
}

func (self *Tokenizer) previewFrom(start int) string {
	const limit = 32
	if self.left() > limit {
		return self.Source[start:start+limit] + " ..."
	}
	return self.Source[start:]
}

func (self *Tokenizer) isNextString(prefix string) bool {
	return strings.HasPrefix(self.rest(), prefix)
}

func (self *Tokenizer) isNextByte(char byte) bool {
	return self.headByte() == char
}

func (self *Tokenizer) isNextWhitespace() bool {
	return self.isNextByteIn(byteMapWhitespace)
}

func (self *Tokenizer) isNextByteIn(byteMap []bool) bool {
	return self.left() > 0 && isByteIn(byteMap, self.headByte())
}

func (self *Tokenizer) skipByte() {
	self.cursor++
}

func (self *Tokenizer) skipChar() {
	_, size := utf8.DecodeRuneInString(self.rest())
	self.cursor += size
}

func (self *Tokenizer) skipString(str string) {
	self.skipNBytes(len(str))
}

func (self *Tokenizer) skipNBytes(n int) {
	self.cursor += n
}

func (self *Tokenizer) skippedByte(char byte) bool {
	if self.isNextByte(char) {
		self.skipByte()
		return true
	}
	return false
}

func (self *Tokenizer) skippedString(prefix string) bool {
	if self.isNextString(prefix) {
		self.skipString(prefix)
		return true
	}
	return false
}

func (self Tokenizer) error(cause error) error {
	return Error{Cause: cause, Tokenizer: self}
}

type Error struct {
	Cause     error
	Tokenizer Tokenizer
}

// TODO: include tokenizer info.
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

/*
Fancier printing.

TODO: support '#' properly; include tokenizer context such as line, column, and
surrounding text.
*/
func (self Error) Format(fms fmt.State, verb rune) {
	switch verb {
	case 'v':
		if fms.Flag('#') || fms.Flag('+') {
			fmt.Fprintf(fms, "position %v: ", self.Tokenizer.cursor)
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

// Self-reminder about non-free conversions.
func bytesToStringAlloc(bytes []byte) string { return string(bytes) }

// Self-reminder about non-free conversions.
func stringToBytesAlloc(input string) []byte { return []byte(input) }

var byteMapWhitespace = stringByteMap(" \n\r\t\v")
var byteMapOperator = stringByteMap(`~!@#$%^&*:<>.?/\|=+-`)
var byteMapIdentStart = stringByteMap(`ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_`)
var byteMapIdent = stringByteMap(`ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_0123456789`)
var byteMapDigitsBin = stringByteMap(`01`)
var byteMapDigitsOct = stringByteMap(`01234567`)
var byteMapDigitsDec = stringByteMap(`0123456789`)
var byteMapDigitsHex = stringByteMap(`0123456789abcdef`)

func isByteIn(chars []bool, char byte) bool {
	index := int(char)
	return index < len(chars) && chars[index]
}

func stringByteMap(str string) []bool {
	var max int
	for _, char := range str {
		if int(char) > max {
			max = int(char)
		}
	}

	byteMap := make([]bool, max+1)
	for _, char := range str {
		byteMap[int(char)] = true
	}
	return byteMap
}

func byteMapString(byteMap []bool) string {
	var buf strings.Builder
	for char, ok := range byteMap {
		if ok {
			buf.WriteRune(rune(char))
		}
	}
	return buf.String()
}

func stringLastChar(str string) rune {
	char, _ := utf8.DecodeLastRuneInString(str)
	return char
}
