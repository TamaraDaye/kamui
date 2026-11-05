# Kamui — Design & Build Plan

**Status:** Draft v0.2  
**Last updated:** 2026-11-05  
**Target release:** `v0.1.0` in 8 working weeks (2 calendar months)  
**Build progress:** [CHECKPOINTS.md](CHECKPOINTS.md)  
**Scope owner:** Kamui maintainers

---

## 1. Summary

Kamui is a **machine-agnostic provisioning engine**. It takes a target Linux machine and a declarative desired-state document, then converges the machine to that state and reports exactly what changed, what was already correct, what failed, and what was left incomplete.

The core engine knows nothing about how a machine was obtained. An existing VM, a libvirt-created VM, or a future EC2 instance is reduced to the same contract:

```text
Target = address + SSH credential + privilege-escalation details
Input = Target + Desired State
Output = per-resource result + verification evidence
```

The engine **must verify the observed outcome after applying changes**. It must not trust command exit codes alone.

### What ships in the first 2 months

- A Go CLI: `validate`, `inspect`, `plan`, `apply`, `status`, `provision`, `destroy`.
- SSH transport with host-key verification, key/agent auth, sudo, and safe file transfer.
- Core declarative resources: packages, users, files, and systemd services.
- Check → Plan → Apply → Verify lifecycle for every resource.
- Idempotent applies, structured human/JSON reports, and a minimal last-run status record.
- A libvirt provider that creates one VM from the curated image catalog, bootstraps SSH, and hands the target to the same engine.

Should-ship if time allows: group/directory/firewall resources, Fedora/dnf, full history, multi-target execution, and bounded parallelism. EC2, remote state, scheduling, and a web UI are explicitly **after** `v0.1.0`.

---

## 2. Goals, non-goals, and principles

### 2.1 Goals

1. **Acquisition-independent configuration.** A machine obtained anywhere can be configured by the same engine.
2. **Declarative and idempotent.** Re-applying an unchanged configuration produces no changes.
3. **Verified outcomes.** Every applied resource is re-read and compared with the desired state.
4. **Honest reporting.** Results distinguish changed, unchanged, failed, skipped, and incomplete.
5. **Safe by default.** Strict host-key checking, least-privilege privilege escalation, no secret leakage.
6. **Operable.** Clear errors, stable exit codes, machine-readable output, persistent run history.
7. **Boringly testable.** Core logic runs against a fake transport; real distros are covered by integration tests.

### 2.2 Non-goals for `v0.1.0`

- Windows targets.
- Container orchestration or Kubernetes provisioning.
- A web UI, RBAC, or multi-tenant control plane.
- Arbitrary imperative script execution as a first-class resource.
- Remote/shared state backends (S3, Consul, etc.).
- High availability, scheduling, or continuous drift remediation.
- Full cloud coverage. EC2 is a post-`v0.1.0` provider.

### 2.3 Design principles

| Principle | Consequence |
|---|---|
| Desired state over scripts | Resources expose `Check`, `Plan`, `Apply`, `Verify`, not ad-hoc shell commands. |
| The target is the source of observed truth | Local state is history/cache, never the authority for whether a resource exists. |
| Check before changing | Plan is generated from live inspection; no blind writes. |
| Verify after changing | Success requires re-reading the resource, not trusting an exit code. |
| Stable resource identity | Every resource has a deterministic ID; changes are diffed by identity. |
| Fail visibly | A failure stops dependants, is reported as `failed` or `skipped`, and never becomes a false success. |
| Safe concurrency | Parallelism is bounded and respects dependency/resource-conflict constraints. |
| No secrets in output | Keys, tokens, and secret file content are redacted from logs, state, and reports. |

---

## 3. Scope for `v0.1.0`

### Must ship

- SSH target with key or agent auth and strict host-key verification.
- YAML config parsing, strict validation, defaults, and actionable errors.
- Dependency-aware execution plan with cycle detection.
- Read-only `inspect` of OS, architecture, package manager, init system, firewall backend, and sudo capability.
- Core resource executors: package, user, file, and systemd service.
- Apply with per-resource verification and idempotency.
- Human-readable and JSON output; stable exit codes.
- `status` backed by a minimal local last-run record.
- Safe re-run after interruption; no duplicate or half-written resources.
- Libvirt provider: catalog lookup, digest verification, single-VM creation, cloud-init SSH bootstrap, destroy, orphan cleanup.
- Unit tests, fake-transport tests, and a real-VM integration path for Ubuntu 24.04 and Debian 13.
- Documentation, example configs, and one recorded end-to-end demo.

### Should ship if the schedule holds

- Additional resource executors: group, directory, and firewall.
- Fedora 44 / dnf integration coverage.
- Multi-target fan-out and bounded parallel execution.
- Full run history, retention, and query commands.
- Basic drift report in `plan` output.
- Resume hints for interrupted runs (not full checkpoint replay).
- Package version pinning.

### Won't ship in this window

- AWS EC2 production support.
- Remote state, scheduling, drift daemons.
- Advanced firewall models (zones, rich rules, nftables direct).
- Secrets manager integration.
- Arbitrary command/script resources.
- Windows.

---

## 4. Core concepts

