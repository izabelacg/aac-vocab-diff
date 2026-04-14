package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/izabelacg/aac-vocab-diff/server"
)

// Fixture .ce files shared with the diff package tests.
const (
	fixtureOld = "../diff/testdata/WordPower60 Basic SS_2026-04-08.ce"
	fixtureNew = "../diff/testdata/WordPower60 Basic SS_Custom_2026-04-08.ce"
)

// ── Analytics log file ────────────────────────────────────────────────────────

func TestWithAnalyticsLog_PageViewWrittenToFile(t *testing.T) {
	f := tempLogFile(t)
	srv := server.New(server.WithAnalyticsLog(f))

	w, r := get("/")
	srv.ServeHTTP(w, r)

	assertLogContains(t, f, "page_view")
}

func TestWithAnalyticsLog_DiffOKWrittenToFile(t *testing.T) {
	f := tempLogFile(t)
	srv := server.New(server.WithAnalyticsLog(f))

	body, ct := buildMultipart(t, fixtureOld, fixtureNew)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	srv.ServeHTTP(w, r)

	assertLogContains(t, f, "diff_ok")
}

func TestWithAnalyticsLog_DiffErrWrittenToFile(t *testing.T) {
	f := tempLogFile(t)
	srv := server.New(server.WithAnalyticsLog(f))

	fakeZIP := []byte("PK\x03\x04 not a real zip")
	body, ct := buildMultipartFromBytes(t, fakeZIP, fakeZIP)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	srv.ServeHTTP(w, r)

	assertLogContains(t, f, "diff_err")
}

func TestWithAnalyticsLog_NoOptionNoFile(t *testing.T) {
	// Server without analytics log must not panic on any handler.
	srv := server.New()
	w, r := get("/")
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

// tempLogFile creates a temp file, registers cleanup, and returns its path.
func tempLogFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "analytics-*.log")
	if err != nil {
		t.Fatalf("create temp log file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

// assertLogContains reads path and fails if want is not found.
func assertLogContains(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), want) {
		t.Errorf("analytics log missing %q, got:\n%s", want, data)
	}
}

// ── GET /metrics ──────────────────────────────────────────────────────────────

func TestMetrics_Returns200(t *testing.T) {
	w, r := get("/metrics")
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestMetrics_ContentTypeIsText(t *testing.T) {
	w, r := get("/metrics")
	server.New().ServeHTTP(w, r)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type: got %q, want text/plain prefix", ct)
	}
}

func TestMetrics_InitialCountsAreZero(t *testing.T) {
	w, r := get("/metrics")
	server.New().ServeHTTP(w, r)

	body := w.Body.String()
	for _, want := range []string{"page_views 0", "diffs_ok   0", "diffs_err  0"} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics body missing %q, got:\n%s", want, body)
		}
	}
}

func TestMetrics_PageViewsIncrementOnRootGet(t *testing.T) {
	srv := server.New()

	// Two successful GET / requests.
	for range 2 {
		w, r := get("/")
		srv.ServeHTTP(w, r)
	}

	w, r := get("/metrics")
	srv.ServeHTTP(w, r)

	if !strings.Contains(w.Body.String(), "page_views 2") {
		t.Errorf("expected page_views 2, got:\n%s", w.Body.String())
	}
}

func TestMetrics_DiffsOKIncrementOnSuccess(t *testing.T) {
	srv := server.New()

	body, ct := buildMultipart(t, fixtureOld, fixtureNew)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	srv.ServeHTTP(w, r)

	mw, mr := get("/metrics")
	srv.ServeHTTP(mw, mr)

	if !strings.Contains(mw.Body.String(), "diffs_ok   1") {
		t.Errorf("expected diffs_ok 1 after successful diff, got:\n%s", mw.Body.String())
	}
}

func TestMetrics_DiffsErrIncrementOnFailure(t *testing.T) {
	srv := server.New()

	fakeZIP := []byte("PK\x03\x04 not a real zip")
	body, ct := buildMultipartFromBytes(t, fakeZIP, fakeZIP)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	srv.ServeHTTP(w, r)

	mw, mr := get("/metrics")
	srv.ServeHTTP(mw, mr)

	if !strings.Contains(mw.Body.String(), "diffs_err  1") {
		t.Errorf("expected diffs_err 1 after failed diff, got:\n%s", mw.Body.String())
	}
}

// ── GET /health ───────────────────────────────────────────────────────────────

func TestHealth_Returns200(t *testing.T) {
	w, r := get("/health")
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestHealth_ContentTypeIsJSON(t *testing.T) {
	w, r := get("/health")
	server.New().ServeHTTP(w, r)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type: got %q, want application/json prefix", ct)
	}
}

