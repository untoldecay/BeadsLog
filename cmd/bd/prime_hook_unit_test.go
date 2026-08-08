package main

import (
	"os"
	"testing"
)

// TestAgentProtocolPresent underpins the SessionStart dedupe: --hook goes lean
// only when an agent file already carries the <beads_protocol> block.
func TestAgentProtocolPresent(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	if agentProtocolPresent() {
		t.Fatal("no agent files present → expected false")
	}

	// A file that exists but has no protocol block must not count.
	if err := os.WriteFile("CLAUDE.md", []byte("just user notes\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if agentProtocolPresent() {
		t.Fatal("agent file without protocol block → expected false")
	}

	// Now add the protocol block.
	body := "intro\n" + ProtocolStartTag + "\nprotocol\n" + ProtocolEndTag + "\noutro\n"
	if err := os.WriteFile("CLAUDE.md", []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	if !agentProtocolPresent() {
		t.Fatal("agent file with protocol block → expected true")
	}
}