| Term | Meaning |
|---|---|
| **Target** | A reachable Linux machine plus SSH and privilege-escalation details. |
| **Machine spec** | A request to *obtain* a machine (`kind: Machine`). It produces a Target. |
| **Server spec** | A desired state applied to a Target (`kind: Server`). This is the acquisition-independent core. |
| **Resource** | One addressable unit of desired state, such as a package, user, file, or service. |
| **Check** | Read-only inspection of the resource's current state. |
| **Plan** | The ordered set of actions needed to move observed state toward desired state. |
| **Apply** | Execution of planned actions only. |
| **Verify** | A post-apply re-read proving the resource now matches the desired state. |
| **Run** | One plan/apply execution against one or more targets, with a stable run ID. |
| **Incomplete** | Work whose outcome is unknown because of interruption, timeout, or transport loss. It is never counted as success. |

### 4.1 The key separation

```text
kind: Machine  ──provision──▶  Target  ──apply──▶  kind: Server
   (how to get it)              (contract)          (what it must be)
```

This separation is the architectural answer to "independent of how the machine was obtained." Providers are replaceable; the configuration model and execution engine are not.

---

## 5. Architecture

### 5.1 Component diagram

```text
┌─────────────────────────────────────────────────────────────────┐
│                         kamui CLI                               │
│  validate  inspect  plan  apply  status  history  provision     │
└───────────────┬───────────────────────────────┬─────────────────┘
                │                               │
                ▼                               ▼
┌───────────────────────────┐       ┌───────────────────────────┐
│  Config & validation      │       │  Machine providers        │
│  YAML -> typed model      │       │  libvirt (v0.1)           │
│  strict fields, defaults  │       │  ec2 (post-v0.1)          │
└─────────────┬─────────────┘       └─────────────┬─────────────┘
              │                                   │
              ▼                                   ▼
┌───────────────────────────┐       ┌───────────────────────────┐
│  Planner                  │       │  Catalog / image cache    │
│  dependency DAG           │       │  pinned digests           │
│  topo order + cycles      │       │  verified artifacts       │
└─────────────┬─────────────┘       └───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Execution engine                          │
│  Check -> Plan -> Apply -> Verify -> Result                     │
│  sequential by default, optional workers, interruption safety   │
└───────────────┬───────────────────────────────┬─────────────────┘
                │                               │
                ▼                               ▼
┌───────────────────────────┐       ┌───────────────────────────┐
│  Resource executors       │       │  Transport                │
│  package, user, group,    │       │  SSH exec + SFTP/upload   │
│  file, directory, service,│       │  sudo -n, timeouts,       │
│  firewall                 │       │  host-key verification    │
└─────────────┬─────────────┘       └─────────────┬─────────────┘
              │                                   │
              └───────────────┬───────────────────┘
                              ▼
                 ┌───────────────────────────┐
                 │  Target Linux machine     │
                 │  sshd, sudo, systemd,     │
                 │  apt/dnf, ufw/firewalld   │
                 └───────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  State & reporting: run records, target fingerprints, results   │
└─────────────────────────────────────────────────────────────────┘
```

### 5.2 Component responsibilities

| Component | Responsibility |
|---|---|
| `internal/cli` | Parse commands/flags, load config, render output, map errors to exit codes. |
| `internal/config` | Strict YAML decode, defaults, validation, typed resource model. |
| `internal/plan` | Build the resource graph, infer/validate dependencies, topologically sort, produce actions. |
| `internal/engine` | Execute plans, enforce ordering, collect results, handle cancellation, and optionally add bounded parallelism. |
| `internal/resource/*` | One package per resource type; implements `Check`, `Plan`, `Apply`, `Verify`. |
| `internal/transport/ssh` | Connection lifecycle, command execution, file upload/download, sudo wrapper. |
| `internal/target` | Capability probing and target fingerprinting. |
| `internal/state` | Atomic, locked persistence of runs and last observed results. |
| `internal/catalog` | Load `img-catalog/images.yaml`, verify digests, manage the image cache. |
| `internal/provider/libvirt` | Domain lifecycle, cloud-init seed, readiness wait, cleanup. |
| `internal/provider/ec2` | Post-`v0.1.0`; must implement the same provider interface. |

### 5.3 End-to-end data flow

```text
1. Load and validate config.
2. Resolve target(s), or provision a machine and receive a target.
3. Connect over SSH and inspect capabilities.
4. For each resource, run Check (read-only).
5. Build the dependency DAG and compute actions (noop/create/update/delete/replace).
6. Render plan; in apply mode, execute in topological order. Start sequentially; add bounded parallelism only as a should-ship feature.
7. For each action: Apply, then Verify by re-reading observed state.
8. Aggregate results: changed, unchanged, failed, skipped, incomplete.
9. Persist the run record and print the report.
10. Exit with the status code matching the worst outcome.
```

### 5.4 Key architecture decisions

| ID | Decision | Rationale |
|---|---|---|
| ADR-001 | Split `Machine` and `Server` config kinds. | Keeps acquisition providers out of the desired-state engine. |
| ADR-002 | Every resource implements Check/Plan/Apply/Verify. | Makes idempotency and verification structural, not optional. |
| ADR-003 | Check live target state during `plan`. | Prevents plans based on stale local state. `--offline` is available for structural plans only. |
| ADR-004 | Local state is history, not authority. | Supports recovery and auditing without becoming a distributed-state problem. |
| ADR-005 | Sequential execution by default; bounded, dependency-aware parallelism is should-ship. | Avoids unsafe concurrent mutation and keeps the must-ship engine simple; the interface leaves room for workers later. |
| ADR-006 | SSH via Go's `x/crypto/ssh` plus SFTP. | Full control over auth, host keys, timeouts, and testability. |
| ADR-007 | Strict host-key verification by default. | Prevents silent machine substitution. TOFU must be explicit. |

