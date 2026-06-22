package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Enabled   bool     `toml:"enabled"`
	MatchMode string   `toml:"match_mode"`
	Models    []string `toml:"models"`
}

func Default() Config {
	return Config{
		Enabled:   true,
		MatchMode: "mixed",
		Models:    []string{"nvidia/"},
	}
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := parseTOML(string(data), &cfg); err != nil {
		return Config{}, err
	}
	if cfg.MatchMode == "" {
		cfg.MatchMode = "mixed"
	}
	if len(cfg.Models) == 0 {
		return Config{}, errors.New("whitelist.toml has no models")
	}
	return cfg, nil
}

func parseTOML(src string, cfg *Config) error {
	lines := strings.Split(src, "\n")
	inModels := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "models") && strings.Contains(line, "[") {
			inModels = true
			continue
		}
		if inModels {
			if strings.Contains(line, "]") {
				inModels = false
				continue
			}
			line = strings.TrimSuffix(strings.TrimSpace(line), ",")
			line = strings.Trim(line, "\"")
			if line != "" {
				cfg.Models = append(cfg.Models, line)
			}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "enabled":
			cfg.Enabled = strings.EqualFold(val, "true")
		case "match_mode":
			cfg.MatchMode = strings.Trim(val, "\"")
		case "models":
			// handled by block parser above
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
	}
	return nil
}

func (c Config) Match(model string) bool {
	if !c.Enabled {
		return false
	}
	m := strings.ToLower(strings.TrimSpace(model))
	for _, raw := range c.Models {
		pat := strings.ToLower(strings.TrimSpace(raw))
		switch c.MatchMode {
		case "prefix":
			if strings.HasPrefix(m, pat) {
				return true
			}
		case "exact":
			if m == pat {
				return true
			}
		default:
			if m == pat || strings.HasPrefix(m, pat) {
				return true
			}
		}
	}
	return false
}
