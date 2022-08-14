package recipe

import (
	"github.com/kkyr/go-recipe"
	"github.com/kkyr/go-recipe/internal/scraper/custom"

	"github.com/PuerkitoBio/goquery"
)

type recipeScraperFunc func(*goquery.Document) (recipe.Scraper, error)

var hostToScraper = map[string]recipeScraperFunc{
	custom.ForksOverKnivesHost: custom.NewForksOverKnivesScraper,
	custom.MinimalistBakerHost: custom.NewMinimalistBakerScraper,
}
