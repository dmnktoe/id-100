package main

import (
	"fmt"
	"html/template"
	"path/filepath"
)

func main() {
	files, _ := filepath.Glob("web/templates/*.html")
	comps, _ := filepath.Glob("web/templates/components/*.html")
	files = append(files, comps...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		fmt.Printf("parse error: %v\n", err)
		return
	}
	fmt.Println("parsed templates:")
	for _, n := range t.Templates() {
		fmt.Println(" -", n.Name())
	}
}
