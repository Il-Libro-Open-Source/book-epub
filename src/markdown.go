package main

import (
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

// NewMarkdownConverter crea un nuovo convertitore Goldmark configurato per ePUB.
// Include supporto per tabelle, frontmatter YAML e syntax highlighting con Monokai.
// Genera output XHTML compatibile con lo standard ePUB.
func NewMarkdownConverter() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			&frontmatter.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithStyle("pastie"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
					chromahtml.WrapLongLines(true),
					chromahtml.TabWidth(2),
				),
			),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}
