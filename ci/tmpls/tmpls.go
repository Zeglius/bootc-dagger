package tmpls

import (
	"crypto/sha256"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

// [template.FuncMap] with functions used in tag templates.
var TmplFuncs = template.FuncMap{

	"nowTag": func() string {
		return time.Now().UTC().Format("20060102")
	},

	"json": func(a any) string {
		b, _ := json.Marshal(a)
		return string(b)
	},

	"uuid": func() string {
		return uuid.NewString()
	},

	"sha256": func(s string) string {
		b := sha256.Sum256([]byte(s))
		return string(b[:])
	},

	"replaceRe": func(pttrn, old, new string) string {
		return regexp.MustCompile(pttrn).ReplaceAllString(old, new)
	},

	"replace": func(s, old, new string) string {
		return strings.ReplaceAll(s, old, new)
	},

	"sort": func(s ...string) []string {
		sort.Strings(s)
		return s
	},

	"parseDockerRef": func(ref string) struct {
		domain string
		owner  string
		image  string
		tag    string
	} {
		// ghcr.io/ublue-os/bazzite:latest -> $domain/$owner/$image:$tag

		itemRe := `[^/]+`

		domainOwnerImageRe := regexp.MustCompile(`^` + itemRe + `/` + itemRe + `/` + itemRe)
		ownerImageRe := regexp.MustCompile(`^` + itemRe + `/` + itemRe)
		imageRe := regexp.MustCompile(`^` + itemRe)

		res := struct {
			domain string
			owner  string
			image  string
			tag    string
		}{}

		switch {
		case domainOwnerImageRe.MatchString(ref):
			a := domainOwnerImageRe.FindStringSubmatch(ref)[:3]
			res.domain, res.owner, res.image = a[1], a[2], a[3]
		case ownerImageRe.MatchString(ref):
			a := ownerImageRe.FindStringSubmatch(ref)[:2]
			res.owner, res.image = a[1], a[2]
		case imageRe.MatchString(ref):
			res.image = imageRe.FindString(ref)
		default:
			return struct {
				domain string
				owner  string
				image  string
				tag    string
			}{}
		}

		// extract tag
		if m := regexp.MustCompile(`:[^:]+$`).FindAllString(ref, 1); len(m) == 1 {
			res.tag = m[0][1:]
		}

		return res
	},
}
