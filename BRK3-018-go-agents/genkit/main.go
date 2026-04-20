package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

var recipeFile = flag.String("recipes", "../recipes.yaml", "Path to the recipes YAML file")
var userPrompt = flag.String("prompt", "", "Addional prompt context")

type RecipeQuery struct {
	Query string
}

func main() {
	flag.Parse()

	ctx := context.Background()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}
	g := genkit.Init(context.Background(),
		genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: apiKey}),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash"))

	recipeTool := genkit.DefineTool(g, "recipe-db",
		"Read available recipes from the database, given a query",
		func(ctx *ai.ToolContext, rq RecipeQuery) (string, error) {
			return QueryRecipes(rq.Query), nil
		},
	)

	// build our prompt
	pb := strings.Builder{}
	pb.WriteString("Choose a recipe, giving the recipe name, description, and a two sentence explanation of your choice.\n\n")
	if len(*userPrompt) > 0 {
		fmt.Fprintf(&pb, "Additional user context: %s", *userPrompt)
	}

	resp, err := genkit.Generate(ctx, g,
		ai.WithPrompt(pb.String()),
		ai.WithTools(recipeTool))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Text())

}

// QueryRecipes returns the contents of our recipe file
func QueryRecipes(_ string) string {
	// Read the recipes file
	recipesData, _ := os.ReadFile(*recipeFile)
	return string(recipesData)

}
