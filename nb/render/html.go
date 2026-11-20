// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"fmt"
	"html"
	"strings"

	"github.com/asymmetric-effort/go-notebook/nb/document"
)

// cspMeta is the Content-Security-Policy meta tag that blocks all remote loads.
const cspMeta = `<meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; img-src data:;">`

// HTML renders a document to a standalone HTML file. Display output is
// embedded. Images are inlined as data URIs. The CSP blocks all remote loads.
func HTML(doc *document.Document, title string) string {
	var body strings.Builder

	for _, e := range doc.Elements {
		switch e.Kind {
		case document.KindProse:
			// Prose is passed through as HTML. A full CommonMark renderer
			// would convert Markdown to HTML here. For now, wrap in a div.
			body.WriteString("<div class=\"prose\">")
			body.WriteString(html.EscapeString(e.Content))
			body.WriteString("</div>\n")
		case document.KindDisplay:
			if e.Mime == "text/markdown" {
				body.WriteString("<div class=\"output\">")
				body.WriteString(html.EscapeString(e.Content))
				body.WriteString("</div>\n")
			} else if strings.HasPrefix(e.Mime, "image/") {
				body.WriteString(fmt.Sprintf("<img src=\"data:%s;base64,%s\" alt=\"%s\">\n",
					e.Mime, e.Content, html.EscapeString(e.CellName)))
			}
		case document.KindError:
			body.WriteString(fmt.Sprintf("<div class=\"error\"><strong>Error in cell %q (line %d):</strong> %s</div>\n",
				html.EscapeString(e.CellName), e.Line, html.EscapeString(e.Content)))
		case document.KindImage:
			body.WriteString(fmt.Sprintf("<img src=\"data:%s;base64,%s\" alt=\"image\">\n",
				e.Mime, e.Content))
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>%s</title>
%s
<style>
body { font-family: system-ui, sans-serif; max-width: 800px; margin: 2em auto; padding: 0 1em; }
.error { background: #fee; border: 1px solid #c00; padding: 1em; margin: 1em 0; }
.output { background: #f8f8f8; padding: 1em; margin: 1em 0; }
pre { background: #f4f4f4; padding: 1em; overflow-x: auto; }
</style>
</head>
<body>
%s</body>
</html>
`, html.EscapeString(title), cspMeta, body.String())
}
