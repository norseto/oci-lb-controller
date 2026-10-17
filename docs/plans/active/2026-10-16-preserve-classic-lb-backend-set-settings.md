---
title: Preserve existing Classic Load Balancer BackendSet settings
status: in_progress
review_status: none
template_version: 3.0.0
skill_version: 2.0.0
created_at: 2026-10-15T21:08:58Z
---

# Preserve existing Classic Load Balancer BackendSet settings

## Objective and Requirement Authority

Observable objective: when Classic Load Balancer backend membership is reconciled, the outgoing `UpdateBackendSet` request retains the existing BackendSet-level settings fetched from OCI while updating backends according to the existing registrar rules.

Classification: `ExecPlan`. No Strategy is required because one Classic LB adapter and its tests own the complete setting-preservation outcome. The other five SecurityCloud findings are separate from this work.

| ID | Authority | Requirement | Approval or source |
| --- | --- | --- | --- |
| U1 | U | Correct the reported BackendSet setting propagation omission: SSL configuration, application-cookie session persistence, LB-cookie session persistence, and backend connection limit. | User: "設定引き継ぎ漏れを修正する。まずは計画を作成して"; preceding finding assessment identified these four omissions. |
| U2 | U | Author the plan first; do not begin implementation in this planning turn. | The same user request, specifically "まずは計画を作成して". The exact implementation-approval record follows the applied `exec-plan/references/implementation-workflow.md` workflow. |
| U3 | U | After implementation, obtain an independent review and resolve all accepted actionable findings within this workstream, repeating review until none remain or an authorized stop condition applies. | User: "実行を承認します。実装してください。実装後、独立レビューによる妥当な指摘がない状態にしてください。" |
| R1 | R | Preserve existing capabilities and contracts outside U1, including backend membership/IP/port/weight construction, Policy and HealthChecker preservation, optional-value distinctions, routing, error handling, and reconciliation behavior. | User-provided AGENTS.md: General and Requirement Authority; current Classic adapter, provider dispatch, and controller code. |
| R2 | R | Keep source comments and authored Markdown prose in English, preserve existing source headers, and format Go changes with `make fmt` or `go fmt ./...`. | AGENTS.md and the supplied Language and Structure rules; existing source header convention. |
| R3 | R | Validate executed behavior, generate and inspect overall coverage with a minimum of 80%, and run `make vet`, `make test`, `make lint`, `make vulcheck`, and `make seccheck` before any commit. Do not create a PR with failing tests. | AGENTS.md: Testing and Test Coverage; supplied Evidence and Scope rules. |
| R4 | R | Keep planning, authorization records, review, and acceptance with the main agent; delegate authorized implementation and corrective edits to `implementer`. Complete the main agent's plan review before implementation. | Supplied AGENTS.md: Task Delegation in the Main Codex Agent; the applied execution workflow records this review as `Planning review result: GO`. |

## Scope and Existing Mechanism

Deliverables and ownership:

- `internal/controller/cloud/oci/loadbalancer/loadbalancer.go`: preserve the four omitted BackendSet-level settings when constructing the existing update request. Implementation ownership belongs to `implementer` after approval.
- `internal/controller/cloud/oci/loadbalancer/loadbalancer_test.go`: exercise `RegisterBackends` with configured and unconfigured settings and inspect the resulting request and serialized body. The same implementation owner maintains these tests.
- This plan: maintained by the main agent, including authorization, review, and observed evidence.

Verified baseline: repository revision `fe50252ca2203cdfd1fc13a884dbc978d38e98f8`; OCI Go SDK `v65.83.1` in `go.mod`. SecurityCloud finding `csf_bd921d2b0bd2435b5277e529` was observed at revision `9d3bac4298b6f4bb09b3208efa3d37661fb49139`; the relevant adapter logic remains unchanged.

`RegisterBackends` fetches the current BackendSet, constructs desired backends from Nodes, and calls `UpdateBackendSet`. It already preserves Policy and all nine HealthChecker fields. It does not populate the four fields in U1. `internal/controller/cloud/oci/provider.go` routes Classic LB requests to this adapter; `internal/controller/lbregistrar_controller.go` owns phase changes and retries.

The pinned SDK provides all required update fields:

| Existing BackendSet field | Update representation | Planned handling |
| --- | --- | --- |
| `BackendMaxConnections` | Same `*int` type | Preserve the pointer value, including nil and explicit zero. |
| `SessionPersistenceConfiguration` | Same `*SessionPersistenceConfigurationDetails` type | Pass through unchanged, including nil. |
| `LbCookieSessionPersistenceConfiguration` | Same `*LbCookieSessionPersistenceConfigurationDetails` type | Pass through unchanged, including nil. |
| `SslConfiguration` | `*SslConfigurationDetails`, distinct from the read type | Keep nil when absent; otherwise explicitly copy every corresponding field. |

SSL fields to copy are `VerifyDepth`, `VerifyPeerCertificate`, `HasSessionResumption`, `TrustedCertificateAuthorityIds`, `CertificateIds`, `CertificateName`, `CipherSuiteName`, `Protocols`, and `ServerOrderPreference`. The last field requires conversion from the read enum to the SDK write enum with the same underlying string value. Preserve explicit false/zero values and nil versus empty slices; do not introduce defaults. A local nil-guarded conversion in the existing function is sufficient; no shared abstraction or new dependency is needed.

The two session persistence modes are mutually exclusive according to the SDK. Tests must use separate valid fixtures for them. The SDK documentation states that explicit null disables session persistence. Actual OCI TLS removal and server-side treatment of omitted fields have not been reproduced. This plan proves preservation of fetched settings in the request, rather than claiming a live TLS downgrade was demonstrated.

Non-goals: per-backend flag preservation, NLB changes, no-op update suppression, Endpoints fan-out, new concurrency control or retries, WorkRequest handling, CRD/RBAC changes, dependency upgrades, future SDK-field automation, and release/development supply-chain findings. No deployment, live OCI mutation, SecurityCloud rescan/status change, commit, or PR is requested by this plan.

The lifecycle boundary remains unchanged: this function returns after the OCI update call returns; it does not wait for Classic LB work-request completion. Existing controller retry behavior remains responsible for registration errors. The fix preserves settings from the fetched snapshot and adds no guarantee about concurrent external edits.

## Change Procedure

All commands run from `/workspaces/oci-lb-controller` with the existing workspace environment.

1. Before implementation, obtain and record the exact approval quote, confirm the current planning review is `GO`, and re-read the two owned source files and pinned SDK types. Check `git status --short` for ownership conflicts. A changed baseline alone does not require a new plan; a material change to these assumptions does. Set `status: in_progress` only when implementation begins.
2. Delegate U1 and R1-R3 implementation to one `implementer`, supplying the approved plan, current revision, the two owned paths, expected behavior, commands, and stop conditions. The main agent keeps the plan and review ownership under R4.
3. Add focused regression tests through the existing fake client and `RegisterBackends`. Run `go test ./internal/controller/cloud/oci/loadbalancer -run TestRegisterBackends -count=1`. At least the configured-settings case must fail because the current request loses the settings; unrelated fixture/setup failures are not evidence of the defect.
4. Populate the two persistence fields and connection limit directly from the fetched BackendSet. Convert non-nil SSL configuration using the field mapping above. Keep backend construction, Policy, HealthChecker, signatures, logging, and error propagation intact. Do not mutate the fetched settings while constructing the request.
5. Run `make fmt`, then `go test ./internal/controller/cloud/oci/loadbalancer -count=1` and `go test ./internal/controller/cloud/oci/...`. Expect the new preservation assertions and existing backend/provider/NLB tests to pass. Check the diff for unintended formatting edits outside the owned paths.
6. Run `make test` to exercise repository tests and provision the existing envtest assets. Its manifest/code-generation prerequisites are existing validation behavior; inspect generated diffs and do not include unrelated changes. Then run `go test ./... -coverprofile=coverage.out` and `go tool cover -func coverage.out`. Expect all tests to pass and overall statement coverage to be at least 80%. The default test invocation does not enable the `e2e` build tag; no live cluster test is required. `internal/controller/suite_test.go` already uses the local `bin/k8s/1.29.0-<os>-<arch>` envtest asset directory.
7. Before any subsequently requested commit, also complete `make vet`, `make lint`, `make vulcheck`, and `make seccheck`, together with the required `make test` result for that revision. Report failures separately and do not fix unrelated findings under this plan. Use the existing environment; if environment/executable selection causes a failure, consult the user before changing it. Follow repository and sandbox requirements for a confirmed cache-permission failure.
8. The main agent reviews the actual diff and executed evidence against U1/R1, tracing controller/provider dispatch, current BackendSet retrieval, conversion, SDK request serialization, and the update/error path. Obtain a separate read-only review from an agent that did not implement the change, with fresh context containing the original requirements and the behaviorally complete review scope. The main agent owns finding classification and acceptance. Use `review-findings` for accepted corrections, return them to the same implementation owner, and repeat the independent review over the complete affected flow until no accepted actionable findings remain. Stop if a correction changes scope or recreates the same failure without progress.
9. Record implementation results and unresolved limitations here. Keep the plan active until explicit human outcome review; only then set terminal status and move it according to `docs/plans/README.md`.

