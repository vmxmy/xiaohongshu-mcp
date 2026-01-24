package rod

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// rodCookie 是 Rod 保存的 cookie 格式（也是我们项目使用的标准格式）
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

func loadCookies(path string) ([]*proto.NetworkCookieParam, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var cookies []rodCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, fmt.Errorf("解析 cookies 文件失败: %w", err)
	}

	return convertToProtoCookies(cookies), nil
}

func setCookies(page *rod.Page, cookies []*proto.NetworkCookieParam) error {
	if len(cookies) == 0 {
		return nil
	}

	// 使用 Rod 的 SetCookies 方法
	return page.SetCookies(cookies)
}

func convertToProtoCookies(in []rodCookie) []*proto.NetworkCookieParam {
	out := make([]*proto.NetworkCookieParam, 0, len(in))
	for _, c := range in {
		pc := toProtoCookie(c)
		if pc != nil {
			out = append(out, pc)
		}
	}
	return out
}

func toProtoCookie(c rodCookie) *proto.NetworkCookieParam {
	if c.Name == "" {
		return nil
	}
	if c.Domain == "" && c.Path == "" {
		return nil
	}

	pc := &proto.NetworkCookieParam{
		Name:  c.Name,
		Value: c.Value,
	}

	if c.Domain != "" {
		pc.Domain = c.Domain
	}

	if c.Path != "" {
		pc.Path = c.Path
	}

	if c.Expires > 0 {
		pc.Expires = proto.TimeSinceEpoch(c.Expires)
	}

	pc.HTTPOnly = c.HTTPOnly
	pc.Secure = c.Secure

	// 映射 SameSite 属性
	if sameSite := mapSameSite(c.SameSite); sameSite != "" {
		pc.SameSite = sameSite
	}

	return pc
}

func mapSameSite(value string) proto.NetworkCookieSameSite {
	switch strings.ToLower(value) {
	case "lax":
		return proto.NetworkCookieSameSiteLax
	case "strict":
		return proto.NetworkCookieSameSiteStrict
	case "none":
		return proto.NetworkCookieSameSiteNone
	default:
		return ""
	}
}
