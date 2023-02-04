package ld

import (
	"encoding/json"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/piprate/json-gold/ld"
)

const (
	contextKey = "@context"
	graphKey   = "@graph"
	typeKey    = "@type"

	recipeType = "Recipe"
	schemaURL  = "http://schema.org/"

	jsonLdSelector = `script[type="application/ld+json"]`
)

// NewRecipeProcessor returns a RecipeProcessor with default settings.
func NewRecipeProcessor() *RecipeProcessor {
	return &RecipeProcessor{
		proc: ld.NewJsonLdProcessor(),
		opts: ld.NewJsonLdOptions(""),
		ctx: map[string]any{
			contextKey: schemaURL,
			typeKey:    recipeType,
		},
	}
}

// RecipeProcessor is a json-ld Schema Recipe processor.
type RecipeProcessor struct {
	proc *ld.JsonLdProcessor
	opts *ld.JsonLdOptions
	ctx  map[string]any
}

// GetRecipeNode searches doc to find a Schema.org Recipe node encoded in ld+json format.
// If found, the Recipe is serialized into a map. Individual recipe fields can be accessed
// in the map using the field names defined in https://schema.org/Recipe.
func (rp *RecipeProcessor) GetRecipeNode(doc *goquery.Document) (map[string]any, error) {
	jsonLdDocs := doc.Find(jsonLdSelector).Map(func(_ int, sel *goquery.Selection) string {
		return sel.Text()
	})
	if len(jsonLdDocs) == 0 {
		return nil, fmt.Errorf("no ld+json document found")
	}

	var (
		node map[string]any
		err  error
	)

	for _, doc := range jsonLdDocs {
		// Some websites (e.g. AllRecipes.com) have their schema wrapped in a list
		// doc = strings.TrimSpace(doc)
		// doc = strings.TrimPrefix(doc, "[")
		// doc = strings.TrimSuffix(doc, "]")

		if node, err = rp.parseJSON(doc); err == nil {
			return node, nil
		}
	}

	return nil, err
}

func (rp *RecipeProcessor) parseJSON(data string) (map[string]any, error) {
	nodeMap, err := unmarshalJSONObjectOrArray(data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data failed: %w", err)
	}

	var nodes []any
	if isGraphNode(nodeMap) {
		nodes = ld.Arrayify(nodeMap[graphKey])
	} else {
		nodes = ld.Arrayify(nodeMap)
	}

	recipeNode, ok := findRecipeNode(nodes)
	if !ok {
		return nil, fmt.Errorf("could not find Recipe node")
	}

	addSchemaCtx(recipeNode)

	recipeNode, err = rp.proc.Compact(recipeNode, rp.ctx, rp.opts)
	if err != nil {
		return nil, fmt.Errorf("could not compact Recipe node: %w", err)
	}

	return recipeNode, nil
}

func unmarshalJSONObjectOrArray(data string) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(data), &m); err == nil {
		return m, nil
	}

	var nodes []any
	if err := json.Unmarshal([]byte(data), &nodes); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			return nil, fmt.Errorf("unmarshal as array failed at byte offset %d", e.Offset)
		}
		return nil, fmt.Errorf("unmarshal as array failed: %w", err)
	}

	for _, node := range nodes {
		if m, ok := node.(map[string]any); ok {
			return m, nil
		}
	}

	return nil, fmt.Errorf("unable to unmarshal data")
}

func isGraphNode(v any) bool {
	vMap, isMap := v.(map[string]any)
	_, containsGraph := vMap[graphKey]

	return isMap && containsGraph
}

func addSchemaCtx(v any) {
	vMap, isMap := v.(map[string]any)
	_, containsCtx := vMap[contextKey]

	if isMap && !containsCtx {
		vMap[contextKey] = schemaURL
	}
}

func findRecipeNode(nodes []any) (map[string]any, bool) {
	for _, node := range nodes {
		if m, ok := node.(map[string]any); ok {
			str, arr, err := ConvertToStringOrArray(m[typeKey])
			if err != nil {
				return nil, false
			}
			if arr != nil {
				if ArrayContains(arr, recipeType) {
					return m, true
				}
			} else {
				if str == recipeType {
					return m, true
				}
			}
		}
	}

	return nil, false
}

func ConvertToStringOrArray(data any) (string, []interface{}, error) {
	if rt, ok := data.(string); ok {
		return rt, nil, nil
	} else if rt, ok := data.([]interface{}); ok {
		return "", rt, nil
	}

	return "", nil, fmt.Errorf("can't convert to string or array")
}

func ArrayContains(array []any, other string) bool {
	for _, v := range array {
		if v == other {
			return true
		}
	}
	return false
}