Stop the affected work if a fetched setting cannot be represented by the pinned SDK, preserving a value requires an API/design/dependency change, another owner has conflicting edits, or validation requires external mutation. Report the exact decision instead of discarding a setting or expanding scope.

## Acceptance and Evidence

### Planned Evidence

| Requirement | Proof and expected result |
| --- | --- |
| U1, R1: configured settings | Execute `RegisterBackends` with non-default SSL fields, a positive connection limit, and each persistence mode in separate cases. The captured `UpdateBackendSetRequest` retains every setting; assert all SSL fields, not only certificate name. Policy and all HealthChecker values also remain unchanged. |
| U1, R1: absent and zero values | Execute cases with nil optional settings, explicit false/zero pointers, unset SSL enum, and nil/empty SSL slices. Request values preserve those distinctions and do not create or enable settings that were absent. |
| U1, R1: request consumer | Construct an HTTP request using the captured request's SDK `HTTPRequest` method without sending it. Decode its JSON body and verify retained SSL, persistence, and connection-limit values. Assert the pinned SDK's actual representation of absent values; do not infer serialization from source text or Go pointers alone. |
| U1, R1: backend changes and removal | Existing IP/port/weight and NodePort fallback tests continue to pass. Execute an empty-target case with existing optional settings and verify an empty backend list with the same BackendSet-level settings retained. |
| R1: failures and surrounding consumers | Preserve existing client/get/update error behavior. A failed current-set fetch must not send an update; an update failure must remain an error. Scoped cloud tests and the repository suite pass. Review provider dispatch and controller phase/retry ownership without editing them. |
| R2 | Review comments and retained source headers; run the repository formatter and inspect the resulting diff. |
| R3 | Run the commands in steps 5-7 at their applicable stage. Coverage reports an overall total of at least 80%; no check is reported as passed when pending, omitted, or failed. |
| U2, R4 | During planning, only this plan is changed. Before implementation, the main agent records approval and gives the bounded write assignment; the main agent retains review and lifecycle records. |
| U3 | The independent reviewer inspects actual source and tests across the affected flow and returns concrete findings or explicitly reports none. The main agent assesses each finding against U1/R1 and records the final review scope, remaining limitations, and absence of accepted actionable findings after any repair cycle. Passing tests alone do not satisfy this requirement. |

The primary regression evidence is executed request construction and serialization. No source-text assertion substitutes for it. Live OCI verification is omitted because it would modify external infrastructure and is unnecessary to prove the requested propagation fix; live TLS impact remains unverified.

### Observed Evidence

