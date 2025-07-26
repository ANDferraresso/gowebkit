package form

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ANDferraresso/gowebkit/validator"
)

func TestFormSetupAndAddField(t *testing.T) {
	form := &Form{}
	form.SetupForm()

	field := &Field{
		Name:      "username",
		Title:     "Username",
		MinLength: "3",
		MaxLength: "20",
	}
	form.AddField(field)

	if len(form.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(form.Fields))
	}
	if form.UIs["username"] == nil {
		t.Error("Expected UI to be initialized for field 'username'")
	}
}

func TestFormValidateField(t *testing.T) {
	form := &Form{}
	form.SetupForm()
	form.Validator = validator.Validator{}
	form.Validator.SetupValidator()

	form.Fields["email"] = &Field{
		Name:      "email",
		MinLength: "5",
		MaxLength: "50",
		Checks:    []validator.Check{},
	}

	ok := form.ValidateField("email", "test@example.com")
	if !ok {
		t.Error("Expected validation to pass for valid email")
	}

	ok = form.ValidateField("email", "abc")
	if ok {
		t.Error("Expected validation to fail for short email")
	}
}

func TestFormValidateAll(t *testing.T) {
	form := &Form{}
	form.SetupForm()
	form.Fields["name"] = &Field{Name: "name", MinLength: "3", MaxLength: "50"}
	form.FieldsOrder = []string{"name"}
	form.Required = []string{"name"}

	form.Prefix = ""

	form.Validator = validator.Validator{}
	form.Validator.SetupValidator()

	data := url.Values{}
	data.Set("name", "John")

	req := httptest.NewRequest("POST", "/", nil)
	req.PostForm = data

	_, fValues, wrong := form.ValidateAll(req)
	if fValues["name"] != "John" {
		t.Errorf("Expected 'John', got %v", fValues["name"])
	}
	if len(wrong) > 0 {
		t.Errorf("Expected no wrong fields, got: %v", wrong)
	}
}
