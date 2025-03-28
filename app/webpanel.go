package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"sync"
)

// Imbed the templates and static files into the binary so its all one happy family or file.

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

// WebPanelEnabled controls whether the web panel is available
var WebPanelEnabled = false

var (
	templates     *template.Template
	templatesOnce sync.Once
)

// InitWebPanel initializes templates for the web panel
func InitWebPanel() {
	templatesOnce.Do(func() {
		// Parse templates from embedded filesystem
		var err error
		templates, err = template.ParseFS(templateFS, "templates/*.html")
		if err != nil {
			logger.Error("Failed to parse templates", "error", err)
			return
		}
	})
}

// SetWebPanelEnabled sets whether the web panel is enabled
func SetWebPanelEnabled(enabled bool) {
	WebPanelEnabled = enabled
	if WebPanelEnabled {
		InitWebPanel()
	}
}

// IsWebPanelEnabled returns whether the web panel is enabled
func IsWebPanelEnabled() bool {
	return WebPanelEnabled
}

// EnableWebPanel enables the web panel interface
func EnableWebPanel() {
	SetWebPanelEnabled(true)
}

// DisableWebPanel disables the web panel interface
func DisableWebPanel() {
	SetWebPanelEnabled(false)
}

// SetupWebHandlers registers all handlers for the web panel
func SetupWebHandlers(mux *http.ServeMux) {
	// Get a sub-filesystem for the static directory
	subFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		logger.Error("Error creating sub-filesystem", "error", err)
		return
	}

	// Static file handler for CSS, JS, and images
	staticFSHandler := http.FileServer(http.FS(subFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFSHandler))

	// Home page handler
	mux.HandleFunc("GET /web", serveWebPanel)

	// API endpoint for video downloads
	mux.HandleFunc("POST /web/download", handleWebDownload)
}

// serveWebPanel serves the web panel interface
func serveWebPanel(w http.ResponseWriter, r *http.Request) {
	if !IsWebPanelEnabled() {
		http.Error(w, "Web panel is disabled", http.StatusForbidden)
		return
	}

	InitWebPanel()
	err := templates.ExecuteTemplate(w, "webpanel.html", nil)
	if err != nil {
		logger.Error("Error executing template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Response represents the JSON response for the web panel
type Response struct {
	URL   string `json:"url,omitempty"`
	Error string `json:"error,omitempty"`
}

// handleWebDownload processes YouTube download requests from the web interface
func handleWebDownload(w http.ResponseWriter, r *http.Request) {
	if !IsWebPanelEnabled() {
		http.Error(w, "Web panel is disabled", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		renderResult(w, Response{Error: "Failed to parse form data"})
		return
	}

	videoURL := r.FormValue("videoUrl")
	quality := r.FormValue("quality")

	if videoURL == "" {
		renderResult(w, Response{Error: "Video URL is required"})
		return
	}

	if quality == "" {
		quality = "Q720P" // Default quality
	}

	// Process the URL to prevent compatability issues with the existing resonite backend api.
	// For URLs that begin with http:// or https://, remove the protocol completely
	// This avoids issues with the urlhelper function's protocol handling
	if strings.HasPrefix(videoURL, "http://") {
		videoURL = strings.TrimPrefix(videoURL, "http://")
	} else if strings.HasPrefix(videoURL, "https://") {
		videoURL = strings.TrimPrefix(videoURL, "https://")
	}

	// Create a custom request to reuse our existing download logic
	downloadPath := "/reso/" + quality + "/" + videoURL
	req, _ := http.NewRequest("GET", downloadPath, nil)
	req.Host = r.Host

	// Use a ResponseRecorder to capture the output
	rw := newResponseRecorder()

	// Process the download
	GetResoVideos(rw, req)

	// Check if there was an error (prefixed with "error:")
	result := rw.Body.String()
	if len(result) > 6 && result[:6] == "error:" {
		renderResult(w, Response{Error: result[7:]})
	} else {
		renderResult(w, Response{URL: result})
	}
}

// renderResult renders the download result using the template
func renderResult(w http.ResponseWriter, response Response) {
	InitWebPanel()
	err := templates.ExecuteTemplate(w, "result.html", response)
	if err != nil {
		logger.Error("Error executing template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// responseRecorder implements http.ResponseWriter to capture the response
type responseRecorder struct {
	Code   int
	Body   *strings.Builder
	header http.Header
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{
		Code:   200,
		Body:   &strings.Builder{},
		header: make(http.Header),
	}
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.Body.Write(b)
}

func (r *responseRecorder) WriteHeader(code int) {
	r.Code = code
}
