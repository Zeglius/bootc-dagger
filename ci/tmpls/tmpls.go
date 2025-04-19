package tmpls

import (
	"crypto/sha256"
	"encoding/json"
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
