package algolia

import (
	"fmt"
	"os"
	"reflect"

	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/hashicorp/terraform/helper/schema"
)

var rankingDefault = []string{"typo", "geo", "words", "filters", "proximity", "attribute", "exact", "custom"}

func resourceIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceIndexCreate,
		Read:   resourceIndexRead,
		Update: resourceIndexUpdate,
		Delete: resourceIndexDelete,

		Schema: map[string]*schema.Schema{
			// Attributes
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of this terraform index",
			},
			"searchable_attributes": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of attributes eligible for textual search.",
			},
			"attributes_for_faceting": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of attributes you want to use for faceting.",
			},
			"unretrievable_attributes": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of attributes that cannot be retrieved at query time.",
			},
			"attribute_for_distinct": &schema.Schema{
				Type:        schema.TypeString,
				Default:     nil,
				Optional:    true,
				Description: "Name of the de-duplication attribute for the distinct feature.",
			},
			"attributes_to_retrieve": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of object attributes you want to retrieve.",
			},
			// Ranking
			"custom_ranking": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Specifies the custom ranking criterion.",
			},
			// TODO distinct as string (integer | bool)
			"ranking": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Controls the way results are sorted.",
			},
			// TODO how do we make this depend on actual resources created in terraform?
			"replicas": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of indices to which you want to replicate all write operations.",
			},
			"max_values_per_facet": &schema.Schema{
				Type:         schema.TypeInt,
				Default:      100,
				Optional:     true,
				Description:  "Maximum number of facet values returned for each facet.",
				ValidateFunc: IntBetween(1, 1000),
			},
			"sort_facet_values_by": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "count",
				Optional:     true,
				Description:  "Controls how facet values are sorted.",
				ValidateFunc: StringInSet([]string{"alpha", "count"}),
			},
			// Highlighting / snippeting
			"attributes_to_highlight": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of attributes to highlight.",
			},
			"attributes_to_snippet": &schema.Schema{ // TODO validate this against valid count format?
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of attributes to snippet, with an optional maximum number of words to snippet.",
			},
			"highlight_pre_tag": &schema.Schema{
				Type:        schema.TypeString,
				Default:     "<em>",
				Optional:    true,
				Description: "String inserted before highlighted parts in highlight and snippet results.",
			},
			"highlight_post_tag": &schema.Schema{
				Type:        schema.TypeString,
				Default:     "</em>",
				Optional:    true,
				Description: "String inserted after highlighted parts in highlight and snippet results.",
			},
			"optional_words": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of words that should be considered as optional when found in the query.",
			},
			"snippet_ellipsis_text": &schema.Schema{
				Type:        schema.TypeString,
				Default:     "…",
				Optional:    true,
				Description: "String used as an ellipsis indicator when a snippet is truncated.",
			},
			"restrict_highlight_and_snippet_arrays": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Restrict arrays in highlight and snippet results to items that matched the query. \nWhen false, all array items are highlighted/snippeted. When true, only array items that matched at least partially are highlighted/snippeted.",
			},
			"advanced_syntax": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enables the advanced query syntax.",
			},
			"hits_per_page": &schema.Schema{
				Type:         schema.TypeInt,
				Default:      20,
				Optional:     true,
				Description:  "Set the number of hits per page.",
				ValidateFunc: IntBetween(1, 1000),
			},
			"pagination_limited_to": &schema.Schema{
				Type:         schema.TypeInt,
				Default:      1000,
				Optional:     true,
				Description:  "Set the number of hits accessible via pagination.",
				ValidateFunc: IntGTE(1), // There's not really an upper cap according to docs, but > 1000 will cause perf issues
			},
			"min_proximity": &schema.Schema{
				Type:         schema.TypeInt,
				Default:      1,
				Optional:     true,
				Description:  "Precision of the proximity ranking criterion.",
				ValidateFunc: IntBetween(1, 7),
			},
			"min_word_size_for_1_typo": &schema.Schema{
				Type:         schema.TypeInt,
				Default:      4,
				Optional:     true,
				Description:  "Minimum number of characters a word in the query string must contain to accept matches with one typo.",
				ValidateFunc: IntGTE(1),
			},
			"min_word_size_for_2_typos": &schema.Schema{
				Type:         schema.TypeInt,
				Default:      8,
				Optional:     true,
				Description:  "Minimum number of characters a word in the query string must contain to accept matches with two typos.",
				ValidateFunc: IntGTE(1),
			},
			"max_facet_hits": &schema.Schema{
				Type:         schema.TypeInt,
				Default:      10,
				Optional:     true,
				Description:  "Maximum number of facet hits to return during a search for facet values.",
				ValidateFunc: IntBetween(1, 100),
			},
			"typo_tolerance": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "true",
				Optional:     true,
				Description:  "Controls whether typo tolerance is enabled and how it is applied",
				ValidateFunc: StringInSet([]string{"true", "false", "min", "strict"}),
			},
			"allow_typos_on_numeric_tokens": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Whether to allow typos on numbers (“numeric tokens”) in the query str",
			},
			"replace_synonyms_in_highlight": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Whether to replace words matched via synonym expansion by the matched synonym in highlight and snippet results.",
			},
			"allow_compression_of_integer_array": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enables compression of large integer arrays.",
			},
			"query_type": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "prefixLast",
				Optional:     true,
				Description:  "Controls if and how query words are interpreted as prefixes.",
				ValidateFunc: StringInSet([]string{"prefixLast", "prefixAll", "prefixNone"}),
			},
			// TODO think of the best way to model this guy, can be string[] or bool
			// Same with RemoveStopWords
			// "ignorePlurals": &schema.Schema{
			// 	Type:        schema.TypeBool,
			// 	Default:     true,
			// 	Optional:    true,
			// 	Description: "Whether to allow typos on numbers (“numeric tokens”) in the query str",
			// },
			// TODO validate is a subset of searchableAttributes
			"disable_typo_tolerance_on_attributes": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of attributes on which you want to disable typo tolerance.",
			},
			"response_fields": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Choose which fields the response will contain. Applies to search and browse queries.",
			},
			"disable_typo_tolerance_on_words": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of words on which typo tolerance will be disabled.",
			},
			"separators_to_index": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Separators (punctuation characters) to index.",
			},
			"remove_words_if_no_results": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				Description:  "Selects a strategy to remove words from the query when it doesn’t match any hits.",
				ValidateFunc: StringInSet([]string{"none", "lastWords", "firstWords", "allOptional"}),
			},
		},
	}
}

