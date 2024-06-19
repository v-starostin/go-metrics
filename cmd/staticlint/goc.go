/*
Package main implements a multichecker that includes standard Go static analyzers,
all SA class analyzers from staticcheck.io, one analyzer from other classes
of staticcheck.io, a public analyzer, and a custom analyzer.

To use this multichecker, run the following command:

	go run cmd/staticlint/main.go ./...

# Analyzers

# Standard Go Analyzers

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

# Staticcheck SA Analyzers

Includes all SA class analyzers from staticcheck.io. For a complete list, visit
https://staticcheck.io/docs/checks#SA.

# Public Analyzer

Includes an analyzer from the go-critic public analyzer package.

# Custom Analyzer

  - osexitanalyzer: Prohibits direct calls to os.Exit in the main function of the main package.
*/
package main