---

## 6. Configuration model

### 6.1 Canonical `Server` example

```yaml
apiVersion: kamui/v1
kind: Server

metadata:
  name: web-01

spec:
  target:
    host: 203.0.113.10
    port: 22
    user: ubuntu
    auth:
      private_key: ~/.ssh/id_ed25519
      # agent: true is also supported
    host_key:
      policy: strict
      known_hosts: ~/.ssh/known_hosts
    become:
      method: sudo
      user: root

  resources:
    - id: admin-user
      type: user
      name: deploy
      state: present
      groups: [sudo]
      shell: /bin/bash
      home: /home/deploy
      ssh_authorized_keys:
        - ssh-ed25519 AAAAC3Nza... operator@workstation

    - id: base-packages
      type: package
      names: [nginx, git, ufw]
      state: present

    - id: nginx-config
      type: file
      path: /etc/nginx/sites-available/default
      state: present
      owner: root
      group: root
      mode: "0644"
      content: |
        server {
          listen 80 default_server;
          server_name _;
          root /var/www/html;
        }
      depends_on: [base-packages]

    - id: nginx-service
      type: service
      name: nginx
      state: running
      enabled: true
      depends_on: [base-packages, nginx-config]

    - id: firewall
      type: firewall
      backend: auto
      default_inbound: deny
      allowed_inbound_ports: [22, 80, 443]
      depends_on: [base-packages]
```

### 6.2 `Machine` example (libvirt)

```yaml
apiVersion: kamui/v1
kind: Machine

metadata:
  name: web-01

spec:
  provider:
    type: libvirt
    uri: qemu:///system
  image:
    catalog: img-catalog/images.yaml
    id: ubuntu-24.04-lts-amd64-20261017
  machine:
    vcpus: 2
    memory_mb: 2048
    disk_gb: 20
  access:
    ssh_user: ubuntu
    private_key: ~/.ssh/id_ed25519
    authorized_keys:
      - ssh-ed25519 AAAAC3Nza... operator@workstation
  network:
    mode: nat
```

`kamui provision` creates the machine, waits for SSH, and emits a target record that `kamui apply` can consume without the user copying IP addresses by hand.

### 6.3 Validation rules

- Unknown fields are rejected (strict decoding), except inside explicitly free-form fields such as `content`.
- `apiVersion` must be `kamui/v1`; `kind` must be `Machine` or `Server`.
- A `Server` must define exactly one of `target` or `targets`.
- Resource IDs are unique and non-empty.
- `depends_on` entries must reference existing resource IDs; cycles are rejected.
- Resource type and fields must be known; every type has a schema.
- File `path` must be absolute; `mode` must be a valid octal string; owner/group are optional.
- Service `state` is `running` or `stopped`; `enabled` defaults to `true`.
- Firewall ports are `1..65535`; default inbound policy is `deny` or `allow`.
- SSH keys must parse as authorized-key lines.
- Duplicate resource identity (`type` + `name`/`path`) is rejected.
- Secrets are allowed only in designated fields and are redacted from all output.

### 6.4 Resource types in `v0.1.0`

| Type | Tier | Required fields | Key behavior |
|---|---|---|---|
| `package` | must | `names`, `state` | Install/remove with apt; verify with the package database. |
| `user` | must | `name`, `state` | Create/update user, groups, shell, home, authorized keys. |
| `file` | must | `path`, `state` | Atomic content write, owner/group/mode; hash verification. |
| `service` | must | `name`, `state` | systemd enable/start/stop/disable; verify enabled and active. |
| `group` | should | `name`, `state` | Create/remove group; verify with `getent`. |
| `directory` | should | `path`, `state` | Create/remove directory, recursive mode/owner options. |
| `firewall` | should | `default_inbound`, `allowed_inbound_ports` | Detect ufw/firewalld; apply and re-read rules. |

`dnf`/Fedora support is a should-ship item. The must-ship path targets apt-based Ubuntu 24.04 and Debian 13.

Planned after `v0.1.0`: `package_repo`, `template`, `mount`, `cron`, `sysctl`, and a deliberate `exec` escape hatch with explicit changed-when semantics.

### 6.5 Dependency handling

The planner uses two sources of order:

1. **Explicit** `depends_on` edges from the config.
2. **Inferred** edges where correctness requires them, for example:
   - a service depends on the package that provides it when names match;
   - a file depends on its parent directory resource when one is declared;
   - a firewall rule that opens a service port depends on the service only if declared.

Inference is conservative and always visible in `plan` output. If an inferred edge is wrong, the user can override it with explicit `depends_on`. Cycles are always a validation error.

---

## 7. Execution engine

### 7.1 Resource lifecycle

```go
type Resource interface {
    ID() string
    Type() string
    DependsOn() []string

    Check(ctx context.Context, t *Target) (Observed, error)
    Plan(ctx context.Context, t *Target, o Observed) (Action, error)
    Apply(ctx context.Context, t *Target, a Action) (ApplyResult, error)
    Verify(ctx context.Context, t *Target) (VerifyResult, error)
}
```

