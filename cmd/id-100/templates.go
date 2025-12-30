package main

import (
	"html/template"
	"log"
	"path/filepath"
)

func LoadTemplates() *Template {
	files, err := filepath.Glob("web/templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	comps, _ := filepath.Glob("web/templates/components/*.html")
	files = append(files, comps...)

	funcs := template.FuncMap{
		"eq": func(a, b string) bool { return a == b },
	}
	tmpl := template.New("").Funcs(funcs)
	tmpls, err := tmpl.ParseFiles(files...)
	if err != nil {
		log.Fatalf("failed to parse templates %v: %v", files, err)
	}
	return &Template{templates: tmpls}
}
