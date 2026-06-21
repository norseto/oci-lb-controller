---
title: Fix security review findings for LBRegistrar secret trust and provider log fields
status: completed
review_status: approved
template_version: 2.12.0
skill_version: 1.13.0
created_at: 2026-06-20T22:54:20Z
---

# Fix security review findings for LBRegistrar secret trust and provider log fields

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.
Treat section ownership as fail-closed: `Progress` owns factual state transitions and stop points, `Surprises & Discoveries` owns still-active findings, `Decision Log` owns durable reusable decisions, and `Outcomes & Retrospective` owns milestone-level outcomes and remaining gaps. `Concrete Steps` owns planned commands, `Validation and Acceptance` owns planned proof, and `Implementation Completion Report` owns completed validation results.
Keep the title in this header and the YAML front matter in sync.
All recorded datetimes in this document must use UTC timestamps in the form `YYYY-MM-DDTHH:MM:SSZ`. Do not use date-only values.
Use `status` to track lifecycle progress. Allowed values are `planning`, `in_progress`, `completed`, and `failed`.
Use `review_status` to track human review state. Allowed values are `none` and `approved`.
Set `status: planning` while authoring or revising the ExecPlan before implementation starts, `status: in_progress` once implementation begins, `status: completed` only after human review confirms that all planned work and validation are finished, and `status: failed` only after human confirmation that the work is stopped due to failure or explicit cancellation.
Set `review_status: none` while authoring, implementing, or waiting for human review. Set `review_status: approved` only after a human explicitly approves the current outcome. Implementation-complete but review-pending work should remain `status: in_progress` with `review_status: none`.
No parent Strategy exists. Strategy was considered and is not required because the work remains one bounded security-review remediation workstream: document one trust boundary and remove one repeated sensitive log field pattern.
Before implementation starts, this ExecPlan must explicitly name its primary deliverables as concrete output files or concrete repository changes. An ExecPlan without concrete deliverables is invalid.
This repository's additional plan directory rules are in `docs/plans/README.md`; those rules define `active/`, `completed/`, and `failed/` storage only and do not replace this template.

## Purpose / Big Picture

A security review identified two accepted issues in this repository. After this work, operators who read `README.md` will understand that permission to create or edit `LBRegistrar` resources is effectively cluster-admin-equivalent because the controller can be directed to read a Kubernetes Secret in any namespace. The controller will also stop passing the OCI configuration provider object to structured log fields, removing a fragile latent credential exposure path where future logging behavior could serialize sensitive provider internals.

## Primary Deliverables

- Update `README.md` with a `## Security Considerations` section that documents the `LBRegistrar` Secret-access trust model and practical RBAC-narrowing caveat.
- Update `internal/controller/cloud/oci/loadbalancer/loadbalancer.go` by removing the `"provider", provider` structured log field from three debug log calls while preserving client creation, backend retrieval, and backend registration behavior.
- Update `internal/controller/cloud/oci/networkloadbalancer/networkloadbalancer.go` by removing the `"provider", provider` structured log field from three debug log calls while preserving client creation, backend retrieval, backend registration, and work request behavior.

## Tracked Decision Record

No tracked decision record required.

## External API Parameters and Capability Checks

No external API used. This work changes documentation and log call arguments only. It adds no dependency, changes no dependency version, and does not call a network API during implementation.

- [x] (2026-06-20T22:54:20Z) Implementation targets and local scoped tests are available in this workspace.
  Capability Check Status: PASS
  Checked at: 2026-06-20T22:54:20Z
  Evidence: `rg` found the README anchor text and all six `"provider", provider` log fields in the expected files. `go test ./internal/controller/cloud/...` completed successfully with `ok` for `cloud/oci`, `cloud/oci/loadbalancer`, and `cloud/oci/networkloadbalancer`.

## Implementation Gate (Hard Stop)

Plan-document creation and revision are allowed as plan-only work before implementation approval. No implementation-target changes are allowed until the required unlock checklist is complete, all items in `External API Parameters and Capability Checks` are completed and marked checked, and explicit user approval is recorded.

Blocked implementation-target actions before gate pass:
- `apply_patch` against implementation-target files
- Any file write/edit operation outside the current ExecPlan document
- Code generation or migration commands that modify implementation-target files

