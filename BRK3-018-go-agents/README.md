## BRK3-018 - Choosing the right stack for agents in Go

This directory contains the complete code used in the Google Cloud Next breakout
session "Choosing the right stack for agents in Go".

### Overview

This session builds a simple recipe suggestion agent, with 3 different
approaches. The code included demonstrates some of the tradeoffs to be
considered when building agents.

### Requirements

The samples require an API Key in the `GEMINI_API_KEY` environment variable.
This API key must have access to the Gemini API, and does not require any other
permissions.

### Usage

In any of the directories containing an agent, `go run . --prompt "your
prompt here"` will run the agent. If used without a prompt, a recipe will be chosen at random.

### Directory Layout

* `direct-api/` contains the fully manual implementation, using AI
services through a simple API client.
* `genkit/` contains the agent built using the [Genkit](http://genkit.dev)
framework
* `adk/` contains the agent built with Google's [Agent Development
Kit](http://adk.dev).
* `recipes.yaml` is our recipe database, queried by all agent variants.

