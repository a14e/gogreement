package indexing

import (
	"go/types"
	"iter"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/util"
)

// We have to use an ugly solution here. This is because in the analysis framework facts are only passed
// within a single analyzer
// therefore we have to declare a new type for each analyzer. But this is just a named type over PackageAnnotations
// that's why we use generics here
// because of this it's important to ensure that the generic type matches the type used for facts in the specific
// analyzer, otherwise it won't work

// BuildImmutableTypesIndex creates an index of immutable types from current and imported packages
func BuildImmutableTypesIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypesMap {
	result := util.NewTypesMap()

	for pkg, ann := range iterOverPackages[T](pass, packageAnnotations) {
		for _, annot := range ann.ImmutableAnnotations {
			result.Add(pkg.Path(), annot.OnType)
		}
	}

	return result
}

// BuildConstructorIndex creates an index of constructor functions for types
// FIXME do we really need this? there's a strong hypothesis that we only need packageAnnotations of the specific package
// TODO think about whether we can delete this or significantly simplify it
func BuildConstructorIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypeAssociationRegistry {
	result := util.NewTypeAssociationRegistry()

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
func BuildTestOnlyFuncsIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypeAssociationRegistry {
	result := util.NewTypeAssociationRegistry()

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
func BuildTestOnlyMethodsIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypeAssociationRegistry {
	result := util.NewTypeAssociationRegistry()

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

// BuildMutableFieldsIndex creates an index of @mutable fields in @immutable types
// Returns a map: packageName -> typeName -> []fieldNames
func BuildMutableFieldsIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) util.TypeAssociationRegistry {
	result := util.NewTypeAssociationRegistry()

	for pkg, ann := range iterOverPackages[T](pass, packageAnnotations) {
		for _, annot := range ann.MutableAnnotations {
			// Add mutable field to the registry (fieldName, typeName)
			result.Add(pkg.Path(), annot.FieldName, annot.OnType)
		}
	}

	return result
}

// BuildPackageOnlyIndex creates an AttachmentsMap of @packageonly annotations from current and imported packages
func BuildPackageOnlyIndex[T annotations.AnnotationWrapper](pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) *util.AttachmentsMap {
	result := &util.AttachmentsMap{}

	for pkg, ann := range iterOverPackages[T](pass, packageAnnotations) {
		pkgPath := pkg.Path()

		// Process package-level @packageonly annotations (functions, types)
		for _, annot := range ann.PackageOnlyAnnotations {
			switch annot.Kind {
			case annotations.TestOnlyOnType:
				// Add allowed packages directly to type
				for _, allowedPkg := range annot.AllowedPackages {
					result.AddPkgTypeAttachment(pkgPath, annot.ObjectName, allowedPkg)
				}
			case annotations.TestOnlyOnFunc:
				// Add allowed packages directly to function
				for _, allowedPkg := range annot.AllowedPackages {
					result.AddPkgFunctionAttachment(pkgPath, annot.ObjectName, allowedPkg)
				}
			case annotations.TestOnlyOnMethod:
				// Add allowed packages directly to method
				for _, allowedPkg := range annot.AllowedPackages {
					result.AddPkgTypeMethodAttachment(pkgPath, annot.ReceiverType, annot.ObjectName, allowedPkg)
				}
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
					if !yield(imp, fact.GetAnnotations()) {
						return
					}
				}
			}
		}

	}
}
