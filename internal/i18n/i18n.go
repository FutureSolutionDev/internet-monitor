package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"

	"internet-monitor/web"
)

const Default = "en"

type Meta struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
	Intl string `json:"intl"`
}

type locale struct {
	Meta     Meta              `json:"meta"`
	Messages map[string]string `json:"messages"`
}

type Vars map[string]any

var (
	locales     = mustLoad(web.Locales)
	placeholder = regexp.MustCompile(`\{(\w+)\}`)
)

func load(fsys fs.FS) (map[string]locale, error) {
	files, err := fs.Glob(fsys, "locales/*.json")
	if err != nil {
		return nil, err
	}
	out := make(map[string]locale, len(files))
	for _, name := range files {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		var l locale
		if err := json.Unmarshal(data, &l); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		out[strings.TrimSuffix(path.Base(name), ".json")] = l
	}
	if _, ok := out[Default]; !ok {
		return nil, fmt.Errorf("locales: %s.json is missing", Default)
	}
	return out, nil
}

func mustLoad(fsys fs.FS) map[string]locale {
	l, err := load(fsys)
	if err != nil {
		panic(err)
	}
	return l
}

func Has(lang string) bool {
	_, ok := locales[lang]
	return ok
}

func Normalize(lang string) string {
	if Has(lang) {
		return lang
	}
	return Default
}

func Languages() []string {
	out := make([]string, 0, len(locales))
	for code := range locales {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

func T(lang, key string, vars Vars) string {
	text, ok := locales[lang].Messages[key]
	if !ok || text == "" {
		text, ok = locales[Default].Messages[key]
	}
	if !ok {
		return key
	}
	if len(vars) == 0 {
		return text
	}
	return placeholder.ReplaceAllStringFunc(text, func(m string) string {
		if v, ok := vars[m[1:len(m)-1]]; ok {
			return fmt.Sprint(v)
		}
		return m
	})
}
