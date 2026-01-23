package configgen

import "fmt"

type PageWaiter interface {
	WaitVisible(selector string) error
}

type Validator struct {
	Required []string
}

func (v Validator) Validate(selectors map[string]string, page PageWaiter) error {
	for _, key := range v.Required {
		selector := selectors[key]
		if selector == "" {
			return fmt.Errorf("missing selector: %s", key)
		}
		if err := page.WaitVisible(selector); err != nil {
			return err
		}
	}
	return nil
}
