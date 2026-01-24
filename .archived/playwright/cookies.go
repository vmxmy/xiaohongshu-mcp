package playwright

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type rodCookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

type storageState struct {
	Cookies []playwright.OptionalCookie `json:"cookies"`
}

func loadCookies(path string) ([]playwright.OptionalCookie, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return parseCookies(data)
}

func parseCookies(data []byte) ([]playwright.OptionalCookie, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	var rodCookies []rodCookie
	if err := json.Unmarshal(data, &rodCookies); err == nil {
		return convertRodCookies(rodCookies), nil
	}
	var state storageState
	if err := json.Unmarshal(data, &state); err == nil {
		if len(state.Cookies) == 0 {
			return nil, nil
		}
		return state.Cookies, nil
	}
	return nil, fmt.Errorf("unsupported cookies json")
}

func convertRodCookies(in []rodCookie) []playwright.OptionalCookie {
	out := make([]playwright.OptionalCookie, 0, len(in))
	for _, c := range in {
		oc, ok := toOptionalCookie(c)
		if !ok {
			continue
		}
		out = append(out, oc)
	}
	return out
}

func toOptionalCookie(c rodCookie) (playwright.OptionalCookie, bool) {
	if c.Name == "" {
		return playwright.OptionalCookie{}, false
	}
	if c.Domain == "" && c.Path == "" {
		return playwright.OptionalCookie{}, false
	}
	oc := playwright.OptionalCookie{
		Name:  c.Name,
		Value: c.Value,
	}
	if c.Domain != "" {
		domain := c.Domain
		oc.Domain = &domain
	}
	if c.Path != "" {
		path := c.Path
		oc.Path = &path
	}
	if c.Expires > 0 {
		expires := c.Expires
		oc.Expires = &expires
	}
	httpOnly := c.HTTPOnly
	secure := c.Secure
	oc.HttpOnly = &httpOnly
	oc.Secure = &secure
	if sameSite := mapSameSite(c.SameSite); sameSite != nil {
		oc.SameSite = sameSite
	}
	return oc, true
}

func mapSameSite(value string) *playwright.SameSiteAttribute {
	switch strings.ToLower(value) {
	case "lax":
		return playwright.SameSiteAttributeLax
	case "strict":
		return playwright.SameSiteAttributeStrict
	case "none":
		return playwright.SameSiteAttributeNone
	default:
		return nil
	}
}
