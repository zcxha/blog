package folio

import (
	"bytes"
	"html"
	"strings"

	latex "github.com/aziis98/goldmark-latex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

var markdownEngine = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Footnote,
		latex.NewLatex(
			latex.WithOutputInlineDelim(`\(`, `\)`),
			latex.WithOutputBlockDelim(`\[`, `\]`),
		),
	),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	// The repository owner authors every post, so embedded HTML is intentional.
	goldmark.WithRendererOptions(
		goldmarkhtml.WithUnsafe(),
		renderer.WithNodeRenderers(util.Prioritized(&safeLatexRenderer{}, 100)),
	),
)

// goldmark-latex deliberately writes raw TeX. Escaping it as HTML text is
// required for expressions such as $x<y$; the browser decodes the entities
// before MathJax reads the text node.
type safeLatexRenderer struct{}

func (r *safeLatexRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(latex.KindInlineLatex, r.renderInline)
	reg.Register(latex.KindLatexBlock, r.renderBlock)
}

func (r *safeLatexRenderer) renderInline(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`\(`)
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			value := child.(*ast.Text).Segment.Value(source)
			_, _ = w.WriteString(html.EscapeString(string(value)))
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString(`\)`)
	return ast.WalkContinue, nil
}

func (r *safeLatexRenderer) renderBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	block := node.(*latex.LatexBlock)
	if entering {
		_, _ = w.WriteString(`<p>\[`)
		for i := 0; i < block.Lines().Len(); i++ {
			line := block.Lines().At(i)
			_, _ = w.WriteString(html.EscapeString(string(line.Value(source))))
		}
	} else {
		_, _ = w.WriteString(`\]</p>` + "\n")
	}
	return ast.WalkContinue, nil
}

func renderMarkdownGoldmark(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}

	var out bytes.Buffer
	if err := markdownEngine.Convert([]byte(input), &out); err != nil {
		// bytes.Buffer writes cannot fail; keeping this fallback makes loading a
		// post total even if a future renderer adds an I/O failure mode.
		return "<pre>" + templateEscape(input) + "</pre>"
	}
	return out.String()
}

func templateEscape(input string) string {
	return html.EscapeString(input)
}
