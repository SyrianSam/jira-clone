package handler

import (
	"html/template"
)

func LoadTemplates(pattern string) *template.Template {
	templates := template.Must(template.ParseGlob(pattern))
	return templates
}