| Phase | Mutates target? | Failure meaning |
|---|---|---|
| Check | No | Cannot determine state; the resource is `failed`. |
| Plan | No | Cannot produce a safe action; the resource is `failed`. |
| Apply | Yes | The change was attempted and failed or timed out. |
| Verify | No | The change appeared to run, but the desired outcome is not proven. |

### 7.2 Actions and statuses

**Plan actions:** `noop`, `create`, `update`, `delete`, `replace`.

**Run result statuses:**

| Status | Meaning |
|---|---|
| `changed` | An action was applied and verification passed. |
| `unchanged` | Check already matched desired state; no action taken. |
| `failed` | Check, apply, or verification failed. |
| `skipped` | A dependency failed, so this resource was not attempted. |
| `incomplete` | Interrupted, timed out, or transport was lost; outcome unknown. |

Verification failure is never `changed`. It is `failed`, even if the apply command exited zero.

### 7.3 Idempotency contract

For every resource, with an unchanged target and unchanged config:

```text
apply #1 -> changed
apply #2 -> unchanged (all resources)
```

This is a release-blocking test, not a nice-to-have. Each executor owns a test that proves it.

### 7.4 Verification matrix

| Resource | Check / Verify method | Apply method |
|---|---|---|
| Package | `dpkg-query` or `rpm -q`; optional version compare | `apt-get install/remove` or `dnf install/remove` |
| User | `getent passwd`, `id`, `getent group`, home/shell checks | `useradd`, `usermod`, `userdel` |
| Group | `getent group` | `groupadd`, `groupdel` |
| File | `stat` for mode/owner/group, SHA-256 for content | atomic upload: temp file, fsync, rename |
| Directory | `stat` type/mode/owner | `mkdir -p`, `chmod`, `chown` |
| Service | `systemctl is-enabled`, `systemctl is-active` | `systemctl enable/disable`, `start/stop/restart` |
| Firewall | `ufw status` or `firewall-cmd --list-all` | add/remove rules, set default policy |
| SSH keys | managed block in `authorized_keys`, parse and compare | write managed block atomically |

### 7.5 Failure, interruption, and recovery

- **Dependency failure:** dependants become `skipped`; independent branches continue.
- **Transport loss:** in-flight resources become `incomplete`; remaining resources are not attempted.
- **SIGINT/SIGTERM:** stop scheduling new work, finish or cancel the current command, flush a partial run record, and exit with code `4`.
- **Re-run:** because every apply re-checks observed state, re-running is the recovery mechanism. Exact checkpoint replay is post-`v0.1.0`.
- **Locking:** one apply per target at a time, enforced by a lock file keyed by target fingerprint.
- **Temp-file safety:** file writes use a temp file in the destination directory plus rename; interrupted writes never replace the original.

### 7.6 Concurrency

Concurrency is a should-ship feature. The must-ship engine executes the plan sequentially; the interface should allow adding workers later without changing resource executors.

- Workers default to `min(4, number of independent resources)`; configurable with `--parallel`.
- A resource never runs until all its dependencies are complete.
- Resources that touch the same path, service, or package set are serialized even without an explicit edge.
- Results are collected per target and per resource; output order is stable regardless of completion order.

---

## 8. SSH transport and target inspection

### 8.1 Transport contract

```go
type Transport interface {
    Run(ctx context.Context, cmd Command) (Result, error)
    Upload(ctx context.Context, local io.Reader, remotePath string, mode os.FileMode) error
    Close() error
}
```

Requirements:

- Key-file and SSH-agent authentication.
- Strict host-key verification against `known_hosts` by default.
- Explicit `accept-new` / TOFU policy only when requested.
- Connection, command, and idle timeouts; context cancellation everywhere.
- Separate stdout/stderr capture, exit code, and bounded output to prevent memory blowups.
- Privilege escalation via `sudo -n`; a preflight check fails fast if passwordless sudo is unavailable.
- Safe command construction: no string-concatenated untrusted shell. Prefer argument quoting or uploaded scripts with fixed paths.
- Uploads are atomic and preserve the requested mode/owner where applicable.

### 8.2 Target inspection

`kamui inspect` collects at least:

```text
hostname, SSH user, privilege-escalation result
OS family / distribution / version
kernel and architecture
package manager (apt, dnf, unknown)
init system (systemd, unknown)
firewall backend (ufw, firewalld, none/unknown)
available shells and default shell
sudo capability and whether a password is required
```

Inspection is cached for the duration of a run and stored with the run record. A capability mismatch produces a clear, early failure rather than a mid-apply surprise.

### 8.3 Supported targets for `v0.1.0`

- **Must:** Ubuntu 24.04 LTS (apt, systemd, ufw).
- **Must:** Debian 13 (apt, systemd, ufw).
- **Should:** Fedora 44 (dnf, systemd, firewalld).

Other systemd-based distributions may work but are not release-blocking.

---

## 9. State, history, and reporting

### 9.1 State store

Default location: `${XDG_STATE_HOME:-~/.local/state}/kamui/`.

```text
runs/<run-id>.json          # immutable run record
targets/<fingerprint>/      # last observed state and lock file
cache/images/               # verified image artifacts
```

