package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(benchCmd)
}

// benchCmd emits a harness-agnostic benchmark protocol (like `bd onboard` /
// `bd prime` emit protocol text). The driving agent reads it and runs an A/B
// comparison — BeadsLog retrieval vs brute-force grep — on THIS repo's devlog
// history, scored by a blind judge. Works in any harness that has bd + an agent.
var benchCmd = &cobra.Command{
	Use:   "bench [prompt]",
	Short: "Emit an A/B benchmark protocol (BeadsLog retrieval vs brute-force grep)",
	Long: `Print a protocol the current agent runs to benchmark how it answers
project-history questions WITH BeadsLog retrieval vs WITHOUT (brute-force grep),
on this repository's own devlog history.

  bd bench "why did we switch storage?"   # benchmark a specific question
  bd bench                                  # agent generates its own questions

The agent spawns two arms (bd vs grep) plus a blind judge, measures tokens and
tool-calls, and reports honestly — a tie or a loss on a small/greppable corpus is
a valid result.`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		prompt := strings.TrimSpace(strings.Join(args, " "))
		fmt.Println(renderBenchProtocol(prompt))
	},
}

func renderBenchProtocol(prompt string) string {
	var questionSection string
	if prompt != "" {
		questionSection = fmt.Sprintf(`The user gave you ONE question/topic to benchmark:

    %q

If it is a single question, use it as-is. If it is a broad topic, expand it into
3–5 concrete questions about this repo's history that are answerable from the
devlogs.`, prompt)
	} else {
		questionSection = `No question was given — YOU generate the test set:

1. Use ` + "`bd devlog entities`" + ` and ` + "`bd devlog search`" + ` to find topics this
   repo actually has history on (decisions, fixes, abandoned/paused work,
   dependencies).
2. Write 5 questions the way a *teammate* would casually ask them — fuzzy, no
   exact keywords from the source (e.g. "did we ever give up on an approach for
   X?" not "what does function Y do?"). Keyword-exact questions make grep trivial
   and don't test retrieval.
3. Each question must be answerable from the devlogs so both arms can reach it.`
	}

	return fmt.Sprintf(benchProtocolTemplate, questionSection)
}

const benchProtocolTemplate = `# BeadsLog Efficiency Benchmark — Agent Protocol

You will benchmark how efficiently an agent answers project-history questions
**with** BeadsLog retrieval versus **without** it (brute-force grep/read), on
THIS repository's own devlog history. Run it honestly — the point is a real
number, not a favorable one.

## 0. Capability check (do this first, then tell the user)

Can you spawn parallel sub-agents (a Task/Agent tool in this harness)?
- Yes -> run each arm as its own sub-agent with a clean, isolated context. Preferred.
- No  -> tell the user you cannot run a clean parallel A/B here, and offer to run
        the arms sequentially in fresh contexts, or stop. State which mode you use.

## 1. The questions

%s

## 2. The two arms (same questions, same model, isolated contexts)

- **Arm A — BeadsLog:** may use ` + "`bd devlog search / graph / impact / resume`" + ` plus
  reading files. Pass ` + "`--json`" + ` to bd commands for compact output, and ` + "`--no-daemon`" + `.
- **Arm B — brute-force:** may use ONLY ` + "`grep` / `ls` / `read`" + ` over the repo's
  devlog directory (e.g. ` + "`_rules/_devlog/`" + `). No ` + "`bd`" + ` commands at all.

Give both arms the identical questions and the same instruction: answer each in
2–3 sentences, include *why it mattered* and what it affects/depends on, and cite
the source.

## 3. Measure, per arm

- **tokens** — total the arm consumed (from your sub-agent usage report)
- **tool-calls** — number of tool invocations
- **correctness** — if the answer is knowable, how many are right

## 4. Blind judge (a 3rd sub-agent)

Give the judge BOTH answer sets **anonymized and shuffled** as "Set 1" / "Set 2"
— never reveal which used bd. It scores each set 1–5 on: correctness,
significance (conveys *why*), impact/dependency awareness, actionability (would it
steer the next move right). It picks the stronger set and says why.

## 5. Report — honestly

Print a table (arm · tokens · tool-calls · correctness · judge score), the
token ratio and tool-call ratio, and the blind judge's verdict. Then an honest
one-line summary.

DO NOT fabricate a BeadsLog win. On a small or easily-grepped corpus, expect
ties on tokens and answer quality — brute-force grep is a capable baseline when
an agent knows good keywords. BeadsLog's repeatable edge is usually **fewer
tool-calls / round-trips**; its decisive advantages grow with corpus scale, with
questions whose wording differs from the source (semantic recall), and on tasks
grep cannot do at all (cross-session persistence, validation/lifecycle badges,
shared team graph, impact analysis). Report what you actually measured.
`
