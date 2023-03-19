package main

import (
	"strings"

	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"

	"github.com/breml/errchkjson"
	"github.com/charithe/durationcheck"
)

func main() {

	mychecks := []*analysis.Analyzer{
		bools.Analyzer,
		deepequalerrors.Analyzer,
		loopclosure.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		sigchanyzer.Analyzer,
		unsafeptr.Analyzer,
		quickfix.Analyzers[0].Analyzer,
		stylecheck.Analyzers[0].Analyzer,
		durationcheck.Analyzer,
		errchkjson.NewAnalyzer(),
		MainOsExitAnalyzer,
	}

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(mychecks...)
}
