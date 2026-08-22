# Running Modes: Team, Solo, and Switching

**Problem:** Beads writes issue and devlog state to git. On a shared repo that's great for a team — but a disaster if you push beads onto a work branch (`develop`/`main`) your teammates share, or if you drop beads into a repo whose team never opted in.

**Solution:** Beads never touches your work branch by default, and every mode is reversible. Pick the one that matches how you collaborate.

## 🧭 Which mode?

| You want to… | Mode | Command | Pushed? |
|---|---|---|---|
| Share beads with a team | **Team** | `bd init` (or `bd init --team`) | Yes — to a dedicated `beads-metadata` branch |
| Use beads privately on a shared repo | **Solo · Invisible** | `bd init --solo` → *Invisible* | Never — invisible to the team |
| Keep beads versioned locally, never shared | **Solo · Local branch** | `bd init --solo` → *Local branch* | Never — local `beads-local` branch |
| Commit beads on your current branch | **Inline** | `bd init --inline` | With your normal commits |

Beads data lands on your **work branch** only if you explicitly choose **Inline**.

---

## 1. Team

**Why:** Your team shares one project memory. Everyone's issues and devlogs converge, but must stay off the branches you ship code on.

**How:**
```bash
bd init          # with a git remote, auto-configures the dedicated sync branch
bd onboard       # wire up your agent
```
Teammates just run `bd init` (or clone and `bd sync`) — they inherit the sync branch from the committed config.

**What happens:**
- Beads commits go to a dedicated **`beads-metadata`** branch via an internal worktree — never `develop`/`main`.
- `bd sync` pulls teammates' changes (per-issue last-write-wins merge) and pushes yours; concurrent pushes auto fetch-rebase-retry.
- Share to mainline only if you want: `bd sync --merge` (or a `beads-metadata → main` PR). The tooling reads the branch directly, so this is optional.

> Rename the branch with `bd config set sync-branch <name>`. See [Protected Branches](PROTECTED_BRANCHES.md) and [Sync](SYNC.md).

---

## 2. Solo · Invisible

**Why:** You want beads for yourself on a repo whose team **doesn't use it**. Nothing beads-related should ever reach the remote or a teammate's checkout — not the data, not the config, not the agent instructions.

**How:**
```bash
bd init --solo          # choose: 1) Invisible
                        # then:   graph → Fresh (clean) or Continuity (carry team devlogs)
```

**What happens:**
- The **entire beads footprint** is hidden from git but kept on disk: `.beads/`, the devlog dir, the `<beads_protocol>` block in `AGENTS.md`/`CLAUDE.md`, `_rules/` scaffolding, and the `.gitattributes` merge line (via `.git/info/exclude` + `skip-worktree`). Your agent still reads everything; the team sees nothing.
- Local-only: `no-push`, no daemon auto-sync. `bd doctor` won't nag about unpushed data.
- Solo settings live in a git-excluded `config.local.yaml`, so they never leak into the committed `config.yaml`.

---

## 3. Solo · Local branch

**Why:** Same privacy as Invisible, but you want beads **versioned in git** (history, backup) — just never pushed.

**How:**
```bash
bd init --solo          # choose: 2) Local branch
```

**What happens:**
- Beads commits to a local **`beads-local`** branch (or continues an existing sync branch if detected), never pushed.
- `no-push` + daemon auto-sync off. Everything stays on your machine.

---

## 4. Inline

**Why:** You *want* beads committed alongside your code on the current branch — a small solo repo, or a project where beads **is** the shared work and lives on `main`.

**How:**
```bash
bd init --inline
```

**What happens:**
- No dedicated sync branch. Beads is committed and pushed with your normal commits on the current branch (the pre-sync-branch behavior).
- Only choose this when landing beads on your work branch is intended.

---

## 5. Switch: Team → Solo

**Why:** You're on a team repo but want to go private — experiment without publishing, or the team stopped using beads.

**How:**
```bash
bd init --solo --force   # Invisible + Fresh|Continuity, or Local branch
```

**What happens:**
- **Fresh** starts a clean private graph; **Continuity** copies the team's committed devlogs into your private graph so history carries over (the team dir is never touched).
- All the Invisible hiding (above) is applied. Fully reversible.

---

## 6. Switch: Solo → Team

**Why:** You're ready to share what you did privately, or rejoining a team that uses beads.

**How:**
```bash
bd init --team --force
```

**What happens:**
- Your new solo devlogs are **published** into the team dir; `devlog_dir` and config are restored; the solo excludes are removed; the private `config.local.yaml` is dropped.
- Beads re-points to the dedicated `beads-metadata` branch. Your local issue state merges with the team (per-issue last-write-wins) on the next sync — bd warns you before it does.

---

## Reversibility

Every mode is switchable with `--force`; solo state is per-machine and never committed. Nothing you do locally can strand a teammate, and nothing they do can expose your private work.
