import crypto from "node:crypto";
import fs from "node:fs/promises";
import path from "node:path";

const MAX_INPUT_BYTES = 256 * 1024;
const DEFAULT_SETTLE_MS = 250;

function now() {
  return new Date().toISOString();
}

function diagnostic(code, message, severity = "error") {
  const apiKey = process.env.MYCELIS_API_KEY?.trim();
  const safeMessage = apiKey ? String(message).replaceAll(apiKey, "[redacted]") : String(message);
  return { code, message: safeMessage.slice(0, 2_000), severity };
}

async function readRequest() {
  const chunks = [];
  let size = 0;
  for await (const chunk of process.stdin) {
    size += chunk.length;
    if (size > MAX_INPUT_BYTES) throw new Error("validation request exceeds 256 KiB");
    chunks.push(chunk);
  }
  return JSON.parse(Buffer.concat(chunks).toString("utf8"));
}

function validateRequest(request) {
  const parsed = new URL(request.launch_url);
  if (!["http:", "https:"].includes(parsed.protocol)) throw new Error("launch_url must use HTTP(S)");
  if (!request.content_digest || typeof request.content_digest !== "string") throw new Error("content_digest is required");
  if (!path.isAbsolute(request.evidence_path || "")) throw new Error("evidence_path must be absolute");
  if (!request.plan || !Array.isArray(request.plan.checks)) throw new Error("plan.checks must be an array");
	request.acceptance_criteria ??= [];
	request.criterion_mappings ??= [];
	if (!Array.isArray(request.acceptance_criteria) || !Array.isArray(request.criterion_mappings) ||
	    request.acceptance_criteria.length !== request.criterion_mappings.length) {
	  throw new Error("every acceptance criterion requires one explicit validation mapping");
	}
	if (request.acceptance_criteria.length > 32) throw new Error("at most 32 acceptance criteria are supported");
	request.acceptance_criteria.forEach((criterion, index) => {
	  const mapping = request.criterion_mappings[index];
	  if (typeof criterion !== "string" || !criterion.trim() || criterion.length > 1_000 || mapping?.criterion !== criterion) {
	    throw new Error("criterion mappings must preserve exact criterion order and text");
	  }
	  if (!["check", "probe", "journey", "unsupported"].includes(mapping.source)) throw new Error("unsupported criterion mapping source");
	});
}

function hash(buffer) {
  return crypto.createHash("sha256").update(buffer).digest("hex");
}

function evidenceRef(kind, filePath, contentType, buffer) {
  return { kind, path: filePath, content_type: contentType, ...(buffer ? { sha256: hash(buffer) } : {}) };
}

function sameOriginAsset(launchURL, resourceURL) {
  try {
    const launch = new URL(launchURL);
    const resource = new URL(resourceURL);
    return resource.origin === launch.origin;
  } catch {
    return false;
  }
}

async function observeBefore(page, probe, evidencePath, evidenceRefs) {
  const observation = probe.observe;
  const target = observation.kind === "url_change" ? null : page.locator(observation.target).first();
  switch (observation.kind) {
    case "visual_change": {
      const buffer = await target.screenshot();
      const filePath = path.join(evidencePath, "probe-before.png");
      await fs.writeFile(filePath, buffer);
      evidenceRefs.push(evidenceRef("probe_before", filePath, "image/png", buffer));
      return { target, value: hash(buffer) };
    }
    case "text_change":
      return { target, value: (await target.textContent()) ?? "" };
    case "value_change":
      return { target, value: await target.inputValue() };
    case "element_visible":
      return { target, value: await target.isVisible() };
    case "url_change":
      return { target, value: page.url() };
    default:
      throw new Error(`unsupported observation: ${observation.kind}`);
  }
}

