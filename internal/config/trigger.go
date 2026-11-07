package config

import (
	"fmt"
	"net/url"
	"strings"

	"syncftp/internal/lang"
)

// Trigger — belirli dosyalar yüklendiğinde (ya da yüklenmeden hemen önce) çalışan eylemler.
// Global (sync.triggers) ve sunucuya özel (servers[].triggers) olabilir; ikisi birleşir.
// Kalıplar sync include/exclude ile aynı eşleştirici (internal/pattern); boş = her yüklemede.
type Trigger struct {
	Name        string   `json:"name"`
	When        string   `json:"when,omitempty"`    // "after" (varsayılan): yüklemeden sonra · "before": öncesinde, hata → sync iptal
	Paths       []string `json:"paths,omitempty"`   // database/migrations/, *.sql, composer.lock …
	Enabled     *bool    `json:"enabled,omitempty"` // nil = açık
	Run         []string `json:"run,omitempty"`     // yerel komutlar (proje klasöründe)
	Remote      []string `json:"remote,omitempty"`  // SSH komutları (yalnızca SFTP; remote_path'te)
	HTTPURL     string   `json:"http_url,omitempty"`
	HTTPMethod  string   `json:"http_method,omitempty"`  // POST (varsayılan) | GET | PUT | PATCH | DELETE
	HTTPHeaders []string `json:"http_headers,omitempty"` // "Ad: değer"
	HTTPBody    string   `json:"http_body,omitempty"`    // boşsa JSON özet; {files} {files_json} {count} {server} {project} {trigger}
	Notify      string   `json:"notify,omitempty"`       // Slack / Discord / genel webhook adresi
}

// IsEnabled nil = açık.
func (t Trigger) IsEnabled() bool { return t.Enabled == nil || *t.Enabled }

// IsBefore yükleme öncesi mi.
func (t Trigger) IsBefore() bool { return strings.EqualFold(strings.TrimSpace(t.When), "before") }

// HasAction en az bir eylem tanımlı mı.
func (t Trigger) HasAction() bool {
	return nonEmpty(t.Run) || nonEmpty(t.Remote) || strings.TrimSpace(t.HTTPURL) != "" || strings.TrimSpace(t.Notify) != ""
}

func nonEmpty(list []string) bool {
	for _, s := range list {
		if strings.TrimSpace(s) != "" {
			return true
		}
	}
	return false
}

func cloneTriggers(in []Trigger) []Trigger {
	if in == nil {
		return nil
	}
	out := make([]Trigger, len(in))
	for i, t := range in {
		t.Paths = append([]string(nil), t.Paths...)
		t.Run = append([]string(nil), t.Run...)
		t.Remote = append([]string(nil), t.Remote...)
		t.HTTPHeaders = append([]string(nil), t.HTTPHeaders...)
		if t.Enabled != nil {
			v := *t.Enabled
			t.Enabled = &v
		}
		out[i] = t
	}
	return out
}

// maskTriggerHeaders başlık değerlerini (Authorization: Bearer …) maskeler; ad kalır.
func maskTriggerHeaders(in []Trigger, mask func(string) string) []Trigger {
	out := cloneTriggers(in)
	for i := range out {
		for j, h := range out[i].HTTPHeaders {
			if k, v, ok := strings.Cut(h, ":"); ok {
				out[i].HTTPHeaders[j] = k + ": " + mask(strings.TrimSpace(v))
			}
		}
	}
	return out
}

var triggerMethods = map[string]bool{"": true, "GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true}

// ValidateTriggers ad / zaman / adres / yöntem / başlık biçimini denetler. where: "sync" ya da sunucu adı.
func ValidateTriggers(list []Trigger, where string) error {
	seen := map[string]bool{}
	for _, t := range list {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			return fmt.Errorf(lang.L.TriggerValNameFmt, where)
		}
		key := strings.ToLower(name)
		if seen[key] {
			return fmt.Errorf(lang.L.GuiValDupFmt, name)
		}
		seen[key] = true
		if w := strings.ToLower(strings.TrimSpace(t.When)); w != "" && w != "after" && w != "before" {
			return fmt.Errorf(lang.L.TriggerValWhenFmt, name, t.When)
		}
		if !t.HasAction() {
			return fmt.Errorf(lang.L.TriggerValActionFmt, name)
		}
		for _, raw := range []string{t.HTTPURL, t.Notify} {
			if raw = strings.TrimSpace(raw); raw == "" {
				continue
			}
			// yer tutucular (ör. {server}) geçerli bir adresin parçası olabilir — biçim kontrolü öncesi at
			u, err := url.Parse(strings.NewReplacer("{", "", "}", "").Replace(raw))
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
				return fmt.Errorf(lang.L.TriggerValURLFmt, name, raw)
			}
		}
		if !triggerMethods[strings.ToUpper(strings.TrimSpace(t.HTTPMethod))] {
			return fmt.Errorf(lang.L.TriggerValMethodFmt, name, t.HTTPMethod)
		}
		for _, h := range t.HTTPHeaders {
			if h = strings.TrimSpace(h); h == "" {
				continue
			}
			if k, _, ok := strings.Cut(h, ":"); !ok || strings.TrimSpace(k) == "" {
				return fmt.Errorf(lang.L.TriggerValHeaderFmt, name, h)
			}
		}
	}
	return nil
}
