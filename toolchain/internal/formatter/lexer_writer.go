package formatter

import "strings"

type fmtWriter struct {
	b           strings.Builder
	indent      int
	lineStarted bool
}

func newFmtWriter() *fmtWriter { return &fmtWriter{} }

func (w *fmtWriter) String() string { return w.b.String() }

func (w *fmtWriter) writeIndent() {
	if w.lineStarted {
		return
	}
	w.b.WriteString(strings.Repeat("  ", w.indent))
	w.lineStarted = true
}

func (w *fmtWriter) line(s string) {
	w.writeIndent()
	w.b.WriteString(s)
	w.b.WriteByte('\n')
	w.lineStarted = false
}

func (w *fmtWriter) lineWithTrailing(s string, trailing *commentNode) {
	if trailing == nil {
		w.line(s)
		return
	}
	w.line(s + " " + trailing.Text)
}

func (w *fmtWriter) blank() {
	if w.b.Len() == 0 {
		return
	}
	if !strings.HasSuffix(w.b.String(), "\n") {
		w.b.WriteByte('\n')
	}
	w.b.WriteByte('\n')
	w.lineStarted = false
}
