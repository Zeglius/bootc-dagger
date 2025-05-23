package tmpls

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"maps"

	"github.com/google/uuid"
)

// [template.FuncMap] with functions used in tag templates.
var TmplFuncs template.FuncMap

func init() {
	t := template.FuncMap{

		"nowTag": func() string {
			return time.Now().UTC().Format("20060102")
		},

		"json": func(a any) string {
			var (
				res []byte
				err error
			)
			switch a.(type) {
			case map[string]any:
				res, err = json.Marshal(a)
			case []any:
				res, err = json.Marshal(a)
			case map[any]any:
				panic("unsupported type")
			default:
				res, err = json.Marshal(a)
			}

			if err != nil {
				panic(err)
			}
			return string(res)
		},

		"uuid": func() string {
			return uuid.NewString()
		},

		"sha256": func(s string) string {
			b := sha256.Sum256([]byte(s))
			return string(b[:])
		},

		"replaceRe": func(pttrn, new, old string) string {
			return regexp.MustCompile(pttrn).ReplaceAllString(old, new)
		},

		"sort": func(s ...string) []string {
			sort.Strings(s)
			return s
		},

		"startsWith": func(prefix, s string) bool {
			return strings.HasPrefix(s, prefix)
		},

		"endsWith": func(suffix, s string) bool {
			return strings.HasSuffix(s, suffix)
		},

		"contains": func(substr, s string) bool {
			return strings.Contains(s, substr)
		},

		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},

		"join": func(sep string, elems ...string) string {
			return strings.Join(elems, sep)
		},

		"chomp": func(s string) string {
			return strings.TrimRight(s, "\n\r")
		},

		"split": func(sep string, s string) []string {
			return strings.Split(s, sep)
		},

		"splitN": func(sep string, n int, s string) []string {
			return strings.SplitN(s, sep, n)
		},

		"first": func(s string) string {
			return strings.Split(s, "=")[0]
		},

		"last": func(s string) string {
			return strings.Split(s, "=")[1]
		},

		"slice": func(elems ...any) []any {
			return elems
		},

		"dict": func(elems ...any) map[any]any {
			// If the num of elements is not even, means we have a
			// mismatched key-value pair.
			if len(elems)%2 != 0 {
				panic("map function requires an even number of arguments")
			}

			m := make(map[any]any)

			for i := range elems {
				if i%2 != 0 {
					continue
				}
				m[elems[i]] = elems[i+1]
			}

			return m
		},

		"index": func(index int, elems []any) (any, error) {
			if index < 0 || index >= len(elems) {
				return nil, errors.New("index out of range")
			}
			return elems[index], nil
		},

		"fields": func(sep string, line string) (map[string]string, error) {
			res := make(map[string]string)
			k, v, _ := strings.Cut(line, sep)
			res[k] = v
			return res, nil
		},

		"append": func(elem []any, elems ...any) []any {
			return append(elem, elems...)
		},

		"lower": func(s string) string {
			return strings.ToLower(s)
		},

		"upper": func(s string) string {
			return strings.ToUpper(s)
		},

		"title": func(s string) string {
			if s == "" {
				return ""
			}
			return strings.ToUpper(string(s[0])) + s[1:]
		},

		"default": func(defaultV any, input any) any {
			switch input.(type) {
			case nil:
				return defaultV
			case int:
				if input == 0 {
					return defaultV
				}
			case string:
				if input == "" {
					return defaultV
				}
			case float64:
				if input == 0.0 {
					return defaultV
				}
			}
			return input
		},

		// "parseDockerRef": func(ref string) struct {
		// 	domain string
		// 	owner  string
		// 	image  string
		// 	tag    string
		// } {
		// 	// ghcr.io/ublue-os/bazzite:latest -> $domain/$owner/$image:$tag

		// 	itemRe := `[^/]+`

		// 	domainOwnerImageRe := regexp.MustCompile(`^` + itemRe + `/` + itemRe + `/` + itemRe)
		// 	ownerImageRe := regexp.MustCompile(`^` + itemRe + `/` + itemRe)
		// 	imageRe := regexp.MustCompile(`^` + itemRe)

		// 	res := struct {
		// 		domain string
		// 		owner  string
		// 		image  string
		// 		tag    string
		// 	}{}

		// 	switch {
		// 	case domainOwnerImageRe.MatchString(ref):
		// 		a := domainOwnerImageRe.FindStringSubmatch(ref)[:3]
		// 		res.domain, res.owner, res.image = a[1], a[2], a[3]
		// 	case ownerImageRe.MatchString(ref):
		// 		a := ownerImageRe.FindStringSubmatch(ref)[:2]
		// 		res.owner, res.image = a[1], a[2]
		// 	case imageRe.MatchString(ref):
		// 		res.image = imageRe.FindString(ref)
		// 	default:
		// 		return struct {
		// 			domain string
		// 			owner  string
		// 			image  string
		// 			tag    string
		// 		}{}
		// 	}

		// 	// extract tag
		// 	if m := regexp.MustCompile(`:[^:]+$`).FindAllString(ref, 1); len(m) == 1 {
		// 		res.tag = m[0][1:]
		// 	}

		// 	return res
		// },
	}

	if TmplFuncs == nil {
		TmplFuncs = make(template.FuncMap)
	}

	maps.Copy(TmplFuncs, t)
}
