// Package devlog answers where a devlog file lives across a repository's
// branches and worktrees.
//
// All worktrees of a repository share one beads.db, but devlog markdown files
// are versioned per branch. "The file is not in this checkout" therefore does
// not mean "the file is gone", and only the latter makes a session a ghost.
package devlog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/untoldecay/BeadsLog/internal/beads"
)

// State is where a devlog file was found.
type State int

const (
	// Absent: the file is in no checkout and on no branch tip.
	Absent State = iota
	// Here: the file is on disk in the current checkout.
	Here
	// Elsewhere: not on disk here, but on a branch tip or in another worktree.
	Elsewhere
	// Unknown: git could not be queried, so absence cannot be proven.
	// Callers must not mark or delete anything on Unknown.
	Unknown
)

// Presence is a lazily built snapshot of every devlog file git knows about.
// git is only queried when a file is missing from the current checkout.
type Presence struct {
	dir string // devlog directory as seen from the current checkout

	loaded   bool
	inRepo   bool
	err      error
	toplevel string
	prefix   string            // dir relative to toplevel, "" or "a/b/"
	blobs    map[string]string // basename -> git object id, on branch tips
	blobRef  map[string]string // basename -> where that blob was found
	files    map[string]string // basename -> path in another worktree

	// For files linked outside the devlog directory (e.g. "../docs/x.md"):
	// the other worktrees and every branch tip, searched by repo path.
	worktrees []string          // other worktree roots
	tips      []string          // commit ids, sorted
	tipLabel  map[string]string // commit id -> ref or worktree label
	pathCache map[string]pathHit

	restriction string
}

// NewPresence returns a snapshot for the given devlog directory.
func NewPresence(devlogDir string) *Presence {
	return &Presence{dir: devlogDir}
}

type pathHit struct {
	file   string // path in another worktree
	blob   string // object id on a branch tip
	source string
}

// ResolvePath returns the on-disk path of an index link or session filename,
// relative to the devlog directory dir. A plain name is a devlog file; a link
// with directories ("../_documentation/x.md", "sub/x.md") is followed, with
// the old base-name reading kept for links that repeat the directory
// ("_rules/_devlog/x.md").
func ResolvePath(dir, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	base := filepath.Join(dir, filepath.Base(filename))
	if filepath.Base(filename) == filename {
		return base
	}
	full := filepath.Join(dir, filename)
	if _, err := os.Stat(full); err == nil {
		return full
	}
	if _, err := os.Stat(base); err == nil {
		return base
	}
	if repeatsDir(dir, filename) {
		return base
	}
	return full
}

// repeatsDir reports whether filename already starts with the devlog
// directory, as in "_rules/_devlog/x.md" for dir "_rules/_devlog".
func repeatsDir(dir, filename string) bool {
	d := filepath.ToSlash(filepath.Clean(dir))
	f := filepath.ToSlash(filepath.Clean(filename))
	return d != "." && strings.HasPrefix(f, d+"/")
}

// nested reports whether filename points outside the plain devlog directory
// listing, so lookups must use its full path rather than its base name.
func (p *Presence) nested(filename string) bool {
	if filepath.IsAbs(filename) {
		dir, err := filepath.Abs(p.dir)
		if err != nil {
			return true
		}
		return canonicalAbove(filepath.Dir(filename)) != canonicalAbove(dir)
	}
	return filepath.Base(filename) != filename && !repeatsDir(p.dir, filename)
}

// Lookup reports where the devlog file lives. A plain name is looked up in
// the devlog directory of every worktree and branch tip; a link that leaves
// the directory ("../_documentation/x.md") is looked up by its own path.
func (p *Presence) Lookup(filename string) State {
	if p.nested(filename) {
		return p.lookupPath(filename)
	}
	base := filepath.Base(filename)
	if _, err := os.Stat(filepath.Join(p.dir, base)); err == nil {
		return Here
	}
	p.load()
	if p.err != nil {
		return Unknown
	}
	if _, ok := p.files[base]; ok {
		return Elsewhere
	}
	if _, ok := p.blobs[base]; ok {
		return Elsewhere
	}
	return Absent
}