Required unlock checklist (must be complete before implementation):
- [x] (2026-06-20T22:54:20Z) Capability checks are complete and recorded as `Capability Check Status: PASS` with `Checked at` and `Evidence`.
- [x] (2026-06-20T22:54:20Z) Design risk review is complete, including likely breakage cases, reuse assumptions, coexistence risks, lifecycle/layout/state risks, boundary cases, and stop conditions that would require replanning or user confirmation.
- [x] (2026-06-20T22:54:20Z) `Validation and Acceptance` includes acceptance criteria derived from user requirements, primary deliverables, and documented repository constraints.
- [x] (2026-06-20T22:54:20Z) Each non-obvious acceptance criterion maps back to its source, or the plan explains why a separate mapping list is unnecessary for this bounded workstream.
- [x] (2026-06-21T02:24:34Z) Explicit user approval is recorded as `User approval quote: "<exact quote>" (<UTC timestamp>)`.

User approval quote: "それでは実行を承認します。実行してください。" (2026-06-21T02:24:34Z)

Approval policy:
- No code changes are allowed without an exact approval quote.
- Interpreted or implicit approval is invalid. Planning requests, scope clarifications, and review comments are not implementation approval.
- When explicit implementation approval is provided, record the user's exact approval quote verbatim, preserving the original language and punctuation.

## Progress

- [x] (2026-06-20T22:54:20Z) Recreated this ExecPlan at the same path using template version 2.12.0 and skill version 1.13.0.
- [x] (2026-06-20T22:54:20Z) Verified current workspace anchors for the planned README insertion and six provider log field removals.
- [x] (2026-06-20T22:54:20Z) Ran scoped capability validation with `go test ./internal/controller/cloud/...`; the command passed.
- [x] (2026-06-21T02:24:34Z) Recorded explicit implementation approval quote and set `status: in_progress`.
- [x] (2026-06-21T02:25:03Z) Applied documentation and log-field changes in `README.md`, `loadbalancer.go`, and `networkloadbalancer.go`.
- [x] (2026-06-21T02:25:50Z) Ran planned validation and recorded results in `Implementation Completion Report`.
- [x] (2026-06-21T02:26:21Z) Re-reviewed the behaviorally complete scope for the original findings and found no remaining or new actionable issues.
- [x] (2026-06-21T02:29:53Z) Fixed review finding P2 by changing the README role name from `lbregistrar-editor` to the actual bundled ClusterRole name `lbregistrar-editor-role`.
- [x] (2026-06-21T06:14:10Z) Human review approved the work; set `review_status: approved` and `status: completed`.
- [x] (2026-06-21T06:14:10Z) Move this plan to `docs/plans/completed/` per `docs/plans/README.md`.

## Surprises & Discoveries

None currently active.

## Decision Log

- Decision: Scope the `LBRegistrar` Secret-access finding to documentation only, not RBAC manifest changes or controller namespace enforcement.
  Rationale: The accepted remediation is to document the current trust model and warn operators. Changing RBAC shape or cache scoping would be a separate behavior and deployment design change beyond this bounded security-review fix.
  Timestamp: 2026-06-20T22:54:20Z
- Decision: Exclude the Network Load Balancer `GetBackendSet` nil-pointer review item from this workstream.
  Rationale: This plan only carries the two accepted findings. The nil-pointer item was previously judged outside scope and would require a separate user decision if reconsidered.
  Timestamp: 2026-06-20T22:54:20Z

## Outcomes & Retrospective

Implementation reached a reviewable milestone on 2026-06-21T02:25:50Z. The planned README section was added, the six provider log fields were removed, validation passed, and the follow-up review finding was fixed by using the actual bundled `lbregistrar-editor-role` ClusterRole name. Human review approved the outcome on 2026-06-21T06:14:10Z, so this plan is complete.

## Design Risk Review

- Risk: The README could imply that narrowing Secret access is a supported runtime toggle when it also requires controller cache scoping or a non-cached Secret reader.
  Evidence or reason: `internal/controller/lbregistrar_controller.go` uses controller-runtime client access for Secrets, and `config/rbac/role.yaml` currently grants cluster-wide `get`, `list`, and `watch` on Secrets.
  Mitigation in this plan: Write the README section as a trust-model warning and clearly describe namespace narrowing as installer-side customization that must account for cached-client `list` and `watch` requirements.
  Acceptance proof: Human review of the new README section confirms it states both admin-equivalence of `LBRegistrar` edit rights and the cached-client caveat.

