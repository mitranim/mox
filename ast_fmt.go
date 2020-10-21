package mox

import (
	"regexp"
	"strings"
)

const (
	FmtCommentSingleLineEnclosingSpaces  = false
	FmtCommentMultiLineEnclosingNewlines = true
)

func FmtMox(nodes []Node) {
	for i, node := range nodes {
		switch node := node.(type) {
		case NodeSpace:
			nodes[i] = FmtNodeSpace(node)
		case NodeComment:
			nodes[i] = FmtNodeComment(node)
			// case NodeBlock:
			// FmtMox([]Node(node))
		}
	}
}

func FmtNodeSpace(node NodeSpace) NodeSpace {
	node = NodeSpace(trimTrailingSpaceMultiline(string(node)))
	node = NodeSpace(trimPolyMultilines(string(node)))
	return node
}

func FmtNodeComment(node NodeComment) NodeComment {
	content := strings.TrimSpace(string(node))
	content = trimTrailingSpaceMultiline(content)

	if countNewlines(content) == 0 {
		if FmtCommentSingleLineEnclosingSpaces {
			return NodeComment(" " + content + " ")
		}
		return NodeComment(content)
	}

	content = string(FmtNodeSpace(NodeSpace(content)))

	if FmtCommentMultiLineEnclosingNewlines {
		return NodeComment("\n" + content + "\n")
	}
	return NodeComment(content)
}

func trimTrailingSpaceMultiline(str string) string {
	return regTrailingSpaceMultiline.ReplaceAllString(str, "$1")
}

/*
Incomplete. A full implementation should probably be equivalent to
`unicode.IsSpace`, excluding newlines.
*/
var regTrailingSpaceMultiline = regexp.MustCompile(`[ \t\v]+(\r\n|\r|\n)`)

func trimPolyMultilines(str string) string {
	return regPolyMultiline.ReplaceAllString(str, "$1")
}

var regPolyMultiline = regexp.MustCompile(`((?:\r\n|\r|\n){2})(?:\r\n|\r|\n)+`)

func countNewlines(str string) int {
	count := 0
	for _, char := range str {
		if char == '\n' {
			count++
		}
	}
	return count
}
