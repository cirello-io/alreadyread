package frontend

import (
	"embed"
	"html/template"
)

//go:embed index.html assets
var Content embed.FS

//go:embed linkTable.html
var linkTableTPL string

var LinkTable = template.Must(template.New("linkTable").Parse(linkTableTPL))
