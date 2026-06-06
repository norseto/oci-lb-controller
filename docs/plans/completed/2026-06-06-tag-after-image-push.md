---
title: Tag Releases After Image Publication
status: completed
review_status: approved
template_version: 2.11.0
skill_version: 1.12.0
created_at: 2026-06-06T12:35:42Z
---

# Tag Releases After Image Publication

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.
Treat section ownership as fail-closed: `Progress` owns factual state transitions and stop points, `Surprises & Discoveries` owns still-active findings, `Decision Log` owns durable reusable decisions, and `Outcomes & Retrospective` owns milestone-level outcomes and remaining gaps. `Concrete Steps` owns planned commands, `Validation and Acceptance` owns planned proof, and `Implementation Completion Report` owns completed validation results.
Keep the title in this header and the YAML front matter in sync.
All recorded datetimes in this document use UTC timestamps in the form `YYYY-MM-DDTHH:MM:SSZ`.
Use `status` to track lifecycle progress. Allowed values are `planning`, `in_progress`, `completed`, and `failed`.
Use `review_status` to track human review state. Allowed values are `none` and `approved`.
Set `status: planning` while authoring or revising the ExecPlan before implementation starts, `status: in_progress` once implementation begins, `status: completed` only after human review confirms that all planned work and validation are finished, and `status: failed` only after human confirmation that the work is stopped due to failure or explicit cancellation.
Set `review_status: none` while authoring, implementing, or waiting for human review. Set `review_status: approved` only after a human explicitly approves the current outcome. Implementation-complete but review-pending work should remain `status: in_progress` with `review_status: none`.
No parent Strategy exists for this workstream. The repository plan directory rules in `docs/plans/README.md` provide additional placement context, but this ExecPlan follows the `exec-plan` skill template as the canonical structure.
Before implementation starts, this ExecPlan names its primary deliverables as concrete repository changes.
This ExecPlan carries one delegated repository-tracked bounded decision about the exact script split for release finalization.

When authoring or validating an ExecPlan, ask the user to resolve ambiguities. During implementation, resolve ambiguities autonomously unless the user has explicitly required confirmation.

## Purpose / Big Picture

The release CI currently creates and pushes the Git release tag before the container image is built, pushed, and signed. If any later image publication step fails, the remote tag already exists and must be deleted manually before retrying.

After this change, a release version update still publishes the same image tag, but the workflow is split so image build/push completes in one job and signing plus Git tag creation happens in a later dependent job. A failed signing attempt should leave the Docker Hub image available for retry, leave no new Git release tag behind, and allow GitHub Actions failed-job rerun to retry only the signing and tag finalization job.

## Primary Deliverables

- Update `.github/workflows/tag-release.yml` so Docker image build/push runs in a `build-and-push` job and cosign signing plus Git tag creation runs in a dependent `sign-and-tag` job.
- Update or split `hack/tag-release.sh` so version validation and post-publication Git tag creation can be run at the correct workflow phases.
- Preserve existing release version parsing from `version.go`, image tag naming, Docker platforms, cosign signing, and `release-${MAJOR.MINOR}` branch creation for versions ending in `.0-beta.1`.
- Preserve the release-image identity by signing only the digest emitted by the `build-and-push` job. Do not resolve the digest back from Docker Hub by tag as a fallback.
- Make release finalization retry-safe so an existing `v${VERSION}` tag does not prevent required `.0-beta.1` release branch creation or push.
- Check remote `origin` refs, not only local checkout refs, when deciding whether release Git tags and release branches already exist.
- When remote release refs already exist, verify that they point to the current release commit; fail closed instead of overwriting mismatched refs.

## Tracked Decision Record

Decision: Whether to keep `hack/tag-release.sh` as the post-publication finalization script with a new validation-only helper, or split it into clearly named validation and finalization scripts.

This decision is bounded to one workstream because both options only affect the same release workflow and helper-script boundary. It does not introduce a second primary outcome, a separate rollout track, or a cross-repository dependency.

Evidence required: inspect all current consumers of `hack/tag-release.sh`, `hack/get-version.sh`, release branch creation, and `git describe --always`; choose the smallest script structure that keeps pre-build metadata generation free of tag creation while preserving existing behavior.

