package main

//go:generate go run gen_protocol.go

const ProtocolStartTag = "<beads_protocol>"
const ProtocolEndTag = "</beads_protocol>"

// Legacy tags for migration
const LegacyProtocolStartTag = "<!-- BD_PROTOCOL_START -->"
const LegacyProtocolEndTag = "<!-- BD_PROTOCOL_END -->"