func TestHealth_JSONContainsBuildMetadata(t *testing.T) {
	w, r := get("/health")
	server.New().ServeHTTP(w, r)

	var payload struct {
		Status    string `json:"status"`
		Version   string `json:"version"`
		Commit    string `json:"commit"`
		BuildTime string `json:"buildTime"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal health json: %v", err)
	}
	if payload.Status != "ok" {
		t.Errorf("status: got %q, want ok", payload.Status)
	}
	if payload.Version == "" {
		t.Error("version: empty")
	}
	if payload.Commit == "" {
		t.Error("commit: empty")
	}
	if payload.BuildTime == "" {
		t.Error("buildTime: empty")
	}
}

// ── GET / — upload form ───────────────────────────────────────────────────────

func TestServeUploadForm_Returns200(t *testing.T) {
	w, r := get("/")
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestServeUploadForm_ContentTypeIsHTML(t *testing.T) {
	w, r := get("/")
	server.New().ServeHTTP(w, r)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type: got %q, want text/html prefix", ct)
	}
}

func TestServeUploadForm_ContainsFileInputs(t *testing.T) {
	w, r := get("/")
	server.New().ServeHTTP(w, r)

	body := w.Body.String()
	for _, want := range []string{`name="old"`, `name="new"`, `action="/diff"`} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}

func TestServeUploadForm_MethodNotAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want 405", w.Code)
	}
}

// ── GET / — how-it-works + footer ────────────────────────────────────────────

func TestServeUploadForm_ContainsHowItWorksSection(t *testing.T) {
	w, r := get("/")
	server.New().ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, "What does this compare?") {
		t.Error("upload form missing 'how it works' details section")
	}
}

func TestServeUploadForm_ContainsFeedbackFormLink(t *testing.T) {
	w, r := get("/")
	server.New().ServeHTTP(w, r)

	const formPath = "docs.google.com/forms/d/11YlQYQ_YzGSYQaFJxZFT8TbPX6vj2AvW1ow5ai4CDOU"
	if !strings.Contains(w.Body.String(), formPath) {
		t.Error("upload form missing Google feedback form link")
	}
}

func TestServeUploadForm_ContainsFooter(t *testing.T) {
	w, r := get("/")
	server.New().ServeHTTP(w, r)

	if !strings.Contains(w.Body.String(), "PRC-Saltillo") {
		t.Error("upload form missing footer with PRC-Saltillo disclaimer")
	}
}

func TestServeUploadForm_ContainsBuildInfo(t *testing.T) {
	w, r := get("/")
	server.New().ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, `class="build-info"`) {
		t.Error("upload form missing build-info line")
	}
	if !strings.Contains(body, "dev") || !strings.Contains(body, "unknown") {
		t.Error("expected default build metadata (dev, unknown) in footer")
	}
}

func TestServeUploadForm_ContainsLoadingIndicator(t *testing.T) {
	w, r := get("/")
	server.New().ServeHTTP(w, r)

	body := w.Body.String()
	for _, want := range []string{
		`@keyframes spin`,              // CSS animation driving the spinner
		`.spinner {`,                   // spinner CSS class
		`id="submit-btn"`,              // button target referenced by the script
		`btn.classList.add('loading')`, // JS marks the button as loading on submit
		`Comparing`,                    // loading label shown to the user
		`requestAnimationFrame`,        // deferred submit so Safari paints the spinner
	} {
		if !strings.Contains(body, want) {
			t.Errorf("upload form missing loading indicator element %q", want)
		}
	}
}

func TestServeDiff_ReportContainsFeedbackLink(t *testing.T) {
	body, ct := buildMultipart(t, fixtureOld, fixtureNew)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	const formPath = "docs.google.com/forms/d/11YlQYQ_YzGSYQaFJxZFT8TbPX6vj2AvW1ow5ai4CDOU"
	if !strings.Contains(w.Body.String(), formPath) {
		t.Error("diff report missing feedback form link")
	}
}

// ── POST /diff — rate limiting ────────────────────────────────────────────────

func TestServeDiff_RateLimitExceeded_Returns429(t *testing.T) {
	srv := server.New()
	const ip = "10.0.0.1:9999"

	// Exhaust the burst (3 requests) — they fail for other reasons but not 429.
	for i := range 3 {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/diff", nil)
		r.RemoteAddr = ip
		srv.ServeHTTP(w, r)
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d should not be rate-limited yet", i+1)
		}
	}

	// 4th request from the same IP must be rejected with 429.
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", nil)
	r.RemoteAddr = ip
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th request: got %d, want 429", w.Code)
	}
}

// ── POST /diff — upload size limit ────────────────────────────────────────────

func TestServeDiff_OversizedBody_Returns413(t *testing.T) {
	// Build a multipart body whose total size exceeds 50 MB.
	// We do this by writing a large "file" directly into the multipart writer
	// without opening a real file — this keeps the test fast and self-contained.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("old", "big.ce")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	// Write 81 MB of zeros — well over the 80 MB cap.
	zeros := make([]byte, 81<<20)
	if _, err := part.Write(zeros); err != nil {
		t.Fatalf("write zeros: %v", err)
	}
	mw.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status: got %d, want 413", w.Code)
	}
}

// ── POST /diff — sanitised errors ────────────────────────────────────────────

func TestServeDiff_InvalidFiles_Returns500WithGenericMessage(t *testing.T) {
	// Files start with PK so they pass the ZIP check, but are otherwise invalid
	// and will cause diff.CompareFiles to fail → 500.
	fakeZIP := []byte("PK\x03\x04 this is not a real zip archive")
	body, ct := buildMultipartFromBytes(t, fakeZIP, fakeZIP)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want 500", w.Code)
	}
	got := w.Body.String()
	if !strings.Contains(got, "Something went wrong") {
		t.Errorf("response body missing generic error message, got: %q", got)
	}
}

func TestServeDiff_InvalidFiles_DoesNotLeakInternalError(t *testing.T) {
	fakeZIP := []byte("PK\x03\x04 this is not a real zip archive")
	body, ct := buildMultipartFromBytes(t, fakeZIP, fakeZIP)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	got := w.Body.String()
	// Internal library names / paths must not reach the client.
	for _, leak := range []string{"zip", "archive", "ExtractC4V", "sql", "/tmp/"} {
		if strings.Contains(strings.ToLower(got), strings.ToLower(leak)) {
			t.Errorf("response body leaks internal detail %q: %s", leak, got)
		}
	}
}

// ── POST /diff — ZIP validation ───────────────────────────────────────────────

func TestServeDiff_NonZIPOldFile_Returns400(t *testing.T) {
	fakeZIP := []byte("PK\x03\x04 valid zip")
	body, ct := buildMultipartFromBytes(t, []byte("not a zip"), fakeZIP)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "does not look like") {
		t.Errorf("response body missing friendly message, got: %q", w.Body.String())
	}
}

func TestServeDiff_NonZIPNewFile_Returns400(t *testing.T) {
	fakeZIP := []byte("PK\x03\x04 valid zip")
	body, ct := buildMultipartFromBytes(t, fakeZIP, []byte("not a zip"))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "does not look like") {
		t.Errorf("response body missing friendly message, got: %q", w.Body.String())
	}
}

// ── POST /diff ────────────────────────────────────────────────────────────────

func TestServeDiff_MethodNotAllowed(t *testing.T) {
	w, r := get("/diff")
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want 405", w.Code)
	}
}

func TestServeDiff_ReturnsDiffReport(t *testing.T) {
	body, ct := buildMultipart(t, fixtureOld, fixtureNew)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200\nbody: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Vocab Diff") {
		t.Error("response body does not look like an HTML report")
	}
}

func TestServeDiff_ReportContainsBackLink(t *testing.T) {
	body, ct := buildMultipart(t, fixtureOld, fixtureNew)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	if !strings.Contains(w.Body.String(), `href="/"`) {
		t.Error("diff report missing back link to upload form")
	}
}

func TestServeDiff_ReportUsesOriginalFilenames(t *testing.T) {
	body, ct := buildMultipart(t, fixtureOld, fixtureNew)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	html := w.Body.String()
	if strings.Contains(html, "/tmp/") || strings.Contains(html, "vocab-upload-") {
		t.Error("report contains temp file paths instead of original filenames")
	}
}

func TestServeDiff_MissingNewFile_ReturnsBadRequest(t *testing.T) {
	body, ct := buildMultipartOne(t, "old", fixtureOld)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestServeDiff_MissingOldFile_ReturnsBadRequest(t *testing.T) {
	body, ct := buildMultipartOne(t, "new", fixtureNew)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/diff", body)
	r.Header.Set("Content-Type", ct)
	server.New().ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func get(path string) (*httptest.ResponseRecorder, *http.Request) {
	return httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil)
}

// buildMultipart encodes two files as a multipart/form-data body.
func buildMultipart(t *testing.T, oldPath, newPath string) (io.Reader, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	addFilePart(t, mw, "old", oldPath)
	addFilePart(t, mw, "new", newPath)
	mw.Close()
	return &buf, mw.FormDataContentType()
}

// buildMultipartOne encodes a single file — simulates a missing field.
func buildMultipartOne(t *testing.T, field, path string) (io.Reader, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	addFilePart(t, mw, field, path)
	mw.Close()
	return &buf, mw.FormDataContentType()
}

// buildMultipartFromBytes builds a two-file multipart body from raw byte slices.
// Used to inject invalid (non-ZIP) content without needing real fixture files.
func buildMultipartFromBytes(t *testing.T, oldData, newData []byte) (io.Reader, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for _, f := range []struct {
		field string
		data  []byte
	}{
		{"old", oldData},
		{"new", newData},
	} {
		part, err := mw.CreateFormFile(f.field, f.field+".ce")
		if err != nil {
			t.Fatalf("create form file %q: %v", f.field, err)
		}
		if _, err := part.Write(f.data); err != nil {
			t.Fatalf("write %q data: %v", f.field, err)
		}
	}
	mw.Close()
	return &buf, mw.FormDataContentType()
}

func addFilePart(t *testing.T, mw *multipart.Writer, field, filePath string) {
	t.Helper()
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("open fixture %q: %v", filePath, err)
	}
	defer f.Close()
	part, err := mw.CreateFormFile(field, f.Name())
	if err != nil {
		t.Fatalf("create form file part: %v", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		t.Fatalf("copy fixture data: %v", err)
	}
}