Return to `plan-strategy` if the implementation reveals that release tagging, image publication, and branch creation must be split into independently reviewed workflows or independently releasable tracks.

## External API Parameters and Capability Checks

The implementation changes GitHub Actions workflow ordering. It does not add a new external API, dependency, package, SDK, or runtime library.

Capability Check Status: PASS

Checked at: 2026-06-06T12:35:42Z

Evidence:

- `.github/workflows/tag-release.yml` already grants `contents: write`, which is required for pushing Git tags and release branches.
- `.github/workflows/tag-release.yml` already grants `id-token: write`, which is required by keyless cosign signing.
- `.github/workflows/tag-release.yml` already uses `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` for Docker Hub publishing; this plan preserves that requirement.
- The planned work uses existing repository scripts and workflow actions rather than introducing a dependency version movement.

## Implementation Gate (Hard Stop)

Plan-document creation and revision are allowed as plan-only work before implementation approval. No implementation-target changes are allowed until the required unlock checklist is complete, all items in `External API Parameters and Capability Checks` are completed and marked checked, and explicit user approval is recorded.

Blocked implementation-target actions before gate pass:

- `apply_patch` against `.github/workflows/tag-release.yml`, `hack/tag-release.sh`, `hack/get-version.sh`, or any new implementation-target test/script file.
- Any file write/edit operation outside this ExecPlan document.
- Any generated implementation changes.

Required unlock checklist:

- [x] Capability checks are complete and recorded as `Capability Check Status: PASS` with `Checked at` and `Evidence`. (2026-06-06T12:35:42Z)
- [x] `Validation and Acceptance` includes acceptance criteria derived from user requirements, primary deliverables, and documented repository constraints. (2026-06-06T12:35:42Z)
- [x] Each non-obvious acceptance criterion maps back to its source in `Requirement to Acceptance Mapping`. (2026-06-06T12:35:42Z)
- [x] Explicit user approval is recorded as `User approval quote: "それでは、実行を承認します。実行してください。"` (2026-06-06T13:25:54Z)

Approval policy:

- No implementation-target code changes are allowed without an exact approval quote.
- Planning requests, scope clarifications, and review comments are not implementation approval.
- When explicit implementation approval is provided, record the user's exact approval quote verbatim, preserving the original language and punctuation.

## Progress

- [x] (2026-06-06T12:35:42Z) Inspected `docs/plans/README.md`, `.github/workflows/tag-release.yml`, `hack/tag-release.sh`, and `hack/get-version.sh`.
- [x] (2026-06-06T12:35:42Z) Confirmed that the current workflow runs `hack/tag-release.sh` before Docker Buildx setup, Docker Hub login, image build/push, and cosign signing.
- [x] (2026-06-06T12:35:42Z) Replaced the initial non-template plan with an `exec-plan` template-compliant plan and renamed it to `docs/plans/active/2026-06-06-tag-after-image-push.md`.
- [x] (2026-06-06T12:51:11Z) Updated the plan to split release publishing into `build-and-push` and `sign-and-tag` jobs so signing failures can be retried with failed-job rerun without rebuilding the image.
- [x] (2026-06-06T12:57:19Z) Recorded that `sign-and-tag` must use only the digest output from `build-and-push`; Docker Hub tag-to-digest re-resolution is not an accepted fallback.
- [x] (2026-06-06T13:02:16Z) Recorded that release finalization must not exit immediately on existing Git tag when `.0-beta.1` release branch finalization may still be required.
- [x] (2026-06-06T13:17:44Z) Recorded that finalization idempotency checks must inspect remote `origin` tag and branch refs instead of relying only on refs present in the GitHub Actions checkout.
- [x] (2026-06-06T13:21:44Z) Recorded that existing remote release tag and branch refs must match the current release commit, using peeled tag commits for annotated tags.
- [x] (2026-06-06T13:25:54Z) Recorded explicit implementation approval and moved the plan to `status: in_progress`.
- [x] (2026-06-06T13:27:35Z) Updated `hack/tag-release.sh` with validation-only mode, remote-ref idempotency checks, peeled annotated-tag commit checks, fail-closed mismatch handling, and retry-safe `.0-beta.1` branch finalization.
- [x] (2026-06-06T13:27:35Z) Updated `.github/workflows/tag-release.yml` into `build-and-push` and `sign-and-tag` jobs, with digest output passing, empty-digest guard, Docker Hub login in both publishing jobs, cosign signing before finalization, and post-signing tag finalization.
- [x] (2026-06-06T13:29:00Z) Preserved container `GITVERSION` behavior by passing the release tag string as the Docker build arg before the Git tag is created.
- [x] (2026-06-06T13:30:15Z) Completed focused validation for shell syntax, version validation, release finalization idempotency, remote-ref visibility, mismatch fail-closed behavior, workflow static structure, and whitespace.
- [x] (2026-06-06T13:32:50Z) Ran `actionlint .github/workflows/tag-release.yml`; it passed.
- [x] (2026-06-06T18:35:45Z) Human review completed; closing the ExecPlan.