// lookupPath finds a file linked outside the devlog directory: on disk here,
// then at the same repository path in the other worktrees and branch tips.
func (p *Presence) lookupPath(filename string) State {
	if _, err := os.Stat(ResolvePath(p.dir, filename)); err == nil {
		return Here
	}
	p.load()
	if p.err != nil {
		return Unknown
	}
	if !p.inRepo {
		return Absent // plain directory: disk is the only truth
	}
	rel, ok := p.repoPath(filename)
	if !ok {
		return Unknown // outside the repository: git cannot prove it is gone
	}
	hit, err := p.findPath(rel)
	switch {
	case err != nil:
		return Unknown
	case hit.file != "" || hit.blob != "":
		return Elsewhere
	}
	return Absent
}

// repoPath turns a filename into a slash path relative to the repository
// root, or false if it points outside the repository.
func (p *Presence) repoPath(filename string) (string, bool) {
	var rel string
	if filepath.IsAbs(filename) {
		r, err := filepath.Rel(canonical(p.toplevel), canonicalAbove(filepath.Dir(filename)))
		if err != nil {
			return "", false
		}
		rel = filepath.ToSlash(filepath.Join(r, filepath.Base(filename)))
	} else {
		rel = filepath.ToSlash(filepath.Clean(filepath.Join(filepath.FromSlash(p.prefix), filename)))
	}
	if rel == ".." || strings.HasPrefix(rel, "../") || rel == "." {
		return "", false
	}
	return rel, true
}

func (p *Presence) findPath(rel string) (pathHit, error) {
	if hit, ok := p.pathCache[rel]; ok {
		return hit, nil
	}
	hit := pathHit{}
	for _, wt := range p.worktrees {
		path := filepath.Join(wt, filepath.FromSlash(rel))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			hit = pathHit{file: path, source: path}
			break
		}
	}
	if hit.file == "" && len(p.tips) > 0 {
		var in bytes.Buffer
		for _, id := range p.tips {
			fmt.Fprintf(&in, "%s:%s\n", id, rel)
		}
		out, err := gitInput(p.toplevel, &in, "cat-file", "--batch-check")
		if err != nil {
			return hit, err
		}
		for i, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f := strings.Fields(line); len(f) == 3 && f[1] == "blob" && i < len(p.tips) {
				hit = pathHit{blob: f[0], source: p.tipLabel[p.tips[i]]}
				break
			}
		}
	}
	p.pathCache[rel] = hit
	return hit, nil
}

// Read returns the content of a file that is Elsewhere and a label saying
// where it came from (a worktree path or a ref).
func (p *Presence) Read(filename string) (content []byte, source string, err error) {
	base := filepath.Base(filename)
	p.load()
	if p.err != nil {
		return nil, "", p.err
	}
	if p.nested(filename) {
		rel, ok := p.repoPath(filename)
		if !ok || !p.inRepo {
			return nil, "", os.ErrNotExist
		}
		hit, err := p.findPath(rel)
		switch {
		case err != nil:
			return nil, "", err
		case hit.file != "":
			content, err = os.ReadFile(hit.file) // #nosec G304 - path enumerated from git worktrees
			return content, hit.source, err
		case hit.blob != "":
			content, err = git(p.toplevel, "cat-file", "blob", hit.blob)
			return content, hit.source, err
		}
		return nil, "", os.ErrNotExist
	}
	if path, ok := p.files[base]; ok {
		if content, err = os.ReadFile(path); err == nil { // #nosec G304 - path enumerated from git worktrees
			return content, path, nil
		}
	}
	if id, ok := p.blobs[base]; ok {
		content, err = git(p.toplevel, "cat-file", "blob", id)
		return content, p.blobRef[base], err
	}
	return nil, "", os.ErrNotExist
}

// Restriction explains why this checkout must not decide that a session is
// gone, or returns "" if it may. Linked worktrees, and checkouts redirected to
// another repository's database, only see part of what the shared database
// covers.
func (p *Presence) Restriction() string {
	p.load()
	return p.restriction
}