Properties:

- Atomic writes (temp + rename), directory mode `0700`, files `0600`.
- Per-target lock file prevents concurrent applies.
- State is never used as the authority for resource existence; it is for history, auditing, and recovery hints.
- Secrets never enter the state store.

### 9.2 Run record

Each run record contains:

- run ID, start/end time, duration, Kamui version;
- config path and SHA-256 of the normalized config;
- target fingerprint and inspected capabilities;
- per-resource: ID, type, action, status, verification evidence, error, duration;
- summary counts and final exit code.

### 9.3 Human output

```text
Target web-01 (203.0.113.10) — ubuntu 24.04, apt, systemd, ufw

  + admin-user      user       created and verified
  ~ base-packages   package    installed nginx, git; verified
  ~ nginx-config    file       content/mode updated; verified
  = nginx-service   service    already enabled and running
  ! firewall        firewall   verification failed: port 443 not open

Summary: 3 changed, 1 unchanged, 1 failed, 0 skipped, 0 incomplete
Run ID: 01J...  Duration: 12.4s
```

### 9.4 Exit codes

| Code | Meaning |
|---|---|
| `0` | All resources are `changed` or `unchanged`, and every change verified. |
| `1` | One or more resources failed verification or apply. |
| `2` | Invalid usage, config, or validation error. |
| `3` | Connection, authentication, or host-key failure. |
| `4` | Run incomplete (interrupted, timeout, transport lost). |

---

## 10. CLI specification

```text
kamui validate -f server.yaml
kamui inspect  -f server.yaml [--json]
kamui plan     -f server.yaml [--offline] [--json]
kamui apply    -f server.yaml [--auto-approve] [--parallel N] [--json]
kamui status   -f server.yaml [--json]
kamui history  [--limit N] [--json]
kamui provision -f machine.yaml [--auto-approve] [--json]
kamui destroy   -f machine.yaml [--auto-approve]
kamui version
```

Behavior notes:

- `validate` is local-only and fast.
- `inspect` is read-only on the target.
- `plan` connects and inspects by default; `--offline` validates structure and renders an unknown-state plan.
- `apply` prints the plan and asks for confirmation unless `--auto-approve` is set.
- `status` reports the most recent run and is must-ship. `history` and `--parallel` are should-ship; without them, apply runs sequentially and status reads the latest local record.
- `--json` is stable, documented, and suitable for CI.
- No command prints private keys, tokens, or secret file contents.

---

## 11. Machine providers

### 11.1 Provider interface

```go
type Provider interface {
    Validate(spec MachineSpec) error
    Provision(ctx context.Context, spec MachineSpec) (Target, error)
    Destroy(ctx context.Context, ref MachineRef) error
    Cleanup(ctx context.Context, ref MachineRef) error
}
```

Provider implementations must return the same `Target` contract. The engine must not import provider packages.

### 11.2 Libvirt provider (`v0.1.0`)

1. Load `img-catalog/images.yaml` and select a pinned image by ID.
2. Download into the cache if absent; verify the catalog digest before use.
3. Generate a cloud-init seed with the configured SSH user/key and hostname.
4. Create the domain with `libvirt.org/go/libvirt` and `libvirt.org/go/libvirtxml`.
5. Wait for an IP address and for SSH to accept the configured key, with a bounded timeout.
6. Return a `Target` and persist a machine reference for `destroy` and orphan cleanup.
7. On failure, clean up partial domains, volumes, and seed files.

### 11.3 EC2 provider (after `v0.1.0`)

Same interface. Requires AWS credential resolution, AMI selection, security groups, key-pair handling, instance readiness, tagging, and partial-infrastructure cleanup. The catalog model should be extended rather than special-cased.

---

## 12. Security model

- **Host authenticity:** strict `known_hosts` verification by default. A mismatch aborts before authentication.
- **Credentials:** prefer SSH agent; key files must be `0600`. Never log, serialize, or echo private keys.
- **Privilege escalation:** `sudo -n` only in `v0.1.0`. A password-required sudo setup fails preflight with a clear message.
- **Command safety:** no concatenation of untrusted values into shell strings. Use fixed command templates with validated arguments, or upload a script with a known path and mode.
- **File safety:** atomic writes, explicit modes, no world-writable temp files.
- **Secret handling:** designated secret fields are redacted from logs, JSON output, and state. File content is hashed, not stored, in reports.
- **State permissions:** state directories are `0700`; files are `0600`.
- **Image integrity:** the libvirt provider refuses to boot an image whose digest does not match the catalog.
- **Least data:** inspect only what is needed; do not collect unrelated host data.

---

## 13. Repository layout

```text
cmd/main.go                        # thin entry point

internal/
  cli/                             # commands, flags, rendering, exit codes
  config/                          # strict YAML decode, defaults, validation
  plan/                            # DAG construction, topological sort, actions
  engine/                          # execution, cancellation, results
  resource/
    resource.go                    # Resource interface and registry
    pkg/                           # apt/dnf package executor
    user/                          # user and group executors
    file/                          # file and directory executors
    service/                       # systemd executor
    firewall/                      # ufw/firewalld executor
  transport/ssh/                   # connection, exec, upload, sudo
  target/                          # inspection and fingerprinting
  state/                           # run records, locks, last observed state
  catalog/                         # catalog loading and digest verification
  provider/
    libvirt.go                     # libvirt machine provider
    ec2.go                         # post-v0.1.0

img-catalog/images.yaml            # curated, pinned image catalog
examples/
  machine.yaml
  server.yaml
  README.md
test/
  integration/                     # real-VM tests, build-tagged
  fixtures/                        # configs and golden plan output
docs/
  config-reference.md
  adr/
Makefile                           # optional convenience targets
CHECKPOINTS.md                     # incremental build tracker
```

