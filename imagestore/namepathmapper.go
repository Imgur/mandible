package imagestore

import (
	"regexp"
	"strings"
)

type NamePathMapper struct {
	regex   *regexp.Regexp
	replace string
}

func NewNamePathMapper(expr string, mapping string) *NamePathMapper {
	var r *regexp.Regexp
	if len(expr) > 0 {
		r = regexp.MustCompile(expr)
	}

	return &NamePathMapper{
		r,
		mapping,
	}
}

func (this *NamePathMapper) mapToPath(obj *StoreObject) string {
	repl := strings.Replace(this.replace, "${ImageName}", obj.Name, -1)
	repl = strings.Replace(repl, "${ImageSize}", obj.Type, -1)

	if this.regex != nil {
		return this.regex.ReplaceAllString(obj.Name, repl)
	}

	return repl
}