- Risk: Removing `"provider", provider` could accidentally remove the provider variable or alter the OCI client construction flow.
  Evidence or reason: The provider argument is required by `newLBClient(provider)` and `newNLBClient(provider)` immediately after the affected log calls.
  Mitigation in this plan: Delete only the structured log field arguments from the six `logger.V(1).Info(...)` calls and preserve all function signatures, provider variables, client construction calls, messages, and error handling.
  Acceptance proof: `go test ./internal/controller/cloud/...` and `go test ./...` pass, and `rg -n '"provider", provider' internal/` returns no matches.

- Risk: The implementation could drift into fixing the excluded NLB nil-pointer issue or broader RBAC behavior.
  Evidence or reason: Adjacent code contains a backend conversion loop in `networkloadbalancer.go`, and RBAC manifests are near the documented Secret trust model.
  Mitigation in this plan: Treat backend conversion changes, RBAC manifest changes, controller cache changes, and non-cached Secret lookup changes as explicit non-goals.
  Acceptance proof: `git diff -- README.md internal/controller/cloud/oci/loadbalancer/loadbalancer.go internal/controller/cloud/oci/networkloadbalancer/networkloadbalancer.go` shows only the planned README section and provider log field removals.

Stop conditions requiring user confirmation or a return to Strategy:
- Any required fix expands into RBAC manifest changes, controller cache configuration, Secret lookup architecture, OCI SDK behavior changes, or new runtime policy controls.
- Any target anchor is missing or the six provider log call sites no longer match the expected pattern.
- Validation reveals a compile or test failure unrelated to the planned edits.

## Context and Orientation

This repository is a Kubernetes controller for registering cluster nodes as Oracle Cloud Infrastructure load balancer backends.

`api/v1alpha1/lbregistrar_types.go` defines `LBRegistrar`, a cluster-scoped custom resource. Its `spec.apiKey.privateKey` fields identify a Kubernetes Secret name, key, and optional namespace. `internal/controller/lbregistrar_controller.go` reads that Secret and builds an OCI `common.ConfigurationProvider`. `config/rbac/role.yaml` grants the manager ClusterRole cluster-wide `get`, `list`, and `watch` permissions on Secrets.

`common.ConfigurationProvider` is the OCI SDK provider object used to create load balancer clients. The current load balancer and network load balancer implementations pass that provider as a structured log field named `provider` in six debug log calls. The current logger does not expose a known leak in normal operation, but logging the provider object is unnecessary and fragile because the provider is derived from raw private-key material.

`README.md` currently has `## LBRegistrar Spec`, then the sentence `See the [docs](docs) directory for design details.`, then `## License`. The planned security section should be inserted before `## License`.

## Observable Behavior and Compatibility

Controller reconciliation behavior, OCI API calls, function signatures, log message text, and exported interfaces must remain unchanged. The intentional observable change is that debug logs no longer include a `provider` key-value field at the six affected call sites. README documentation gains a new operator-facing security note; no migration or compatibility break is intended.

## Lifecycle and Consistency Semantics

This work does not add lifecycle, shutdown, cleanup, retry, persistence, cache, or recovery behavior. The code change is a synchronous log argument removal, and the documentation change has no runtime state.

## Plan of Work

1. In `README.md`, insert a new `## Security Considerations` section between the existing design-details sentence and `## License`. The section must explain that `LBRegistrar` edit rights should be treated as cluster-admin-equivalent because the controller can read referenced Secrets, and that narrowing Secret RBAC requires installer-side care for controller-runtime cached-client `list` and `watch` behavior.
2. In `internal/controller/cloud/oci/loadbalancer/loadbalancer.go`, remove only the `, "provider", provider` arguments from the `logger.V(1).Info(...)` calls in `loadBalancerClient`, `GetBackendSet`, and `RegisterBackends`.
3. In `internal/controller/cloud/oci/networkloadbalancer/networkloadbalancer.go`, remove only the `, "provider", provider` arguments from the `logger.V(1).Info(...)` calls in `loadBalancerClient`, `GetBackendSet`, and `RegisterBackends`.
4. Do not edit `config/rbac/role.yaml`, controller cache configuration, Secret lookup code, generated API files, or backend conversion logic.

## Concrete Steps

All commands run in `/workspaces/oci-lb-controller`.

