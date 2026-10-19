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

A question about what a file says, rather than how it is built, is refused
(see `REVIEW.md`). A released plugin is changed by a new version; a changed
question is a new major version.

Releases are made by a maintainer: a tag `<name>/vX.Y.Z` builds the plugin
from that tag, signs it with cosign (keyless, GitHub Actions identity) and
publishes it with a signed index.
