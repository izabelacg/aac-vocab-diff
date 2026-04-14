package server

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"

	"github.com/izabelacg/aac-vocab-diff/diff"
	"github.com/izabelacg/aac-vocab-diff/internal/version"
	"github.com/izabelacg/aac-vocab-diff/report"
)

//go:embed templates/upload.html
var uploadFS embed.FS

// Option configures a Server.
type Option func(*Server)

// WithAnalyticsLog enables append-only event logging to path. Each event is
// written as "<RFC3339> <event>\n" with no personal data. If the file cannot
// be opened the error is logged and the server starts without analytics logging.
func WithAnalyticsLog(path string) Option {
	return func(s *Server) {
		el, err := newEventLogger(path)
		if err != nil {
			log.Printf("server: open analytics log %q: %v", path, err)
			return
		}
		s.eventLog = el
	}
}

// Server is an http.Handler with all routes registered.
type Server struct {
	mux      *http.ServeMux
	tmpl     *template.Template
	counters counters     // zero value is ready to use
	eventLog *eventLogger // nil when analytics logging is disabled
}

// New creates a Server with all routes registered. Options are applied in order.
func New(opts ...Option) *Server {
	tmpl, err := template.ParseFS(uploadFS, "templates/upload.html")
	if err != nil {
		// upload.html is compiled into the binary; a parse failure is a programming error.
		panic("server: parse upload template: " + err.Error())
	}
	s := &Server{
		mux:  http.NewServeMux(),
		tmpl: tmpl,
	}
	for _, o := range opts {
		o(s)
	}
	// One diff per 5 seconds per IP, burst of 3. The upload form is cheap
	// (static HTML) so it is not rate-limited.
	lim := newIPLimiter(rate.Every(5*time.Second), 3)
	s.mux.HandleFunc("/", s.serveUploadForm)
	s.mux.Handle("/diff", rateLimitMiddleware(http.HandlerFunc(s.serveDiff), lim))
	s.mux.HandleFunc("/metrics", s.serveMetrics)
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			Status    string `json:"status"`
			Version   string `json:"version"`
			Commit    string `json:"commit"`
			BuildTime string `json:"buildTime"`
		}{
			Status:    "ok",
			Version:   version.Version,
			Commit:    version.Commit,
			BuildTime: version.BuildTime,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("health: encode json: %v", err)
		}
	})
	return s
}

// ListenAndServe starts the HTTP server on addr (e.g. ":8080").
func ListenAndServe(addr string, opts ...Option) error {
	return http.ListenAndServe(addr, New(opts...))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) serveUploadForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.counters.pageViews.Add(1)
	s.eventLog.logEvent("page_view")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.Execute(w, struct {
		Version   string
		Commit    string
		BuildTime string
	}{
		Version:   version.Version,
		Commit:    version.Commit,
		BuildTime: version.BuildTime,
	}); err != nil {
		log.Printf("serveUploadForm: %v", err)
	}
}

func (s *Server) serveDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Cap total upload at 80 MB — .ce files are typically 2–15 MB each.
	r.Body = http.MaxBytesReader(w, r.Body, 80<<20)

	oldPath, oldName, cleanOld, err := saveUploadedFile(r, "old")
	if err != nil {
		s.counters.diffsErr.Add(1)
		s.eventLog.logEvent("diff_err")
		if errors.As(err, new(*http.MaxBytesError)) {
			http.Error(w, "Upload too large. Each file must be under 80 MB.", http.StatusRequestEntityTooLarge)
			return
		}
		if errors.Is(err, errNotZIP) {
			http.Error(w, "The 'old' file does not look like a .ce file. Please check you selected the right file.", http.StatusBadRequest)
			return
		}
		log.Printf("serveDiff: save 'old' file: %v", err)
		http.Error(w, "Please attach a valid 'old' file and try again.", http.StatusBadRequest)
		return
	}
	defer cleanOld()

	newPath, newName, cleanNew, err := saveUploadedFile(r, "new")
	if err != nil {
		s.counters.diffsErr.Add(1)
		s.eventLog.logEvent("diff_err")
		if errors.As(err, new(*http.MaxBytesError)) {
			http.Error(w, "Upload too large. Each file must be under 80 MB.", http.StatusRequestEntityTooLarge)
			return
		}
		if errors.Is(err, errNotZIP) {
			http.Error(w, "The 'new' file does not look like a .ce file. Please check you selected the right file.", http.StatusBadRequest)
			return
		}
		log.Printf("serveDiff: save 'new' file: %v", err)
		http.Error(w, "Please attach a valid 'new' file and try again.", http.StatusBadRequest)
		return
	}
	defer cleanNew()

	d, err := diff.CompareFiles(oldPath, newPath)
	if err != nil {
		s.counters.diffsErr.Add(1)
		s.eventLog.logEvent("diff_err")
		log.Printf("serveDiff: %v", err)
		http.Error(w, "Something went wrong processing the files. Please try again.", http.StatusInternalServerError)
		return
	}

	// Override labels with the user's original filenames, not the temp paths.
	d.OldLabel = oldName
	d.NewLabel = newName

	data := report.NewHTMLData(d)
	data.BackURL = "/"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := report.WriteHTML(w, data); err != nil {
		// Headers are already sent; can only log at this point.
		s.counters.diffsErr.Add(1)
		s.eventLog.logEvent("diff_err")
		log.Printf("serveDiff WriteHTML: %v", err)
		return
	}
	s.counters.diffsOK.Add(1)
	s.eventLog.logEvent("diff_ok")
}

// errNotZIP is returned by saveUploadedFile when the uploaded file does not
// begin with the ZIP magic bytes (PK).
var errNotZIP = errors.New("not a ZIP archive")

// saveUploadedFile reads one multipart field by name, copies it to a temp
// .ce file, and returns (tempPath, originalFilename, cleanupFunc, error).
// The cleanup func removes the temp file; the caller must defer it.
// Returns errNotZIP if the file does not start with the ZIP magic bytes.
func saveUploadedFile(r *http.Request, field string) (path, name string, cleanup func(), err error) {
	f, hdr, err := r.FormFile(field)
	if err != nil {
		return "", "", nil, err
	}
	defer f.Close()

	// Peek at the first 4 bytes to check for the ZIP magic number (PK).
	header := make([]byte, 4)
	n, _ := io.ReadFull(f, header)
	if n < 2 || header[0] != 'P' || header[1] != 'K' {
		return "", "", nil, errNotZIP
	}

	tmp, err := os.CreateTemp("", "vocab-upload-*.ce")
	if err != nil {
		return "", "", nil, err
	}

	// Reconstruct the full stream by prepending the peeked bytes.
	src := io.MultiReader(bytes.NewReader(header[:n]), f)
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", "", nil, err
	}
	tmp.Close()

	tmpName := tmp.Name()
	return tmpName, hdr.Filename, func() { os.Remove(tmpName) }, nil
}