## Surprises & Discoveries

- Observation: The initial local skill search missed `/home/vscode/.agents/skills/exec-plan/SKILL.md` because only `/home/vscode/.codex/skills` was checked.
  Evidence: `/home/vscode/.agents/skills/exec-plan/SKILL.md` exists and requires the template at `/home/vscode/.agents/skills/exec-plan/assets/PLANS.md`.

- Observation: After splitting the workflow into separate jobs, the signing job also needs Docker Hub login.
  Evidence: The previous single job performed `docker/login-action@v3` before both `docker/build-push-action@v6` and `cosign sign`; the split job loses that login state unless `sign-and-tag` logs in again.

- Observation: Computing `git describe --always` before creating the Git release tag would change the Docker build arg from the release tag string to another description.
  Evidence: The previous workflow created the local Git tag before `image-meta`, so `git describe --always` could resolve to `v${VERSION}`. The updated workflow now emits `gitversion=${tag}` directly to preserve the release image metadata while still delaying remote Git tag creation.

## Decision Log

- Decision: Keep this as a single ExecPlan and do not create a parent Strategy.
  Rationale: The requested outcome is one bounded release-CI workstream: move Git tag creation after image publication while preserving existing version, image, signing, and release-branch behavior.
  Timestamp: 2026-06-06T12:35:42Z

- Decision: Split the workflow into `build-and-push` and `sign-and-tag` jobs instead of only reordering steps inside one job.
  Rationale: GitHub Actions can rerun failed jobs or a specific job, not only a failed step inside an otherwise single job. Separating signing and tag finalization lets a signing failure retry against the already pushed image digest without rebuilding or pushing the image again.
  Timestamp: 2026-06-06T12:51:11Z

- Decision: Do not add a Docker Hub tag-to-digest re-resolution fallback in `sign-and-tag`.
  Rationale: The release image tag is expected to be immutable, and the signing step must prove it signs the exact digest emitted by the current workflow run's `build-and-push` job. Re-resolving a digest by tag from Docker Hub would weaken that provenance link and is unnecessary for the intended retry path.
  Timestamp: 2026-06-06T12:57:19Z

- Decision: Treat Git tag finalization and `.0-beta.1` release branch finalization as separate idempotent responsibilities.
  Rationale: The current script exits immediately when `v${VERSION}` exists. If tag push succeeds but release branch creation or push fails, a rerun would skip the still-required branch finalization. Separating these checks makes the finalization job retry-safe.
  Timestamp: 2026-06-06T13:02:16Z

- Decision: Use remote `origin` refs for release finalization idempotency checks.
  Rationale: A fresh GitHub Actions checkout may not contain all remote tags or release branches locally. Checking only `git tag -l` or local branches can miss an existing remote tag or branch during rerun and cause a duplicate push attempt before remaining finalization work runs.
  Timestamp: 2026-06-06T13:17:44Z

- Decision: Existing remote release refs must be verified against the current release commit.
  Rationale: Treating remote ref existence alone as success can hide a stale or wrong tag or branch. For annotated tags, the tag object is not the commit, so finalization must compare the peeled tag commit to the release commit.
  Timestamp: 2026-06-06T13:21:44Z

