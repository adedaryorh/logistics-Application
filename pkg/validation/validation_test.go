package validation

import "testing"

type sampleRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=3"`
}

func TestValidateStruct(t *testing.T) {
	errs := ValidateStruct(sampleRequest{
		Email: "not-an-email",
		Name:  "ab",
	})
	if len(errs) != 2 {
		t.Fatalf("expected 2 validation errors, got %d", len(errs))
	}
}
