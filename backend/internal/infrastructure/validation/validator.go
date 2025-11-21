package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// InitValidator menginisialisasi validator dengan custom validations
func InitValidator() {
	validate = validator.New()

	// Register custom validation untuk alphanumeric + underscore
	validate.RegisterValidation("alphanum_underscore", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		matched, _ := regexp.MatchString(`^[a-z0-9_]+$`, strings.ToLower(value))
		return matched
	})

	// Register custom validation untuk password strength
	validate.RegisterValidation("password_strength", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
		return hasLetter && hasNumber
	})
}

// GetValidator mengembalikan instance validator
func GetValidator() *validator.Validate {
	if validate == nil {
		InitValidator()
	}
	return validate
}

// ValidateStruct memvalidasi struct menggunakan validator
func ValidateStruct(s interface{}) error {
	if err := GetValidator().Struct(s); err != nil {
		// Format error messages menjadi lebih user-friendly
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			tag := err.Tag()
			
			var message string
			switch tag {
			case "required":
				message = fmt.Sprintf("%s is required", field)
			case "min":
				message = fmt.Sprintf("%s must be at least %s characters", field, err.Param())
			case "max":
				message = fmt.Sprintf("%s must be less than %s characters", field, err.Param())
			case "email":
				message = fmt.Sprintf("%s must be a valid email address", field)
			case "len":
				message = fmt.Sprintf("%s must be exactly %s characters", field, err.Param())
			case "numeric":
				message = fmt.Sprintf("%s must contain only numbers", field)
			case "alphanum_underscore":
				message = fmt.Sprintf("%s can only contain lowercase letters, numbers, and underscores", field)
			case "password_strength":
				message = fmt.Sprintf("%s must contain at least one letter and one number", field)
			default:
				message = fmt.Sprintf("%s is invalid", field)
			}
			errors = append(errors, message)
		}
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

