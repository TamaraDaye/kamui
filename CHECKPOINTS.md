# Kamui Build Checkpoints

This file tracks small, verifiable increments. Each checkpoint has a clear goal, acceptance criteria, and a suggested commit. Update the status and commit hash when a checkpoint passes.

**Status legend:** `done` · `in progress` · `blocked` · `todo`

**Current focus:** Slice 1 — Foundation and offline plan  
**Next checkpoint:** C2 — finish CLI cleanup, then C3 config model

**Scope note:** this is a 2-month resume project. The must-ship path is intentionally narrow: config/plan/apply/verify, SSH, four core executors, one libvirt VM, and a recorded demo. Should-ship items are listed separately and only start after C12.

---

## Checkpoint overview

| ID | Increment | Milestone | Status | Commit |
|---|---|---|---|---|
| C0 | Repository initialized, catalog and design doc committed | M0 | done | `449a232` |
| C1 | Compile-clean skeleton and package boundaries | M0 | done | `5051533` |
| C2 | CLI root, `version`, and help | M0 | done | `834cc1b` |
| C3 | Config type model and strict YAML decoding | M1 | in progress | — |
| C4 | Validation rules and `kamui validate` | M1 | todo | — |
| C5 | Resource registry and dependency DAG | M1 | todo | — |
| C6 | `kamui plan --offline` and golden tests | M1 | todo | — |

---

## Validation findings (2026-11-05)

- C2 CLI cleanup is done: `version` prints one line and the custom help command is gone.
- `internal/engine.go` has been removed.
- `make build` still fails: `main_package_path = ./` should point at `./cmd`, and the Windows target currently uses `GOOS=darwin`.
- `go build ./...`, `go vet ./...`, and `go test ./...` now pass at commit `162a8a1`.
- The parser builds, but it does not populate the typed `MachineSpec` or `ServerSpec`; `spec` currently decodes into `map[string]interface{}`. See C3.
- `examples/machine.yaml`, `examples/server.yaml`, and `examples/README.md` exist and parse as valid YAML.

---

## C0 — Repository initialized

**Status:** done (commit `449a232`)

**Achieved**

- Initial git commit.
- `Kamui.md` design and 8-week plan.
- Curated `img-catalog/images.yaml`.
- Provider scaffold under `internal/provider/`.
- Empty placeholder files for config, engine, target, state, catalog, and CLI.

**Known follow-up**

- `go build ./...` currently fails because `internal/catalog.go` and
  `internal/provider/ec2.go` have no package declaration.
- `internal/cli.go` duplicates the `internal/cli` package and should be removed.

---

## C1 — Compile-clean skeleton and package boundaries

**Status:** done (commit `5051533`; delete `internal/engine.go` as cleanup)

**Goal:** settle package boundaries while the files are still empty, and get a green build.

**Why this first:** moving empty files is free. Moving files after they contain code and tests creates churn, import rewrites, and package-cycle risk.

**Recommended layout**

```text
cmd/kamui/main.go              package main
internal/cli/commands.go       package cli
internal/config/config.go      package config
internal/engine/engine.go      package engine
internal/target/target.go      package target
internal/state/state.go        package state
internal/catalog/catalog.go    package catalog
internal/provider/libvirt.go   package provider
internal/provider/ec2.go       package provider
```

**Steps**

1. Delete `internal/cli.go` (redundant with `internal/cli`).
2. Move the flat root files into the matching subpackages:
   - `internal/config.go` → `internal/config/config.go`
   - `internal/engine.go` → `internal/engine/engine.go`
   - `internal/target.go` → `internal/target/target.go`
   - `internal/state.go` → `internal/state/state.go`
   - `internal/catalog.go` → `internal/catalog/catalog.go`
3. Add the correct `package` line to every moved file.
4. Add `package provider` to `internal/provider/ec2.go`.
5. Add a minimal `internal/cli.Execute()` stub and call it from `cmd/main.go`.
6. Document the exact build, test, and vet commands. A `Makefile` is optional and not part of this checkpoint.
7. Run formatting, build, and vet.