1. Confirm the implementation gate has a recorded explicit user approval quote. Stop if `User approval quote` is still `Not yet provided`.
2. Edit `README.md` at the anchor before `## License` to add the planned `## Security Considerations` section. Stop if the anchor text is missing or the surrounding README structure has changed enough to make the placement ambiguous.
3. Edit `internal/controller/cloud/oci/loadbalancer/loadbalancer.go` and remove only `, "provider", provider` from the three affected debug log calls. Stop if any call site has changed shape or the provider is used only by the removed log field.
4. Edit `internal/controller/cloud/oci/networkloadbalancer/networkloadbalancer.go` and remove only `, "provider", provider` from the three affected debug log calls. Stop if any call site has changed shape or the provider is used only by the removed log field.
5. Run scoped tests:

       go test ./internal/controller/cloud/...

   Expected transcript shape:

       ok   github.com/norseto/oci-lb-controller/internal/controller/cloud/oci
       ok   github.com/norseto/oci-lb-controller/internal/controller/cloud/oci/loadbalancer
       ok   github.com/norseto/oci-lb-controller/internal/controller/cloud/oci/networkloadbalancer

6. Run scoped vet:

       go vet ./internal/controller/cloud/... ./api/...

   Expected result: no output and exit code 0.

7. Confirm provider log fields are removed:

       rg -n '"provider", provider' internal/

   Expected result: no matches.

8. Run full repository tests:

       go test ./...

   Expected result: all packages pass.

9. Before commit or PR only, run repository-required broader checks:

       make vet
       make test
       make lint
       make vulcheck
       make seccheck

   Expected result: all commands pass. If sandbox cache restrictions affect Go or envtest paths, rerun with the writable cache pattern documented in `AGENTS.md`, such as `GOPATH=/tmp/go GOCACHE=/tmp/go-build HOME=/tmp/envtest-home make test`, and record the exact command and result.

10. After implementation is validated and human review confirms completion, update this plan according to `docs/plans/README.md`: set final status only after human confirmation and move the file from `docs/plans/active/` to `docs/plans/completed/`.

## Validation and Acceptance

### Requirement to Acceptance Mapping

- User requirement: fix accepted security review findings for documentation and provider log fields. Acceptance: `README.md` contains a `## Security Considerations` section that states `LBRegistrar` create/update rights are effectively cluster-admin-equivalent and explains the cached-client caveat for narrowing Secret access; the six planned log calls no longer pass the provider object.
- Primary deliverables: update `README.md`, `loadbalancer.go`, and `networkloadbalancer.go` only. Acceptance: `git diff -- README.md internal/controller/cloud/oci/loadbalancer/loadbalancer.go internal/controller/cloud/oci/networkloadbalancer/networkloadbalancer.go` shows only the planned documentation and log-field edits, with no backend conversion, RBAC, cache, or Secret lookup changes.
- Repository constraint: code comments and Markdown prose authored in repository files must be English. Acceptance: the new README text is written in English.
- Repository constraint: format Go code with `go fmt ./...` or `make fmt` when formatting changes are needed. Acceptance: this plan changes only log call arguments and should remain gofmt-compatible; if formatting differs after edit, run `go fmt ./...` and record it.
- Validation requirement: existing behavior must keep compiling and tests must pass. Acceptance: `go test ./internal/controller/cloud/...`, `go vet ./internal/controller/cloud/... ./api/...`, and `go test ./...` pass.
- Design risk mitigation: provider removal must be limited to log fields. Acceptance: `rg -n '"provider", provider' internal/` returns no matches, while the provider remains passed into OCI client constructors.

No overlap or interleaving validation is required because this work does not introduce multiple runtime producers, phases, asynchronous operations, or state transitions that can affect one caller-visible condition.

## Testing Strategy

Use existing unit and package tests. The code changes only remove structured log field arguments, so new unit tests would assert logging internals rather than controller behavior and are not required. `go test ./internal/controller/cloud/...` provides focused compile and regression coverage for the touched packages. `go vet ./internal/controller/cloud/... ./api/...` checks nearby Go correctness. `go test ./...` provides full-repository regression coverage after the documentation and log edits.

No integration, envtest, or end-to-end validation is required for implementation because no reconciliation logic, Kubernetes API interaction, OCI request construction, RBAC manifest, controller cache configuration, or public API changes. Repository-required `make vet`, `make test`, `make lint`, `make vulcheck`, and `make seccheck` are reserved for commit or PR readiness unless the user asks for that phase.

