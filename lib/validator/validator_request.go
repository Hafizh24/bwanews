package validatorLib

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	var errorMessages []string
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				errorMessages = append(errorMessages, err.Field()+" is required")
			case "email":
				errorMessages = append(errorMessages, err.Field()+" must be a valid email")
			case "min":
				errorMessages = append(errorMessages, err.Field()+" must be at least "+err.Param()+" characters long")
			case "eqfield":
				errorMessages = append(errorMessages, err.Field()+" must be equal to "+err.Param())
			default:
				errorMessages = append(errorMessages, err.Field()+" is not valid")
			}
		}
		return errors.New("validation error: " + joinMessages(errorMessages))

	}
	return nil
}

func joinMessages(messages []string) string {
	result := ""
	for i, msg := range messages {
		if i > 0 {
			result += "; "
		}
		result += msg
	}

	return result
}
