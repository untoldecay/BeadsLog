package main

//go:generate go run gen_protocol.go

const ProtocolStartTag = "<beads_protocol>"
const ProtocolEndTag = "</beads_protocol>"

// Legacy tags for migration
const LegacyProtocolStartTag = "<!-- BD_PROTOCOL_START -->"
const LegacyProtocolEndTag = "<!-- BD_PROTOCOL_END -->"

// BootstrapTrap is the pre-onboard trigger line bd writes into agent files at
// init (replaced by the full <beads_protocol> block once 'bd onboard' runs).
// Detecting it lets solo mode hide agent files that carry beads content but
// haven't been finalized yet.
const BootstrapTrap = "BEFORE ANYTHING ELSE: run 'bd onboard'"