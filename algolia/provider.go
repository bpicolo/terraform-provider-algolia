package algolia

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Algolia application id",
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Algolia api key",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"algolia_index": resourceIndex(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {
	config := Config{
		ApplicationId: data.Get("application_id").(string),
		ApiKey:        data.Get("api_key").(string),
	}

	log.Println("[INFO] Initializing Algolia client")
	return config.Client(), nil
}