- Decision: Use the release tag string for the Docker `GITVERSION` build arg.
  Rationale: The original release workflow created the local Git tag before computing image metadata, so release images used the release tag as the version string. Passing `v$(hack/get-version.sh)` directly preserves that observable metadata while still delaying remote Git tag creation until after signing.
  Timestamp: 2026-06-06T13:29:00Z

## Outcomes & Retrospective

This plan has been converted to the `exec-plan` template. Implementation has not started, and no implementation-target files have been changed.
Implementation has completed, validation has passed, and human review has completed. The release CI now delays Git release tag finalization until after image build, push, and signing, while preserving retry-safe finalization behavior.

## Context and Orientation

The release workflow is `.github/workflows/tag-release.yml`. It runs on pushes to `main` and `release-*` branches when `version.go` changes.

The helper script `hack/get-version.sh` reads `RELEASE_VERSION` from `version.go` and prints the raw version string. The helper script `hack/tag-release.sh` reads the same version, validates it, creates annotated Git tag `v${VERSION}`, pushes that tag to `origin`, and creates and pushes `release-${MAJOR.MINOR}` when the version ends in `.0-beta.1`.

The current workflow creates the Git tag before image publication. The correctness condition for this workstream is the ordering of one release finalization sequence: a successful image build, image push, and image signing must happen before the Git tag is pushed. The retry condition for this workstream is that signing failure should be isolated to a failed job that can be rerun without re-running image build and push. The finalization idempotency condition is that an already pushed Git tag must not block the still-required release branch finalization for `.0-beta.1` versions. Because GitHub Actions checkouts may not contain every remote ref locally, idempotency checks must query `origin`. Because existing refs could be stale or wrong, idempotency checks must also verify that those refs point to the current release commit.

Non-goals:

- Do not change the release trigger branches or path filters.
- Do not change Docker image name `norseto/oci-lb-registrar`.
- Do not change Docker platforms `linux/arm64,linux/amd64`.
- Do not change cosign signing from digest-based signing.
- Do not change accepted release version formats.
- Do not add Docker Hub tag-to-digest re-resolution as a signing fallback.
- Do not treat an existing Git tag as proof that all release finalization responsibilities are complete for `.0-beta.1` versions.
- Do not rely only on local Git tags or branches to decide whether remote release finalization already happened.
- Do not overwrite, move, or force-push an existing mismatched release tag or release branch.

## Observable Behavior and Compatibility

Intentionally changed behavior: a failed Docker build, Docker push, or cosign signing step must not leave a newly pushed Git release tag. A cosign signing failure may leave the pushed Docker Hub image tag in place so the signing job can be rerun against the same digest.

Behavior that must remain unchanged: successful releases still publish `norseto/oci-lb-registrar:v${RELEASE_VERSION}`, sign the pushed image digest, create annotated Git tag `v${RELEASE_VERSION}`, and create `release-${MAJOR.MINOR}` for versions ending in `.0-beta.1`.

Intentionally strengthened behavior: for versions ending in `.0-beta.1`, rerunning release finalization after the Git tag already exists should still verify or complete `release-${MAJOR.MINOR}` branch creation and push.

Compatibility risk is limited to release automation. No controller runtime behavior, Kubernetes API behavior, Go module behavior, or release image version metadata should change.

## Lifecycle and Consistency Semantics

This work changes CI lifecycle ordering and job boundaries. The release finalization sequence has multiple phases: derive metadata, build and push image, pass the image tag and digest to a dependent job, sign image digest, then push Git tag and any release branch.

Invariants:

