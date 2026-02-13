package app

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var uiTmpl *template.Template

func (a *App) loadTemplates() (*template.Template, error) {
	if uiTmpl != nil {
		return uiTmpl, nil
	}
	base := "templates"
	pattern := filepath.Join(base, "*.html")
	t, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}
	uiTmpl = t
	return uiTmpl, nil
}

func (a *App) renderTemplate(w http.ResponseWriter, r *http.Request, name string, data any, status int) {
	t, err := a.loadTemplates()
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
}
