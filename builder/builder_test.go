package builder

import (
	"strings"
	"testing"

	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/schema/common"
	"github.com/stretchr/testify/assert"
)

func TestGuessType(t *testing.T) {
	data := map[string]common.PrimitiveTypeList{
		`{ "items": { "type": "number" } }`:                 {common.ArrayType},
		`{ "properties": { "foo": { "type": "number" } } }`: {common.ObjectType},
		`{ "additionalProperties": false }`:                 {common.ObjectType},
		`{ "maxLength": 10 }`:                               {common.StringType},
		`{ "pattern": "^[a-fA-F0-9]+$" }`:                   {common.StringType},
		`{ "format": "email" }`:                             {common.StringType},
		`{ "multipleOf": 1 }`:                               {common.IntegerType},
		`{ "multipleOf": 1.1 }`:                             {common.NumberType},
		`{ "minimum": 0, "pattern": "^[a-z]+$" }`:           {common.StringType, common.IntegerType},
	}

	for src, expected := range data {
		t.Run(src, func(t *testing.T) {
			s, err := schema.Parse(strings.NewReader(src))
			if !assert.NoError(t, err, "schema.Parse should succeed") {
				return
			}

			if !assert.Equal(t, guessSchemaType(s), expected, "types match") {
				return
			}
		})
	}
}