- Original failure condition: the Classic adapter fetched a BackendSet but omitted SSL configuration, both session persistence fields, and the connection limit from its update request. Policy and all nine HealthChecker fields were already preserved.
- Implemented correction: the owned adapter now converts all nine SSL fields when SSL is present and passes through the two persistence configurations and `BackendMaxConnections`. Backend construction, request target, Policy/HealthChecker, error handling, and surrounding controller/provider behavior are unchanged. No helper, dependency, or external API was added.
- Regression evidence: after adding tests and before editing the adapter, `go test ./internal/controller/cloud/oci/loadbalancer -run TestRegisterBackends -count=1` exited 1 with six cases failing on omitted settings in the captured request and SDK-generated body. After the fix, `go test ./internal/controller/cloud/oci/loadbalancer -count=1` and `go test ./internal/controller/cloud/oci/...` exited 0. The tests execute `RegisterBackends` and SDK `HTTPRequest`, covering configured and absent values, false/zero/empty values, both persistence modes separately, all SSL and HealthChecker fields, empty backend membership, non-mutation of the fetched set, and fetch/update errors.
- Repository validation: `make fmt`, `make test` (including its `vet` prerequisite), and `go test ./... -coverprofile=coverage.out` exited 0. No out-of-scope generated or formatting changes were produced. The profile passed the local secret precheck; `go tool cover -func coverage.out` exited 0 and reported overall coverage of `80.6%` and Classic `RegisterBackends` coverage of `96.7%`. The main agent separately inspected the profile's totals and actual diff. `git diff --check` exited 0.
- Main-agent review: checked the actual two-file diff against U1/R1, all read/write SDK fields, the unchanged provider and single/multi-service caller paths, request serialization, and the update/error path. Invariants checked were preservation of controller-unowned BackendSet settings, absence of invented defaults or input mutation, unchanged backend construction, and unchanged routing/retry ownership. The final test cleanup correction was also inspected; no actionable issue remains.
- Independent review: a separate agent with fresh context and no implementation role reviewed the original contract, actual diff, SDK types/serialization, provider/controller callers, and error paths. After the accepted lint correction, the same reviewer repeated the review over the behaviorally complete scope and confirmed that the close-error finding was resolved and no new or remaining actionable finding existed. The reviewer inspected executed-function/HTTP-body test design without redundantly running the same tests; the main agent evaluated that review together with the successful validation evidence.
- Accepted correction and convergence: additional lint validation identified one unchecked `httpReq.Body.Close` return value in the new test (`ACTIONABLE / AUTOFIX`). The same implementation owner added deferred error checking and test failure reporting. The final `make fmt`, scoped package test, `make lint` (`0 issues`), full `go test ./... -coverprofile=coverage.out`, profile precheck, and coverage inspection all exited 0; overall coverage remains `80.6%`. The production fix did not change during this correction. The original propagation finding and the accepted test cleanup finding are resolved, with no recurrence, oscillation, or new actionable finding at final review.
- Remaining limits: `make vulcheck` and `make seccheck` remain reserved for a subsequently requested commit under step 7. No live OCI or Kubernetes deployment, SecurityCloud mutation, commit, or PR was performed. Live OCI server-side semantics remain unverified.

## Authorization, Current State, and Recovery

Implementation approval quote: "実行を承認します。実装してください。実装後、独立レビューによる妥当な指摘がない状態にしてください。" (2026-10-15T23:29:00Z)

Planning review result: GO (2026-10-15T23:29:29Z). The main agent reviewed the complete U1-U3/R1-R4 contract and ordered procedure after adding the explicitly requested independent review. The current revision changes only the committed plan from the previously tested implementation baseline; the owned paths and SDK pin remain unchanged and the baseline package test passes. Independent read-only review adds no implementation write boundary or external mutation. No blocking finding remains.

Current state: in progress; implementation and validation are complete, and independent re-review has no remaining actionable findings. Ready for human outcome review.

Completed material outcomes: all four omitted settings are preserved in the outgoing Classic LB request; regression, full-suite, and lint validation passed with coverage above 80%; the accepted cleanup correction is complete and both final reviews have no remaining actionable issue.

Active failures or unresolved decisions: no active failure or actionable review finding. The implementation is an uncommitted diff against `aa33ce08f0fef31d08d1cdd5079e588844f466c2`. Live OCI TLS-removal behavior remains unverified and is not an implementation acceptance claim; explicit human outcome confirmation remains required for terminal plan closure.

Safe retry or rollback posture: changes are limited to this plan and the two owned implementation files, with ignored validation profiles/tool caches. Future test failures leave local edits available for inspection; preserve unrelated user edits. Any authorized rollback must be limited to this work's hunks in the two implementation files. No cloud state or credentials were changed by validation.

Next safe action: present the completed implementation, validation, and clean independent re-review for human outcome review. Keep the current lifecycle state until explicit outcome confirmation.