- The Git release tag must not be pushed before image build, image push, and cosign signing have completed successfully.
- Image tag derivation must happen before image build and must not require a pre-existing Git release tag.
- The image digest produced by the build/push job must be passed to the signing job as a job output and used for cosign signing.
- The signing job must not resolve the digest from Docker Hub by image tag; it must sign only the digest emitted by `build-and-push`.
- The `GITVERSION` build argument must remain the release tag string even though remote Git tag creation moves later.
- If image publication fails before finalization, no new Git release tag should exist.
- If cosign signing fails after image publication, no new Git release tag should exist, and rerunning failed jobs should retry the signing and tag finalization job without rerunning the successful image build/push job.
- If failed-job rerun cannot access the original `build-and-push` digest output, the recovery path is to delete the immutable release image tag from Docker Hub and rerun the full workflow, not to sign a digest re-resolved from Docker Hub by tag.
- If finalization succeeds and the workflow is retried, existing Git tag behavior must remain idempotent.
- For `.0-beta.1` versions, release branch finalization must be idempotent independently from Git tag finalization.
- Existing Git tag and release branch checks must inspect `origin` refs, such as `refs/tags/v${VERSION}` and `refs/heads/release-${MINOR}`, instead of relying only on local refs in the checkout.
- Existing remote release refs must point to the current release commit. For annotated tags, compare the peeled tag commit from `refs/tags/v${VERSION}^{}` with the current release commit.

If an intermediate image publication step fails, later Git tag and release branch creation must be skipped by GitHub Actions job dependency ordering. If signing fails after image publication, the image remains on Docker Hub by design and the failed `sign-and-tag` job can be rerun against the digest output from the original successful `build-and-push` job. If that digest output is unavailable during rerun, the workflow must fail rather than re-resolve by tag; manual recovery is to delete the Docker Hub release image tag and rerun the whole workflow. If finalization fails after tag creation but before `.0-beta.1` release branch push, rerun must still complete or verify the release branch instead of exiting only because the tag exists. If an existing remote tag or branch points to a different commit than the release commit, finalization must fail and require human investigation instead of overwriting the ref.

## Plan of Work

First, inspect the release workflow and scripts again immediately before implementation to guard against concurrent edits. Confirm all consumers of the release version, image tag, Git SHA, image digest, Git tag, and release branch.

Next, separate validation and metadata generation from Git tag creation. The implementation should either add a validation-only helper or add a mode to the existing helper so the workflow can validate `RELEASE_VERSION` before image publication without creating a Git tag. Update finalization logic so tag creation and release branch creation are independent idempotent responsibilities. Use remote `origin` ref checks for both responsibilities, and compare existing remote refs to the current release commit before treating them as complete.

Then, update `.github/workflows/tag-release.yml` so it has a `build-and-push` job that computes `tag=v$(hack/get-version.sh)` and `gitversion=${tag}` before Docker build, builds and pushes `norseto/oci-lb-registrar:${tag}`, and exposes the image tag, Git version string, and pushed image digest as job outputs. Add a dependent `sign-and-tag` job that checks out the same ref, configures Git, installs cosign, signs `norseto/oci-lb-registrar@${digest}` using only `needs.build-and-push.outputs.digest`, fails if that digest output is empty, and only then runs the release finalization step that creates and pushes the Git tag and any release branch.

Finally, add or update tests for the shell scripts if a lightweight script test can be added without changing the release interface. The tests should cover valid versions, invalid versions, existing tag idempotency, and `.0-beta.1` release branch creation.

## Concrete Steps

Work from `/workspaces/oci-lb-controller`.

1. Re-read the relevant files:

       cat .github/workflows/tag-release.yml
       cat hack/tag-release.sh
       cat hack/get-version.sh
       rg -n "tag-release|get-version|cosign|norseto/oci-lb-registrar|release-" .github hack Makefile

2. Edit `hack/tag-release.sh` or add a narrowly named helper under `hack/` so validation can run before image publication without creating or pushing a Git tag. Preserve the current validation regex. Preserve Git tag idempotency, but do not exit before `.0-beta.1` release branch finalization has been verified or completed. Use remote checks such as `git ls-remote --tags origin "refs/tags/v${VERSION}"`, `git ls-remote --tags origin "refs/tags/v${VERSION}^{}"`, and `git ls-remote --heads origin "refs/heads/${RELEASE_BRANCH}"` when deciding whether remote tag or branch finalization already exists. Compare existing remote refs with `git rev-parse HEAD` or the equivalent current release commit before treating them as complete.

3. Edit `.github/workflows/tag-release.yml` so the workflow has two jobs:

       build-and-push:
         checkout
         release metadata or validation
         docker/setup-buildx-action
         docker/login-action
         docker/build-push-action
         expose tag, gitversion, and digest as job outputs

       sign-and-tag:
         needs: build-and-push
         checkout
         git config
         sigstore/cosign-installer
         fail if needs.build-and-push.outputs.digest is empty
         cosign sign using needs.build-and-push.outputs.digest
         post-publication Git tag and release branch finalization

