package algolia

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
)

type Config struct {
	ApplicationId string
	ApiKey        string
}

func (c *Config) Client() *algoliasearch.Client {
	client := algoliasearch.NewClient(c.ApplicationId, c.ApiKey)

	return &client
}