The provider files currently live directly in `internal/provider/`. If the provider grows, split it into `internal/provider/libvirt/` and `internal/provider/ec2/` without changing the provider interface.

---

## 14. Testing strategy

### 14.1 Test pyramid

| Layer | What it covers | How |
|---|---|---|
| Unit | Config parsing, validation, defaults, error messages | Table tests, golden files. |
| Unit | Planner ordering, cycle detection, inference, action selection | In-memory fake resources. |
| Unit | Executors | Fake transport returning scripted observed states and command results. |
| Contract | Transport behavior, timeouts, quoting, upload atomicity | Fake SSH server or interface-level fakes. |
| Integration | Real distro behavior, package manager, systemd | Disposable libvirt VMs for Ubuntu and Debian; Fedora if the should-ship dnf path lands. |
| End-to-end | Provision → apply → verify → status → destroy | Scripted demo run. |
| Failure injection | SSH drop, sudo denial, bad package, failing service, bad host key | Scripted faults. |
| Security | Secret redaction, host-key mismatch, unsafe quoting | Regression tests. |

### 14.2 Release-blocking tests

1. **Idempotency:** apply twice; the second run reports zero `changed`.
2. **Verification:** deliberately corrupt a file, stop a service, and remove a package; the next plan detects drift and apply fixes it.
3. **Dependency failure:** fail a package install and confirm dependants are `skipped`, not `failed` or falsely successful.
4. **Interruption:** kill apply mid-run, re-run, and confirm convergence with no half-written files.
5. **Host-key safety:** a changed host key aborts with exit code `3`.
6. **Secret safety:** scan all output and state for private-key material and secret content.

Should-ship tests, required only if the feature is implemented:

7. **Firewall drift:** close a managed port and confirm the next plan detects and reopens it.
8. **Multi-target isolation:** a failure on one target does not corrupt or block the other target's state.

### 14.3 Integration environment

- The integration suite runs against libvirt locally through a documented command; a Makefile or task runner is optional.
- The test harness provisions from the same catalog used by the product, which keeps the provider path exercised.
- CI runs unit and static checks on every commit; integration runs on a self-hosted runner or nightly.
- Each integration test destroys its VMs even on failure; a leak check fails the suite if orphan domains remain.

---

## 15. Milestones and 2-month timeline

### 15.1 Assumptions

- One primary engineer, full-time equivalent.
- The schedule is **8 working weeks**. Week 8 includes buffer and release work.
- Calendar anchor: if work starts on **Monday 9 November 2026**, `v0.1.0` lands by **Friday 2 January 2027**. If the holiday period reduces capacity, either move the release to 9 January 2027 or cut the "Should ship" items; do not cut verification, security, or hardening work.
- Existing libvirt dependencies and `img-catalog/images.yaml` are reused; they are not rebuilt from scratch.

### 15.2 Timeline at a glance

| Week | Primary focus | Milestone gate |
|---|---|---|
| 1 | Contracts, repo skeleton, config model, plan skeleton | **M0 complete**, **M1 in progress** |
| 2 | Validation, DAG/plan, SSH transport start | **M1 complete** |
| 3 | SSH transport, inspection, first executors | **M2 complete** |
| 4 | Packages, users, and files | **M3 in progress** |
| 5 | Core executors complete; services and verification start | **M3 complete** |
| 6 | Verification, reporting, and last-run status | **M4 complete** |
| 7 | Recovery, optional history/multi-target, libvirt start | **M5 complete** |
| 8 | Libvirt end-to-end, hardening, docs, release | **M6 complete / v0.1.0** |

### 15.3 Milestone details

#### M0 — Foundation and contracts (Week 1, days 1–3)

**Tasks**

- Create the repository layout; rename `internal/vm-provisoner` to `internal/provider/libvirt`.
- Commit the first working skeleton and remove placeholder dead code.
- Freeze the `v1` config schema, result statuses, CLI verbs, and exit codes (this document).
- Define core interfaces: `Transport`, `Resource`, `Target`, `Provider`, `StateStore`.
- Define one-command local verification with plain `go` commands. A `Makefile` is optional and should be added only if repeated multi-step workflows justify it.
- Add CI for `gofmt`, `go vet`, unit tests, and build.
- Create `examples/server.yaml` and `examples/machine.yaml`; stop ignoring the example config.
- Write ADR-001 through ADR-004.

**Deliverables:** compiling CLI skeleton, interfaces, example configs, ADRs, green CI.

**Gate:** `go build ./...`, `go vet ./...`, and `go test ./...` pass; `kamui version` works; the config schema has no unresolved contradictions.

#### M1 — Config, validation, and planning (Weeks 1–2)

**Tasks**

