package validator_test

import (
	"bytes"
	"testing"

	"github.com/go-json-schema/validator"
	"github.com/stretchr/testify/assert"
)

func TestGenerator_StringMinLength(t *testing.T) {
	v := validator.New().SetRoot(validator.String().MinLength(1))
	g := validator.NewGenerator()

	buf := bytes.Buffer{}
	if !assert.NoError(t, g.Process(&buf, v), "Process() succeeds") {
		return
	}

	t.Logf("%s", buf.String())
}