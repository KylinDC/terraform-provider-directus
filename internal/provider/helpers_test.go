package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSetStringField(t *testing.T) {
	t.Run("adds non-null non-unknown value", func(t *testing.T) {
		input := make(map[string]interface{})
		setStringField(input, "name", types.StringValue("hello"))
		assert.Equal(t, "hello", input["name"])
	})

	t.Run("skips null value", func(t *testing.T) {
		input := make(map[string]interface{})
		setStringField(input, "name", types.StringNull())
		assert.NotContains(t, input, "name")
	})

	t.Run("skips unknown value", func(t *testing.T) {
		input := make(map[string]interface{})
		setStringField(input, "name", types.StringUnknown())
		assert.NotContains(t, input, "name")
	})

	t.Run("adds empty string value", func(t *testing.T) {
		input := make(map[string]interface{})
		setStringField(input, "name", types.StringValue(""))
		assert.Equal(t, "", input["name"])
	})
}

func TestSetBoolField(t *testing.T) {
	t.Run("adds true value", func(t *testing.T) {
		input := make(map[string]interface{})
		setBoolField(input, "hidden", types.BoolValue(true))
		assert.Equal(t, true, input["hidden"])
	})

	t.Run("adds false value", func(t *testing.T) {
		input := make(map[string]interface{})
		setBoolField(input, "hidden", types.BoolValue(false))
		assert.Equal(t, false, input["hidden"])
	})

	t.Run("skips null value", func(t *testing.T) {
		input := make(map[string]interface{})
		setBoolField(input, "hidden", types.BoolNull())
		assert.NotContains(t, input, "hidden")
	})

	t.Run("skips unknown value", func(t *testing.T) {
		input := make(map[string]interface{})
		setBoolField(input, "hidden", types.BoolUnknown())
		assert.NotContains(t, input, "hidden")
	})
}

func TestStringOrNull(t *testing.T) {
	t.Run("returns value for non-empty string", func(t *testing.T) {
		result := stringOrNull("hello")
		assert.Equal(t, "hello", result.ValueString())
		assert.False(t, result.IsNull())
	})

	t.Run("returns null for empty string", func(t *testing.T) {
		result := stringOrNull("")
		assert.True(t, result.IsNull())
	})
}

func TestStringListOrNull(t *testing.T) {
	t.Run("returns list for non-empty slice", func(t *testing.T) {
		result := stringListOrNull([]string{"a", "b", "c"})
		assert.False(t, result.IsNull())

		var elements []string
		result.ElementsAs(context.Background(), &elements, false)
		assert.Equal(t, []string{"a", "b", "c"}, elements)
	})

	t.Run("returns null for empty slice", func(t *testing.T) {
		result := stringListOrNull([]string{})
		assert.True(t, result.IsNull())
	})

	t.Run("returns null for nil slice", func(t *testing.T) {
		result := stringListOrNull(nil)
		assert.True(t, result.IsNull())
	})

	t.Run("returns list for single element", func(t *testing.T) {
		result := stringListOrNull([]string{"only"})
		assert.False(t, result.IsNull())

		var elements []string
		result.ElementsAs(context.Background(), &elements, false)
		assert.Equal(t, []string{"only"}, elements)
	})
}

func TestBuildInputMap(t *testing.T) {
	t.Run("includes non-nil values only", func(t *testing.T) {
		fields := map[string]interface{}{
			"name":  "test",
			"icon":  nil,
			"note":  "a note",
			"count": nil,
		}

		result := buildInputMap(fields)

		assert.Equal(t, "test", result["name"])
		assert.Equal(t, "a note", result["note"])
		assert.NotContains(t, result, "icon")
		assert.NotContains(t, result, "count")
		assert.Len(t, result, 2)
	})

	t.Run("returns empty map for all nil values", func(t *testing.T) {
		fields := map[string]interface{}{
			"a": nil,
			"b": nil,
		}

		result := buildInputMap(fields)
		assert.Empty(t, result)
	})

	t.Run("returns empty map for empty input", func(t *testing.T) {
		result := buildInputMap(map[string]interface{}{})
		assert.Empty(t, result)
	})

	t.Run("preserves various value types", func(t *testing.T) {
		fields := map[string]interface{}{
			"str":  "value",
			"num":  42,
			"bool": true,
			"list": []string{"a"},
		}

		result := buildInputMap(fields)
		assert.Equal(t, "value", result["str"])
		assert.Equal(t, 42, result["num"])
		assert.Equal(t, true, result["bool"])
		assert.Equal(t, []string{"a"}, result["list"])
	})
}
