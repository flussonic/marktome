package md2json

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type superlinksParser struct{}

var defaultSuperlinksParser = &superlinksParser{}

func NewSuperlinksParser() parser.InlineParser {
	return defaultSuperlinksParser
}

func (s *superlinksParser) Trigger() []byte {
	return []byte{'<'}
}

var superlinksContextKey = parser.NewContextKey()

var ending = "</link>"
var beginning = "<link"
var tagRegexp = regexp.MustCompile(`([a-z]+)="([^"]+)"`) //nolint:golint,lll

func (s *superlinksParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if !strings.HasPrefix(string(line), beginning) {
		return nil
	}
	start := strings.Index(string(line), ">")
	stop := strings.Index(string(line), ending)
	if stop < 0 {
		return nil
	}
	link := ast.NewLink()
	tag := line[len(beginning)+1 : start]
	attrs := tagRegexp.FindAllSubmatch(tag, -1)
	for i := range attrs {
		link.SetAttribute(attrs[i][1], attrs[i][2])
	}

	fmt.Println("Checking link", string(line[:stop]), stop+len(ending))
	// link := ast.NewTextSegment(seg)
	link.Title = line[start+1 : stop]
	block.Advance(stop + len(ending))
	return link
}

type superlinks struct{}

var Superlinks = &superlinks{}

func (e *superlinks) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewSuperlinksParser(), 210),
	))
}
