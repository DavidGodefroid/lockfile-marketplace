# Contributing a plugin

1. Fork the repository and add `plugins/<name>/`, its own Go module:
   - `plugin.toml`: `name`, `version`, `protocol = 1`, `question`, `one`,
     `zero`, `accepts`, `limits`. Nothing else: plugins take no parameters.
   - `main.go`: one call to `sdk.Main`, whose function gets the file as an
     `*os.File` and returns the answer or an error. Never print anything.
   - `fixtures_test.go`: writes synthetic fixtures into `testdata/` (they
     are not committed; `lib/pdfx/pdfxtest` helps for PDF files).
   - `testdata/expected.toml`: one `[[case]]` per fixture with its
     expected `answer`, at least one `0` and one `1`.
   - `plugin_test.go`: `TestMain` calls `sandbox.Main()`, then
     `harness.Check(t, ".")`.
2. Add the module to `go.work`.
3. Run the tests: `go test ./...` in the plugin's directory (Linux: the
   tests run your plugin in the lockfile sandbox).
4. Open a pull request. CI runs the checks listed in `REVIEW.md`, then a
   maintainer reviews the question and the code.

## Protocol 2: typed inputs and a detail

A plugin may instead declare `protocol = 2` with, in `plugin.toml`:

- `inputs = [{ name = "needle", kind = "text", max_bytes = 64 }]`: values
  the human types on the console (at most 4; `kind` is `text`, `email` or
  `digits`), each named in the question as `${needle}`. The agent binds
  the placeholder, never a value. Read them with `sdk.Inputs()`; check
  what the question needs beyond `kind` and `max_bytes` yourself.
- `detail = true`: a short text for the human only (the agent never sees
  it), written to `sdk.Detail()` and capped at 4 KiB. Pass
  `sdk.WithDetail()` to `sdk.Main`: without it `sdk.Detail()` discards.

Never write an input's value anywhere, not even to the detail: the harness
fails a plugin whose detail echoes it. Each `[[case]]` gives every input a
value, `inputs = { needle = "zebra-77" }`; the harness passes it on fd 6
like the console, and also fuzzes random values within each input's `kind`
and `max_bytes`.

A question about what a file says, rather than how it is built, is refused
(see `REVIEW.md`). A released plugin is changed by a new version; a changed
question is a new major version.

Releases are made by a maintainer: a tag `<name>/vX.Y.Z` builds the plugin
from that tag, signs it with cosign (keyless, GitHub Actions identity) and
publishes it with a signed index.
