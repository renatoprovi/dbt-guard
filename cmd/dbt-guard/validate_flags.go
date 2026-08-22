package main

import (
	"fmt"
	"strings"

	"github.com/renatoprovi/dbt-guard/internal/config"
)

type validateOptions struct {
	ManifestPath string
	ConfigPath   string
	Allowed      []string
	Restricted   []string
}

func parseValidateArgs(args []string) (validateOptions, error) {
	var opts validateOptions
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--config":
			if i+1 >= len(args) {
				return validateOptions{}, fmt.Errorf("--config requires a path")
			}
			i++
			opts.ConfigPath = args[i]
		case "--allowed":
			if i+1 >= len(args) {
				return validateOptions{}, fmt.Errorf("--allowed requires a comma-separated list")
			}
			i++
			opts.Allowed = splitCSV(args[i])
		case "--restricted":
			if i+1 >= len(args) {
				return validateOptions{}, fmt.Errorf("--restricted requires a comma-separated list")
			}
			i++
			opts.Restricted = splitCSV(args[i])
		default:
			if strings.HasPrefix(arg, "-") {
				return validateOptions{}, fmt.Errorf("unknown flag: %s", arg)
			}
			positional = append(positional, arg)
		}
	}
	if len(positional) != 1 {
		return validateOptions{}, fmt.Errorf("usage: dbt-guard validate [--config path] [--allowed layers] [--restricted layers] <manifest.json>")
	}
	opts.ManifestPath = positional[0]
	return opts, nil
}

func resolveLayerPolicy(opts validateOptions) (config.LayerPolicy, error) {
	policy := config.DefaultLayerPolicy()
	if opts.ConfigPath != "" {
		loaded, err := config.LoadLayerPolicy(opts.ConfigPath)
		if err != nil {
			return config.LayerPolicy{}, err
		}
		policy = loaded
	}
	if len(opts.Allowed) > 0 || len(opts.Restricted) > 0 {
		policy = policy.WithCLIOverrides(opts.Allowed, opts.Restricted)
	}
	return policy, nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
