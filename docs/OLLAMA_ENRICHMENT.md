# Ollama AI Enrichment Configuration

BeadsLog uses a tiered extraction pipeline to transform your devlog narratives into architectural graphs. While the "Tier 1" Regex extractor is always active, "Tier 2" Ollama enrichment provides deep semantic analysis in the background.

## Initial Setup

When you initialize a new project with `bd init`, the setup wizard will offer to enable background AI enrichment.

### Interactive Model Selection
If you choose "Yes", BeadsLog will:
1. **Detect Ollama:** Verify that Ollama is installed and running.
2. **List Models:** Fetch all available models from your local Ollama instance.
3. **Selection Menu:** Allow you to choose which model to use for extraction.
4. **Auto-Pull:** If no models are found, offer to automatically pull a lightweight model (`llama3.2:1b`).

This ensures that the AI enrichment is correctly configured from day one.

## Core Configuration

The AI enrichment system is managed via the unified `bd config` interface. These settings are persistent and stored in your project's `.beads/config.yaml`.

### Enable/Disable Enrichment

To toggle the entire extraction system (including Regex fallback):

> [!CAUTION]
**Disabling the extraction system is not recommended.** Turning off the Regex tier (`entity_extraction.enabled = false`) defeats the primary purpose of BeadsLog, as no architectural metadata will be captured from your prose. Most users should keep extraction enabled and only toggle the background enrichment if necessary.

```bash
# Enable
bd config set entity_extraction.enabled true

# Disable
bd config set entity_extraction.enabled false
```

To toggle specifically the **Background AI Worker** (Ollama):
```bash
# Enable background AI processing
bd config set entity_extraction.background_enrichment true

# Disable (keeps Regex active but stops AI processing)
bd config set entity_extraction.background_enrichment false
```

### Ollama Model Settings

You can customize the model and connection parameters:

```bash
# Set the model (default: llama3.2:3b)
bd config set ollama.model llama3.2:1b

# Set connection URL (default: http://localhost:11434)
bd config set ollama.url http://localhost:11434

# Set timeout (default: 5s)
bd config set ollama.timeout 10s
```

## How it Works

1. **Extraction:** When a session is synced, it is queued for enrichment (`enrichment_status = 1`).
2. **Processing:** The `bd daemon` picks up queued sessions and sends them to your local Ollama instance.
3. **Crystallization:** Discovered relationships are written back into your Markdown files and the architectural graph is updated.
4. **Monitoring:** You can check the queue status by running `bd status` or see current configuration with `bd config list`.

## Prerequisites

Ollama enrichment requires:
- [Ollama](https://ollama.ai/) installed and running locally.
- The configured model pulled (e.g., `ollama pull llama3.2:3b`).
- The `bd daemon` running (it runs automatically, run `bd daemon start` if not sure).

For more architectural details, see [Devlog Architecture](DEVLOG_ARCHITECTURE.md).
