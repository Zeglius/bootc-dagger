package tmpls

import (
	"context"
	"dagger/ci/internal/dagger"
	"regexp"
	"strings"
)

// SecretsToMap converts a list of dagger Secrets to a map[string]string.
// Returns nil when input is empty, error parsing URI or error reading plaintext value.
// Secret URI must be in the format `[scheme]://[key]` where key does not contain `:` or `$`.
func SecretsToMap(ctx context.Context, secrets []*dagger.Secret) map[string]string {

	if len(secrets) == 0 {
		return nil
	}

	secretsMap := map[string]string{}
	for _, s := range secrets {
		uri, err := s.URI(ctx)
		if err != nil {
			return nil
		}

		name := ""
		if matches := regexp.MustCompile(`^.*://(.*)$`).FindStringSubmatch(uri); matches != nil {
			if len(matches) == 0 {
				continue
			} else if strings.ContainsAny(matches[1], `:$`) {
				continue
			} else {
				name = matches[1]
			}

			value, _ := s.Plaintext(ctx)

			secretsMap[name] = value
		}
	}

	return secretsMap

}
