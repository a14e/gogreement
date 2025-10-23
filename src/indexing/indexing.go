package indexing

import (
	"go/types"
	"iter"

	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
	"goagreement/src/util"
)

// BuildImmutableTypesIndex creates an index of immutable types from current and imported packages
func BuildImmutableTypesIndex[T annotations.AnnotationWrapper](pass *analysis.Pass) util.TypesMap {
	result := util.NewTypesMap()

	for pkg, ann := range iterOverPackages2[T](pass) {
		for _, annot := range ann.ImmutableAnnotations {
			result.Add(pkg.Path(), annot.OnType)
		}
	}

	return result
}

// BuildConstructorIndex creates an index of constructor functions for types
func BuildConstructorIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypeFuncRegistry {
	result := util.NewTypeFuncRegistry()

	for pkg, ann := range iterOverPackages[T](pass, packageAnnotations) {
		for _, annot := range ann.ConstructorAnnotations {
			for _, constructorName := range annot.ConstructorNames {
				result.Add(pkg.Path(), constructorName, annot.OnType)
			}
		}
	}

	return result
}

// BuildTestOnlyTypesIndex creates an index of @testonly types from current and imported packages
func BuildTestOnlyTypesIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypesMap {
	result := util.NewTypesMap()

	for pkg, ann := range iterOverPackages[T](pass, packageAnnotations) {
		for _, annot := range ann.TestonlyAnnotations {
			if annot.Kind == annotations.TestOnlyOnType {
				result.Add(pkg.Path(), annot.ObjectName)
			}
		}
	}

	return result
}

// BuildTestOnlyFuncsIndex creates an index of @testonly functions from current and imported packages
func BuildTestOnlyFuncsIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypeFuncRegistry {
	result := util.NewTypeFuncRegistry()

	for pkg, ann := range iterOverPackages[T](pass, packageAnnotations) {
		for _, annot := range ann.TestonlyAnnotations {
			if annot.Kind == annotations.TestOnlyOnFunc {
				// Store function as funcName -> funcName mapping
				result.Add(pkg.Path(), annot.ObjectName, annot.ObjectName)
			}
		}
	}

	return result
}

// BuildTestOnlyMethodsIndex creates an index of @testonly methods from current and imported packages
func BuildTestOnlyMethodsIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypeFuncRegistry {
	result := util.NewTypeFuncRegistry()

	for pkg, ann := range iterOverPackages[T](pass, packageAnnotations) {
		for _, annot := range ann.TestonlyAnnotations {
			if annot.Kind == annotations.TestOnlyOnMethod {
				// Store method as methodName -> receiverType mapping
				result.Add(pkg.Path(), annot.ObjectName, annot.ReceiverType)
			}
		}
	}

	return result
}

// iterOverPackages just iter over packageAnnotations + facts over imported packages
func iterOverPackages[T annotations.AnnotationWrapper](
	pass *analysis.Pass,
	packageAnnotations *annotations.PackageAnnotations,
) iter.Seq2[*types.Package, *annotations.PackageAnnotations] {

	return func(yield func(*types.Package, *annotations.PackageAnnotations) bool) {
		if pass.Pkg == nil {
			return
		}

		if !yield(pass.Pkg, packageAnnotations) {
			return
		}

		if pass.ImportPackageFact != nil {
			var zero T
			for _, imp := range pass.Pkg.Imports() {
				fact := zero.Empty()
				if pass.ImportPackageFact(imp, fact) {
					yield(imp, fact.GetAnnotations())
				}
			}
		}

	}
}

func iterOverPackages2[T annotations.AnnotationWrapper](
	pass *analysis.Pass,
) iter.Seq2[*types.Package, *annotations.PackageAnnotations] {

	return func(yield func(*types.Package, *annotations.PackageAnnotations) bool) {
		if pass.Pkg == nil {
			return
		}

		var zero T
		fact := zero.Empty()
		if pass.ImportPackageFact(pass.Pkg, fact) {
			if !yield(pass.Pkg, fact.GetAnnotations()) {
				return
			}
		}

		if pass.ImportPackageFact != nil {
			for _, imp := range pass.Pkg.Imports() {
				if pass.ImportPackageFact(imp, fact) {
					if !yield(imp, fact.GetAnnotations()) {
						return
					}
				}
			}
		}

	}
}
