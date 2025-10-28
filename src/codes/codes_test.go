package codes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCodeUniqueness verifies that all error codes are unique
func TestCodeUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	var duplicates []string

	for category, codes := range CodesByCategory {
		for _, code := range codes {
			if seen[code.ID] {
				duplicates = append(duplicates, code.ID)
			}
			seen[code.ID] = true

			// Also verify that the code has a description
			assert.NotEmpty(t, code.Description, "Code %s must have a description", code.ID)

			// Verify code belongs to its category
			assert.True(t, strings.HasPrefix(code.ID, category),
				"Code %s must start with category prefix %s", code.ID, category)
		}
	}

	if len(duplicates) > 0 {
		t.Errorf("Found duplicate error codes: %v", duplicates)
	}
}

// TestCodePrefixConsistency verifies that all codes contain the prefix of their category
func TestCodePrefixConsistency(t *testing.T) {
	for category, codes := range CodesByCategory {
		for _, code := range codes {
			assert.True(t, strings.HasPrefix(code.ID, category),
				"Code %s does not start with category prefix %s", code.ID, category)
		}
	}
}

// TestCategoryPrefixConstants verifies that category prefix constants match the keys
func TestCategoryPrefixConstants(t *testing.T) {
	// Verify all expected categories exist
	expectedCategories := map[string]string{
		ImmutableCategoryPrefix:   "IMM",
		ConstructorCategoryPrefix: "CTOR",
		TestOnlyCategoryPrefix:    "TONL",
	}

	for constant, expected := range expectedCategories {
		assert.Equal(t, expected, constant, "Category prefix constant mismatch")
		_, exists := CodesByCategory[expected]
		assert.True(t, exists, "Category %s not found in CodesByCategory", expected)
	}
}

// TestGetCodesForCheck verifies the iterator returns correct codes hierarchy
func TestGetCodesForCheck(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name:     "IMM01 returns ALL, IMM, IMM01",
			code:     ImmutableFieldAssignment,
			expected: []string{"ALL", "IMM", "IMM01"},
		},
		{
			name:     "CTOR02 returns ALL, CTOR, CTOR02",
			code:     ConstructorNewCall,
			expected: []string{"ALL", "CTOR", "CTOR02"},
		},
		{
			name:     "TONL01 returns ALL, TONL, TONL01",
			code:     TestOnlyTypeUsage,
			expected: []string{"ALL", "TONL", "TONL01"},
		},
		{
			name:     "Unknown code returns ALL and code itself",
			code:     "UNKNOWN99",
			expected: []string{"ALL", "UNKNOWN99"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []string
			for code := range GetCodesForCheck(tt.code) {
				result = append(result, code)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCodeToCheckListInitialization verifies the reverse map is built correctly
func TestCodeToCheckListInitialization(t *testing.T) {
	// Verify all codes have entries in the reverse map
	for category, codes := range CodesByCategory {
		for _, code := range codes {
			checkList, exists := codeToCheckList[code.ID]
			require.True(t, exists, "Code %s not found in codeToCheckList", code.ID)

			// Verify structure: [ALL, category, code]
			require.Len(t, checkList, 3, "Check list for %s should have 3 elements", code.ID)
			assert.Equal(t, "ALL", checkList[0])
			assert.Equal(t, category, checkList[1])
			assert.Equal(t, code.ID, checkList[2])
		}
	}
}
