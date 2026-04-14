package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/izabelacg/aac-vocab-diff/diff"
	"github.com/izabelacg/aac-vocab-diff/report"
	"github.com/izabelacg/aac-vocab-diff/server"
)

type cliOpts struct {
	oldCE, newCE string
	reportPath   string
	addr         string
	analyticsLog string
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  aac-vocab-diff <old.ce> <new.ce> [--report out.html]")
	fmt.Fprintln(os.Stderr, "  aac-vocab-diff serve [--addr :8080]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Flags:")
	fmt.Fprintln(os.Stderr, "  --report <file>        write HTML report to this file (diff mode only)")
	fmt.Fprintln(os.Stderr, "  --addr <addr>          address to listen on (serve mode, default :8080)")
	fmt.Fprintln(os.Stderr, "  --analytics-log <file> append event log to this file (serve mode only)")
}

func main() {
	mode, opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		usage()
		os.Exit(1)
	}

	switch mode {
	case "diff":
		runDiffMode(opts)
	case "serve":
		runServeMode(opts)
	}
}

func runDiffMode(opts cliOpts) {
	d, err := diff.CompareFiles(opts.oldCE, opts.newCE)
	if err != nil {
		log.Fatal(err)
	}
	report.PrintDiff(d)

	if opts.reportPath != "" {
		f, err := os.Create(opts.reportPath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := report.WriteHTML(f, report.NewHTMLData(d)); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stderr, "HTML report written to %s\n", opts.reportPath)
	}
}

func runServeMode(opts cliOpts) {
	fmt.Fprintf(os.Stderr, "Listening on http://localhost%s\n", opts.addr)
	var serverOpts []server.Option
	if opts.analyticsLog != "" {
		serverOpts = append(serverOpts, server.WithAnalyticsLog(opts.analyticsLog))
	}
	if err := server.ListenAndServe(opts.addr, serverOpts...); err != nil {
		log.Fatal(err)
	}
}

// parseArgs handles subcommand dispatch and flag parsing.
// args should be os.Args[1:].
func parseArgs(args []string) (string, cliOpts, error) {
	if len(args) == 0 {
		return "", cliOpts{}, errors.New("no arguments provided")
	}

	if args[0] == "serve" {
		fs := flag.NewFlagSet("serve", flag.ContinueOnError)
		addr := fs.String("addr", ":8080", "address to listen on")
		analyticsLog := fs.String("analytics-log", "", "append event log to this file")
		if err := fs.Parse(args[1:]); err != nil {
			return "", cliOpts{}, err
		}
		return "serve", cliOpts{addr: *addr, analyticsLog: *analyticsLog}, nil
	}

	// Diff mode: accept --report flag anywhere in the argument list so that
	//   aac-vocab-diff old.ce new.ce --report out.html
	// and
	//   aac-vocab-diff --report out.html old.ce new.ce
	// both work.
	var reportPath string
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--report" || arg == "-report":
			if i+1 >= len(args) {
				return "", cliOpts{}, errors.New("--report requires a file path argument")
			}
			reportPath = args[i+1]
			i++
		case strings.HasPrefix(arg, "--report="):
			reportPath = strings.TrimPrefix(arg, "--report=")
		case strings.HasPrefix(arg, "-report="):
			reportPath = strings.TrimPrefix(arg, "-report=")
		default:
			positional = append(positional, arg)
		}
	}

	if len(positional) != 2 {
		return "", cliOpts{}, fmt.Errorf("expected 2 file arguments, got %d", len(positional))
	}
	return "diff", cliOpts{oldCE: positional[0], newCE: positional[1], reportPath: reportPath}, nil
}
