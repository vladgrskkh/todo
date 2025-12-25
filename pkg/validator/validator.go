package validator

// Validator is a struct that holds a map of validation errors.
type Validator struct {
	Errors map[string]string
}

// New returns a new Validator instance with an empty Errors map.
// The Errors map is used to store validation errors.
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

func (v *Validator) Error() string {
	if !v.Valid() {
		return "validation error"
	}
	return ""
}

// Valid returns true if the validator has no errors, false otherwise.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error to the validator for the given key and message.
// If the key already exists in the Errors map, the error is not added.
func (v *Validator) AddError(key, message string) {
	if _, exist := v.Errors[key]; !exist {
		v.Errors[key] = message
	}
}

// Check adds an error to the validator for the given key and message
// if the condition ok is false(calls the AddError method).
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}
