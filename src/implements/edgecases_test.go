package implements

import (
	"testing"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/testutil"

	"github.com/stretchr/testify/assert"
)

func TestImplementsEdgeCases(t *testing.T) {
	pass := testutil.CreateTestPass(t, "implementsedgecases")
	cfg := config.Empty()
	ann := annotations.ReadAllAnnotations(cfg, pass)

	interfaces := LoadInterfaces(pass, ann.ToInterfaceQuery())
	typeModels := LoadTypes(pass, ann.ToTypeQuery())
	missing := FindMissingMethods(ann.ImplementsAnnotations, interfaces, typeModels)

	missingByType := make(map[string]bool)
	for _, m := range missing {
		missingByType[m.TypeName] = true
		t.Logf("missing: %s does not implement %s", m.TypeName, m.InterfaceName)
	}

	t.Run("value type implements an interface via pointer embedding", func(t *testing.T) {
		assert.False(t, missingByType["Outer"],
			"Outer should implement Reader via the promoted *inner.Foo method (no false positive)")
	})

	t.Run("generic type arguments are distinguished", func(t *testing.T) {
		assert.True(t, missingByType["StringBoxImpl"],
			"StringBoxImpl returns Box[string], not Box[int], so it must NOT implement IntBoxer")
	})

	t.Run("pointer depth is significant", func(t *testing.T) {
		assert.True(t, missingByType["SinglePtrImpl"],
			"SinglePtrImpl takes *int, not **int, so it must NOT implement DoublePtr")
	})
}

func TestImplementsUnexportedMethodCrossPackage(t *testing.T) {
	pass := testutil.CreateTestPass(t, "unexpconsumer")
	cfg := config.Empty()
	ann := annotations.ReadAllAnnotations(cfg, pass)

	interfaces := LoadInterfaces(pass, ann.ToInterfaceQuery())
	typeModels := LoadTypes(pass, ann.ToTypeQuery())
	missing := FindMissingMethods(ann.ImplementsAnnotations, interfaces, typeModels)

	found := false
	for _, m := range missing {
		if m.TypeName == "Impl" {
			found = true
		}
	}

	assert.True(t, found,
		"a type with its own read() does not satisfy an interface whose unexported read() belongs to another package")
}
