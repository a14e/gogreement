package ignore

import (
	"bytes"
	"encoding/gob"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goagreement/src/util"
)

func TestIgnoreResultGobSerialization(t *testing.T) {
	// Register the concrete type for gob serialization
	gob.Register(&IgnoreAnnotation{})

	t.Run("Empty IgnoreResult", func(t *testing.T) {
		original := IgnoreResult{
			IgnoreSet: &util.IgnoreSet{},
		}

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		err := encoder.Encode(&original)
		require.NoError(t, err, "encoding should succeed")

		var decoded IgnoreResult
		decoder := gob.NewDecoder(&buf)
		err = decoder.Decode(&decoded)
		require.NoError(t, err, "decoding should succeed")

		assert.Equal(t, 0, decoded.IgnoreSet.Len())
	})

	t.Run("IgnoreResult with annotations", func(t *testing.T) {
		ignoreSet := &util.IgnoreSet{}

		ann1 := &IgnoreAnnotation{
			Codes:    []string{"CODE1"},
			StartPos: token.Pos(10),
			EndPos:   token.Pos(50),
		}
		ignoreSet.Add(ann1)

		ann2 := &IgnoreAnnotation{
			Codes:    []string{"CODE2", "CODE3"},
			StartPos: token.Pos(100),
			EndPos:   token.Pos(200),
		}
		ignoreSet.Add(ann2)

		original := IgnoreResult{
			IgnoreSet: ignoreSet,
		}

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		err := encoder.Encode(&original)
		require.NoError(t, err, "encoding should succeed")

		t.Logf("Serialized size: %d bytes", buf.Len())

		var decoded IgnoreResult
		decoder := gob.NewDecoder(&buf)
		err = decoder.Decode(&decoded)
		require.NoError(t, err, "decoding should succeed")

		// Check length
		require.Equal(t, 2, decoded.IgnoreSet.Len(), "expected 2 annotations")

		// Check that Contains works
		assert.True(t, decoded.IgnoreSet.Contains("CODE1", token.Pos(30)))
		assert.True(t, decoded.IgnoreSet.Contains("CODE2", token.Pos(150)))
		assert.True(t, decoded.IgnoreSet.Contains("CODE3", token.Pos(150)))
		assert.False(t, decoded.IgnoreSet.Contains("CODE1", token.Pos(5)))
	})

	t.Run("IgnoreResult with ALL code", func(t *testing.T) {
		ignoreSet := &util.IgnoreSet{}

		ann := &IgnoreAnnotation{
			Codes:    []string{"ALL"},
			StartPos: token.Pos(1),
			EndPos:   token.Pos(1000),
		}
		ignoreSet.Add(ann)

		original := IgnoreResult{
			IgnoreSet: ignoreSet,
		}

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		err := encoder.Encode(&original)
		require.NoError(t, err, "encoding should succeed")

		var decoded IgnoreResult
		decoder := gob.NewDecoder(&buf)
		err = decoder.Decode(&decoded)
		require.NoError(t, err, "decoding should succeed")

		// Check that ALL code works
		assert.True(t, decoded.IgnoreSet.Contains("ALL", token.Pos(500)))
		assert.True(t, decoded.IgnoreSet.Contains("ALL", token.Pos(1)))
		assert.True(t, decoded.IgnoreSet.Contains("ALL", token.Pos(1000)))
	})
}