- Strict YAML decoding with unknown-field rejection and line-aware errors.
- Defaults and validation for both `Server` and `Machine` kinds.
- Resource registry and per-type schemas.
- Dependency graph: explicit edges, conservative inference, cycle detection, topological sort.
- Action selection: `noop`, `create`, `update`, `delete`, `replace`.
- Human and JSON plan rendering.
- `kamui validate` and `kamui plan --offline`.
- Golden-file tests for plans and validation errors.

**Deliverables:** `validate` and offline `plan`; deterministic plan output; 20+ invalid-config cases covered.

**Gate:** the canonical example produces the expected order; cycles and dangling dependencies fail clearly; no network calls in offline mode.

#### M2 — SSH transport and inspection (Weeks 2–3)

**Tasks**

- SSH connection with key and agent auth.
- Strict host-key verification and explicit TOFU policy.
- Command execution with stdout/stderr/exit code, timeouts, output caps, and context cancellation.
- `sudo -n` preflight and become wrapper.
- Atomic upload/download.
- Capability inspection: OS, version, architecture, package manager, init, firewall, sudo, shell.
- `kamui inspect` with human and JSON output.
- Fake transport for unit tests; real-VM smoke test.

**Deliverables:** reliable SSH layer; inspection output for Ubuntu 24.04 and Debian 13. Fedora 44 is should-ship.

**Gate:** connect/inspect works on Ubuntu and Debian; bad host key and bad auth fail with exit code `3`; command-quoting tests pass.

#### M3 — Core resource executors (Weeks 3–5)

**Tasks**

- Resource lifecycle and result plumbing in the engine.
- `package`: apt install/remove, verify via the package database. Add dnf only if Fedora is in scope.
- `user`: create/update/delete, groups, shell, home, authorized keys.
- `file`: content, mode, owner, group, atomic writes.
- Per-resource dependency and conflict handling.
- Idempotency tests for every executor.
- First end-to-end scenario: admin user + packages + config file.

**Deliverables:** three core resource executors; scenario passes on Ubuntu and Debian.

**Gate:** second apply is entirely `unchanged`; failures are isolated to their dependency branch; no unverified success.

#### M4 — Services, verification, and reporting (Weeks 5–6)

**Tasks**

- `service`: systemd enable/disable, start/stop/restart, active-state wait with retries.
- Verification evidence captured per resource.
- Human report with stable ordering and summary counts.
- JSON report schema and exit-code mapping.
- Drift demonstration: mutate the target manually, then plan/apply.
- Failure-injection tests for service and verification.
- Should-ship: `group`, `directory`, and `firewall` executors.

**Deliverables:** canonical scenario with package, user, file, and service converges; reports are accurate under partial failure.

**Gate:** the four core resource types work end-to-end; verification failure forces `failed`; exit codes match the documented contract. Extra executors do not block this gate.

#### M5 — State, status, and recovery (Weeks 6–7)

**Tasks**

- Atomic, permissioned state store with per-target locking.
- Minimal last-run record with target fingerprint and config hash.
- `kamui status`.
- Interruption handling: partial run record, `incomplete` status, safe re-run.
- Should-ship: full `history`, multi-target fan-out, bounded parallelism, and concurrency tests.

**Deliverables:** `status` from the last run; kill-and-rerun test. Full history and multi-target only if the schedule holds.

**Gate:** interrupting a run never yields a false success; a second apply cannot run concurrently on the same target. Multi-target isolation is only required if that should-ship item is implemented.

#### M6 — Libvirt provider and `v0.1.0` release (Weeks 7–8)

**Tasks**

- Catalog loader with schema validation and pinning.
- Artifact download/cache and digest verification; refuse mismatches.
- cloud-init seed generation and domain XML via `libvirtxml`.
- Domain creation, IP discovery, SSH readiness wait.
- Target emission and machine reference persistence.
- `destroy` plus orphan cleanup for partial provisioning.
- End-to-end demo: `provision` → `apply` → `status` → `destroy`.
- Documentation: quick start, config reference, architecture, troubleshooting.
- Tag `v0.1.0`.

**Deliverables:** repeatable one-command demo; release artifacts and docs.

**Gate:** provision/apply/destroy succeeds from a clean environment; a corrupt digest aborts before boot; failure cleanup leaves no domains or volumes behind.

### 15.4 Weekly cadence

- **Monday:** confirm milestone checklist, identify the week's riskiest item.
- **Tuesday–Thursday:** implementation and unit tests; no milestone scope changes mid-week.
- **Friday:** integration run, demo, update docs, short retro.
- **Every day:** keep `main` buildable; land work in small commits.
- **Buffer:** Week 8 contains at least three working days of hardening and release buffer. Protect it.

### 15.5 `v0.1.0` release criteria

- [ ] All "Must ship" scope is complete.
- [ ] Idempotency, verification, interruption, host-key, and secret tests are green.
- [ ] Ubuntu 24.04 and Debian 13 integration suites pass. Fedora 44 is required only if the should-ship dnf path lands.
- [ ] Canonical example and libvirt demo run from a clean checkout.
- [ ] Any should-ship feature that ships has its own tests; unfinished should-ship work is documented as deferred.
- [ ] No known high-severity security issue in the transport, sudo, or secret-handling paths.
- [ ] CLI help, config reference, quick start, and troubleshooting docs exist.
- [ ] `go vet`, formatting, and unit tests pass in CI.
- [ ] A tagged release and changelog are published.

