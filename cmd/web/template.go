package main

import (
	"text/template"

	renderpkg "github.com/aspandyar/forum/internal/transport/http/render"
)

type templateData = renderpkg.TemplateData

func newTemplateCache() (map[string]*template.Template, error) {
	return renderpkg.NewTemplateCache()
}
