/*
Package main implements a multichecker that includes standard Go static analyzers,
all SA class analyzers from staticcheck.io, at least one analyzer from other classes
of staticcheck.io, two public analyzers, and a custom analyzer
that prohibits direct calls to os.Exit in the main function of the main package.

# Usage

To use this multichecker, run the following command:

	go run cmd/staticlint/main.go ./...

# Analyzers

## Standard Go Analyzers

- asmdecl: Checks assembly declarations.
- assign: Detects useless assignments.
- atomic: Checks for common mistakes using the sync/atomic package.
- atomicalign: Checks for non-64-bits-aligned arguments to sync/atomic functions.
- bools: Detects common mistakes involving booleans.
- buildtag: Ensures correct use of build tags.
- buildssa: Builds SSA-form IR for later passes.
- cgocall: Detects direct calls to C code.
- composites: Checks for unkeyed composite literals.
- copylock: Detects locks passed by value.
- ctrlflow: Builds a control-flow graph.
- deepequalerrors: Checks for calls of reflect.DeepEqual on error values.
- directives: Extracts linter directives.
- errorsas: Reports passing non-pointer or non-error values to errors.As.
- httpresponse: Checks for mistakes using http.Response.
- loopclosure: Detects incorrect uses of loop variables in closures.
- lostcancel: Checks for failure to cancel context.
- nilfunc: Detects useless comparisons of functions to nil.
- printf: Checks for errors in Printf calls.
- shift: Checks for mistaken shifts.
- stdmethods: Detects methods that do not satisfy interface requirements.
- structtag: Ensures struct field tags are well-formed.
- tests: Detects common mistakes in tests.
- unmarshal: Checks for mistakes using encoding/json.Unmarshal.
- unreachable: Detects unreachable code.
- unsafeptr: Checks for invalid uses of unsafe.Pointer.
- unusedresult: Checks for unused results of function calls.

## Staticcheck SA Analyzers

Includes all SA class analyzers from staticcheck.io. For a complete list, visit
https://staticcheck.io/docs/checks#SA.

## Public Analyzers

Includes analyzers from go-critic or other public analyzer packages of your choice.

## Custom Analyzer

- osexitanalyzer: Prohibits direct calls to os.Exit in the main function of the main package.
*/
package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/analysis/facts/directives"
	"honnef.co/go/tools/analysis/facts/nilness"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/v-starostin/go-metrics/internal/analyzer"
)

func main() {
	var analyzers []*analysis.Analyzer

	// Add standard analyzers
	analyzers = append(analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directives.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analysis,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	)

	// Add staticcheck SA analyzers
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Add one analyzer from stylecheck
	for _, v := range stylecheck.Analyzers {
		if !strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
			break
		}
	}

	// Add one analyzer from stylecheck, simple, and quickfix
	analyzers = append(analyzers,
		stylecheck.Analyzers[0].Analyzer,
		simple.Analyzers[0].Analyzer,
		quickfix.Analyzers[0].Analyzer,
	)

	// Add two public analyzers (e.g., from go-critic)
	// Make sure to import the necessary packages for additional analyzers
	// go get github.com/go-critic/go-critic/checkers
	// import "github.com/go-critic/go-critic/checkers/analyzer"

	// analyzers = append(analyzers,
	//     analyzer.Analyzer, // Replace with actual analyzers
	//     anotherAnalyzer.Analyzer,
	// )

	// Add custom analyzer
	analyzers = append(analyzers, analyzer.OSExit)

	multichecker.Main(analyzers...)
}