**Acceptance**

```bash
gofmt -w ./cmd ./internal
go build ./...
go vet ./...
go test ./...
```

All four commands exit `0`.

**Suggested commit**

```text
chore: establish package boundaries and green build
```

**Note:** do not create empty files for `config`, `plan`, `resource`, etc. Add each file when it has a package declaration and at least one real symbol. Empty `.go` files break the build.

---

## C2 — CLI root, `version`, and help

**Status:** in progress

**Goal:** a usable command surface before any business logic.

**Steps**

- Choose Cobra or stdlib `flag`. Recommended: Cobra.
- Add `kamui version`.
- Add `kamui --help` and `kamui help`.
- Establish exit-code helpers (`0` success, `2` usage/config error).
- Add one CLI test that runs the root command in-process.

**Remaining before done**

- Remove `fmt.Println("test build")` from `cli.Execute`.
- Remove the custom `help` command and rely on Cobra's built-in help, or make it `cobra.NoArgs`.
- Change the CLI test to assert `version` output and command errors, not just call `rootCmd.Execute()`.
- Make `go run ./cmd version` print exactly one line.

**Acceptance**

```bash
go run ./cmd version
go run ./cmd --help
```

Both exit `0`, and `version` prints exactly one version string.

**Suggested commit**

```text
feat(cli): add root command and version
```

---

## C3 — Config type model and strict YAML decoding

**Status:** in progress (commit `162a8a1`)

**Goal:** load `examples/machine.yaml` with strict decoding, then add `examples/server.yaml`. Machine first keeps the checkpoint small.

**Review findings from `162a8a1`**

- Build, vet, and tests pass.
- Header parsing, `apiVersion` checking, and kind dispatch are on the right track.
- **Typed spec is not populated.** `decodeStrict(data, cfg)` decodes `Config.Spec`
  as `map[string]interface{}` because `Spec` is `any`. Decode a concrete
  `machineDocument` or `serverDocument` with `Spec MachineSpec` /
  `Spec ServerSpec`, then assign the result to `Config.Spec`.
- `ProviderSpec.Type` has `yaml:"provider"`; the example uses `provider.type`.
  Change the tag to `yaml:"type"`.
- `NetworkSpec.mode` is unexported, so `network.mode` is ignored. Export it as
  `Mode string \`yaml:"mode"\``.
- `Metadata.Name` has no explicit tag. Add `yaml:"name"`.
- `Resources.Name` uses `yaml:"names"`, which collides with `Names`. Use
  `yaml:"name,omitempty"`. `SSHAuthorizedKeys` must be `[]string` to match the
  example. `Content` is missing, so `examples/server.yaml` will fail strict
  decoding once the typed spec path is fixed.
- `Load` returns only an error and prints the config. It should return
  `(*Config, error)` and leave printing to the CLI.
- `decodeStrict` should reject trailing YAML documents.
- No tests exist yet. The examples are ready to serve as the first fixtures.

**Steps**

- Define `Document`, `Metadata`, `ServerSpec`, `MachineSpec`, `Target`, and the raw resource model.
- Implement `config.Load(path string) (*Document, error)`.
- Decode with `yaml.Decoder.KnownFields(true)`.
- Add defaults that are safe (for example, `port: 22`).
- Keep validation separate from decoding for this checkpoint.
- Add valid and invalid fixtures under `internal/config/testdata/`.
- `examples/server.yaml` and `examples/machine.yaml` are already added; use them as the first decode fixtures.

**Acceptance**

- `examples/machine.yaml` decodes successfully into a typed `MachineSpec`.
- An unknown top-level or machine-spec field returns an error.
- `kind: Server` returns either a decoded `ServerSpec` or a clear
  "not implemented yet" error until the server parser lands.
- Tests run without network access.
- `go test ./internal/config/...` passes and `go build ./...` is green.

**Suggested commit**

```text
feat(config): add v1 type model and strict YAML decoding
```

---

## C4 — Validation rules and `kamui validate`

