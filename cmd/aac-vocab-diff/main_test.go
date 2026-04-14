package main

import "testing"

// ── Diff mode ─────────────────────────────────────────────────────────────────

func TestParseArgs_DiffMode(t *testing.T) {
	mode, opts, err := parseArgs([]string{"old.ce", "new.ce"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != "diff" {
		t.Errorf("mode: got %q, want \"diff\"", mode)
	}
	if opts.oldCE != "old.ce" || opts.newCE != "new.ce" {
		t.Errorf("paths: got %q %q, want \"old.ce\" \"new.ce\"", opts.oldCE, opts.newCE)
	}
	if opts.reportPath != "" {
		t.Errorf("reportPath should be empty by default, got %q", opts.reportPath)
	}
}

func TestParseArgs_DiffModeWithReport_FlagAfterFiles(t *testing.T) {
	_, opts, err := parseArgs([]string{"old.ce", "new.ce", "--report", "out.html"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.reportPath != "out.html" {
		t.Errorf("reportPath: got %q, want \"out.html\"", opts.reportPath)
	}
}

func TestParseArgs_DiffModeWithReport_FlagBeforeFiles(t *testing.T) {
	_, opts, err := parseArgs([]string{"--report", "out.html", "old.ce", "new.ce"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.reportPath != "out.html" {
		t.Errorf("reportPath: got %q, want \"out.html\"", opts.reportPath)
	}
}

func TestParseArgs_DiffModeWithReport_EqualSyntax(t *testing.T) {
	_, opts, err := parseArgs([]string{"old.ce", "--report=out.html", "new.ce"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.reportPath != "out.html" {
		t.Errorf("reportPath: got %q, want \"out.html\"", opts.reportPath)
	}
}

func TestParseArgs_NoArgs_ReturnsError(t *testing.T) {
	_, _, err := parseArgs([]string{})
	if err == nil {
		t.Error("expected error for no arguments, got nil")
	}
}

func TestParseArgs_OneFileArg_ReturnsError(t *testing.T) {
	_, _, err := parseArgs([]string{"only.ce"})
	if err == nil {
		t.Error("expected error for single file argument, got nil")
	}
}

func TestParseArgs_MissingReportValue_ReturnsError(t *testing.T) {
	_, _, err := parseArgs([]string{"old.ce", "new.ce", "--report"})
	if err == nil {
		t.Error("expected error when --report has no value, got nil")
	}
}

// ── Serve mode ────────────────────────────────────────────────────────────────

func TestParseArgs_ServeMode(t *testing.T) {
	mode, opts, err := parseArgs([]string{"serve", "--addr", ":9090"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != "serve" {
		t.Errorf("mode: got %q, want \"serve\"", mode)
	}
	if opts.addr != ":9090" {
		t.Errorf("addr: got %q, want \":9090\"", opts.addr)
	}
}

func TestParseArgs_ServeModeDefaultAddr(t *testing.T) {
	_, opts, err := parseArgs([]string{"serve"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.addr != ":8080" {
		t.Errorf("default addr: got %q, want \":8080\"", opts.addr)
	}
}
