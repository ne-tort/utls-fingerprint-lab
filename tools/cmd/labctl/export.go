package main

import (
	"crypto/sha256"
	"encoding/hex"
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
	Track     string `json:"track,omitempty"`
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

func runExport(labRoot string, cfg *fileRoot, checkDedup bool) error {
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
		if m.ja4 == "" {
			fmt.Fprintf(os.Stderr, "warn: skip %s — empty JA4\n", t.ID)
			continue
		}
		members = append(members, m)
	}

	byFamily := map[string][]rankedMember{}
	for _, m := range members {
		byFamily[m.t.Family] = append(byFamily[m.t.Family], m)
	}

	aliases := map[string]string{}
	for k, v := range defaultAliases {
		aliases[k] = v
	}

	var outProfiles []exportProfile
	families := map[string][]string{}

	for family, list := range byFamily {
		sort.SliceStable(list, func(i, j int) bool {
			return betterMember(list[i], list[j])
		})

		// Dedup within family by JA4: keep best member, alias the rest.
		type uniq struct {
			canonical rankedMember
			dupes     []rankedMember
		}
		var order []string
		byJA4 := map[string]*uniq{}
		for _, m := range list {
			u, ok := byJA4[m.ja4]
			if !ok {
				byJA4[m.ja4] = &uniq{canonical: m}
				order = append(order, m.ja4)
				continue
			}
			u.dupes = append(u.dupes, m)
		}

		var shorts []string
		rank := 0
		for _, ja4 := range order {
			u := byJA4[ja4]
			m := u.canonical
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
				Track:     m.t.Track,
				Builtin:   m.t.Kind == "emit-builtin",
				JA4:       m.ja4,
				Format:    m.format,
			})
			shorts = append(shorts, short)
			// Lab IDs of duplicates → this short name.
			for _, d := range u.dupes {
				aliases[d.t.ID] = short
				fmt.Fprintf(os.Stderr, "dedup %s: %s → %s (same JA4 as %s)\n", family, d.t.ID, short, m.t.ID)
			}
			rank++
		}
		families[family] = shorts
	}

	sort.Slice(outProfiles, func(i, j int) bool {
		if outProfiles[i].Family != outProfiles[j].Family {
			return outProfiles[i].Family < outProfiles[j].Family
		}
		return outProfiles[i].Rank < outProfiles[j].Rank
	})

	if checkDedup {
		seen := map[string]string{} // family|ja4 → short
		for _, p := range outProfiles {
			key := p.Family + "|" + p.JA4
			if prev, ok := seen[key]; ok {
				return fmt.Errorf("dedup check failed: %s and %s share family=%s ja4=%s", prev, p.ShortName, p.Family, p.JA4)
			}
			seen[key] = p.ShortName
		}
		fmt.Fprintf(os.Stderr, "dedup check OK (%d unique profiles)\n", len(outProfiles))
	}

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
		for _, name := range []string{"clienthello.bin", "profile.json", "meta.json", "slot.json"} {
			if err := copyFile(filepath.Join(src, name), filepath.Join(dst, name)); err != nil {
				if name == "meta.json" || name == "slot.json" {
					continue
				}
				return fmt.Errorf("%s: %w", p.LabID, err)
			}
		}
		side, _ := json.MarshalIndent(map[string]any{
			"short_name": p.ShortName,
			"lab_id":     p.LabID,
			"family":     p.Family,
			"rank":       p.Rank,
			"track":      p.Track,
			"ja4":        p.JA4,
		}, "", "  ")
		_ = os.WriteFile(filepath.Join(dst, "short.json"), append(side, '\n'), 0o644)
	}

	cat := exportCatalog{
		Version:     1,
		GeneratedAt: catalogStamp(outProfiles, aliases),
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
	md.WriteString("Within each family, identical JA4 values are collapsed to one short name;\n")
	md.WriteString("duplicate lab IDs appear under `aliases` in `catalog.json`.\n\n")
	md.WriteString("| Short | Lab ID | Family | Rank | Track | Builtin | JA4 |\n")
	md.WriteString("|-------|--------|--------|------|-------|---------|-----|\n")
	for _, p := range outProfiles {
		md.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %d | %s | %v | `%s` |\n",
			p.ShortName, p.LabID, p.Family, p.Rank, p.Track, p.Builtin, p.JA4))
	}
	if err := os.WriteFile(filepath.Join(exportRoot, "NAMES.md"), []byte(md.String()), 0o644); err != nil {
		return err
	}

	fmt.Printf("wrote %s (%d profiles, %d families, %d aliases)\n",
		exportRoot, len(outProfiles), len(families), len(aliases))
	return nil
}

// catalogStamp is a stable content hash so re-export without profile changes
// does not dirty lx-utls-check / git.
func catalogStamp(profiles []exportProfile, aliases map[string]string) string {
	h := sha256.New()
	for _, p := range profiles {
		fmt.Fprintf(h, "p\t%s\t%s\t%s\t%d\t%v\t%s\n",
			p.ShortName, p.LabID, p.Family, p.Rank, p.Builtin, p.JA4)
	}
	keys := make([]string, 0, len(aliases))
	for k := range aliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(h, "a\t%s\t%s\n", k, aliases[k])
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))[:16]
}

// betterMember reports whether a should rank before b (newer / preferred).
func betterMember(a, b rankedMember) bool {
	// track=latest always beats pinned/others at equal kind tier preference.
	la, lb := a.t.Track == trackLatest, b.t.Track == trackLatest
	if la != lb {
		return la
	}
	bi, bj := a.t.Kind == "emit-builtin", b.t.Kind == "emit-builtin"
	if bi != bj {
		return !bi && bj // non-builtin first
	}
	if a.t.Version != b.t.Version {
		return a.t.Version > b.t.Version
	}
	if !a.capturedAt.Equal(b.capturedAt) {
		return a.capturedAt.After(b.capturedAt)
	}
	return a.t.ID < b.t.ID
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