func buildSettingsFromResourceData(d *schema.ResourceData) algoliasearch.Settings {
	ranking := castStringList(d.Get("ranking").([]interface{}))
	// Normalize how defaults are handled on both ends, so that terraform properly stores
	// the status in tfstate.
	if reflect.DeepEqual(ranking, rankingDefault) {
		ranking = []string{}
	}

	settings := algoliasearch.Settings{
		AdvancedSyntax:                    d.Get("advanced_syntax").(bool),
		AllowCompressionOfIntegerArray:    d.Get("allow_compression_of_integer_array").(bool),
		AllowTyposOnNumericTokens:         d.Get("allow_typos_on_numeric_tokens").(bool),
		AttributeForDistinct:              d.Get("attribute_for_distinct").(string),
		AttributesForFaceting:             castStringList(d.Get("attributes_for_faceting").([]interface{})),
		AttributesToHighlight:             castStringList(d.Get("attributes_to_highlight").([]interface{})),
		AttributesToRetrieve:              castStringList(d.Get("attributes_to_retrieve").([]interface{})),
		AttributesToSnippet:               castStringList(d.Get("attributes_to_snippet").([]interface{})),
		CustomRanking:                     castStringList(d.Get("custom_ranking").([]interface{})),
		DisableTypoToleranceOnAttributes:  castStringList(d.Get("disable_typo_tolerance_on_attributes").([]interface{})),
		DisableTypoToleranceOnWords:       castStringList(d.Get("disable_typo_tolerance_on_words").([]interface{})),
		HighlightPostTag:                  d.Get("highlight_post_tag").(string),
		HighlightPreTag:                   d.Get("highlight_pre_tag").(string),
		HitsPerPage:                       d.Get("hits_per_page").(int),
		MaxFacetHits:                      d.Get("max_facet_hits").(int),
		MaxValuesPerFacet:                 d.Get("max_values_per_facet").(int),
		MinProximity:                      d.Get("min_proximity").(int),
		MinWordSizefor1Typo:               d.Get("min_word_size_for_1_typo").(int),
		MinWordSizefor2Typos:              d.Get("min_word_size_for_2_typos").(int),
		PaginationLimitedTo:               d.Get("pagination_limited_to").(int),
		OptionalWords:                     castStringList(d.Get("optional_words").([]interface{})),
		QueryType:                         d.Get("query_type").(string),
		Ranking:                           ranking,
		RemoveWordsIfNoResults:            d.Get("remove_words_if_no_results").(string),
		ReplaceSynonymsInHighlight:        d.Get("replace_synonyms_in_highlight").(bool),
		Replicas:                          castStringList(d.Get("replicas").([]interface{})),
		ResponseFields:                    castStringList(d.Get("response_fields").([]interface{})),
		RestrictHighlightAndSnippetArrays: d.Get("restrict_highlight_and_snippet_arrays").(bool),
		SearchableAttributes:              castStringList(d.Get("searchable_attributes").([]interface{})),
		SeparatorsToIndex:                 d.Get("separators_to_index").(string),
		SnippetEllipsisText:               d.Get("snippet_ellipsis_text").(string),
		SortFacetValuesBy:                 d.Get("sort_facet_values_by").(string),
		TypoTolerance:                     d.Get("typo_tolerance").(string),
		UnretrievableAttributes:           castStringList(d.Get("unretrievable_attributes").([]interface{})),
	}

	return settings
}

