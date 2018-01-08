package validator_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/schema/draft04"
	"github.com/go-json-schema/validator"
	"github.com/go-json-schema/validator/builder"
	"github.com/stretchr/testify/assert"
)

// test against github#2
func TestArrayItemsReference(t *testing.T) {
	const src = `{
	"definitions": {
	  "uint": {
			"type": "integer",
			"minimum": 0
		}
	},
	"type": "object",
	"properties": {
		"numbers": {
			"type": "array",
			"items": { "$ref": "#/definitions/uint" }
		},
		"tuple": {
			"items": [ { "type": "string" }, { "type": "boolean" }, { "type": "number" } ]
		}
	}
}`
	s, err := schema.Parse(strings.NewReader(src), schema.WithSchemaID(draft04.SchemaID))
	if !assert.NoError(t, err, "schema.Parseer should succeed") {
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if !assert.NoError(t, err, "builder.Build should succeed") {
		return
	}

	var buf bytes.Buffer
	g := validator.NewGenerator()
	if !assert.NoError(t, g.Process(&buf, v), "Generator.Process should succeed") {
		return
	}

	code := buf.String()
	if !assert.True(t, strings.Contains(code, "\tItems("), "Generated code chould contain `.Items()`") {
		t.Logf("%s", code)
		return
	}

	if !assert.True(t, strings.Contains(code, "\tPositionalItems("), "Generated code should contain `.PositionalItems()`") {
		t.Logf("%s", code)
		return
	}
}