async function performAction(page, probe, observeDuringHold) {
  const action = probe.action;
  const actionTarget = action.target ? page.locator(action.target).first() : null;
  switch (action.kind) {
    case "click":
      await actionTarget.click();
      return;
    case "fill":
      await actionTarget.fill(action.value ?? "");
      return;
    case "key_press":
      if (actionTarget) await actionTarget.focus();
      await page.keyboard.press(action.key);
      return;
    case "key_hold":
      if (actionTarget) await actionTarget.focus();
      await page.keyboard.down(action.key);
      try {
        await page.waitForTimeout(Math.min(Math.max(action.duration_ms || 500, 1), 10_000));
        await observeDuringHold?.();
      } finally {
        await page.keyboard.up(action.key);
      }
      return;
    case "pointer": {
      const box = await actionTarget.boundingBox();
      if (!box) throw new Error(`pointer target is not visible: ${action.target}`);
      await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      await page.mouse.down();
      await page.mouse.up();
      return;
    }
    default:
      throw new Error(`unsupported action: ${action.kind}`);
  }
}

async function observeAfter(page, probe, before, evidencePath, evidenceRefs) {
  let after;
  switch (probe.observe.kind) {
    case "visual_change": {
      const buffer = await before.target.screenshot();
      const filePath = path.join(evidencePath, "probe-after.png");
      await fs.writeFile(filePath, buffer);
      evidenceRefs.push(evidenceRef("probe_after", filePath, "image/png", buffer));
      after = hash(buffer);
      break;
    }
    case "text_change":
      after = (await before.target.textContent()) ?? "";
      break;
    case "value_change":
      after = await before.target.inputValue();
      break;
    case "element_visible":
      after = await before.target.isVisible();
      break;
    case "url_change":
      after = page.url();
      break;
  }
  const passed = probe.observe.kind === "element_visible" ? after === true : after !== before.value;
  return {
    action: probe.action.kind,
    observation: probe.observe.kind,
    passed,
    before: String(before.value).slice(0, 1_000),
    after: String(after).slice(0, 1_000),
  };
}

async function retainReport(report, evidencePath) {
  await fs.mkdir(evidencePath, { recursive: true });
  const reportPath = path.join(evidencePath, "validation-report.json");
  report.evidence_refs.push(evidenceRef("validation_report", reportPath, "application/json"));
  await fs.writeFile(reportPath, `${JSON.stringify(report, null, 2)}\n`, { encoding: "utf8" });
  return report;
}

