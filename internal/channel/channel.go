// Package channel implements release channels: which published versions exist,
// which one a machine follows, and how versions are ordered.
//
// stable follows vX.Y.Z tags. dev additionally follows vX.Y.Z-dev.N tags, so a
// dev machine moves to X.Y.Z when that is released.
package channel

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	Stable = "stable"
	Dev    = "dev"

	// defaultSource is the repository whose tags are the published versions.
	defaultSource = "https://github.com/untoldecay/BeadsLog"
)

var (
	stableRe = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	devRe    = regexp.MustCompile(`^\d+\.\d+\.\d+-dev\.\d+$`)
)

// Parse validates a channel name.
func Parse(name string) (string, error) {
	switch name {
	case Stable, Dev:
		return name, nil
	}
	return "", fmt.Errorf("unknown channel %q (use %q or %q)", name, Stable, Dev)
}

// IsDev reports whether version is a dev pre-release.
func IsDev(version string) bool {
	return devRe.MatchString(strings.TrimPrefix(version, "v"))
}

// Compare orders two versions (with or without "v"): -1, 0 or 1. A pre-release
// sorts before its release, and dev.10 after dev.9.
func Compare(a, b string) int {
	a, b = strings.TrimPrefix(a, "v"), strings.TrimPrefix(b, "v")
	if semver.IsValid("v"+a) && semver.IsValid("v"+b) {
		return semver.Compare("v"+a, "v"+b)
	}
	// Not semver: compare the dotted numbers, ignoring anything else.
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		na, nb := number(pa, i), number(pb, i)
		if na != nb {
			if na < nb {
				return -1
			}
			return 1
		}
	}
	return 0
}

func number(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	digits := parts[i]
	if j := strings.IndexFunc(digits, func(r rune) bool { return r < '0' || r > '9' }); j >= 0 {
		digits = digits[:j]
	}
	n, _ := strconv.Atoi(digits)
	return n
}

// Latest returns the newest of versions on the channel, or "".
func Latest(channel string, versions []string) string {
	best := ""
	for _, v := range versions {
		v = strings.TrimPrefix(v, "v")
		if !stableRe.MatchString(v) && !(channel == Dev && devRe.MatchString(v)) {
			continue
		}
		if best == "" || Compare(v, best) > 0 {
			best = v
		}
	}
	return best
}

// FetchLatest returns the newest published version on the channel.
func FetchLatest(channel string) (string, error) {
	source := os.Getenv("BD_RELEASE_SOURCE") // tests point this at a local repository
	if source == "" {
		source = defaultSource
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", "--refs", source, "v*") // #nosec G204 - fixed command
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("listing releases from %s: %w %s", source, err, strings.TrimSpace(stderr.String()))
	}
	var versions []string
	for _, line := range strings.Split(string(out), "\n") {
		if _, tag, ok := strings.Cut(line, "refs/tags/"); ok {
			versions = append(versions, strings.TrimSpace(tag))
		}
	}
	latest := Latest(channel, versions)
	if latest == "" {
		return "", fmt.Errorf("no %s release found at %s", channel, source)
	}
	return latest, nil
}

// Current returns the channel this machine follows. Unless one was chosen, it
// follows from the running version: a dev build follows dev.
func Current(version string) string {
	if path, err := file(); err == nil {
		if data, err := os.ReadFile(path); err == nil { // #nosec G304 - fixed per-user path
			if c, err := Parse(strings.TrimSpace(string(data))); err == nil {
				return c
			}
		}
	}
	if IsDev(version) {
		return Dev
	}
	return Stable
}

// Save stores the channel for this machine. It is deliberately not in
// .beads/config.yaml, which is committed and shared with the team.
func Save(name string) error {
	if _, err := Parse(name); err != nil {
		return err
	}
	path, err := file()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(name+"\n"), 0600)
}

func file() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", errors.New("no user config directory")
	}
	return filepath.Join(dir, "bd", "channel"), nil
}
