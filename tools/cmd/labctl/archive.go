package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	trackLatest = "latest"
	trackPinned = "pinned"

	imagePull  = "pull"
	imageCache = "cache"

	// Rank sentinel so track=latest always sorts above numeric pins.
	versionLatest = 9999

	archiveFile = "targets.archive.yaml"
)

var majorFromSemver = regexp.MustCompile(`(\d+)`)

func normalizeTarget(t *target) {
	if t.Track == "" {
		t.Track = trackPinned
	}
	if t.Track == trackLatest {
		if t.Version < versionLatest {
			t.Version = versionLatest
		}
		if t.ImagePolicy == "" {
			t.ImagePolicy = imagePull
		}
		return
	}
	if t.ImagePolicy == "" {
		t.ImagePolicy = imageCache
	}
}

func loadLab(labRoot string) (*fileRoot, error) {
	cfg, err := load(filepath.Join(labRoot, "targets.yaml"))
	if err != nil {
		return nil, err
	}
	archPath := filepath.Join(labRoot, archiveFile)
	if b, err := os.ReadFile(archPath); err == nil {
		var arch fileRoot
		if err := yaml.Unmarshal(b, &arch); err != nil {
			return nil, fmt.Errorf("%s: %w", archiveFile, err)
		}
		cfg.Targets = append(cfg.Targets, arch.Targets...)
	}
	for i := range cfg.Targets {
		normalizeTarget(&cfg.Targets[i])
	}
	return cfg, nil
}

type profileSlot struct {
	JA4             string `json:"ja4"`
	SoftwareVersion string `json:"software_version,omitempty"`
	SoftwareMajor   int    `json:"software_major,omitempty"`
	Track           string `json:"track,omitempty"`
	CapturedAt      string `json:"captured_at,omitempty"`
}

func readProfileJA4(profileDir string) string {
	b, err := os.ReadFile(filepath.Join(profileDir, "profile.json"))
	if err != nil {
		return ""
	}
	var p struct {
		Expected struct {
			JA4 string `json:"ja4"`
		} `json:"expected"`
	}
	if json.Unmarshal(b, &p) != nil {
		return ""
	}
	return p.Expected.JA4
}

func readSlot(profileDir string) profileSlot {
	var s profileSlot
	if b, err := os.ReadFile(filepath.Join(profileDir, "slot.json")); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	if s.JA4 == "" {
		s.JA4 = readProfileJA4(profileDir)
	}
	if s.SoftwareMajor == 0 {
		if b, err := os.ReadFile(filepath.Join(profileDir, "meta.json")); err == nil {
			var m struct {
				SoftwareMajor   int    `json:"software_major"`
				SoftwareVersion string `json:"software_version"`
				ResolvedVersion string `json:"resolved_version"`
			}
			if json.Unmarshal(b, &m) == nil {
				s.SoftwareMajor = m.SoftwareMajor
				if s.SoftwareVersion == "" {
					s.SoftwareVersion = m.SoftwareVersion
					if s.SoftwareVersion == "" {
						s.SoftwareVersion = m.ResolvedVersion
					}
				}
				if s.SoftwareMajor == 0 {
					s.SoftwareMajor = parseMajor(s.SoftwareVersion)
				}
			}
		}
	}
	return s
}

func writeSlot(profileDir string, s profileSlot) error {
	s.CapturedAt = time.Now().UTC().Format(time.RFC3339)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(profileDir, "slot.json"), append(b, '\n'), 0o644)
}

func parseMajor(ver string) int {
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return 0
	}
	if m := majorFromSemver.FindStringSubmatch(ver); len(m) > 1 {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 0
}

func snapshotProfile(profileDir string) (string, error) {
	if _, err := os.Stat(filepath.Join(profileDir, "clienthello.bin")); err != nil {
		return "", nil
	}
	tmp, err := os.MkdirTemp("", "utls-slot-*")
	if err != nil {
		return "", err
	}
	for _, name := range []string{"clienthello.bin", "profile.json", "meta.json", "slot.json", "README.md"} {
		src := filepath.Join(profileDir, name)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := copyFile(src, filepath.Join(tmp, name)); err != nil {
			_ = os.RemoveAll(tmp)
			return "", err
		}
	}
	return tmp, nil
}

func maybeArchiveAfterCapture(labRoot string, cfg *fileRoot, t target, beforeDir string) error {
	if beforeDir == "" {
		return nil
	}
	defer os.RemoveAll(beforeDir)
	if t.Track != trackLatest {
		return nil
	}

	profDir := filepath.Join(labRoot, "profiles", t.ID)
	old := readSlot(beforeDir)
	if old.JA4 == "" {
		old.JA4 = readProfileJA4(beforeDir)
	}
	newJA4 := readProfileJA4(profDir)
	if old.JA4 == "" || newJA4 == "" || old.JA4 == newJA4 {
		slot := readSlot(profDir)
		slot.JA4 = newJA4
		slot.Track = trackLatest
		if slot.SoftwareMajor == 0 {
			slot.SoftwareMajor = old.SoftwareMajor
			slot.SoftwareVersion = old.SoftwareVersion
		}
		_ = writeSlot(profDir, slot)
		return nil
	}

	major := old.SoftwareMajor
	if major == 0 {
		major = parseMajor(old.SoftwareVersion)
	}
	archID := ""
	if major == 0 || major >= versionLatest {
		sum := old.JA4
		if len(sum) > 12 {
			sum = sum[:12]
		}
		archID = fmt.Sprintf("%s-prev-%s", t.Family, sanitizeID(sum))
	} else {
		archID = fmt.Sprintf("%s-%d", t.Family, major)
	}
	return archiveSnapshot(labRoot, t, beforeDir, old, archID)
}

