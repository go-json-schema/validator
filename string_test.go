package validator_test

import (
	"strings"
	"testing"

	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/schema/draft04"
	"github.com/go-json-schema/validator"
	"github.com/go-json-schema/validator/builder"
	"github.com/stretchr/testify/assert"
)

func TestStringFromSchema(t *testing.T) {
	const src = `{
  "type": "string",
  "maxLength": 15,
  "minLength": 5,
  "default": "Hello, World!"
}`

	s, err := schema.Parse(strings.NewReader(src), schema.WithSchemaID(draft04.SchemaID))
	if !assert.NoError(t, err, "schema.Parse should succeed") {
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if !assert.NoError(t, err, "Builder.Build should succeed") {
		return
	}

	c2 := validator.String()
	c2.Default("Hello, World!").MaxLength(15).MinLength(5)
	if !assert.Equal(t, c2, v.Root(), "constraints are equal") {
		return
	}
}

func TestString(t *testing.T) {
	var s string
	c := validator.String()
	c.Default("Hello, World!").MaxLength(15)

	if !assert.True(t, c.HasDefault(), "HasDefault is true") {
		return
	}

	if !assert.Equal(t, c.DefaultValue(), "Hello, World!", "DefaultValue returns expected value") {
		return
	}

	if !assert.NoError(t, c.Validate(s), "validate should succeed") {
		return
	}

	c.MinLength(5)
	if !assert.Error(t, c.Validate(s), "validate should fail") {
		return
	}

	s = "Hello"
	if !assert.NoError(t, c.Validate(s), "validate should succeed") {
		return
	}
}