func readResourceFromSettings(d *schema.ResourceData, s algoliasearch.Settings) {
	ranking := s.Ranking
	if reflect.DeepEqual(s.Ranking, rankingDefault) {
		ranking = []string{}
	}

	d.Set("advanced_syntax", s.AdvancedSyntax)
	d.Set("allow_compression_of_integer_array", s.AllowCompressionOfIntegerArray)
	d.Set("allow_typos_on_numeric_tokens", s.AllowTyposOnNumericTokens)
	d.Set("attribute_for_distinct", s.AttributeForDistinct)
	d.Set("attributes_for_faceting", s.AttributesForFaceting)
	d.Set("attributes_to_highlight", s.AttributesToHighlight)
	d.Set("attributes_to_retrieve", s.AttributesToRetrieve)
	d.Set("attributes_to_snippet", s.AttributesToSnippet)
	d.Set("custom_ranking", s.CustomRanking)
	d.Set("disable_typo_tolerance_on_attributes", s.DisableTypoToleranceOnAttributes)
	d.Set("disable_typo_tolerance_on_words", s.DisableTypoToleranceOnWords)
	d.Set("highlight_post_tag", s.HighlightPostTag)
	d.Set("highlight_pre_tag", s.HighlightPreTag)
	d.Set("hits_per_page", s.HitsPerPage)
	d.Set("max_facet_hits", s.MaxFacetHits)
	d.Set("max_values_per_facet", s.MaxValuesPerFacet)
	d.Set("min_proximity", s.MinProximity)
	d.Set("min_word_size_for_1_typo", s.MinWordSizefor1Typo)
	d.Set("min_word_size_for_2_typos", s.MinWordSizefor2Typos)
	d.Set("optional_words", s.OptionalWords)
	d.Set("pagination_limited_to", s.PaginationLimitedTo)
	d.Set("query_type", s.QueryType)
	d.Set("ranking", ranking)
	d.Set("remove_words_if_no_results", s.RemoveWordsIfNoResults)
	d.Set("replace_synonyms_in_highlight", s.ReplaceSynonymsInHighlight)
	d.Set("replicas", s.Replicas)
	d.Set("response_fields", s.ResponseFields)
	d.Set("restrict_highlight_and_snippet_arrays", s.RestrictHighlightAndSnippetArrays)
	d.Set("searchable_attributes", s.SearchableAttributes)
	d.Set("separators_to_index", s.SeparatorsToIndex)
	d.Set("snippet_ellipsis_text", s.SnippetEllipsisText)
	d.Set("sort_facets_values_by", s.SortFacetValuesBy)
	d.Set("typo_tolerance", s.TypoTolerance)
	d.Set("unretrievable_attributes", s.UnretrievableAttributes)
}

// Takes an array of interface and casts to string
func castStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}