4. If script behavior is split, update any script references found by `rg` so existing callers continue to work or intentionally call the new helper.

5. Run focused validation commands selected during implementation. At minimum, run syntax checks for changed shell scripts and inspect the workflow diff:

       bash -n hack/tag-release.sh
       git diff -- .github/workflows/tag-release.yml hack

6. If tests are added for shell behavior, run those exact tests and record the command and outcome in `Implementation Completion Report`.

7. If the implementation is prepared for commit readiness, run repository-required checks:

       make vet
       make test
       make lint
       make vulcheck
       make seccheck

## Validation and Acceptance

### Requirement to Acceptance Mapping

- Source: user requirement that CI should tag only after build and push complete.
  Acceptance: `.github/workflows/tag-release.yml` must place Git tag creation in a job that depends on a successful Docker image build/push job and runs after cosign signing, and no earlier job or step may call a script path that creates or pushes `v${VERSION}`.
  Proof: workflow diff inspection plus script call inspection with `rg`.

- Source: user requirement to avoid manual tag deletion when build or push fails.
  Acceptance: if Docker build, Docker push, or cosign signing fails, GitHub Actions job and step ordering skips the tag creation step.
  Proof: workflow step ordering review and, if feasible, a workflow-level static check that confirms finalization is after the signing step.

- Source: user requirement to make signing failure retryable without rebuilding when possible.
  Acceptance: if cosign signing fails after image push, rerunning failed jobs should rerun the `sign-and-tag` job while the successful `build-and-push` job remains successful and does not need to rebuild or repush the image.
  Proof: workflow structure review confirms signing and Git tag finalization are isolated in a dependent job and consume `needs.build-and-push.outputs.digest`.

- Source: user requirement to preserve the identity of the image built by CI.
  Acceptance: `sign-and-tag` must sign only `needs.build-and-push.outputs.digest`, must fail when that output is unavailable or empty, and must not include any Docker Hub tag-to-digest re-resolution fallback.
  Proof: workflow diff inspection confirms there is no tag-based digest lookup in `sign-and-tag` and that an explicit empty-digest guard exists before `cosign sign`.

- Source: primary deliverables preserving release behavior.
  Acceptance: successful release still publishes image tag `norseto/oci-lb-registrar:v$(hack/get-version.sh)`, signs by digest, and then pushes annotated Git tag `v${VERSION}`.
  Proof: workflow diff inspection and script test or temporary Git repository exercise.

- Source: existing `hack/tag-release.sh` behavior.
  Acceptance: valid versions `X.Y.Z`, `X.Y.Z-alpha.N`, and `X.Y.Z-beta.N` remain accepted; invalid versions remain rejected.
  Proof: script tests or direct shell exercises against temporary `version.go` contents.

- Source: existing release branch behavior.
  Acceptance: versions ending in `.0-beta.1` still create and push `release-${MAJOR.MINOR}` after image publication; other versions do not create that branch.
  Proof: script tests or direct shell exercises in a temporary Git repository.

- Source: release finalization retry safety.
  Acceptance: if `v${VERSION}` already exists for a `.0-beta.1` version but `release-${MAJOR.MINOR}` is missing, rerunning finalization creates and pushes the missing branch instead of exiting after tag detection.
  Proof: script test or temporary Git repository exercise where the tag exists before finalization and the release branch is absent.

- Source: CI checkout consistency risk.
  Acceptance: finalization uses remote `origin` ref checks to detect existing `v${VERSION}` tags and `release-${MAJOR.MINOR}` branches, and does not rely only on `git tag -l` or local branch state.
  Proof: script inspection plus temporary Git repository exercise where the remote contains a tag or branch that is not present in the local checkout before finalization starts.

- Source: release ref integrity.
  Acceptance: if remote `v${VERSION}` already exists, finalization verifies the peeled tag commit matches the current release commit; if remote `release-${MAJOR.MINOR}` already exists, finalization verifies the branch head matches the current release commit; mismatches fail without overwrite or force push.
  Proof: script tests or temporary Git repository exercises covering matching remote refs and mismatched remote refs.

