package validator_test

import (
	"log"

	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/validator"
	"github.com/go-json-schema/validator/builder"
)

func ExampleBuild() {
	s, err := schema.ParseFile(`/path/to/schema.json`)
	if err != nil {
		log.Printf("failed to open schema: %s", err)
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if err != nil {
		log.Printf("failed to build validator: %s", err)
		return
	}

	var input interface{}
	if err := v.Validate(input); err != nil {
		log.Printf("validation failed: %s", err)
		return
	}
}

func ExampleManual() {
	v := validator.Object().
		AddProp(`zip`, validator.String().RegexpString(`^\d{5}$`)).
		AddProp(`address`, validator.String()).
		AddProp(`name`, validator.String()).
		AddProp(`phone_number`, validator.String().RegexpString(`^[\d-]+$`)).
		Required(`zip`, `address`, `name`)

	var input interface{}
	if err := v.Validate(input); err != nil {
		log.Printf("validation failed: %s", err)
		return
	}
}