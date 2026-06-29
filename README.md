
# cc-simplerouter

`simplerouter` launches [Claude Code](https://claude.com/claude-code) against any
[OpenRouter](https://openrouter.ai) model — configured for that one process
only, so your normal Claude Code setup is untouched.


```powershell
simplerouter                              # first run: pick a key + model
simplerouter --model z-ai/glm-5.2 .       # launch with a specific model in the current dir
```

On first run you paste an OpenRouter API key. It is validated against OpenRouter
and saved (with your last model) in `~/.simplerouter/config.json`. Nothing is
written to your global Claude Code config — `simplerouter` only sets environment
variables for the child `claude` process.

## Install

Requires [Go](https://go.dev/dl/) and an installed `claude` CLI.

Don't have Go? Install it with winget (then open a new terminal so it's on `PATH`):

```powershell
winget install --id GoLang.Go
```

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1
```

This builds `simplerouter.exe` and copies it to `~\.local\bin`. The script locates
Go even if it isn't on the current shell's `PATH`.

## The model picker

Run `simplerouter` to open the interactive picker.

<img width="675" height="462" alt="image" src="https://github.com/user-attachments/assets/1f15087a-ef63-4cf4-b875-54b1bb2052ce" />


- **↑ / ↓** — move the highlight (auto-pages at the top/bottom of a page)
- **← / →** — flip pages
- **type** — filter live by id or name
- **↵** — launch the highlighted model
- **p** — open provider selection for the highlighted model (see below)
- **esc** — cancel

The list is pre-filtered to models usable by Claude Code and ordered by
OpenRouter popularity, with recommended models are pinned to the top.

## Provider / endpoint selection

simplerouter default's to OpenRouter's choice of provider. If you want to select a specific inference provider, press **`p`** on a highlighted model:

<img width="674" height="461" alt="image" src="https://github.com/user-attachments/assets/d2093cc0-270a-43ef-a980-b972e93439dc" />

OpenRouter only honors provider routing in the request **body**, and Claude Code doesn't let you add body fields. So when you pin a provider, `simplerouter` starts a tiny localhost proxy for the session and points`ANTHROPIC_BASE_URL` at it; the proxy injects `provider.only` into each request before forwarding to OpenRouter. It binds to `127.0.0.1`, makes no changes to
your OpenRouter account, and shuts down when `claude` exits.

> **Note:** pinning sets `allow_fallbacks: false`, so a transient error from the
> chosen provider isn't absorbed by OpenRouter's fallback and Claude Code will
> retry. If a provider is flaky, just pick another (or skip provider selection
> and let OpenRouter route).

## Flags

```
simplerouter [--model MODEL] [--reset-key] [--disable-thinking] [path-or-prompt] [-- CLAUDE_ARGS...]
```

- `--model MODEL` — OpenRouter model id, name, or unique suffix (skips the picker)
- `--reset-key` — forget the saved OpenRouter API key, then prompt again
- `--disable-thinking` — drop Claude Code's Anthropic-specific thinking/beta
  request fields (see below)

## What it sets in Claude Code's environment

Only for the launched process. Notably:

- `ANTHROPIC_BASE_URL` → OpenRouter (or the local provider proxy, when pinned)
- `ANTHROPIC_AUTH_TOKEN` → your OpenRouter key; all model tiers (opus/sonnet/
  haiku/subagent) point at your chosen model
- `CLAUDE_CODE_AUTO_COMPACT_WINDOW` → the model's context length
- `CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false` → disables the "suggest what to
  type next" feature, which otherwise re-sends the whole conversation each turn
  just to predict your next prompt and wastes money.

## Model compatibility

`simplerouter` targets OpenRouter models that work through Claude Code's
Anthropic-compatible API path — i.e. text models that support tool calling
(which the picker already filters for).

By default it preserves Claude Code's normal thinking behavior. If a provider
chokes on Claude Code's thinking/beta request fields, retry with
`--disable-thinking`:

```powershell
simplerouter --disable-thinking --model XXX.
```

**Known issue:** OpenAI's GPT-5-family models (e.g. `openai/gpt-5-mini`) don't
currently work through Claude Code here due to how they return encrypted
reasoning blocks.