- Source: release consistency invariant.
  Acceptance: `GITVERSION` remains the release tag string for the image build even though remote Git tag creation happens after signing.
  Proof: workflow diff inspection confirms `gitversion=${tag}` is emitted before build and passed as the Docker `GITVERSION` build arg.

## Testing Strategy

Use focused shell-script tests for version validation, tag idempotency, and release branch behavior because the risky behavior lives in shell helpers and Git operations. A temporary Git repository test is sufficient to prove tag and branch side effects without contacting the real remote if remote push commands are isolated or replaced by a safe local remote.

The shell-script tests should include a partial-finalization retry case: pre-create `v${VERSION}` for a `.0-beta.1` version, leave `release-${MAJOR.MINOR}` absent, run finalization, and expect the branch to be created or pushed while the existing tag is left unchanged.

The shell-script tests should include a remote-ref visibility case: create a local checkout that does not have the remote tag or release branch locally, ensure the bare `origin` has the relevant ref, run finalization, and expect the script to detect the remote ref rather than attempting a duplicate push.

The shell-script tests should include ref-integrity cases: an existing remote annotated tag whose peeled commit matches `HEAD` should be accepted, an existing remote annotated tag whose peeled commit differs from `HEAD` should fail, an existing remote release branch whose head matches `HEAD` should be accepted, and an existing remote release branch whose head differs from `HEAD` should fail.

Use workflow diff review to prove ordering and job-output data flow because GitHub Actions itself is declarative and the requested behavior depends on job dependency sequence. If an existing workflow linter is available in the repository, run it; otherwise record that no workflow linter exists and rely on static inspection plus script tests.

Full Go unit tests are not expected to exercise release CI ordering because this change should not alter Go code or controller runtime behavior. Run repository-required `make` checks only when the implementation is intended to be commit-ready, following repository policy.

## Naming and Semantic Changes

No parameter or variable is planned to change from singular to collection form. Any new helper script or step name must describe its single responsibility, such as validation before publication or release finalization after publication.

## Idempotence and Recovery

The post-publication finalization step must remain safe to retry when `v${VERSION}` already exists on `origin`. Existing Git tag idempotency must be preserved, but an existing remote tag must not short-circuit `.0-beta.1` release branch finalization.

Retry after failed image build or failed image push should not require deleting a Git tag because the `sign-and-tag` job will not have run. Retry after failed cosign signing should not require rebuilding or repushing the image when `needs.build-and-push.outputs.digest` remains available because `sign-and-tag` is a separate failed job that consumes the successful `build-and-push` job outputs. Docker Hub may retain the pushed immutable release image tag when signing fails; that is intentional for retryability in this plan.

If `sign-and-tag` cannot access the original digest output during failed-job rerun, it must fail closed. The recovery path is to delete the Docker Hub release image tag and rerun the whole workflow so a new `build-and-push` job produces a fresh digest output. The workflow must not silently sign a digest re-resolved from Docker Hub by tag.

If Git tag push succeeds but `.0-beta.1` release branch finalization fails, rerunning `sign-and-tag` must verify that the tag exists and continue to create or push the release branch. The recovery path must not require deleting the Git tag solely to retry branch creation.

If a remote tag or release branch already exists but is absent from the local checkout, finalization must detect the remote ref and continue idempotently. The recovery path must not depend on fetching all tags and branches into local refs before each check.

If a remote tag or release branch exists and points to a different commit than the current release commit, finalization must fail closed. The recovery path is human investigation and manual correction, not automatic overwrite.

No destructive local commands are planned. Any command that would delete remote tags, delete remote branches, or rewrite Git history is outside this plan.

## Artifacts and Notes

Relevant current workflow excerpt:

    - run: hack/tag-release.sh
    - uses: docker/setup-buildx-action@v3
    ...
    - id: build
      uses: docker/build-push-action@v6
    ...
    - run: cosign sign --yes norseto/oci-lb-registrar@${{ steps.build.outputs.digest }}

