package cli

import (
	"fmt"

	"github.com/owainlewis/pair-cli/internal/api"
	pairconfig "github.com/owainlewis/pair-cli/internal/config"
)

func newAPIClient(opts *Options) (api.Client, error) {
	resolved, err := pairconfig.Resolve(pairconfig.Overrides{
		BaseURL: opts.BaseURL,
		Token:   opts.Token,
	})
	if err != nil {
		return api.Client{}, err
	}
	if resolved.BaseURL == "" {
		return api.Client{}, fmt.Errorf("missing base URL: set PAIR_BASE_URL or run pair config set base-url <url>")
	}
	if resolved.Token == "" {
		return api.Client{}, fmt.Errorf("missing token: set PAIR_TOKEN or run pair config set token <token>")
	}

	return api.Client{
		BaseURL: resolved.BaseURL,
		Token:   resolved.Token,
	}, nil
}