## Naming and Semantic Changes

None. No identifiers, parameters, exported interfaces, or data cardinality change.

## Idempotence and Recovery

The planned edits are idempotent text changes. Reapplying the README insertion must be avoided if the `## Security Considerations` section already exists; in that case, update the existing section instead of adding a duplicate. Removing `, "provider", provider` is safe to repeat because the absence of matches is the expected final state.

Rollback is limited to the three implementation-target files and this plan document. Do not use destructive repository-wide commands. If rollback is requested, revert only the specific edited files after confirming the user's intended scope.

## Artifacts and Notes

Expected log edit shape:

    before: logger.V(1).Info("Creating Load Balancer client", "provider", provider)
    after:  logger.V(1).Info("Creating Load Balancer client")

Expected provider-field search after implementation:

    $ rg -n '"provider", provider' internal/
    <no output>

## Interfaces and Dependencies

No dependencies change. No external service interface changes. The following Go function signatures must remain unchanged:

    func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec) ([]*models.LoadBalanceTarget, error)
    func RegisterBackends(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec, targets *corev1.NodeList) error

The package-local `loadBalancerClient` helpers in both OCI load balancer packages must still accept `common.ConfigurationProvider` and pass it to the OCI SDK client constructor.

## Implementation Completion Report

- Completed implementation-target edits:
  - `README.md`: added `## Security Considerations` documenting `LBRegistrar` Secret-access trust and installer-side RBAC/cache narrowing caveats.
  - `internal/controller/cloud/oci/loadbalancer/loadbalancer.go`: removed the `provider` structured log field from three debug log calls.
  - `internal/controller/cloud/oci/networkloadbalancer/networkloadbalancer.go`: removed the `provider` structured log field from three debug log calls.
- Validation run:
  - `go test ./internal/controller/cloud/...`: passed; `cloud/oci`, `cloud/oci/loadbalancer`, and `cloud/oci/networkloadbalancer` reported `ok`.
  - `go vet ./internal/controller/cloud/... ./api/...`: passed with no output.
  - `rg -n '"provider", provider' internal/`: returned no matches.
  - `go test ./...`: passed for all repository packages.
- Post-fix review:
  - Original failure condition: `LBRegistrar` Secret-access trust was not documented, and debug log calls passed the OCI provider object as a structured field.
  - Reviewed behaviorally complete scope: `README.md`, `internal/controller/cloud/oci/loadbalancer/loadbalancer.go`, `internal/controller/cloud/oci/networkloadbalancer/networkloadbalancer.go`, and the provider use that feeds OCI client construction.
  - Invariants checked: no RBAC/cache/Secret lookup behavior changed; no backend conversion logic changed; provider values still reach `newLBClient(provider)` and `newNLBClient(provider)`; no `internal/` log call still passes `"provider", provider`.
  - Result: original findings are resolved, and no remaining or new actionable findings were found in the reviewed scope.
- Review finding follow-up:
  - Finding: README named `lbregistrar-editor`, but the bundled ClusterRole is `lbregistrar-editor-role` in `config/rbac/lbregistrar_editor_role.yaml`.
  - Fix made: changed the README warning to name `lbregistrar-editor-role`.
  - Re-review scope: README security section and bundled LBRegistrar RBAC role manifest names.
  - Result: the review finding is resolved, and no remaining or new actionable findings were found in the re-reviewed scope.
- Planned validation omitted:
  - `make vet`, `make test`, `make lint`, `make vulcheck`, and `make seccheck` were not run because this work has not proceeded to a commit or PR phase; the plan marks them as before-commit-or-PR checks.

## Change Notes

- 2026-06-20T22:54:20Z: Recreated the plan at the existing path using the current ExecPlan template and skill versions. Added the required `Design Risk Review`, refreshed capability evidence, clarified the implementation gate, and kept implementation approval unset.
- 2026-06-21T02:25:50Z: Recorded implementation approval, applied the planned README and provider log-field changes, and documented passing validation results. Kept the plan active pending human review.
- 2026-06-21T02:26:21Z: Added post-fix review evidence for the behaviorally complete scope.
- 2026-06-21T02:29:53Z: Fixed review finding P2 by using the actual bundled `lbregistrar-editor-role` ClusterRole name in README.
- 2026-06-21T06:14:10Z: Closed the plan after human review approval and prepared it for move to `docs/plans/completed/`.
