package indexing

import (
	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
	"goagreement/src/util"
)

// BuildImmutableTypesIndex creates an index of immutable types from current and imported packages
func BuildImmutableTypesIndex(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) util.TypesMap {
	result := util.NewTypesMap()

	if pass.Pkg == nil {
		return result
	}

	for _, annot := range packageAnnotations.ImmutableAnnotations {
		result.Add(pass.Pkg.Path(), annot.OnType)
	}

	if pass.ImportPackageFact != nil {
		for _, imp := range pass.Pkg.Imports() {
			var importedAnnotations annotations.PackageAnnotations
			if pass.ImportPackageFact(imp, &importedAnnotations) {
				for _, annot := range importedAnnotations.ImmutableAnnotations {
					result.Add(imp.Path(), annot.OnType)
				}
			}
		}
	}

	return result
}

// BuildConstructorIndex creates an index of constructor functions for types
func BuildConstructorIndex(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) util.FuncMap {
	result := util.NewFuncMap()

	if pass.Pkg == nil {
		return result
	}

	for _, annot := range packageAnnotations.ConstructorAnnotations {
		for _, constructorName := range annot.ConstructorNames {
			result.Add(pass.Pkg.Path(), constructorName, annot.OnType)
		}
	}

	if pass.ImportPackageFact != nil {
		for _, imp := range pass.Pkg.Imports() {
			var importedAnnotations annotations.PackageAnnotations
			if pass.ImportPackageFact(imp, &importedAnnotations) {
				for _, annot := range importedAnnotations.ConstructorAnnotations {
					for _, constructorName := range annot.ConstructorNames {
						result.Add(imp.Path(), constructorName, annot.OnType)
					}
				}
			}
		}
	}

	return result
}

// BuildTestOnlyTypesIndex creates an index of @testonly types from current and imported packages
func BuildTestOnlyTypesIndex(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) util.TypesMap {
	result := util.NewTypesMap()

	if pass.Pkg == nil {
		return result
	}

	for _, annot := range packageAnnotations.TestonlyAnnotations {
		if annot.Kind == annotations.TestOnlyOnType {
			result.Add(pass.Pkg.Path(), annot.ObjectName)
		}
	}

	if pass.ImportPackageFact != nil {
		for _, imp := range pass.Pkg.Imports() {
			var importedAnnotations annotations.PackageAnnotations
			if pass.ImportPackageFact(imp, &importedAnnotations) {
				for _, annot := range importedAnnotations.TestonlyAnnotations {
					if annot.Kind == annotations.TestOnlyOnType {
						result.Add(imp.Path(), annot.ObjectName)
					}
				}
			}
		}
	}

	return result
}

// BuildTestOnlyFuncsIndex creates an index of @testonly functions from current and imported packages
func BuildTestOnlyFuncsIndex(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) util.FuncMap {
	result := util.NewFuncMap()

	if pass.Pkg == nil {
		return result
	}

	for _, annot := range packageAnnotations.TestonlyAnnotations {
		if annot.Kind == annotations.TestOnlyOnFunc {
			// Store function as funcName -> funcName mapping
			result.Add(pass.Pkg.Path(), annot.ObjectName, annot.ObjectName)
		}
	}

	if pass.ImportPackageFact != nil {
		for _, imp := range pass.Pkg.Imports() {
			var importedAnnotations annotations.PackageAnnotations
			if pass.ImportPackageFact(imp, &importedAnnotations) {
				for _, annot := range importedAnnotations.TestonlyAnnotations {
					if annot.Kind == annotations.TestOnlyOnFunc {
						result.Add(imp.Path(), annot.ObjectName, annot.ObjectName)
					}
				}
			}
		}
	}

	return result
}

// BuildTestOnlyMethodsIndex creates an index of @testonly methods from current and imported packages
func BuildTestOnlyMethodsIndex(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) util.FuncMap {
	result := util.NewFuncMap()

	if pass.Pkg == nil {
		return result
	}

	for _, annot := range packageAnnotations.TestonlyAnnotations {
		if annot.Kind == annotations.TestOnlyOnMethod {
			// Store method as methodName -> receiverType mapping
			result.Add(pass.Pkg.Path(), annot.ObjectName, annot.ReceiverType)
		}
	}

	if pass.ImportPackageFact != nil {
		for _, imp := range pass.Pkg.Imports() {
			var importedAnnotations annotations.PackageAnnotations
			if pass.ImportPackageFact(imp, &importedAnnotations) {
				for _, annot := range importedAnnotations.TestonlyAnnotations {
					if annot.Kind == annotations.TestOnlyOnMethod {
						result.Add(imp.Path(), annot.ObjectName, annot.ReceiverType)
					}
				}
			}
		}
	}

	return result
}