func archiveSnapshot(labRoot string, latest target, beforeDir string, old profileSlot, archID string) error {
	archID = sanitizeID(archID)
	dst := filepath.Join(labRoot, "profiles", archID)
	if _, err := os.Stat(filepath.Join(dst, "clienthello.bin")); err == nil {
		existing := readProfileJA4(dst)
		if existing == old.JA4 {
			fmt.Printf("archive: %s already has JA4 (skip copy)\n", archID)
			goto register
		}
		suffix := old.JA4
		if len(suffix) > 12 {
			suffix = suffix[:12]
		}
		archID = fmt.Sprintf("%s-%s", archID, sanitizeID(suffix))
		dst = filepath.Join(labRoot, "profiles", archID)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, name := range []string{"clienthello.bin", "profile.json", "meta.json", "README.md"} {
		src := filepath.Join(beforeDir, name)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := copyFile(src, filepath.Join(dst, name)); err != nil {
			return err
		}
	}
	if b, err := os.ReadFile(filepath.Join(dst, "profile.json")); err == nil {
		var m map[string]any
		if json.Unmarshal(b, &m) == nil {
			m["id"] = archID
			if nb, err := json.MarshalIndent(m, "", "  "); err == nil {
				_ = os.WriteFile(filepath.Join(dst, "profile.json"), append(nb, '\n'), 0o644)
			}
		}
	}
	_ = writeSlot(dst, profileSlot{
		JA4:             old.JA4,
		SoftwareVersion: old.SoftwareVersion,
		SoftwareMajor:   old.SoftwareMajor,
		Track:           trackPinned,
	})
	fmt.Printf("archive: saved previous %s → profiles/%s\n", latest.ID, archID)

register:
	if err := ensureArchiveTarget(labRoot, latest, archID, old); err != nil {
		return err
	}
	slot := readSlot(filepath.Join(labRoot, "profiles", latest.ID))
	slot.JA4 = readProfileJA4(filepath.Join(labRoot, "profiles", latest.ID))
	slot.Track = trackLatest
	_ = writeSlot(filepath.Join(labRoot, "profiles", latest.ID), slot)
	return nil
}

func ensureArchiveTarget(labRoot string, latest target, archID string, old profileSlot) error {
	cfg, err := loadLab(labRoot)
	if err == nil {
		for _, t := range cfg.Targets {
			if t.ID == archID {
				return nil
			}
		}
	}
	major := old.SoftwareMajor
	if major == 0 {
		major = parseMajor(old.SoftwareVersion)
	}
	if major == 0 {
		major = 1
	}
	pin := strconv.Itoa(major)
	entry := fmt.Sprintf(`
  - id: %s
    group: browsers
    kind: wishlist
    status: active
    utls_ready: true
    family: %s
    version: %d
    track: pinned
    pin: %q
    image_policy: cache
    notes: "auto-archived from %s when JA4 changed"
`, archID, latest.Family, major, pin, latest.ID)

	path := filepath.Join(labRoot, archiveFile)
	b, err := os.ReadFile(path)
	if err != nil {
		b = []byte("# Auto-generated pinned archives from track=latest captures.\n# labctl appends here on JA4 change.\n\nversion: 1\ntargets:\n")
	}
	if !strings.Contains(string(b), "targets:") {
		b = append(b, []byte("\ntargets:\n")...)
	}
	b = append(b, []byte(entry)...)
	return os.WriteFile(path, b, 0o644)
}

func sanitizeID(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	out := b.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return strings.Trim(out, "-")
}

func probeSoftwareVersion(labRoot string, t target) (string, int) {
	svc := t.Service
	if svc == "" {
		svc = t.ID
	}
	var cmdArgs []string
	switch t.ID {
	case "chrome-linux":
		cmdArgs = []string{"run", "--rm", "--no-deps", "--entrypoint", "google-chrome", svc, "--version"}
	case "brave":
		cmdArgs = []string{"run", "--rm", "--no-deps", "--entrypoint", "brave-browser", svc, "--version"}
	case "firefox-release":
		cmdArgs = []string{"run", "--rm", "--no-deps", "--entrypoint", "bash", svc, "-lc", "firefox-esr --version 2>/dev/null || true"}
	default:
		return "", 0
	}
	cmd := compose(labRoot, cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if l == "" {
			continue
		}
		fields := strings.Fields(l)
		for _, f := range fields {
			if strings.Count(f, ".") >= 1 {
				if maj := parseMajor(f); maj > 0 {
					return f, maj
				}
			}
		}
		if maj := parseMajor(l); maj > 0 {
			return l, maj
		}
	}
	return "", 0
}

func updateSlotAfterCapture(labRoot string, t target) {
	profDir := filepath.Join(labRoot, "profiles", t.ID)
	if _, err := os.Stat(filepath.Join(profDir, "clienthello.bin")); err != nil {
		return
	}
	slot := readSlot(profDir)
	slot.JA4 = readProfileJA4(profDir)
	slot.Track = t.Track
	if t.Track == trackLatest {
		if ver, maj := probeSoftwareVersion(labRoot, t); maj > 0 {
			slot.SoftwareVersion = ver
			slot.SoftwareMajor = maj
		}
	} else if t.Pin != "" {
		slot.SoftwareVersion = t.Pin
		slot.SoftwareMajor = parseMajor(t.Pin)
	} else if t.Version > 0 && t.Version < versionLatest {
		slot.SoftwareMajor = t.Version
	}
	_ = writeSlot(profDir, slot)
}
