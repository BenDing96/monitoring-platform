package pricing

// ModelPrice holds per-token costs for a model in USD.
type ModelPrice struct {
	InputPerMToken  float64 // cost per 1M input tokens
	OutputPerMToken float64 // cost per 1M output tokens
}

// table is the in-process pricing table, keyed by model name / alias.
// Values as of 2025 Q2 — update via Pull Request or future DB-backed table.
var table = map[string]ModelPrice{
	// Claude models
	"claude-opus-4-6":            {InputPerMToken: 15.0, OutputPerMToken: 75.0},
	"claude-sonnet-4-6":          {InputPerMToken: 3.0, OutputPerMToken: 15.0},
	"claude-haiku-4-5-20251001":  {InputPerMToken: 0.8, OutputPerMToken: 4.0},
	"claude-3-5-sonnet-20241022": {InputPerMToken: 3.0, OutputPerMToken: 15.0},
	"claude-3-5-haiku-20241022":  {InputPerMToken: 0.8, OutputPerMToken: 4.0},
	// OpenAI models
	"gpt-4o":      {InputPerMToken: 2.5, OutputPerMToken: 10.0},
	"gpt-4o-mini": {InputPerMToken: 0.15, OutputPerMToken: 0.6},
	"gpt-4-turbo": {InputPerMToken: 10.0, OutputPerMToken: 30.0},
	"o1":          {InputPerMToken: 15.0, OutputPerMToken: 60.0},
	// Gemini models
	"gemini-2.0-flash": {InputPerMToken: 0.1, OutputPerMToken: 0.4},
	"gemini-1.5-pro":   {InputPerMToken: 1.25, OutputPerMToken: 5.0},
}

// Calculate returns the USD cost for the given model and token counts.
// Returns 0 if the model is not in the table.
func Calculate(model string, inputTokens, outputTokens uint32) float64 {
	p, ok := table[model]
	if !ok {
		return 0
	}
	return (float64(inputTokens)*p.InputPerMToken + float64(outputTokens)*p.OutputPerMToken) / 1_000_000
}

// Lookup returns the price entry for a model and whether it was found.
func Lookup(model string) (ModelPrice, bool) {
	p, ok := table[model]
	return p, ok
}
