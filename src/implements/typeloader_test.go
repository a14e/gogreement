package implements

import (
	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTypesVariousKinds(t *testing.T) {
	pass := testutil.CreateTestPass(t, "withimports")

	tests := []struct {
		name               string
		queries            []annotations.TypeQuery
		expectedCount      int
		expectedType       string
		expectedUnderlying string
		expectedMethods    []string
	}{
		{
			name: "load struct type",
			queries: []annotations.TypeQuery{
				{TypeName: "MyReader"},
			},
			expectedCount:      1,
			expectedType:       "MyReader",
			expectedUnderlying: "struct",
			expectedMethods:    []string{"Read"},
		},
		{
			name: "load int alias",
			queries: []annotations.TypeQuery{
				{TypeName: "Duration"},
			},
			expectedCount:      1,
			expectedType:       "Duration",
			expectedUnderlying: "int64",
			expectedMethods:    []string{"Seconds", "String"},
		},
		{
			name: "load string alias",
			queries: []annotations.TypeQuery{
				{TypeName: "MyString"},
			},
			expectedCount:      1,
			expectedType:       "MyString",
			expectedUnderlying: "string",
			expectedMethods:    []string{"Upper", "Append"},
		},
		{
			name: "load func type",
			queries: []annotations.TypeQuery{
				{TypeName: "HandlerFunc"},
			},
			expectedCount:      1,
			expectedType:       "HandlerFunc",
			expectedUnderlying: "func",
			expectedMethods:    []string{"ServeHTTP"},
		},
		{
			name: "load slice type",
			queries: []annotations.TypeQuery{
				{TypeName: "ByteSlice"},
			},
			expectedCount:      1,
			expectedType:       "ByteSlice",
			expectedUnderlying: "slice",
			expectedMethods:    []string{"Len", "Append"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LoadTypes(pass, tt.queries)

			assert.Len(t, result, tt.expectedCount)

			if len(result) > 0 {
				typeModel := result[0]
				assert.Equal(t, tt.expectedType, typeModel.Name)
				assert.Equal(t, tt.expectedUnderlying, typeModel.UnderlyingType)
				assert.Len(t, typeModel.Methods, len(tt.expectedMethods))

				methodNames := make([]string, len(typeModel.Methods))
				for i, method := range typeModel.Methods {
					methodNames[i] = method.Name
				}

				for _, expectedMethod := range tt.expectedMethods {
					assert.Contains(t, methodNames, expectedMethod)
				}
			}
		})
	}
}
