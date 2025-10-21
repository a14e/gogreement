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
