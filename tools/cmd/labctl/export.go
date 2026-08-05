package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Legacy upstream fingerprint aliases → newest short name of that family.
var defaultAliases = map[string]string{
	"":                           "chrome",
	"chrome_psk":                 "chrome",
	"chrome_psk_shuffle":         "chrome",
	"chrome_padding_psk_shuffle": "chrome",
	"chrome_pq":                  "chrome",
	"chrome_pq_psk":              "chrome",
}

type exportCatalog struct {
	Version     int                       `json:"version"`
	GeneratedAt string                    `json:"generated_at"`
	Profiles    []exportProfile           `json:"profiles"`
	Aliases     map[string]string         `json:"aliases"`
	Families    map[string][]string       `json:"families"`
}

type exportProfile struct {
	ShortName string `json:"short_name"`
	LabID     string `json:"lab_id"`
	Family    string `json:"family"`
	Rank      int    `json:"rank"`
	Version   int    `json:"version"`
	Builtin   bool   `json:"builtin"`
	JA4       string `json:"ja4,omitempty"`
	Format    string `json:"format,omitempty"`
}

type rankedMember struct {
	t          target
	capturedAt time.Time
	ja4        string
	format     string
	hasProfile bool
}

func runExport(labRoot string, cfg *fileRoot) error {
	profilesDir := filepath.Join(labRoot, "profiles")
	var members []rankedMember
	for _, t := range cfg.Targets {
		if t.Status != "active" || !t.UTLSReady {
			continue
		}
		if t.Family == "" {
			return fmt.Errorf("target %s: family required for export", t.ID)
		}
		m := rankedMember{t: t}
		metaPath := filepath.Join(profilesDir, t.ID, "meta.json")
		profPath := filepath.Join(profilesDir, t.ID, "profile.json")
		binPath := filepath.Join(profilesDir, t.ID, "clienthello.bin")
		if _, err := os.Stat(binPath); err == nil {
			m.hasProfile = true
		}
		if b, err := os.ReadFile(metaPath); err == nil {
			var meta struct {
				CapturedAt string `json:"captured_at"`
				JA4        string `json:"ja4"`
			}
			if json.Unmarshal(b, &meta) == nil {
				if ts, err := time.Parse(time.RFC3339, meta.CapturedAt); err == nil {
					m.capturedAt = ts
				}
				m.ja4 = meta.JA4
			}
		}
		if b, err := os.ReadFile(profPath); err == nil {
			var p struct {
				Format   string `json:"format"`
				Expected struct {
					JA4 string `json:"ja4"`
				} `json:"expected"`
			}
			if json.Unmarshal(b, &p) == nil {
				m.format = p.Format
				if m.ja4 == "" {
					m.ja4 = p.Expected.JA4
				}
			}
		}
		if !m.hasProfile {
			fmt.Fprintf(os.Stderr, "warn: skip %s — no clienthello.bin\n", t.ID)
			continue
		}
		members = append(members, m)
	}

	byFamily := map[string][]rankedMember{}
	for _, m := range members {
		byFamily[m.t.Family] = append(byFamily[m.t.Family], m)
	}

	var outProfiles []exportProfile
	families := map[string][]string{}
	for family, list := range byFamily {
		sort.SliceStable(list, func(i, j int) bool {
			bi, bj := list[i].t.Kind == "emit-builtin", list[j].t.Kind == "emit-builtin"
			if bi != bj {
				return !bi && bj // non-builtin first (newer side)
			}
			if list[i].t.Version != list[j].t.Version {
				return list[i].t.Version > list[j].t.Version
			}
			if !list[i].capturedAt.Equal(list[j].capturedAt) {
				return list[i].capturedAt.After(list[j].capturedAt)
			}
			return list[i].t.ID < list[j].t.ID
		})
		var shorts []string
		for rank, m := range list {
			short := family
			if rank > 0 {
				short = fmt.Sprintf("%s-%d", family, rank)
			}
			outProfiles = append(outProfiles, exportProfile{
				ShortName: short,
				LabID:     m.t.ID,
				Family:    family,
				Rank:      rank,
				Version:   m.t.Version,
				Builtin:   m.t.Kind == "emit-builtin",
				JA4:       m.ja4,
				Format:    m.format,
			})
			shorts = append(shorts, short)
		}
		families[family] = shorts
	}

	sort.Slice(outProfiles, func(i, j int) bool {
		if outProfiles[i].Family != outProfiles[j].Family {
			return outProfiles[i].Family < outProfiles[j].Family
		}
		return outProfiles[i].Rank < outProfiles[j].Rank
	})

	aliases := map[string]string{}
	for k, v := range defaultAliases {
		aliases[k] = v
	}
	// Stock family names without suffix already point at newest; keep for docs.

	exportRoot := filepath.Join(labRoot, "dist", "export")
	if err := os.RemoveAll(exportRoot); err != nil {
		return err
	}
	profOut := filepath.Join(exportRoot, "profiles")
	if err := os.MkdirAll(profOut, 0o755); err != nil {
		return err
	}

	for _, p := range outProfiles {
		src := filepath.Join(profilesDir, p.LabID)
		dst := filepath.Join(profOut, p.ShortName)
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		for _, name := range []string{"clienthello.bin", "profile.json", "meta.json"} {
			if err := copyFile(filepath.Join(src, name), filepath.Join(dst, name)); err != nil {
				if name == "meta.json" {
					continue
				}
				return fmt.Errorf("%s: %w", p.LabID, err)
			}
		}
		// Stamp short name into a sidecar for consumers.
		side, _ := json.MarshalIndent(map[string]any{
			"short_name": p.ShortName,
			"lab_id":     p.LabID,
			"family":     p.Family,
			"rank":       p.Rank,
		}, "", "  ")
		_ = os.WriteFile(filepath.Join(dst, "short.json"), append(side, '\n'), 0o644)
	}

	cat := exportCatalog{
		Version:     1,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Profiles:    outProfiles,
		Aliases:     aliases,
		Families:    families,
	}
	catBytes, err := json.MarshalIndent(cat, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(exportRoot, "catalog.json"), append(catBytes, '\n'), 0o644); err != nil {
		return err
	}

	var md strings.Builder
	md.WriteString("# Short fingerprint names\n\n")
	md.WriteString("| Short | Lab ID | Family | Rank | Builtin | JA4 |\n")
	md.WriteString("|-------|--------|--------|------|---------|-----|\n")
	for _, p := range outProfiles {
		md.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %d | %v | `%s` |\n",
			p.ShortName, p.LabID, p.Family, p.Rank, p.Builtin, p.JA4))
	}
	if err := os.WriteFile(filepath.Join(exportRoot, "NAMES.md"), []byte(md.String()), 0o644); err != nil {
		return err
	}

	fmt.Printf("wrote %s (%d profiles, %d families)\n", exportRoot, len(outProfiles), len(families))
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
