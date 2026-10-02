#!/usr/bin/env node

const repository = process.env.GITHUB_REPOSITORY;
const pullRequest = process.env.MCPFATHER_PR_NUMBER;
const token = process.env.GH_TOKEN || process.env.GITHUB_TOKEN;
const phase = process.env.MCPFATHER_COMMENT_PHASE;
const runUrl = process.env.MCPFATHER_RUN_URL;
const buildResult = process.env.MCPFATHER_BUILD_RESULT;
const integrationResult = process.env.MCPFATHER_INTEGRATION_RESULT;

if (!repository || !pullRequest || !token || !phase || !runUrl) {
  throw new Error("Missing MCPFather PR CI comment context.");
}
if (!["started", "build", "final"].includes(phase)) {
  throw new Error(`Unsupported MCPFather PR CI comment phase: ${phase}`);
}

const [owner, repo] = repository.split("/", 2);
if (!owner || !repo) {
  throw new Error(`Invalid GITHUB_REPOSITORY: ${repository}`);
}

const marker = "<!-- mcpfather-ci-status -->";
const apiRoot = process.env.GITHUB_API_URL || "https://api.github.com";
const commentsUrl = `${apiRoot}/repos/${owner}/${repo}/issues/${pullRequest}/comments`;

const resultIcon = (result) => {
  if (result === "success") return "✅";
  if (result === "skipped") return "⏭️";
  if (["failure", "cancelled", "timed_out", "action_required", "startup_failure"].includes(result)) {
    return "❌";
  }
  return "⏳";
};

const displayResult = (result) => result || "pending";

const renderComment = () => {
  const lines = [marker, "## MCPFather PR CI", "", `- Run: [Actions log](${runUrl})`];

  if (phase === "started") {
    lines.push("- Status: ⏳ Build, unit tests, and integration tests started.");
    return lines.join("\n");
  }

  lines.push(`- Build and unit tests: ${resultIcon(buildResult)} ${displayResult(buildResult)}`);

  if (phase === "build") {
    lines.push("- Integration tests: ⏳ pending");
    lines.push("", buildResult === "success"
      ? "- Status: ⏳ Build and unit tests passed; integration tests are still running."
      : "- Status: ❌ Build or unit tests failed; integration tests may still be running.");
    return lines.join("\n");
  }

  lines.push(`- Integration tests: ${resultIcon(integrationResult)} ${displayResult(integrationResult)}`);
  const passed = buildResult === "success" && integrationResult === "success";
  lines.push("", passed
    ? "**CI completed successfully.**"
    : `**CI failed.** Review the [Actions log](${runUrl}).`);
  return lines.join("\n");
};

const body = renderComment();
if (process.env.MCPFATHER_COMMENT_DRY_RUN === "true") {
  console.log(body);
  process.exit(0);
}

const headers = {
  Accept: "application/vnd.github+json",
  Authorization: `Bearer ${token}`,
  "X-GitHub-Api-Version": "2022-11-28",
  "Content-Type": "application/json",
};

const request = async (url, options = {}) => {
  const response = await fetch(url, { ...options, headers: { ...headers, ...options.headers } });
  if (!response.ok) {
    throw new Error(
      `GitHub API ${options.method || "GET"} ${url} failed: ${response.status} ${await response.text()}`,
    );
  }
  return {
    data: response.status === 204 ? undefined : await response.json(),
    link: response.headers.get("link"),
  };
};

const nextPage = (linkHeader) => {
  if (!linkHeader) return undefined;
  const next = linkHeader.split(",").find((link) => link.includes('rel="next"'));
  return next?.match(/<([^>]+)>/)?.[1];
};

let existing;
let pageUrl = `${commentsUrl}?per_page=100`;
while (pageUrl && !existing) {
  const page = await request(pageUrl);
  existing = page.data.find(
    (comment) => comment.user?.login === "github-actions[bot]" && comment.body?.includes(marker),
  );
  pageUrl = nextPage(page.link);
}

if (existing) {
  await request(`${apiRoot}/repos/${owner}/${repo}/issues/comments/${existing.id}`, {
    method: "PATCH",
    body: JSON.stringify({ body }),
  });
} else {
  await request(commentsUrl, { method: "POST", body: JSON.stringify({ body }) });
}
