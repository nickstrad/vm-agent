# VM Agent

An AI-built personal agent platform for a Linux VM, with enforced permissions, deterministic testing, and first-class analytics.

Start at [the docs index](docs/index.md) or skim [the roadmap](docs/roadmap.md).
The repository is currently in the planning stage.

## Knowledge base

A pinned copy of [nickstrad/kb](https://github.com/nickstrad/kb) lives in `tools/kb`.
It is copied source, including its CLI, tests, and knowledge-store skill. The repo's store is `docs/knowledge`: tracked entries live in `data/`, while `.kb/` holds the ignored local database.
The [project skill](.agents/skills/update-knowledge-store/SKILL.md) is available under `.agents/skills`, with a `.claude/skills` symlink; both agent guides require its knowledge workflow.
Use it only for this repository, always setting `KB_ROOT` explicitly; follow [the repo-scoped invocation recipe](docs/knowledge-base.md#run-the-repo-scoped-cli) and [store contract](docs/knowledge/AGENTS.md). No computer-wide tool installation or shell-profile changes are needed.

## Environment variables

This section is the configuration reference and will grow as platform variables are implemented, including analytics configuration.
The variables below are supported by the imported knowledge tool; platform analytics currently has planned flags in the roadmap, with no implemented environment variables yet.

### Knowledge store

| Variable | Default or precedence | Purpose |
| --- | --- | --- |
| `KB_ROOT` | Required here: absolute path to this repo's `docs/knowledge`. Upstream falls back to `~/knowledge` if unset; do not use that fallback. | Root containing tracked `data/` and ignored `.kb/kb.sqlite`. |
| `KB_CALLER` | `--caller`, then this variable, then `unknown`. | Identify the caller in search logs and feedback. |
| `KB_EMBEDDER` | Set `none` for this repo's initial full-text store; `--embedder` overrides it. Upstream automatically selects a provider if unset. | Choose `none`, `ollama`, `openrouter`, or `openai`; change an existing store's provider only deliberately. |
| `KB_EMBED_MODEL` | `--embed-model` overrides it; otherwise provider default. | Select the embedding model. |
| `EDITOR` | `vi`. | Editor command used by `kb edit`. |

Automatic selection uses OpenRouter when `KB_OPENROUTER_API_KEY` or `OPENROUTER_API_KEY` is set, otherwise Ollama. Set `KB_EMBEDDER=none` for full-text-only use without an embedding service.
Changing the store's provider or model requires `kb reindex --all`.

### Optional embeddings

| Variable | Default or precedence | Purpose |
| --- | --- | --- |
| `KB_OPENROUTER_API_KEY` | Takes precedence over `OPENROUTER_API_KEY`. | Knowledge-specific OpenRouter credential. |
| `OPENROUTER_API_KEY` | Used when the knowledge-specific key is unset. | General OpenRouter credential; also enables automatic OpenRouter selection. |
| `KB_EMBED_API_KEY` | Used by `openai`; fallback credential for explicitly selected `openrouter`. | Credential for an OpenAI-style embedding endpoint. |
| `KB_EMBED_URL` | OpenRouter: `https://openrouter.ai/api/v1`; required for `openai`. | Base URL serving the embeddings API. |
| `KB_OLLAMA_URL` | `http://127.0.0.1:11434`. | Ollama endpoint when selected. |
| `KB_EMBED_DIM` | `768`; the stored vector width in this build. | Requested embedding dimensions. |
| `KB_EMBED_DOC_PREFIX` | Provider-specific; set empty to remove it. | Prefix applied to indexed text. |
| `KB_EMBED_QUERY_PREFIX` | Provider-specific; set empty to remove it. | Prefix applied to search queries. |
| `KB_EMBED_SEND_DIMENSIONS` | `true`. | Include the dimensions field in OpenAI-style requests. |
| `KB_EMBED_TRUNCATE` | OpenRouter `text-embedding-3` models: `true`; otherwise `false`. | Truncate longer vectors to the configured width. |

Provider model defaults are `openai/text-embedding-3-small` for OpenRouter and `nomic-embed-text` for Ollama; `openai` requires an explicit model.
Set credentials in the trusted development environment and keep them out of Git and the future agent runtime.

### Tool installation and testing

| Variable | Default | Purpose |
| --- | --- | --- |
| `KB_DEFAULT_ROOT` | Remembered install value, otherwise `~/knowledge`. | Build-time root setting used by the upstream installer. |
| `KB_LIVE_OLLAMA` | Unset. | Set to `1` to opt into upstream tests against a real Ollama service. |

The upstream installer and live-service tests have not been run for this project.