async function validate(request, chromium) {
  const report = {
    status: "error",
    content_digest: request.content_digest,
    launch_url: request.launch_url,
    started_at: now(),
    finished_at: now(),
    diagnostics: [],
    evidence_refs: [],
    checks: [],
	criterion_evidence: [],
  };
	const unsupported = request.criterion_mappings.filter((mapping) => mapping.source === "unsupported");
	if (unsupported.length > 0) {
	  report.status = "failed";
	  report.finished_at = now();
	  report.diagnostics.push(diagnostic("unsupported_semantic_criteria", `No deterministic browser observation is defined for: ${unsupported.map((item) => item.criterion).join(" | ")}`));
	  return retainReport(report, request.evidence_path);
	}
  const pageErrors = [];
  const failedAssets = new Set();
  let browser;
  try {
    browser = await chromium.launch({ headless: true });
    const apiKey = process.env.MYCELIS_API_KEY?.trim();
    const page = await browser.newPage({
      ...(apiKey ? {
        extraHTTPHeaders: {
          Authorization: apiKey.startsWith("Bearer ") ? apiKey : `Bearer ${apiKey}`,
        },
      } : {}),
    });
    page.on("pageerror", (error) => pageErrors.push(String(error.message || error)));
    page.on("requestfailed", (requestFailure) => {
      if (sameOriginAsset(request.launch_url, requestFailure.url())) failedAssets.add(requestFailure.url());
    });
    page.on("response", (response) => {
      if (response.status() >= 400 && sameOriginAsset(request.launch_url, response.url())) failedAssets.add(response.url());
    });

    let loadResponse;
    let loadError;
    try {
      loadResponse = await page.goto(request.launch_url, { waitUntil: "domcontentloaded", timeout: 15_000 });
    } catch (error) {
      loadError = error;
    }
    await page.waitForTimeout(DEFAULT_SETTLE_MS);

    if (request.plan.probe && !loadError) {
      try {
        const before = await observeBefore(page, request.plan.probe, request.evidence_path, report.evidence_refs);
        let observedDuringAction = false;
        await performAction(page, request.plan.probe, async () => {
          report.probe = await observeAfter(page, request.plan.probe, before, request.evidence_path, report.evidence_refs);
          observedDuringAction = true;
        });
        if (!observedDuringAction) {
          await page.waitForTimeout(DEFAULT_SETTLE_MS);
          report.probe = await observeAfter(page, request.plan.probe, before, request.evidence_path, report.evidence_refs);
        }
        if (!report.probe.passed) report.diagnostics.push(diagnostic("probe_observation_unchanged", "The requested interaction did not produce its expected observable effect."));
      } catch (error) {
        report.probe = {
          action: request.plan.probe.action.kind,
          observation: request.plan.probe.observe.kind,
          passed: false,
        };
        report.diagnostics.push(diagnostic("probe_execution_failed", error.message || error));
      }
    }

    for (const check of request.plan.checks) {
      if (check === "load") {
        const passed = !loadError && (!loadResponse || loadResponse.status() < 400);
        report.checks.push({ check, passed, detail: passed ? "Page loaded." : String(loadError?.message || `HTTP ${loadResponse?.status()}`) });
      } else if (check === "no_page_errors") {
        report.checks.push({ check, passed: pageErrors.length === 0, detail: pageErrors.join(" | ").slice(0, 2_000) });
      } else if (check === "no_failed_local_assets") {
        report.checks.push({ check, passed: failedAssets.size === 0, detail: [...failedAssets].join(" | ").slice(0, 2_000) });
      }
    }
    for (const item of report.checks.filter((item) => !item.passed)) {
      report.diagnostics.push(diagnostic(`check_${item.check}_failed`, item.detail || `${item.check} failed`));
    }
    report.status = report.checks.every((item) => item.passed) && (!report.probe || report.probe.passed) ? "passed" : "failed";
	const reportRef = path.join(request.evidence_path, "validation-report.json");
	for (const mapping of request.criterion_mappings) {
	  let passed;
	  if (mapping.source === "journey") {
	    await page.reload({ waitUntil: "domcontentloaded" });
	    passed = true;
	    for (const probe of mapping.journey || []) {
	      try {
	        const before = await observeBefore(page, probe, request.evidence_path, report.evidence_refs);
	        let result;
	        await performAction(page, probe, async () => { result = await observeAfter(page, probe, before, request.evidence_path, report.evidence_refs); });
	        if (!result) { await page.waitForTimeout(DEFAULT_SETTLE_MS); result = await observeAfter(page, probe, before, request.evidence_path, report.evidence_refs); }
	        passed = passed && result.passed;
	      } catch (error) {
	        passed = false;
	        report.diagnostics.push(diagnostic("criterion_journey_failed", `${mapping.criterion}: ${error.message || error}`));
	        break;
	      }
	    }
	  } else {
	    passed = mapping.source === "probe"
	      ? report.probe?.passed === true
	      : report.checks.some((item) => item.check === mapping.check && item.passed);
	  }
	  report.criterion_evidence.push({ criterion: mapping.criterion, passed, ...(passed ? { evidence_refs: [reportRef] } : {}) });
	}
	if (report.criterion_evidence.some((item) => !item.passed)) report.status = "failed";
  } catch (error) {
    report.status = "unavailable";
    report.diagnostics.push(diagnostic("browser_runtime_unavailable", error.message || error));
  } finally {
    await browser?.close().catch(() => {});
    report.finished_at = now();
  }
  return retainReport(report, request.evidence_path);
}

let request;
try {
  request = await readRequest();
  validateRequest(request);
  await fs.mkdir(request.evidence_path, { recursive: true });
  const { chromium } = await import("@playwright/test");
  const report = await validate(request, chromium);
  process.stdout.write(JSON.stringify(report));
} catch (error) {
  const report = {
    status: "unavailable",
    content_digest: request?.content_digest ?? "",
    launch_url: request?.launch_url ?? "",
    started_at: now(),
    finished_at: now(),
    diagnostics: [diagnostic("validator_unavailable", error.message || error)],
    evidence_refs: [],
    checks: [],
  };
  if (request?.evidence_path && path.isAbsolute(request.evidence_path)) await retainReport(report, request.evidence_path).catch(() => {});
  process.stdout.write(JSON.stringify(report));
  process.exitCode = 2;
}