// The client's default ToMap() removes empty array attributes, which isn't correct behavior
// when we explicitly want to create empty settings e.g. to clear out settings that are currently set.
// Original: https://github.com/algolia/algoliasearch-client-go/blob/master/algoliasearch/types_settings.go#L81
func settingsAsMap(s algoliasearch.Settings) algoliasearch.Map {
	m := algoliasearch.Map{
		// Indexing parameters
		"allowCompressionOfIntegerArray": s.AllowCompressionOfIntegerArray,
		"attributeForDistinct":           s.AttributeForDistinct,
		"attributesForFaceting":          s.AttributesForFaceting,
		"attributesToIndex":              s.AttributesToIndex,
		"customRanking":                  s.CustomRanking,
		"numericAttributesToIndex":       s.NumericAttributesToIndex,
		"numericAttributesForFiltering":  s.NumericAttributesForFiltering,
		"ranking":                        s.Ranking,
		"replicas":                       s.Replicas,
		"searchableAttributes":           s.SearchableAttributes,
		"separatorsToIndex":              s.SeparatorsToIndex,
		"unretrievableAttributes":        s.UnretrievableAttributes,

		// Query expansion
		"disableTypoToleranceOnAttributes": s.DisableTypoToleranceOnAttributes,
		"disableTypoToleranceOnWords":      s.DisableTypoToleranceOnWords,

		// Default query parameters (can be overridden at query-time)
		"advancedSyntax":             s.AdvancedSyntax,
		"allowTyposOnNumericTokens":  s.AllowTyposOnNumericTokens,
		"attributesToHighlight":      s.AttributesToHighlight,
		"attributesToRetrieve":       s.AttributesToRetrieve,
		"attributesToSnippet":        s.AttributesToSnippet,
		"highlightPostTag":           s.HighlightPostTag,
		"highlightPreTag":            s.HighlightPreTag,
		"hitsPerPage":                s.HitsPerPage,
		"maxFacetHits":               s.MaxFacetHits,
		"maxValuesPerFacet":          s.MaxValuesPerFacet,
		"minProximity":               s.MinProximity,
		"minWordSizefor1Typo":        s.MinWordSizefor1Typo,
		"minWordSizefor2Typos":       s.MinWordSizefor2Typos,
		"optionalWords":              s.OptionalWords,
		"queryType":                  s.QueryType,
		"replaceSynonymsInHighlight": s.ReplaceSynonymsInHighlight,
		"snippetEllipsisText":        s.SnippetEllipsisText,
		"typoTolerance":              s.TypoTolerance,
		"responseFields":             s.ResponseFields,
		"removeWordsIfNoResults":     s.RemoveWordsIfNoResults,
	}

	// Handle `Distinct` separately as it may be either a `bool` or a `float64`
	// which is in fact a `int`.
	switch v := s.Distinct.(type) {
	case bool:
		m["distinct"] = v
	case float64:
		m["distinct"] = int(v)
	}

	// Handle `IgnorePlurals` separately as it may be either a `bool` or a
	// `[]interface{}` which is in fact a `[]string`.
	switch v := s.IgnorePlurals.(type) {

	case bool:
		m["ignorePlurals"] = v

	case []interface{}:
		var languages []string
		for _, itf := range v {
			lang, ok := itf.(string)
			if ok {
				languages = append(languages, lang)
			} else {
				fmt.Fprintln(os.Stderr, "Settings.ToMap(): `ignorePlurals` slice doesn't only contain strings")
			}
		}
		if len(languages) > 0 {
			m["ignorePlurals"] = languages
		}

	default:
		fmt.Fprintf(os.Stderr, "Settings.ToMap(): Wrong type for `ignorePlurals`: %v\n", reflect.TypeOf(s.IgnorePlurals))

	}

	// Handle `RemoveStopWords` separately as it may be either a `bool` or a
	// `[]interface{}` which is in fact a `[]string`.
	switch v := s.RemoveStopWords.(type) {

	case bool:
		m["removeStopWords"] = v

	case []interface{}:
		var languages []string
		for _, itf := range v {
			lang, ok := itf.(string)
			if ok {
				languages = append(languages, lang)
			} else {
				fmt.Fprintln(os.Stderr, "Settings.ToMap(): `removeStopWords` slice doesn't only contain strings")
			}
		}
		if len(languages) > 0 {
			m["removeStopWords"] = languages
		}

	default:
		fmt.Fprintf(os.Stderr, "Settings.ToMap(): Wrong type for `removeStopWords`: %v\n", reflect.TypeOf(s.RemoveStopWords))

	}

	return m
}

func resourceIndexCreate(d *schema.ResourceData, m interface{}) error {
	client := *m.(*algoliasearch.Client)
	index := client.InitIndex(d.Get("name").(string))
	settings := buildSettingsFromResourceData(d)
	_, err := index.SetSettings(settingsAsMap(settings))
	if err != nil {
		return fmt.Errorf("Error creating index %s: %v", d.Get("name").(string), err)
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceIndexRead(d *schema.ResourceData, m interface{}) error {
	client := *m.(*algoliasearch.Client)
	index := client.InitIndex(d.Id())
	settings, err := index.GetSettings()
	if err != nil && err.Error() == "{\"message\":\"ObjectID does not exist\",\"status\":404}\n" {
		d.SetId("")
		return nil
	}

	readResourceFromSettings(d, settings)
	return nil
}

func resourceIndexUpdate(d *schema.ResourceData, m interface{}) error {
	client := *m.(*algoliasearch.Client)
	index := client.InitIndex(d.Id())
	settings := buildSettingsFromResourceData(d)
	_, err := index.SetSettings(settingsAsMap(settings))
	if err != nil {
		return fmt.Errorf("Error updating index %s: %v", d.Id(), err)
	}

	return nil
}

func resourceIndexDelete(d *schema.ResourceData, m interface{}) error {
	client := *m.(*algoliasearch.Client)
	index := client.InitIndex(d.Get("name").(string))
	_, err := index.Delete()
	return err
}
