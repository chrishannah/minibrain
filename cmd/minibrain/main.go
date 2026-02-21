package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chrishannah/minibrain/internal/agent"
	"github.com/chrishannah/minibrain/internal/userconfig"
)

func main() {
	var useCLI bool
	flag.BoolVar(&useCLI, "cli", false, "run in CLI mode")
	flag.Parse()

	if useCLI {
		prompt := strings.TrimSpace(strings.Join(flag.Args(), " "))
		if prompt == "" {
			fmt.Println("usage: minibrain -cli \"I want you to build X\" @file")
			os.Exit(1)
		}

		if handled, err := runCLICommand(prompt); handled {
			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
			fmt.Println("done")
			return
		}

		_, err := runAgent(prompt)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("done")
		return
	}

	runTUI()
}

func runAgent(prompt string) (agent.Result, error) {
	root, err := os.Getwd()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to get working directory: %w", err)
	}
	brainDir, err := agent.ResolveBrainDir()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to resolve brain dir: %w", err)
	}
	perms := agent.ResolvePermissionState(root, readAllowedFromEnv(), writeAllowedFromEnv())
	cfg := buildConfig(root, brainDir, configOptions{
		allowRead:  perms.AllowRead,
		allowWrite: perms.AllowWrite,
	})

	return agent.Run(prompt, cfg)
}

func runAgentWithAllow(prompt string, allowRead, allowWrite bool) (agent.Result, error) {
	root, err := os.Getwd()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to get working directory: %w", err)
	}
	brainDir, err := agent.ResolveBrainDir()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to resolve brain dir: %w", err)
	}
	cfg := buildConfig(root, brainDir, configOptions{
		allowRead:  allowRead,
		allowWrite: allowWrite,
	})

	return agent.Run(prompt, cfg)
}

func runAgentWithAllowAndReads(prompt string, allowRead, allowWrite bool, readPaths []string) (agent.Result, error) {
	root, err := os.Getwd()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to get working directory: %w", err)
	}
	brainDir, err := agent.ResolveBrainDir()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to resolve brain dir: %w", err)
	}
	cfg := buildConfig(root, brainDir, configOptions{
		allowRead:  allowRead,
		allowWrite: allowWrite,
		readPaths:  readPaths,
	})

	return agent.Run(prompt, cfg)
}

func runCLICommand(prompt string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(prompt)) {
	case "/model":
		cfg, err := userconfig.Load()
		if err != nil {
			fmt.Println("model: gpt-4.1")
			return true, nil
		}
		if strings.TrimSpace(cfg.Model) == "" {
			fmt.Println("model: gpt-4.1")
			return true, nil
		}
		fmt.Println("model:", cfg.Model)
		return true, nil
	case "/usage":
		cfg, err := baseConfig()
		if err != nil {
			return true, err
		}
		usage, err := agent.GetUsageStats(cfg)
		if err != nil {
			return true, err
		}
		fmt.Printf("LTM bytes: %d\n", usage.LtmBytes)
		fmt.Printf("STM bytes: %d\n", usage.StmBytes)
		fmt.Printf("STM context bytes: %d\n", usage.StmContextBytes)
		fmt.Printf("Conversation bytes: %d\n", usage.ConvBytes)
		fmt.Printf("Conversation context bytes: %d\n", usage.ConvContextBytes)
		fmt.Printf("Approx tokens: %d/%d\n", usage.ApproxTokens, usage.BudgetTokens)
		return true, nil
	case "/clear":
		cfg, err := baseConfig()
		if err != nil {
			return true, err
		}
		return true, agent.ClearShortTerm(cfg)
	case "/condense":
		cfg, err := baseConfig()
		if err != nil {
			return true, err
		}
		_, err = agent.CondenseShortTerm(cfg)
		return true, err
	default:
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(prompt)), "/model ") {
			fields := strings.Fields(prompt)
			if len(fields) < 2 {
				return true, nil
			}
			newModel := strings.TrimSpace(fields[1])
			if newModel == "" {
				return true, nil
			}
			cfg, err := userconfig.Load()
			if err != nil {
				cfg = userconfig.Config{}
			}
			cfg.Model = newModel
			if err := userconfig.Save(cfg); err != nil {
				return true, err
			}
			fmt.Println("model set:", newModel)
			return true, nil
		}
		return false, nil
	}
}

func readAllowedFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("MINIBRAIN_ALLOW_READ")))
	return v == "1" || v == "true" || v == "yes"
}

func writeAllowedFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("MINIBRAIN_ALLOW_WRITE")))
	return v == "1" || v == "true" || v == "yes"
}
