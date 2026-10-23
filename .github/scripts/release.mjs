import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { appendFile, readFile, readdir, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';

const versionTag = /^v\d+\.\d+\.\d+$/;
const workflows = [
  'test.yml',
  'golangci.yml',
  'gogenerate.yml',
  'govulncheck.yml',
  'release-checks.yml',
];

// Only missing or still-running work is retryable. Failed checks and inconsistent
// release identity require an operator to resolve the cause before rerunning.
class Pending extends Error {}

async function waitFor(operation, timeout) {
  const deadline = Date.now() + timeout;
  let previous;
  for (;;) {
    try {
      return await operation();
    } catch (error) {
      if (!(error instanceof Pending) || Date.now() >= deadline) throw error;
      if (error.message !== previous) console.log(error.message);
      previous = error.message;
      await new Promise(resolve => setTimeout(resolve, 15_000));
    }
  }
}

async function github(path, { method = 'GET', body, accept = 'application/vnd.github+json' } = {}) {
  const repository = process.env.GITHUB_REPOSITORY;
  assert.match(repository, /^[\w.-]+\/[\w.-]+$/);
  assert.ok(process.env.GITHUB_TOKEN, 'GITHUB_TOKEN is required');
  const url = new URL(`/repos/${repository}/${path}`, process.env.GITHUB_API_URL || 'https://api.github.com');
  const response = await fetch(url, {
    method,
    headers: {
      Authorization: `Bearer ${process.env.GITHUB_TOKEN}`,
      Accept: accept,
      'X-GitHub-Api-Version': '2022-11-28',
      'Content-Type': 'application/json',
    },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal: AbortSignal.timeout(30_000),
  });
  assert.ok(response.ok, `${method} ${url.pathname} returned HTTP ${response.status}`);
  return response.json();
}

async function pages(path, key) {
  const results = [];
  for (let page = 1; ; page++) {
    const data = await github(`${path}${path.includes('?') ? '&' : '?'}per_page=100&page=${page}`);
    const items = key ? data[key] : data;
    assert.ok(Array.isArray(items), `Invalid collection: ${path}`);
    results.push(...items);
    if (items.length < 100) return results;
  }
}

async function tagCommit(tag) {
  let { object } = await github(`git/ref/tags/${tag}`);
  for (let depth = 0; object.type === 'tag' && depth < 10; depth++) {
    ({ object } = await github(`git/tags/${object.sha}`));
  }
  assert.equal(object.type, 'commit', `${tag} must resolve to a commit`);
  return object.sha;
}

async function approvals(sha) {
  const repository = process.env.GITHUB_REPOSITORY;
  const candidates = (await pages(`commits/${sha}/pulls`)).filter(pr =>
    pr.base.ref === 'main' && pr.base.repo.full_name === repository &&
    pr.head.repo?.full_name === repository && pr.head.ref === 'release-please--branches--main');
  const merged = [];
  for (const candidate of candidates) {
    const pr = await github(`pulls/${candidate.number}`);
    if (pr.merged && pr.merge_commit_sha === sha) merged.push(pr);
  }
  return merged;
}

function notes(body) {
  const lines = body.split('\n');
  const delimiters = lines.flatMap((line, index) => line.trim() === '---' ? [index] : []);
  assert.ok(delimiters.length >= 2, 'Release PR body must contain the Release Please notes delimiters');
  return lines.slice(delimiters[0] + 1, delimiters.at(-1)).join('\n').trim();
}

async function candidate() {
  const tag = process.env.GITHUB_REF?.replace(/^refs\/tags\//, '');
  assert.match(process.env.GITHUB_REF || '', /^refs\/tags\/v\d+\.\d+\.\d+$/, 'Run this workflow on the release tag');
  const sha = execFileSync('git', ['rev-parse', 'HEAD'], { encoding: 'utf8' }).trim();
  assert.equal(sha, process.env.GITHUB_SHA, 'Checkout must match the workflow commit');
  assert.equal(await tagCommit(tag), sha, 'Remote tag must match the checked-out commit');
  const comparison = await github(`compare/${sha}...main`);
  assert.ok(['ahead', 'identical'].includes(comparison.status), 'Release commit must belong to main');
  return { tag, sha };
}

async function releaseFor(tag) {
  const releases = await pages('releases');
  const matches = releases.filter(release => release.tag_name === tag);
  if (matches.length === 0) throw new Pending(`Waiting for the existing draft for ${tag}; no release will be created here`);
  assert.equal(matches.length, 1, `More than one release uses ${tag}`);
  const release = matches[0];
  assert.equal(release.name, tag, 'Release name must equal its tag');
  assert.equal(release.prerelease, false, 'Prereleases are not supported');
  // GoReleaser's existing-draft lookup uses the name, so reject ambiguous names too.
  assert.equal(releases.filter(other => other.name === tag).length, 1, 'Release name must be unique');
  if (process.env.RELEASE_ID) assert.equal(String(release.id), process.env.RELEASE_ID, 'Release ID changed');
  return { release, releases };
}

async function validateDraft({ tag, sha }, release, releases) {
  const manifest = JSON.parse(await readFile('.release-please-manifest.json', 'utf8'));
  assert.equal(`v${manifest['.']}`, tag, 'Tag must match the approved release manifest');
  const prs = await approvals(sha);
  assert.equal(prs.length, 1, `${sha} must be the final commit of one merged Release Please PR`);
  const pr = prs[0];
  assert.ok(release.body?.trim(), 'Release notes must not be empty');
  assert.equal(release.body.trim(), notes(pr.body), 'Draft notes must match the merged release PR');
  for (const other of releases.filter(item => item.draft && item.id !== release.id && versionTag.test(item.tag_name))) {
    const otherSha = await tagCommit(other.tag_name);
    // Only approved provider releases participate in the one-in-flight policy.
    if ((await approvals(otherSha)).length === 0) continue;
    const comparison = await github(`compare/${otherSha}...${sha}`);
    assert.equal(comparison.status, 'behind', `Resolve earlier approved draft ${other.tag_name} before ${tag}`);
  }
  return pr;
}

async function ready(sha, mergedAt, timeout = 0) {
  const checks = workflows.map(workflow => async () => {
    const runs = await pages(`actions/workflows/${workflow}/runs?head_sha=${sha}&event=push&branch=main`, 'workflow_runs');
    const run = runs.filter(item => item.head_sha === sha && item.head_branch === 'main' &&
      item.event === 'push' && item.path === `.github/workflows/${workflow}`)
      .sort((a, b) => b.id - a.id)[0];
    if (!run || run.status !== 'completed') throw new Pending(`Waiting for ${workflow} on ${sha}`);
    assert.equal(run.conclusion, 'success', `${workflow} failed on ${sha}: ${run.conclusion}`);
    const jobs = await pages(`actions/runs/${run.id}/attempts/${run.run_attempt}/jobs`, 'jobs');
    assert.ok(jobs.length > 0, `${workflow} has no job evidence`);
    for (const job of jobs) assert.equal(job.conclusion, 'success', `${workflow}: ${job.name} did not succeed`);
    if (workflow === 'test.yml') {
      const live = jobs.filter(job => job.name.startsWith('testacc ('));
      assert.ok(live.length > 0, 'Live acceptance job is missing');
      for (const job of live) {
        const step = job.steps.find(item => item.name === 'Live acceptance tests');
        assert.equal(step?.conclusion, 'success', `${job.name} must execute live acceptance tests`);
      }
    }
  });

  checks.push(async () => {
    const statuses = await pages(`commits/${sha}/statuses`);
    const coverage = statuses.find(status => status.context === 'codecov/project');
    if (!coverage || coverage.state === 'pending') throw new Pending(`Waiting for Codecov on ${sha}`);
    assert.equal(coverage.creator?.login, 'codecov[bot]', 'Coverage status must come from Codecov');
    assert.equal(coverage.state, 'success', `Coverage failed on ${sha}`);
  });

  // Alerts queried at a SHA are not supported by GitHub's ref filter. Select
  // actual analyses by commit, then inspect their SARIF rather than main's latest alerts.
  for (const category of ['/language:go', '/language:actions']) {
    checks.push(async () => {
      let analysis;
      for (let page = 1; ; page++) {
        const items = await github(`code-scanning/analyses?ref=refs/heads/main&tool_name=CodeQL&sort=created&direction=desc&per_page=100&page=${page}`);
        analysis = items.find(item => item.commit_sha === sha && item.ref === 'refs/heads/main' &&
          item.tool.name === 'CodeQL' && item.category === category);
        // Main-branch analysis of this commit cannot predate the PR's merge. Stop
        // there instead of rescanning the entire repository history on every poll.
        if (analysis || items.length < 100 || items.some(item => item.created_at < mergedAt)) break;
      }
      if (!analysis) throw new Pending(`Waiting for CodeQL ${category} on ${sha}`);
      assert.equal(analysis.error, '', `CodeQL ${category} failed: ${analysis.error}`);
      const sarif = await github(`code-scanning/analyses/${analysis.id}`, { accept: 'application/sarif+json' });
      assert.ok(Array.isArray(sarif.runs) && sarif.runs.length > 0, 'CodeQL returned no SARIF runs');
      for (const run of sarif.runs) {
        for (const result of run.results || []) {
          const rule = run.tool.driver.rules?.find(item => item.id === result.ruleId);
          assert.ok(rule, `CodeQL result has no rule: ${result.ruleId}`);
          const level = result.level || rule.defaultConfiguration?.level;
          const severity = Number(rule.properties?.['security-severity'] || 0);
          assert.ok(level !== 'error' && severity < 7, `Blocking CodeQL finding: ${result.ruleId}`);
        }
      }
    });
  }
  // Poll each incomplete requirement independently to avoid repeatedly spending
  // the Actions token's rate limit on checks that already succeeded. Publication
  // calls this again with a zero timeout to revalidate all of them immediately.
  const deadline = Date.now() + timeout;
  for (const check of checks) await waitFor(check, Math.max(0, deadline - Date.now()));
}

async function verifyAssets(release, tag) {
  const prefix = `terraform-provider-contentful_${tag.slice(1)}_`;
  const checksum = `${prefix}SHA256SUMS`;
  const files = (await readdir('dist')).filter(name => name.endsWith('.zip') || name.endsWith('_manifest.json'));
  assert.ok(files.some(name => name.endsWith('.zip')), 'No provider archives were built');
  assert.ok(files.includes(`${prefix}manifest.json`), 'Registry manifest is missing');
  assert.ok(files.every(name => name.startsWith(prefix)), 'Unexpected artifact version');
  const expected = new Map();
  for (const line of (await readFile(join('dist', checksum), 'utf8')).trim().split('\n')) {
    const match = /^([a-f0-9]{64}) [ *]([^/\\\s]+)$/.exec(line);
    assert.ok(match, 'Invalid signed-checksum entry');
    assert.ok(!expected.has(match[2]), `Duplicate checksum for ${match[2]}`);
    expected.set(match[2], match[1]);
  }
  assert.deepEqual([...expected.keys()].sort(), [...files].sort(), 'Checksums must cover exactly the archives and manifest');
  const assets = await pages(`releases/${release.id}/assets`);
  files.push(checksum, `${checksum}.sig`);
  assert.equal(assets.length, files.length, 'Remote release has missing or extra assets');
  for (const file of files) {
    const bytes = await readFile(join('dist', file));
    const digest = createHash('sha256').update(bytes).digest('hex');
    if (expected.has(file)) assert.equal(digest, expected.get(file), `Checksum mismatch: ${file}`);
    const matches = assets.filter(asset => asset.name === file);
    assert.equal(matches.length, 1, `Missing or duplicate remote asset: ${file}`);
    assert.equal(matches[0].state, 'uploaded', `Upload incomplete: ${file}`);
    assert.equal(matches[0].size, bytes.length, `Remote size differs: ${file}`);
    assert.equal(matches[0].digest, `sha256:${digest}`, `Remote digest differs: ${file}`);
  }
}

async function summary(message) {
  console.log(message);
  if (process.env.GITHUB_STEP_SUMMARY) await appendFile(process.env.GITHUB_STEP_SUMMARY, `${message}\n\n`);
}

async function published(release) {
  assert.equal(release.draft, false, 'Release is still a draft');
  assert.equal(release.immutable, true, 'Published release is not immutable; do not rebuild or replace its assets');
  await summary(`GitHub release ${release.tag_name} is published and immutable: ${release.html_url}`);
}

export async function main(command) {
  assert.ok(['prepare', 'verify', 'publish', 'registry'].includes(command), 'Expected prepare, verify, publish, or registry');
  if (command === 'registry') {
    const version = process.env.GITHUB_REF_NAME.slice(1);
    try {
      await waitFor(async () => {
        const response = await fetch('https://registry.terraform.io/v1/providers/cysp/contentful/versions', { signal: AbortSignal.timeout(30_000) });
        assert.ok(response.ok, `Registry returned HTTP ${response.status}`);
        const data = await response.json();
        if (!data.versions.some(item => item.version === version)) throw new Pending(`Registry has not indexed ${version}`);
      }, 300_000);
      await summary(`Terraform Registry: ${version} is available. Inspect its rendered documentation before announcing it.`);
    } catch (error) {
      await summary(`Terraform Registry availability is unconfirmed: ${error.message}. GitHub publication is a separate outcome; do not rebuild the release.`);
      throw error;
    }
    return;
  }

  const source = await candidate();
  const { release, releases } = command === 'prepare'
    ? await waitFor(() => releaseFor(source.tag), 300_000)
    : await releaseFor(source.tag);
  if (!release.draft) {
    await published(release);
    if (command === 'prepare') await appendFile(process.env.GITHUB_OUTPUT, `id=${release.id}\npublished=true\n`);
    return;
  }
  const pr = await validateDraft(source, release, releases);
  if (command === 'prepare') {
    await ready(source.sha, pr.merged_at, 2_700_000);
    await writeFile(join(process.env.RUNNER_TEMP, 'release-notes.txt'), release.body);
    await appendFile(process.env.GITHUB_OUTPUT, `id=${release.id}\npublished=false\n`);
    await summary(`Approved draft ${source.tag}: commit ${source.sha}; required CI succeeded.`);
    return;
  }

  assert.equal(release.body, await readFile(join(process.env.RUNNER_TEMP, 'release-notes.txt'), 'utf8'), 'Release notes changed after preparation');
  await verifyAssets(release, source.tag);
  if (command === 'verify') return;
  await ready(source.sha, pr.merged_at);
  try {
    await github(`releases/${release.id}`, { method: 'PATCH', body: { draft: false, make_latest: 'true' } });
  } catch (error) {
    // A lost response may follow a successful publish. Reconcile before deciding
    // to fail, never repeat publication or start another build automatically.
    if ((await github(`releases/${release.id}`)).draft) throw error;
  }
  const final = await github(`releases/${release.id}`);
  assert.equal(final.tag_name, source.tag, 'Published tag changed');
  await published(final);
  await verifyAssets(final, source.tag);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main(process.argv[2]).catch(error => {
    console.error(error.message);
    process.exitCode = 1;
  });
}
