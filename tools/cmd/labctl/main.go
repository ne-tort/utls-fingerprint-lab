package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type fileRoot struct {
	Version  int `yaml:"version"`
	Defaults struct {
		SNISuffix            string `yaml:"sni_suffix"`
		Dial                 string `yaml:"dial"`
		CurlImpersonateImage string `yaml:"curl_impersonate_image"`
		ProfileFormat        string `yaml:"profile_format"`
	} `yaml:"defaults"`
	Targets []target `yaml:"targets"`
}

type target struct {
	ID            string `yaml:"id"`
	Group         string `yaml:"group"`
	Kind          string `yaml:"kind"`
	Service       string `yaml:"service"`
	EmitID        string `yaml:"emit_id"`
	Wrapper       string `yaml:"wrapper"`
	Status        string `yaml:"status"`
	NeedsDNSAlias bool   `yaml:"needs_dns_alias"`
	UTLSReady     bool   `yaml:"utls_ready"`
	Notes         string `yaml:"notes"`
	Why           string `yaml:"why"`
}

func main() {
	root := flag.String("root", ".", "lab root directory")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	labRoot, err := filepath.Abs(*root)
	if err != nil {
		fatal(err)
	}
	cfg, err := load(filepath.Join(labRoot, "targets.yaml"))
	if err != nil {
		fatal(err)
	}

	switch args[0] {
	case "list":
		group, status := "", ""
		for i := 1; i < len(args); i++ {
			if args[i] == "-group" && i+1 < len(args) {
				group = args[i+1]
				i++
			} else if args[i] == "-status" && i+1 < len(args) {
				status = args[i+1]
				i++
			}
		}
		for _, t := range cfg.Targets {
			if group != "" && t.Group != group {
				continue
			}
			if status != "" && t.Status != status {
				continue
			}
			fmt.Printf("%-32s %-16s %-12s %s\n", t.ID, t.Kind, t.Status, t.Group)
		}
	case "capture":
		idFilter, groupFilter := "", ""
		for i := 1; i < len(args); i++ {
			if args[i] == "-id" && i+1 < len(args) {
				idFilter = args[i+1]
				i++
			} else if args[i] == "-group" && i+1 < len(args) {
				groupFilter = args[i+1]
				i++
			}
		}
		if err := ensureCaptureUp(labRoot); err != nil {
			fatal(err)
		}
		fail := 0
		for _, t := range cfg.Targets {
			if t.Status != "active" {
				continue
			}
			if idFilter != "" && t.ID != idFilter {
				continue
			}
			if groupFilter != "" && t.Group != groupFilter {
				continue
			}
			fmt.Printf("\n=== capture %s (%s) ===\n", t.ID, t.Kind)
			if err := runCapture(labRoot, cfg, t); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", t.ID, err)
				fail++
				continue
			}
			fmt.Printf("OK %s\n", t.ID)
		}
		if fail > 0 {
			os.Exit(1)
		}
	case "verify":
		idFilter := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "-id" && i+1 < len(args) {
				idFilter = args[i+1]
				i++
			}
		}
		if err := ensureCaptureUp(labRoot); err != nil {
			fatal(err)
		}
		profilesDir := filepath.Join(labRoot, "profiles")
		entries, err := os.ReadDir(profilesDir)
		if err != nil {
			fatal(err)
		}
		fail := 0
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), "verify-") {
				continue
			}
			if idFilter != "" && e.Name() != idFilter {
				continue
			}
			if _, err := os.Stat(filepath.Join(profilesDir, e.Name(), "clienthello.bin")); err != nil {
				continue
			}
			fmt.Printf("=== verify %s ===\n", e.Name())
			cmd := compose(labRoot, "run", "--rm", "-e", "PROFILE_ID="+e.Name(), "verify-profile")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL verify %s: %v\n", e.Name(), err)
				fail++
			}
		}
		if fail > 0 {
			os.Exit(1)
		}
	case "catalog":
		type row struct {
			ID     string `json:"id"`
			JA4    string `json:"ja4"`
			Format string `json:"format"`
		}
		var rows []row
		profilesDir := filepath.Join(labRoot, "profiles")
		entries, _ := os.ReadDir(profilesDir)
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), "verify-") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(profilesDir, e.Name(), "profile.json"))
			if err != nil {
				continue
			}
			var m struct {
				Format   string `json:"format"`
				Expected struct {
					JA4 string `json:"ja4"`
				} `json:"expected"`
			}
			if json.Unmarshal(b, &m) != nil {
				continue
			}
			rows = append(rows, row{ID: e.Name(), JA4: m.Expected.JA4, Format: m.Format})
		}
		out, _ := json.MarshalIndent(map[string]any{"profile_count": len(rows), "profiles": rows}, "", "  ")
		path := filepath.Join(labRoot, "catalog", "lab-profiles.json")
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
			fatal(err)
		}
		fmt.Printf("wrote %s (%d profiles)\n", path, len(rows))
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `labctl — utls fingerprint lab controller

Usage:
  labctl [-root DIR] list [-group G] [-status S]
  labctl [-root DIR] capture [-id ID] [-group G]
  labctl [-root DIR] verify [-id ID]
  labctl [-root DIR] catalog
`)
}

func load(path string) (*fileRoot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg fileRoot
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Defaults.SNISuffix == "" {
		cfg.Defaults.SNISuffix = ".fp.lab.local"
	}
	if cfg.Defaults.Dial == "" {
		cfg.Defaults.Dial = "capture:8443"
	}
	if cfg.Defaults.CurlImpersonateImage == "" {
		cfg.Defaults.CurlImpersonateImage = "lexiforest/curl-impersonate:alpine"
	}
	return &cfg, nil
}

func compose(labRoot string, args ...string) *exec.Cmd {
	all := append([]string{"compose", "-f", "compose.yaml", "--project-name", "utls-lab"}, args...)
	cmd := exec.Command("docker", all...)
	cmd.Dir = labRoot
	cmd.Env = append(os.Environ(),
		"DOCKER_BUILDKIT=1",
		"COMPOSE_DOCKER_CLI_BUILD=1",
		"COMPOSE_PROFILES=capture,verify,tools",
	)
	return cmd
}

func ensureCaptureUp(labRoot string) error {
	cmd := compose(labRoot, "up", "-d", "--build", "capture")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	bcmd := compose(labRoot, "build", "tools")
	bcmd.Stdout = os.Stdout
	bcmd.Stderr = os.Stderr
	return bcmd.Run()
}

func runCapture(labRoot string, cfg *fileRoot, t target) error {
	switch t.Kind {
	case "compose":
		svc := t.Service
		if svc == "" {
			svc = t.ID
		}
		cmd := compose(labRoot, "run", "--rm", "--build", svc)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "emit-builtin":
		if t.EmitID == "" {
			return fmt.Errorf("emit_id required")
		}
		cmd := compose(labRoot, "run", "--rm", "--no-deps", "--entrypoint", "emit-builtin", "tools",
			"-id", t.EmitID, "-dial", cfg.Defaults.Dial)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "curl-impersonate":
		if t.Wrapper == "" {
			return fmt.Errorf("wrapper required")
		}
		cmd := compose(labRoot, "run", "--rm",
			"-e", "TARGET_ID="+t.ID,
			"-e", "WRAP="+t.Wrapper,
			"curl-impersonate-one")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported kind %q", t.Kind)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
