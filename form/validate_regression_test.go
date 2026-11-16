package form

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ValidateAll must report ok=false when at least one field is wrong,
// consistently with ValidateAllMap.
func TestValidateAllReportsFailure(t *testing.T) {
	f := &Form{}
	f.SetupForm()
	f.AddField(&Field{Name: "age", MinLength: "1", MaxLength: "3"})
	f.Required = []string{"age"}

	body := url.Values{"age": {"toolong"}}.Encode()
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ok, _, wrong := f.ValidateAll(req)
	if ok {
		t.Errorf("ValidateAll ok = true, want false (wrong fields: %v)", wrong)
	}
	if len(wrong) != 1 || wrong[0] != "age" {
		t.Errorf("wrong fields = %v, want [age]", wrong)
	}

	okMap, _, _ := f.ValidateAllMap(map[string]interface{}{"age": "toolong"})
	if okMap != ok {
		t.Errorf("ValidateAll and ValidateAllMap disagree: %v vs %v", ok, okMap)
	}
}