**Goal:** reject invalid requests before any target contact.

**Validation checklist**

- `apiVersion` is `kamui/v1`; `kind` is `Machine` or `Server`.
- `Server` has exactly one of `target` or `targets`.
- Resource IDs are unique and non-empty.
- Resource types and fields are known.
- File paths are absolute; modes are valid octal strings.
- Service states and firewall policies are valid enums.
- `depends_on` references exist and contain no cycles.
- Duplicate resource identity (`type` + name/path) is rejected.
- Errors include the config path, for example `spec.resources[2].path`.

**Acceptance**

```bash
go run ./cmd/kamui validate -f examples/server.yaml   # exit 0
go run ./cmd/kamui validate -f bad.yaml               # exit 2
```

At least 10 invalid fixtures cover the rules above.

**Suggested commit**

```text
feat(config): validate desired-state documents
```

---

## C5 — Resource registry and dependency DAG

**Goal:** turn validated resources into a stable execution order.

**Steps**

- Define the `Resource` interface with `ID`, `Type`, `DependsOn`, `Check`,
  `Plan`, `Apply`, and `Verify`.
- Add a resource registry keyed by type.
- Add explicit `depends_on` edges and conservative inference.
- Implement cycle detection and dangling-edge validation.
- Implement stable topological sort.
- Add table-driven graph tests, including cycles and independent branches.

**Acceptance**

- The canonical example produces the expected order.
- A cycle returns a clear error naming the participating resources.
- Independent resources have a deterministic order.
- `go test ./internal/plan/...` passes.

**Suggested commit**

```text
feat(plan): add resource registry and dependency graph
```

---

## C6 — `kamui plan --offline` and golden tests

**Goal:** render the structural plan without touching a target.

**Steps**

- Add `kamui plan --offline -f <file>`.
- Render human-readable output with resource ID, type, and order.
- Render stable JSON output.
- Use an explicit `unknown`/`pending inspection` action in offline mode.
- Add golden tests for the human and JSON output.
- Document the difference between offline and live plans.

**Acceptance**

```bash
go run ./cmd/kamui plan --offline -f examples/server.yaml
go run ./cmd/kamui plan --offline --json -f examples/server.yaml
```

Both exit `0`; output order matches the DAG; no network calls occur.

**Suggested commit**

```text
feat(cli): add offline plan output
```

**Slice 1 complete when C1–C6 are done.** At that point, the config and planning contract is frozen enough to start the SSH transport (Slice 2).

---

## Post-Slice-1 checkpoints (provisional)

These are the trimmed must-ship path. Details are intentionally lighter until Slice 1 closes.

| ID | Increment | Milestone | Status |
|---|---|---|---|
| C7 | SSH transport + `kamui inspect` on Ubuntu/Debian | M2 | todo |
| C8 | `file` executor end-to-end: Check/Plan/Apply/Verify | M3 | todo |
| C9 | `package` and `user` executors | M3 | todo |
| C10 | `service` executor + reporting + `status` | M4 | todo |
| C11 | Interruption handling + safe re-run | M5 | todo |
| C12 | Libvirt single-VM provision/destroy + demo | M6 | todo |

### Optional should-ship checkpoints

Only start these after C12 passes, or when the must-ship path is ahead of schedule:

| ID | Increment |
|---|---|
| S1 | `group`, `directory`, and `firewall` executors |
| S2 | Full `history` and retention |
| S3 | Multi-target fan-out and bounded parallelism |
| S4 | Fedora 44 / dnf coverage |
| S5 | Drift report polish, resume hints, package version pinning |

**Resume-project rule:** a small, complete, verified story beats a broad, half-finished one. Prefer C12 plus a recorded demo over starting S1–S5 early.

---

## Working agreement

- One checkpoint per commit or small pull request.
- Do not start the next checkpoint until the current acceptance commands pass.
- Keep `main` buildable at every checkpoint.
- When a checkpoint changes the design, update `Kamui.md` and this file together.
- Checkpoints are reviewed for correctness, testability, and scope control, not for code volume.
