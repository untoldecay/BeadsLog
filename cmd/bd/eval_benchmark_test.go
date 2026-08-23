package main

import (
	"strings"
	"testing"
)

// The emitted protocol must always carry the load-bearing parts (both arms,
// blind judge, honesty guard) and switch its question section on whether a
// prompt was given.
func TestRenderBenchProtocol(t *testing.T) {
	must := func(s, sub, label string) {
		t.Helper()
		if !strings.Contains(s, sub) {
			t.Errorf("protocol missing %s: %q", label, sub)
		}
	}

	// With a prompt: the prompt is embedded, self-generate path is absent.
	withPrompt := renderBenchProtocol("why did we switch storage?")
	must(withPrompt, "why did we switch storage?", "the user's prompt")
	must(withPrompt, "Arm A — BeadsLog", "arm A")
	must(withPrompt, "Arm B — brute-force", "arm B")
	must(withPrompt, "Blind judge", "judge stage")
	must(withPrompt, "DO NOT fabricate", "honesty guard")
	if strings.Contains(withPrompt, "YOU generate the test set") {
		t.Error("prompt path should not include the self-generate instructions")
	}

	// Without a prompt: the agent is told to generate its own questions.
	noPrompt := renderBenchProtocol("")
	must(noPrompt, "YOU generate the test set", "self-generate instructions")
	must(noPrompt, "Arm A — BeadsLog", "arm A")
	must(noPrompt, "DO NOT fabricate", "honesty guard")
}
