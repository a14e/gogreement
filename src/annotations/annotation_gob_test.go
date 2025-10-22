package annotations

import (
	"bytes"
	"encoding/gob"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageAnnotationsGobSerialization(t *testing.T) {
	t.Run("Empty PackageAnnotations", func(t *testing.T) {
		original := PackageAnnotations{}

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		err := encoder.Encode(&original)
		require.NoError(t, err, "encoding should succeed")

		var decoded PackageAnnotations
		decoder := gob.NewDecoder(&buf)
		err = decoder.Decode(&decoded)
		require.NoError(t, err, "decoding should succeed")

		assert.Equal(t, original, decoded)
	})

	t.Run("PackageAnnotations with TestOnlyAnnotation", func(t *testing.T) {
		original := PackageAnnotations{
			TestonlyAnnotations: []TestOnlyAnnotation{
				{
					Kind:         TestOnlyOnFunc,
					ObjectName:   "ReportProblems",
					Pos:          token.Pos(100),
					ReceiverType: "",
				},
				{
					Kind:         TestOnlyOnMethod,
					ObjectName:   "Reset",
					Pos:          token.Pos(200),
					ReceiverType: "Worker",
				},
				{
					Kind:         TestOnlyOnType,
					ObjectName:   "TestHelper",
					Pos:          token.Pos(300),
					ReceiverType: "",
				},
			},
		}

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		err := encoder.Encode(&original)
		require.NoError(t, err, "encoding should succeed")

		var decoded PackageAnnotations
		decoder := gob.NewDecoder(&buf)
		err = decoder.Decode(&decoded)
		require.NoError(t, err, "decoding should succeed")

		assert.Equal(t, len(original.TestonlyAnnotations), len(decoded.TestonlyAnnotations))
		for i, origAnnot := range original.TestonlyAnnotations {
			decodedAnnot := decoded.TestonlyAnnotations[i]
			assert.Equal(t, origAnnot.Kind, decodedAnnot.Kind, "Kind should match")
			assert.Equal(t, origAnnot.ObjectName, decodedAnnot.ObjectName, "ObjectName should match")
			assert.Equal(t, origAnnot.Pos, decodedAnnot.Pos, "Pos should match")
			assert.Equal(t, origAnnot.ReceiverType, decodedAnnot.ReceiverType, "ReceiverType should match")
		}
	})

	t.Run("PackageAnnotations with all annotation types", func(t *testing.T) {
		original := PackageAnnotations{
			ImplementsAnnotations: []ImplementsAnnotation{
				{
					OnType:          "MyStruct",
					OnTypePos:       token.Pos(50),
					InterfaceName:   "MyInterface",
					PackageName:     "io",
					IsPointer:       true,
					PackageFullPath: "io",
					PackageNotFound: false,
				},
			},
			ConstructorAnnotations: []ConstructorAnnotation{
				{
					OnType:           "MyStruct",
					OnTypePos:        token.Pos(150),
					ConstructorNames: []string{"New", "NewMyStruct"},
				},
			},
			ImmutableAnnotations: []ImmutableAnnotation{
				{
					OnType:    "Config",
					OnTypePos: token.Pos(250),
				},
			},
			TestonlyAnnotations: []TestOnlyAnnotation{
				{
					Kind:         TestOnlyOnFunc,
					ObjectName:   "CreateMockData",
					Pos:          token.Pos(350),
					ReceiverType: "",
				},
			},
		}

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		err := encoder.Encode(&original)
		require.NoError(t, err, "encoding should succeed")

		t.Logf("Serialized size: %d bytes", buf.Len())

		var decoded PackageAnnotations
		decoder := gob.NewDecoder(&buf)
		err = decoder.Decode(&decoded)
		require.NoError(t, err, "decoding should succeed")

		// Check all annotation types
		assert.Equal(t, len(original.ImplementsAnnotations), len(decoded.ImplementsAnnotations))
		assert.Equal(t, len(original.ConstructorAnnotations), len(decoded.ConstructorAnnotations))
		assert.Equal(t, len(original.ImmutableAnnotations), len(decoded.ImmutableAnnotations))
		assert.Equal(t, len(original.TestonlyAnnotations), len(decoded.TestonlyAnnotations))

		// Check specific values
		if len(decoded.TestonlyAnnotations) > 0 {
			assert.Equal(t, "CreateMockData", decoded.TestonlyAnnotations[0].ObjectName)
			assert.Equal(t, TestOnlyOnFunc, decoded.TestonlyAnnotations[0].Kind)
		}
	})
}
