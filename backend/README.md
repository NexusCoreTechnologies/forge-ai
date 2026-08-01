# backend

Purpose: Forge AI backend services and CLI.

Requirements for OpenAI execution

- Set `OPENAI_API_KEY` in your environment before running with provider `openai`.

Example:

```bash
export OPENAI_API_KEY="sk-..."
FORGEAI_PROVIDER=openai go run ./cmd/forgeai execute --project "Example" --prompt "sprints/sprint003.md" --config config/config.yaml
```

The CLI exposes:

- `execute` — run a prompt execution (sends execution context and plan when available to the provider).
- `build-context` — build `execution_context.json` in the workspace.
- `plan-sprint` — generate `execution_plan.json` from `MASTER_BACKLOG.md`.

State: Implemented provider framework, context builder, and planner skeletons.