// Err reports why git could not be queried, or nil.
func (p *Presence) Err() error {
	p.load()
	return p.err
}

func (p *Presence) load() {
	if p.loaded {
		return
	}
	p.loaded = true
	p.blobs = map[string]string{}
	p.blobRef = map[string]string{}
	p.files = map[string]string{}
	p.tipLabel = map[string]string{}
	p.pathCache = map[string]pathHit{}
	p.err = p.scan()
}

func (p *Presence) scan() error {
	// The directory may be missing on this branch: start from the nearest
	// ancestor that exists.
	abs, err := filepath.Abs(p.dir)
	if err != nil {
		return err
	}
	rest := ""
	start := abs
	for {
		if info, statErr := os.Stat(start); statErr == nil && info.IsDir() {
			break
		}
		parent := filepath.Dir(start)
		if parent == start {
			return fmt.Errorf("no existing directory above %s", abs)
		}
		rest = filepath.Join(filepath.Base(start), rest)
		start = parent
	}

	inside, err := insideWorkTree(start)
	if err != nil {
		return err
	}
	if !inside {
		return nil // plain directory: disk is the only truth
	}
	p.inRepo = true

	top, err := gitLine(start, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	p.toplevel = filepath.FromSlash(top)
	sub, err := gitLine(start, "rev-parse", "--show-prefix")
	if err != nil {
		return err
	}
	p.prefix = sub
	if rest != "" {
		p.prefix += filepath.ToSlash(rest) + "/"
	}

	if err := p.scanRestriction(); err != nil {
		return err
	}
	commits, err := p.scanWorktrees()
	if err != nil {
		return err
	}
	if err := p.scanRefs(commits); err != nil {
		return err
	}
	return nil
}

func (p *Presence) scanRestriction() error {
	gitDir, err := gitLine(p.toplevel, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return err
	}
	common, err := commonDir(p.toplevel)
	if err != nil {
		return err
	}
	if canonical(gitDir) != common {
		// A bare repository has no main checkout to defer to; there the
		// branch/worktree scan alone decides.
		bare, _ := gitLine(common, "rev-parse", "--is-bare-repository")
		if bare != "true" {
			p.restriction = "this is a linked git worktree"
			return nil
		}
	}

	local := filepath.Join(p.toplevel, ".beads")
	if _, err := os.Stat(filepath.Join(local, beads.RedirectFileName)); err == nil {
		if target := beads.FollowRedirect(local); target != local {
			if other, err := commonDir(target); err == nil && other != common {
				p.restriction = "this checkout is redirected to another repository's database"
			}
		}
	}
	return nil
}

// scanWorktrees records devlog files on disk in the other worktrees (which
// covers uncommitted devlogs) and returns every worktree's HEAD (which covers
// detached ones).
func (p *Presence) scanWorktrees() (map[string]string, error) {
	out, err := git(p.toplevel, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	commits := map[string]string{}
	self := canonical(p.toplevel)
	path := ""
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			path = strings.TrimPrefix(line, "worktree ")
			if canonical(path) == self {
				continue
			}
			p.worktrees = append(p.worktrees, path)
			dir := filepath.Join(path, filepath.FromSlash(p.prefix))
			entries, err := os.ReadDir(dir)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				if !e.IsDir() {
					p.files[e.Name()] = filepath.Join(dir, e.Name())
				}
			}
		case strings.HasPrefix(line, "HEAD "):
			id := strings.TrimPrefix(line, "HEAD ")
			if strings.Trim(id, "0") != "" { // unborn branches report all zeros
				commits[id] = "worktree " + path
			}
		}
	}
	return commits, nil
}

