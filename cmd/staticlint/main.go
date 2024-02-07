// Локальный multichecker, состоящий из стандартных статических анализаторов пакета
// golang.org/x/tools/go/analysis/passes, всех анализаторов класса SA пакета staticcheck.io,
// двух анализаторов класса S пакета staticcheck.io, двух анализаторов класса ST пакета staticcheck.io,
// двух анализаторов класса QF пакета staticcheck.io, а также кастомного анализатора проверки
// наличия os.Exit() в пакете main в ф-ции main(). Должен был бы также состоять из двух или более любых
// публичных анализаторов на мой выбор, но я таковых не нашёл.
//
// Запуск в командной строке:
//
//	mycheck <директория, содержащая пакеты .go>
package main

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
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
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
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
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {

	var mychecks []*analysis.Analyzer

	// Package staticcheck contains analyzes that find bugs and performance issues.
	// Barring the rare false positive, any code flagged by these analyzes needs to be fixed.
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	// Package stylecheck contains analyzes that enforce style rules.
	// Most of the recommendations made are universally agreed upon by
	// the wider Go community. Some analyzes, however, implement stricter
	// rules that not everyone will agree with. In the context of Staticcheck,
	// these analyzes are not enabled by default.
	mychecks = append(mychecks, stylecheck.Analyzers[0].Analyzer)
	mychecks = append(mychecks, stylecheck.Analyzers[1].Analyzer)

	// Package simple contains analyzes that simplify code.
	// All suggestions made by these analyzes are intended to result in objectively simpler code,
	// and following their advice is recommended.
	mychecks = append(mychecks, simple.Analyzers[0].Analyzer)
	mychecks = append(mychecks, simple.Analyzers[1].Analyzer)

	// Package quickfix contains analyzes that implement code refactorings.
	// None of these analyzers produce diagnostics that have to be followed.
	// Most of the time, they only provide alternative ways of doing things,
	// requiring users to make informed decisions.
	mychecks = append(mychecks, quickfix.Analyzers[0].Analyzer)
	mychecks = append(mychecks, quickfix.Analyzers[1].Analyzer)

	// asmdecl defines an Analyzer that reports mismatches between assembly files and Go declarations.
	mychecks = append(mychecks, asmdecl.Analyzer)

	// appends defines an Analyzer that detects if there is only one variable in append.
	mychecks = append(mychecks, appends.Analyzer)

	// assign defines an Analyzer that detects useless assignments.
	mychecks = append(mychecks, assign.Analyzer)

	// atomic defines an Analyzer that checks for common mistakes using the sync/atomic package.
	mychecks = append(mychecks, atomic.Analyzer)

	// atomicalign defines an Analyzer that checks for non-64-bit-aligned arguments to
	// sync/atomic functions. On non-32-bit platforms, those functions panic if their
	// argument variables are not 64-bit aligned. It is therefore the caller's responsibility
	// to arrange for 64-bit alignment of such variables.
	mychecks = append(mychecks, atomicalign.Analyzer)

	// bools defines an Analyzer that detects common mistakes involving boolean operators.
	mychecks = append(mychecks, bools.Analyzer)

	// buildssa defines an Analyzer that constructs the SSA representation of an error-free
	// package and returns the set of all functions within it. It does not report any
	// diagnostics itself but may be used as an input to other analyzers
	mychecks = append(mychecks, buildssa.Analyzer)

	// buildtag defines an Analyzer that checks build tags.
	mychecks = append(mychecks, buildtag.Analyzer)

	// cgocall defines an Analyzer that detects some violations of
	// the cgo pointer passing rules.
	mychecks = append(mychecks, cgocall.Analyzer)

	// composite defines an Analyzer that checks for unkeyed
	// composite literals.
	mychecks = append(mychecks, composite.Analyzer)

	// copylock defines an Analyzer that checks for locks
	// erroneously passed by value.
	mychecks = append(mychecks, copylock.Analyzer)

	// ctrlflow is an analysis that provides a syntactic
	// control-flow graph (CFG) for the body of a function.
	// It records whether a function cannot return.
	// By itself, it does not report any diagnostics.
	mychecks = append(mychecks, ctrlflow.Analyzer)

	// deepequalerrors defines an Analyzer that checks for the use
	// of reflect.DeepEqual with error values.
	mychecks = append(mychecks, deepequalerrors.Analyzer)

	// defers defines an Analyzer that checks for common mistakes in defer
	// statements.
	mychecks = append(mychecks, defers.Analyzer)

	// directive defines an Analyzer that checks known Go toolchain directives.
	mychecks = append(mychecks, directive.Analyzer)

	// errorsas package defines an Analyzer that checks that the second argument to
	// errors.As is a pointer to a type implementing error.
	mychecks = append(mychecks, errorsas.Analyzer)

	// fieldalignment defines an Analyzer that detects structs that would use less
	// memory if their fields were sorted.
	mychecks = append(mychecks, fieldalignment.Analyzer)

	// findcall defines an Analyzer that serves as a trivial
	// example and test of the Analysis API. It reports a diagnostic for
	// every call to a function or method of the name specified by its
	// -name flag. It also exports a fact for each declaration that
	// matches the name, plus a package-level fact if the package contained
	// one or more such declarations.
	mychecks = append(mychecks, findcall.Analyzer)

	// framepointer defines an Analyzer that reports assembly code
	// that clobbers the frame pointer before saving it.
	mychecks = append(mychecks, framepointer.Analyzer)

	// httpmux analysis is active for Go modules configured to run with Go 1.21 or
	// earlier versions.
	mychecks = append(mychecks, httpmux.Analyzer)

	// httpresponse defines an Analyzer that checks for mistakes
	// using HTTP responses.
	mychecks = append(mychecks, httpresponse.Analyzer)

	// ifaceassert defines an Analyzer that flags impossible interface-interface type assertions.
	mychecks = append(mychecks, ifaceassert.Analyzer)

	// inspect defines an Analyzer that provides an AST inspector for the syntax
	// trees of a package. It is only a building block for other analyzers.
	mychecks = append(mychecks, inspect.Analyzer)

	// loopclosure defines an Analyzer that checks for references to
	// enclosing loop variables from within nested functions.
	mychecks = append(mychecks, loopclosure.Analyzer)

	// lostcancel defines an Analyzer that checks for failure to
	// call a context cancellation function.
	mychecks = append(mychecks, lostcancel.Analyzer)

	// nilfunc defines an Analyzer that checks for useless
	// comparisons against nil.
	mychecks = append(mychecks, nilfunc.Analyzer)

	// nilness inspects the control-flow graph of an SSA function
	// and reports errors such as nil pointer dereferences and degenerate
	// nil pointer comparisons.
	mychecks = append(mychecks, nilness.Analyzer)

	// The pkgfact is a demonstration and test of the package fact
	// mechanism.
	//
	// The output of the pkgfact analysis is a set of key/values pairs
	// gathered from the analyzed package and its imported dependencies.
	// Each key/value pair comes from a top-level constant declaration
	// whose name starts and ends with "_".
	mychecks = append(mychecks, pkgfact.Analyzer)

	// printf defines an Analyzer that checks consistency
	// of Printf format strings and arguments.
	mychecks = append(mychecks, printf.Analyzer)

	// reflectvaluecompare defines an Analyzer that checks for accidentally
	// using == or reflect.DeepEqual to compare reflect.Value values.
	mychecks = append(mychecks, reflectvaluecompare.Analyzer)

	// shadow defines an Analyzer that checks for shadowed variables.
	mychecks = append(mychecks, shadow.Analyzer)

	// shift defines an Analyzer that checks for shifts that exceed
	// the width of an integer.
	mychecks = append(mychecks, shift.Analyzer)

	// sigchanyzer defines an Analyzer that detects
	// misuse of unbuffered signal as argument to signal.Notify.
	mychecks = append(mychecks, sigchanyzer.Analyzer)

	// slog defines an Analyzer that checks for
	// mismatched key-value pairs in log/slog calls.
	mychecks = append(mychecks, slog.Analyzer)

	// sortslice defines an Analyzer that checks for calls
	// to sort.Slice that do not use a slice type as first argument.
	mychecks = append(mychecks, sortslice.Analyzer)

	// stdmethods defines an Analyzer that checks for misspellings
	// in the signatures of methods similar to well-known interfaces.
	mychecks = append(mychecks, stdmethods.Analyzer)

	// stringintconv defines an Analyzer that flags type conversions
	// from integers to strings.
	mychecks = append(mychecks, stringintconv.Analyzer)

	// structtag defines an Analyzer that checks struct field tags
	// are well formed.
	mychecks = append(mychecks, structtag.Analyzer)

	// testinggoroutine defines an Analyzerfor detecting calls to
	// Fatal from a test goroutine.
	mychecks = append(mychecks, testinggoroutine.Analyzer)

	// tests defines an Analyzer that checks for common mistaken
	// usages of tests and examples.
	mychecks = append(mychecks, tests.Analyzer)

	// timeformat defines an Analyzer that checks for the use
	// of time.Format or time.Parse calls with a bad format.
	mychecks = append(mychecks, timeformat.Analyzer)

	// The unmarshal package defines an Analyzer that checks for passing
	// non-pointer or non-interface types to unmarshal and decode functions.
	mychecks = append(mychecks, unmarshal.Analyzer)

	// unreachable defines an Analyzer that checks for unreachable code.
	mychecks = append(mychecks, unreachable.Analyzer)

	// unsafeptr defines an Analyzer that checks for invalid
	// conversions of uintptr to unsafe.Pointer.
	mychecks = append(mychecks, unsafeptr.Analyzer)

	// unusedresult defines an analyzer that checks for unused
	// results of calls to certain pure functions.
	mychecks = append(mychecks, unusedresult.Analyzer)

	// unusedwrite checks for unused writes to the elements of a struct or array object.
	mychecks = append(mychecks, unusedwrite.Analyzer)

	// usesgenerics defines an Analyzer that checks for usage of generic
	// features added in Go 1.18.
	mychecks = append(mychecks, usesgenerics.Analyzer)

	var a = &analysis.Analyzer{
		Name: "checkOsExitInMain",
		Doc:  "Checks os.Exit func allocation in package main in func main()",
		Run:  checkOsExitInMain,
	}
	mychecks = append(mychecks, a)

	multichecker.Main(mychecks...)
}

func checkOsExitInMain(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	currentFuncName := ""

	isOsExitInsideMain := func(x *ast.ExprStmt) {
		if call, ok := x.X.(*ast.CallExpr); ok {
			s := fmt.Sprint(call.Fun)
			if s == "&{os Exit}" {
				pass.Reportf(x.Pos(), "os.Exit in main() function")
			}
		}
	}

	fude := func(x *ast.FuncDecl) {
		currentFuncName = x.Name.Name
	}

	for _, file := range pass.Files {
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.ExprStmt: // выражение
				if currentFuncName == "main" {
					isOsExitInsideMain(x)
				}
			case *ast.FuncDecl: // декларация ф-ции
				fude(x)
			}
			return true
		})
	}
	return nil, nil
}