This excerpt demonstrates the current incorrect ordering: tag creation happens before image publication.

Target workflow shape:

    build-and-push -> sign-and-tag

The `sign-and-tag` job must consume the image digest from `build-and-push` and must create the Git tag only after `cosign sign` succeeds.

The `sign-and-tag` job must not perform a Docker Hub lookup that converts `norseto/oci-lb-registrar:${tag}` back to a digest.

## Interfaces and Dependencies

The implementation should use existing shell scripts under `hack/` and existing GitHub Actions in `.github/workflows/tag-release.yml`.

Required script behavior:

- A validation or metadata step must read `RELEASE_VERSION` from `version.go` and reject unsupported formats before image publication.
- The `build-and-push` job must expose the image tag and digest through job outputs.
- The `sign-and-tag` job must consume the digest through the `needs` context, fail if that digest is empty, and sign the pushed image digest.
- The `sign-and-tag` job must not use Docker Hub tag-to-digest re-resolution.
- A finalization step must create annotated tag `v${VERSION}` and push it to `origin` only after image publication and signing succeed.
- A finalization step must create and push `release-${MAJOR.MINOR}` only for versions ending in `.0-beta.1`, and it must still verify or complete that branch when `v${VERSION}` already exists.
- Release finalization idempotency checks must use remote `origin` refs for `refs/tags/v${VERSION}` and `refs/heads/release-${MINOR}`.
- Existing remote release refs must be compared to the current release commit before finalization treats them as already complete. Annotated tags must use the peeled commit from `refs/tags/v${VERSION}^{}`.

No new Go package, controller interface, Kubernetes API type, Docker image name, or GitHub Action version is planned.

## Implementation Completion Report

Implementation changed `.github/workflows/tag-release.yml` and `hack/tag-release.sh`.

Validation run:

- `bash -n hack/tag-release.sh`: passed.
- `hack/tag-release.sh --validate-only`: passed for the repository's current `version.go` value `0.7.0-alpha.5`.
- Temporary Git repository integration script: passed. It covered normal annotated tag creation, existing remote tag with missing `.0-beta.1` release branch, remote tag mismatch fail-closed behavior, and remote release branch mismatch fail-closed behavior.
- Temporary invalid-version validation script: passed. `RELEASE_VERSION = "1.2"` was rejected by `hack/tag-release.sh --validate-only`.
- `rg -n "build-and-push:|sign-and-tag:|needs: build-and-push|digest:|gitversion=|GITVERSION=|cosign sign|hack/tag-release.sh --validate-only|hack/tag-release.sh|docker/login-action" .github/workflows/tag-release.yml`: passed static structure inspection for the split workflow.
- `actionlint .github/workflows/tag-release.yml`: passed.
- `git diff --check`: passed.

Validation omitted:

- `make vet`, `make test`, `make lint`, `make vulcheck`, and `make seccheck` were not run because this implementation changed release workflow and shell release automation only, with no Go code changes and no commit requested.

Copyright/header check:

- `.github/workflows/tag-release.yml` and nearby workflow files do not use copyright headers.
- `hack/tag-release.sh` and nearby shell scripts do not use copyright headers.

## Change Notes

Replaced the initial plan with an `exec-plan` template-compliant plan, added required front matter, recorded the implementation gate, documented CI lifecycle invariants, and renamed the plan file to the required `YYYY-MM-DD-<name>.md` format.

Updated the plan to split the workflow into `build-and-push` and `sign-and-tag` jobs. This records the decision that Docker Hub image retention after signing failure is acceptable because it enables rerunning only the failed signing and tag finalization job.

Updated the plan to require digest-output-only signing, reject Docker Hub tag-to-digest re-resolution fallback, and document full workflow rerun after manual image-tag deletion as the fail-closed recovery path when the original digest output is unavailable.

Updated the plan to require retry-safe release branch finalization for `.0-beta.1` versions even when the Git release tag already exists.

Updated the plan to require remote `origin` ref checks for existing release tags and release branches so fresh CI checkouts do not miss already pushed refs during reruns.

Updated the plan to require current-release-commit verification for existing remote release refs, including peeled commit comparison for annotated tags, and to fail closed on mismatches.
