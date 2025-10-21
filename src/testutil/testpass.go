package testutil

import (
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

var (
	testPackageCache     = make(map[string]*analysis.Pass)
	testPackageCacheLock sync.RWMutex
)

func getCachedPass(pkgName string) *analysis.Pass {
	testPackageCacheLock.RLock()
	defer testPackageCacheLock.RUnlock()
	return testPackageCache[pkgName]
}

func setCachedPass(pkgName string, pass *analysis.Pass) {
	testPackageCacheLock.Lock()
	testPackageCache[pkgName] = pass
	testPackageCacheLock.Unlock()
}

// getTestdataPath returns absolute path to testdata directory
func getTestdataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "testdata")
}

// createTestPass creates a minimal analysis.Pass for testing
func CreateTestPass(t *testing.T, pkgName string) *analysis.Pass {
	if cached := getCachedPass(pkgName); cached != nil {
		t.Logf("Using cached package: %s", pkgName)
		return cached
	}

	testdataPath := getTestdataPath()

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes |
			packages.NeedImports | packages.NeedDeps | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir: testdataPath,
	}

	pattern := "./" + pkgName
	pkgs, err := packages.Load(cfg, pattern)
	require.NoError(t, err, "failed to load package")
	require.NotEmpty(t, pkgs, "no packages loaded")

	if len(pkgs[0].Errors) > 0 {
		for _, e := range pkgs[0].Errors {
			t.Logf("Error: %v", e)
		}
	}
	require.Empty(t, pkgs[0].Errors, "package has errors")

	// Debug: print imports
	t.Logf("Package: %s", pkgs[0].Types.Path())
	t.Logf("Imports count: %d", len(pkgs[0].Types.Imports()))
	for _, imp := range pkgs[0].Types.Imports() {
		t.Logf("  Import: %s (name: %s)", imp.Path(), imp.Name())
	}

	// Create analysis.Pass
	pass := &analysis.Pass{
		Pkg:       pkgs[0].Types,
		Files:     pkgs[0].Syntax,
		TypesInfo: pkgs[0].TypesInfo,
	}

	// Cache the result
	setCachedPass(pkgName, pass)

	return pass
}
