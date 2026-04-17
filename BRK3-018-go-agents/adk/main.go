package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/artifact"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/server/adkrest"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
)

var recipeFile string
var userPrompt string

type RecipeQuery struct {
	Query string
}

func main() {
	fs := flag.NewFlagSet("myagent", flag.ContinueOnError)
	fs.StringVar(&recipeFile, "recipes", "../recipes.yaml", "Path to the recipes YAML file")
	fs.StringVar(&userPrompt, "prompt", "", "Addional prompt context")
	fs.Parse(os.Args[1:])

	ctx := context.Background()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// build our prompt
	pb := strings.Builder{}
	pb.WriteString("Choose a recipe, giving the recipe name, description, and a two sentence explanation of your choice.\n\n")
	if len(userPrompt) > 0 {
		fmt.Fprintf(&pb, "Additional user context: %s", userPrompt)
	}

	// define recipe-db tool
	var _ functiontool.Func[RecipeQuery, string]
	recipeTool, err := functiontool.New(functiontool.Config{
		Name:        "recipe-db",
		Description: "Read available recipes from the database, given a query",
		InputSchema: &jsonschema.Schema{},
	}, func(ctx tool.Context, rq RecipeQuery) (string, error) {
		return QueryRecipes(rq.Query), nil
	})

	recipeAgent, err := llmagent.New(llmagent.Config{
		Name:        "recipe-agent",
		Model:       model,
		Description: "chooses recipes from a database",
		Instruction: pb.String(),
		Tools: []tool.Tool{
			recipeTool,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// use httptest to make a local http endpoint here.
	agentapi := adkrest.NewHandler(
		&launcher.Config{
			SessionService:  session.InMemoryService(),
			ArtifactService: artifact.InMemoryService(),
			MemoryService:   memory.InMemoryService(),
			AgentLoader:     agent.NewSingleLoader(recipeAgent),
			PluginConfig:    runner.PluginConfig{},
		},
		30*time.Second,
	)
	svr := http.NewServeMux()
	svr.Handle("/api/", http.StripPrefix("/api", agentapi))
	testsvr := httptest.NewServer(agentapi)
	defer testsvr.Close()

	adkclient := &ADKClient{
		App:     "recipe-agent",
		User:    "user1",
		BaseURL: testsvr.URL,
	}
	err = adkclient.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	res, err := adkclient.Run(
		*genai.NewContentFromText(pb.String(), genai.RoleUser))
	if err != nil {
		log.Fatalf("failed to run: %v", err)
	}
	fmt.Println(res[len(res)-1].Content.Parts[0].Text)

}

// QueryRecipes returns the contents of our recipe file
func QueryRecipes(_ string) string {
	// Read the recipes file
	recipesData, _ := os.ReadFile(recipeFile)
	return string(recipesData)

}
