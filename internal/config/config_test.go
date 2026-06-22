package config

import "testing"

func TestMatchMixed(t *testing.T) {
	cfg := Config{
		Enabled:   true,
		MatchMode: "mixed",
		Models:    []string{"nvidia/", "nvidia/nemotron-4-340b-instruct"},
	}

	if !cfg.Match("nvidia/llama-3.1-nemotron-70b-instruct") {
		t.Fatalf("expected prefix match")
	}

	if !cfg.Match("nvidia/nemotron-4-340b-instruct") {
		t.Fatalf("expected exact match")
	}

	if cfg.Match("deepseek-ai/deepseek-v4-flash") {
		t.Fatalf("unexpected match")
	}
}
