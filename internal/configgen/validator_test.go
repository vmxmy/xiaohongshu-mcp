package configgen

import "testing"

type fakePage struct{ calls int }

func (p *fakePage) WaitVisible(selector string) error {
	p.calls++
	return nil
}

func TestValidator_ChecksRequiredSelectors(t *testing.T) {
	v := Validator{Required: []string{"submit"}}
	if err := v.Validate(map[string]string{"submit": "button"}, &fakePage{}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
