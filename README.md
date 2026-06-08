<p align="center"><img src="build/appicon.png" alt="askd icon" width="96"></p>

# askd

**askd** is an MCP server that gives your coding agents a way to ask you questions, and you - a sleek GUI to answer them.

Most agents either can't ask you questions at all; and even if they do, the interface of it is locked in the TUI, making you deal with downsides of terminal input: no mouse, no copy-paste, no spellcheck, and a cramped interface. **askd** changes that. Once installed as MCP server, the tray icon appears as soon as your agent starts. When the agent needs your input, you get a notification, you click the icon, and a sleek GUI opens with the question and options. You can click to select answers, copy-paste into the textarea, and even write free-form input if allowed - finally returning you the freedom of GUI.

"There must be a catch," you say. "If it has a GUI, it must be resource-hungry, right?" Wrong!
Since it's written in lightweight Go on top of [Wails v3](https://v3.wails.io), **askd** uses a fraction of the resources of typical Node.js MCP servers. Whole thing fits in a less than 25 MB binary with 8 MB of RAM while used!

## Install

Grab the binary for your system from [Releases](https://github.com/FanaticExplorer/askd/releases) and put it in your `PATH`.

Or, if you have Go installed:

```bash
go install github.com/FanaticExplorer/askd@latest
```

This places `askd` in `$GOPATH/bin` — make sure that's in your `PATH`.

## Configure your agent

Add askd to your agent's MCP server list. The config is the same across agents:

```json
{
  "mcpServers": {
    "askd": {
      "command": "/path/to/askd"
    }
  }
}
```

Replace `/path/to/askd` with the actual path to the binary. Consult your agent's docs for where to place this config.

Most agents cap MCP tool calls at a short timeout. askd waits up to **15 minutes** for an answer — if your agent gives up sooner, increase its MCP timeout accordingly.

**Limits:**
- One question session at a time (rejects new calls while pending)
- 15-minute timeout (returns an error if unanswered)
- Single-instance lock (only one askd process at a time)

### The tool: `ask_user_questions`

The agent passes an array of questions, each with a `title`, `options`, and optional `description`, `multiSelect`, and `allowCustom`. Answers come back as `{selectedIndexes, customText}` per question. See the source for the full schema.

## Build from source

```bash
git clone https://github.com/FanaticExplorer/askd
cd askd
cd frontend && npm install && npm run build && cd ..
go build -o bin/askd.exe .
```

Requirements: Go 1.25+, Node.js, and platform WebView (WebView2 on Windows, WebKit on macOS/Linux).

## Support

I'm a student building tools like this in my free time - if it's useful to you, a [donation](https://donate.fanaticexplorer.dev/) helps me keep going.

## License

MIT — see [LICENSE](LICENSE).
