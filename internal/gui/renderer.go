package gui

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

// TemplateRenderer handles HTML template rendering
type TemplateRenderer struct {
	templates *template.Template
	devMode   bool
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(devMode bool) *TemplateRenderer {
	renderer := &TemplateRenderer{
		devMode: devMode,
	}
	renderer.loadTemplates()
	return renderer
}

// loadTemplates loads all HTML templates
func (r *TemplateRenderer) loadTemplates() {
	tmpl := template.New("")
	
	// Add template functions
	tmpl.Funcs(template.FuncMap{
		"formatDuration": func(d time.Duration) string {
			if d == 0 {
				return "0s"
			}
			return d.String()
		},
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatBytes": func(bytes int64) string {
			const unit = 1024
			if bytes < unit {
				return fmt.Sprintf("%d B", bytes)
			}
			div, exp := int64(unit), 0
			for n := bytes / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
		},
	})
	
	// Load template files using ParseGlob for easier loading
	var err error
	tmpl, err = tmpl.ParseGlob("internal/gui/templates/**/*.html")
	if err != nil {
		// Fallback: load templates individually
		templateDir := "internal/gui/templates"
		
		// Load base template
		baseTemplate := filepath.Join(templateDir, "base.html")
		tmpl = template.Must(tmpl.ParseFiles(baseTemplate))
		
		// Load component templates
		componentDir := filepath.Join(templateDir, "components")
		navbarTemplate := filepath.Join(componentDir, "navbar.html")
		tmpl = template.Must(tmpl.ParseFiles(navbarTemplate))
		
		// Load page templates
		pageDir := filepath.Join(templateDir, "pages")
		pageTemplates := []string{
			"index.html",
			"new-test.html", 
			"test-details.html",
			"test-list.html",
			"docs.html",
			"api-docs.html",
		}
		
		for _, pageTemplate := range pageTemplates {
			templatePath := filepath.Join(pageDir, pageTemplate)
			tmpl = template.Must(tmpl.ParseFiles(templatePath))
		}
	}
	
	r.templates = tmpl
}

// RenderTemplate renders a template with the given data
func (r *TemplateRenderer) RenderTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	if r.devMode {
		// Reload templates in dev mode for faster development
		r.loadTemplates()
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Execute the base template which will include the page content
	return r.templates.ExecuteTemplate(w, "base.html", data)
}

// PageData represents common data passed to all page templates
type PageData struct {
	Title      string
	ActivePage string
	Data       interface{}
}

// NewPageData creates a new PageData instance
func NewPageData(title, activePage string, data interface{}) *PageData {
	return &PageData{
		Title:      title,
		ActivePage: activePage,
		Data:       data,
	}
}