// scanRefs lists the devlog directory on every local and remote branch tip.
// Tips, not history: a devlog that was committed and later deleted everywhere
// really is gone.
func (p *Presence) scanRefs(commits map[string]string) error {
	out, err := git(p.toplevel, "for-each-ref", "--format=%(objectname) %(refname)", "refs/heads", "refs/remotes")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if id, ref, ok := strings.Cut(line, " "); ok {
			commits[id] = shortRef(ref)
		}
	}
	if len(commits) == 0 {
		return nil
	}
	for id, label := range commits {
		p.tips = append(p.tips, id)
		p.tipLabel[id] = label
	}
	sort.Strings(p.tips)

	// Resolve <commit>:<dir> for all tips in one call, then list each distinct
	// tree once: most branches share the same devlog directory.
	ids := make([]string, 0, len(commits))
	for id := range commits {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var in bytes.Buffer
	for _, id := range ids {
		fmt.Fprintf(&in, "%s:%s\n", id, strings.TrimSuffix(p.prefix, "/"))
	}
	out, err = gitInput(p.toplevel, &in, "cat-file", "--batch-check")
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != len(ids) {
		return fmt.Errorf("git cat-file: expected %d answers, got %d", len(ids), len(lines))
	}
	seen := map[string]bool{}
	for i, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[1] != "tree" || seen[fields[0]] {
			continue // "<name> missing": no devlog directory on that tip
		}
		seen[fields[0]] = true
		tree, err := git(p.toplevel, "ls-tree", "-z", fields[0])
		if err != nil {
			return err
		}
		for _, entry := range strings.Split(string(tree), "\x00") {
			meta, name, ok := strings.Cut(entry, "\t")
			f := strings.Fields(meta)
			if ok && len(f) == 3 && f[1] == "blob" {
				if _, dup := p.blobs[name]; !dup {
					p.blobs[name] = f[2]
					p.blobRef[name] = commits[ids[i]]
				}
			}
		}
	}
	return nil
}

func shortRef(ref string) string {
	ref = strings.TrimPrefix(ref, "refs/heads/")
	return strings.TrimPrefix(ref, "refs/remotes/")
}

func canonical(path string) string {
	path = filepath.FromSlash(path)
	if r, err := filepath.EvalSymlinks(path); err == nil {
		path = r
	}
	return filepath.Clean(path)
}

// canonicalAbove canonicalizes a path that may not exist (a directory that
// is only on another branch): it resolves the nearest existing ancestor.
func canonicalAbove(path string) string {
	path = filepath.Clean(path)
	rest := ""
	for d := path; ; d = filepath.Dir(d) {
		if _, err := os.Lstat(d); err == nil {
			return filepath.Join(canonical(d), rest)
		}
		if filepath.Dir(d) == d {
			return path
		}
		rest = filepath.Join(filepath.Base(d), rest)
	}
}

// commonDir returns the canonical shared git directory for dir.
func commonDir(dir string) (string, error) {
	c, err := gitLine(dir, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}
	c = filepath.FromSlash(c)
	if !filepath.IsAbs(c) {
		c = filepath.Join(dir, c)
	}
	return canonical(c), nil
}

// insideWorkTree distinguishes "not a repository" (false, nil) from a git
// failure, which must not be mistaken for it.
func insideWorkTree(dir string) (bool, error) {
	out, err := gitLine(dir, "rev-parse", "--is-inside-work-tree")
	if err == nil {
		return out == "true", nil
	}
	for d := canonical(dir); ; d = filepath.Dir(d) {
		if _, statErr := os.Lstat(filepath.Join(d, ".git")); statErr == nil {
			return false, err // there is a repository, git just failed on it
		}
		if filepath.Dir(d) == d {
			return false, nil
		}
	}
}

func gitLine(dir string, args ...string) (string, error) {
	out, err := git(dir, args...)
	return strings.TrimRight(string(out), "\r\n"), err
}

func git(dir string, args ...string) ([]byte, error) {
	return gitInput(dir, nil, args...)
}

func gitInput(dir string, stdin *bytes.Buffer, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...) // #nosec G204 - fixed subcommands, no shell
	cmd.Dir = dir
	// Git hooks export GIT_DIR & co, which would override the directory.
	for _, kv := range os.Environ() {
		switch name, _, _ := strings.Cut(kv, "="); name {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_PREFIX", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY":
		default:
			cmd.Env = append(cmd.Env, kv)
		}
	}
	if stdin != nil {
		cmd.Stdin = stdin
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}
