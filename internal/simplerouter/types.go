package simplerouter

type Config struct {
	OpenRouterAPIKey string `json:"openrouter_api_key,omitempty"`
	LastModel        string `json:"last_model,omitempty"`
}

type Model struct {
	ID                  string
	Name                string
	ContextLength       int
	PromptPrice         string
	OutputPrice         string
	SupportedParameters []string
}

// Endpoint is one provider serving a model (from /models/:id/endpoints).
type Endpoint struct {
	ProviderName  string
	Tag           string // OpenRouter routing slug, e.g. "deepinfra/fp4"
	Quantization  string
	ContextLength int
	PromptPrice   string
	OutputPrice   string
}

type openRouterEndpointsResponse struct {
	Data struct {
		Endpoints []struct {
			ProviderName  string `json:"provider_name"`
			Tag           string `json:"tag"`
			Quantization  string `json:"quantization"`
			ContextLength int    `json:"context_length"`
			Pricing       struct {
				Prompt     string `json:"prompt"`
				Completion string `json:"completion"`
			} `json:"pricing"`
		} `json:"endpoints"`
	} `json:"data"`
}

type openRouterModelsResponse struct {
	Data []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		ContextLength int    `json:"context_length"`
		Pricing       struct {
			Prompt     string `json:"prompt"`
			Completion string `json:"completion"`
		} `json:"pricing"`
		SupportedParameters []string `json:"supported_parameters"`
	} `json:"data"`
}

type openRouterKeyResponse struct {
	Data map[string]any `json:"data"`
}

type launchSpec struct {
	Path string
	Dir  string
	Args []string
	Env  []string
}