---

## 16. Risks and mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Distro/package-manager differences | Executors work on one distro only | Capability detection; apt first-class; add dnf only with Fedora coverage and a two-distro-plus-Fedora matrix. |
| Sudo requires a password | Every apply fails late | Preflight `sudo -n` check with an actionable error; document the requirement. |
| SSH host-key handling blocks users | Users disable safety | Strict default plus a documented, explicit TOFU path; clear mismatch diagnostics. |
| systemd/firewall state is eventually consistent | False verification failures | Bounded retries with backoff; re-read after settle time. |
| Service or firewall verification is flaky | CI instability | Deterministic fixtures; retry only verification, never apply. |
| Parallel apply causes conflicts (should-ship only) | Corrupted target state | Keep sequential apply as the default; if parallelism lands, add dependency barriers, conflict serialization, and a per-target lock. |
| Interruption leaves unknown state | Silent corruption | Atomic writes, `incomplete` status, re-check on every run, signal handling. |
| Libvirt environment variance | Provider demo fails on another machine | Preflight libvirt connectivity; pin catalog; document prerequisites; clean failure messages. |
| Scope creep from EC2 or extra resources | Misses the 2-month date | Build the four core executors first; "Won't ship" list is binding; stretch work only after M6 gate. |
| Secrets leak into logs/state | Security incident | Redaction tests; hash file content; never serialize credentials. |
| Verification is mistaken for apply success | False confidence | Verification is mandatory and separate; exit code reflects verification, not command exit. |

---

## 17. After `v0.1.0`

In rough priority order:

1. **Deferred should-ship work:** full run history, multi-target fan-out, bounded parallelism, and the `group`, `directory`, and `firewall` executors.
2. **EC2 provider** behind the existing provider interface.
3. **Remote state and locking** for team use.
4. **Drift detection and scheduled runs** with notifications.
5. **More resources:** package repositories, templates, mounts, cron, sysctl, `exec` with explicit semantics.
6. **Advanced firewall support:** zones, rich rules, nftables.
7. **Secrets integration:** Vault, AWS Secrets Manager, SOPS.
8. **Policy and approvals** for production applies.
9. **Web UI / API** on top of the same engine.
10. **Additional platforms** and container targets.

---

## 18. Open decisions

| Decision | Recommendation | Deadline |
|---|---|---|
| CLI framework | `spf13/cobra` for subcommands; add only if it reduces boilerplate | M0 |
| YAML library | `gopkg.in/yaml.v3` with strict decoding | M0 |
| SSH library | `golang.org/x/crypto/ssh` + `github.com/pkg/sftp` | M0 |
| Config kinds | Split `Machine` and `Server` (ADR-001) | M0 |
| Live inspection in `plan` | Yes by default; `--offline` for structure only | M1 |
| State format | Versioned JSON under XDG state directory | M1 |
| Firewall default | Auto-detect; fail clearly for unknown backends; should-ship only | M4 if implemented |
| Package versions | Installed-state only in `v0.1.0`; version pinning as stretch | M3 |
| Sudo model | Passwordless `sudo -n` only in `v0.1.0` | M2 |
| Parallelism default | Sequential must-ship; `min(4, independent resources)` if the should-ship feature lands | Post-M6 if deferred |
| Multi-target definition | Multiple targets in one run with per-target isolation; should-ship only | Post-M6 if deferred |

---

## Appendix A — Current repository state

As of 2026-11-05:

- `cmd/main.go` calls `cli.Execute()`.
- `internal/cli` has Cobra root, `version`, and `help` commands.
- `internal/config`, `internal/engine`, `internal/catalog`, `internal/state`, `internal/target`, and `internal/provider` are package scaffolds.
- `go build ./...`, `go vet ./...`, and `go test ./...` pass.
- `img-catalog/images.yaml` contains a curated, digest-pinned image catalog for Ubuntu, Debian, and Fedora.
- `examples/machine.yaml`, `examples/server.yaml`, and `examples/README.md` exist.
- `.template.yaml` is the legacy desired-state sketch and remains ignored by git.
- The repository has three commits, including the initial skeleton and CLI scaffold.

### Immediate cleanup before M0 closes

1. Delete the leftover `internal/engine.go` root-level package file.
2. Remove the `test build` debug print from `cli.Execute`; rely on Cobra's built-in help.
3. Fix `Makefile`: build `./cmd`, and use `GOOS=windows` for the Windows artifact.
4. Add CI and the `docs/adr/` directory.
5. Add local state paths (`runs/`, `cache/`) and build output (`bin/`) to `.gitignore`.
6. Remove `.template.yaml` once the two example kinds fully replace it.

## Appendix B — Definition of done for any resource executor

A resource is not complete until it has:

- [ ] Schema validation and a documented example.
- [ ] A read-only `Check` that returns structured observed state.
- [ ] `Plan` action selection for create/update/delete/noop.
- [ ] `Apply` that performs only the planned action.
- [ ] `Verify` that re-reads the target and compares desired vs observed.
- [ ] An idempotency test: second apply is `unchanged`.
- [ ] A drift test: manually mutate, then detect and correct.
- [ ] A failure test: apply fails and the result is `failed`, not `changed`.
- [ ] Redaction review for any sensitive input.
- [ ] Integration coverage on at least one supported distro.
