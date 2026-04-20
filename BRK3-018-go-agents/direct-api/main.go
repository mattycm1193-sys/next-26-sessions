package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"
)

var recipeFile = flag.String("recipes", "../recipes.yaml", "Path to the recipes YAML file")
var userPrompt = flag.String("prompt", "", "Addional prompt context")

func main() {
	flag.Parse()

	ctx := context.Background()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	chat, err := client.Chats.Create(ctx, "gemini-2.5-flash",
		&genai.GenerateContentConfig{
			Tools: []*genai.Tool{
				RecipeTool(),
			},
		}, nil)

	// build our prompt
	pb := strings.Builder{}
	pb.WriteString("Choose a recipe, giving the recipe name, description, and a two sentence explanation of your choice.\n\n")
	if len(*userPrompt) > 0 {
		fmt.Fprintf(&pb, "Additional user context: %s", *userPrompt)
	}

	// Ask for a recipe, without providing recipes.
	resp, err := chat.SendMessage(ctx, *genai.NewPartFromText(pb.String()))
	if err != nil {
		log.Fatalf("failed to generate content: %v", err)
	}

	if len(resp.Text()) > 0 {
		fmt.Println(resp.Text())
	}

	// Receive the tool call
	if len(resp.FunctionCalls()) > 0 {
		for _, fn := range resp.FunctionCalls() {
			if fn.Name == "recipe-db" {
				rs := QueryRecipes("")
				// Send the recipe results, and get the next response.
				resp, _ := chat.SendMessage(ctx, genai.Part{Text: rs})
				fmt.Println(resp.Text())
			}
		}
	}
}

func RecipeTool() *genai.Tool {
	t := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Description: "read available recipes from the database",
				Name:        "recipe-db",
			},
		},
	}
	return t
}

// QueryRecipes returns the contents of our recipe file
func QueryRecipes(_ string) string {
	// Read the recipes file
	recipesData, _ := os.ReadFile(*recipeFile)
	return string(recipesData)

}
