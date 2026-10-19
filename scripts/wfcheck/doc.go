// Package wfcheck holds the tests that pin the security properties of the
// GitHub workflows: every action pinned by commit SHA, read-only defaults,
// the signing permissions confined to the jobs that sign, the tag checked to
// be on main before anything is signed, no concurrency that drops a tag run,
// and the cosign major that writes the bundles lockfile verifies.
package wfcheck
