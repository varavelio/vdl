package formatter

import (
	"strings"

	"github.com/varavelio/gen"
)

func newFormatterOutput() *gen.Generator {
	return gen.New().WithSpaces(2)
}

func lineWithTrailing(output *gen.Generator, content string, trailing *commentNode) {
	if trailing == nil {
		output.Line(content)
		return
	}
	output.Line(content + " " + trailing.Text)
}

func blankLine(output *gen.Generator) {
	if output.String() == "" {
		return
	}
	if !strings.HasSuffix(output.String(), "\n") {
		output.Break()
	}
	output.Break()
}

func writeRenderedValue(output *gen.Generator, prefix, value string, trailing *commentNode) {
	if !strings.Contains(value, "\n") {
		lineWithTrailing(output, prefix+value, trailing)
		return
	}

	lines := strings.Split(value, "\n")
	output.Line(prefix + lines[0])
	for i := 1; i < len(lines)-1; i++ {
		if lines[i] == "" {
			blankLine(output)
			continue
		}
		output.Line(lines[i])
	}

	lineWithTrailing(output, lines[len(lines)-1], trailing)
}
