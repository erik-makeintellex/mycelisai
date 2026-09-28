#!/usr/bin/env python3
"""Live user-journey probe against the running Mycelis stack. It prints PASS/FAIL per journey.

Checks real outcomes (file content on disk, recall, citations), not status labels (rule 12).
Run from the repo root against a running stack: `uv run python ops/live_journey_probe.py`.
Reads MYCELIS_API_KEY from .env and never prints it. Idempotent (no duplicate memory rows).
This is the delivery metric: a slice that changes a user journey must keep it passing.
"""
import json, re, sys, time, urllib.request, urllib.error
from pathlib import Path

BASE = "http://127.0.0.1:8081"
KEY = next(l.split("=", 1)[1].strip() for l in Path(".env").read_text().splitlines() if l.startswith("MYCELIS_API_KEY="))
results = []

def call(method, path, body=None, timeout=240):
    req = urllib.request.Request(BASE + path, method=method, data=json.dumps(body).encode() if body is not None else None,
                                 headers={"Authorization": f"Bearer {KEY}", "Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            raw = r.read(); ct = r.headers.get("content-type", "")
            return r.status, ct, raw
    except urllib.error.HTTPError as e:
        return e.code, e.headers.get("content-type", ""), e.read()

def js(raw):
    try: return json.loads(raw)
    except Exception: return {}

def record(name, ok, detail):
    results.append((name, "PASS" if ok else "FAIL", detail[:300]))

def chat(text):
    s, _, raw = call("POST", "/api/v1/chat", {"messages": [{"role": "user", "content": text}]})
    d = js(raw); return s, d, ((d.get("data") or {}).get("payload") or {})

# J1: plain question
s, d, p = chat("In one sentence, what is Mycelis for?")
record("J1 ask question", s == 200 and len((p.get("text") or "").strip()) > 20, f"http={s} text={(p.get('text') or '')[:120]!r}")

# J2: deliverable with real content, preview, confirm, readback
fname = f"probe-brief-{int(time.time())}.md"
s, d, p = chat(f"Create a markdown file named {fname} listing three concrete ways a small bakery can use AI assistants.")
prop = p.get("proposal") or {}
previews = prop.get("draft_previews") or p.get("draft_previews") or []
tok = prop.get("confirm_token")
prev_text = " ".join((x.get("preview") or "") for x in previews)
template = "What you asked for" in prev_text or "Soma created this file from your request" in prev_text
record("J2a proposal has real draft preview", s == 200 and bool(tok) and bool(previews) and not template,
       f"http={s} mode={(d.get('data') or {}).get('mode')} previews={len(previews)} template={template} code={d.get('code') or (d.get('data') or {}).get('code')}")
if tok:
    s2, _, raw2 = call("POST", "/api/v1/intent/confirm-action", {"confirm_token": tok})
    d2 = js(raw2).get("data") or {}
    record("J2b confirm executes", s2 == 200 and d2.get("confirmed") is True, f"http={s2} state={d2.get('execution_state')}")
    time.sleep(3)
    s3, ct3, raw3 = call("GET", f"/api/v1/workspace/files/view?path={fname}")
    body = raw3.decode("utf-8", "replace")
    real = s3 == 200 and "baker" in body.lower() and "What you asked for" not in body and len(body) > 120
    record("J2c file content is real (readback)", real, f"http={s3} bytes={len(body)} head={body[:100]!r}")
    record("J2d text file served as text/plain", s3 == 200 and ct3.startswith("text/plain"), f"content-type={ct3}")
    record("J2e proof honest (verified only if real)", (d2.get("execution_state") == "verified") == real,
           f"state={d2.get('execution_state')} real={real}")

# J3: maintained memory save (no embedding engine) and recall with citation (idempotent: save only if absent)
s0, _, raw0 = call("GET", "/api/v1/memory/deployment-context")
existing = "Rye Bakery" in raw0.decode("utf-8", "replace")
if existing:
    record("J3a memory save without embeddings", s0 == 200, f"http={s0} (profile already saved; list verified, no duplicate written)")
s, _, raw = (200, "", b"skipped") if existing else call("POST", "/api/v1/memory/deployment-context", {
    "title": "Company profile: Juniper & Rye Bakery", "source_label": "owner-provided company profile",
    "content": "Juniper & Rye Bakery is on East 6th Street in Austin. Weekend special: blueberry-lavender scones, 3 for $10. Always sign off with \"See you at the counter.\"",
    "knowledge_class": "company_knowledge", "source_kind": "user_note", "content_domain": "operations",
    "visibility": "global", "sensitivity_class": "role_scoped", "trust_class": "user_provided"})
if not existing:
    record("J3a memory save without embeddings", s in (200, 201), f"http={s} body={raw[:160]!r}")
s, d, p = chat("Write a two-line promo for our weekend special.")
text = (p.get("text") or "") + json.dumps(p.get("provenance") or {}) + json.dumps(p.get("execution_summary") or {})
record("J3b recall uses stored facts", "blueberry-lavender" in text.lower() or "blueberry lavender" in text.lower(), f"text={(p.get('text') or '')[:160]!r}")
record("J3b2 brand voice followed (sign-off)", "counter" in (p.get("text") or "").lower(), "soft check: stored sign-off rule")
cs = p.get("context_sources") or (d.get("data") or {}).get("context_sources") or []
record("J3c answer cites the source", any(("Juniper" in (c.get("title") or "")) and c.get("used") for c in cs), f"context_sources={json.dumps(cs)[:220]}")

# J3d: archive/restore/delete of saved memory (M2). Throwaway entries carry a timestamp; leftovers from an
# interrupted run are deleted first, and both entries are gone at the end (idempotent, no rows accumulate).
MEM = "/api/v1/memory/deployment-context"
def mem_entries(archived=True):
    s, _, raw = call("GET", f"{MEM}?limit=100" + ("&include_archived=true" if archived else ""))
    return s, js(raw).get("entries") or []
for e in mem_entries()[1]:
    if (e.get("title") or "").startswith("Probe J3d "): call("DELETE", f"{MEM}/{e['artifact_id']}")
ts = int(time.time())
def throwaway(label):
    s, _, raw = call("POST", MEM, {"title": f"Probe J3d {label} {ts}", "knowledge_class": "customer_context", "visibility": "global",
        "source_kind": "user_note", "content": f"Weekend special promo note {ts}: two-line promo idea, the weekend special cardamom knot."})
    return js(raw).get("artifact_id") if s in (200, 201) else None
def promo_cites(aid):
    _, d, p = chat("Write a two-line promo for our weekend special.")
    cs = p.get("context_sources") or (d.get("data") or {}).get("context_sources") or []
    return any(c.get("artifact_id") == aid for c in cs)
arch, gone = throwaway("archive"), throwaway("delete")
if arch and gone:
    cited_before = promo_cites(arch)
    s1, _, _ = call("POST", f"{MEM}/{arch}/archive")
    cited_archived = promo_cites(arch)
    record("J3d1 archived entry is not cited", s1 == 200 and cited_before and not cited_archived,
           f"archive http={s1} cited before={cited_before} while archived={cited_archived}")
    s2, _, _ = call("POST", f"{MEM}/{arch}/restore")
    cited_restored = promo_cites(arch)
    record("J3d2 restored entry is cited again", s2 == 200 and cited_restored, f"restore http={s2} cited={cited_restored}")
    s3, _, raw3 = call("DELETE", f"{MEM}/{gone}")
    listed = [e.get("artifact_id") for e in mem_entries()[1]]
    record("J3d3 deleted entry is gone from the list", s3 == 200 and gone not in listed,
           f"delete http={s3} chunks_removed={(js(raw3).get('data') or {}).get('chunks_removed')} still listed={gone in listed}")
else:
    record("J3d memory archive/restore/delete", False, f"could not save throwaway entries archive={arch} delete={gone}")
for aid in (arch, gone):
    if aid: call("DELETE", f"{MEM}/{aid}")


# J4: name-only surfaces (placeholder audit PH-D)
s, _, raw = call("GET", "/api/v1/sensors")
blob = raw.decode("utf-8", "replace").lower()
record("J4 sensors not hard-coded online", not (s == 200 and ("gmail" in blob or "weather" in blob) and "online" in blob), f"http={s} head={blob[:140]!r}")

# J5: web search honesty (blocked or real results, never fake success)
s, d, p = chat("Search the web for the current population of Austin, Texas and cite a source.")
t = p.get("text") or ""
record("J5 web search is real or honestly blocked", s in (200, 422, 503) and (bool(re.search(r"https?://", t)) or "search" in json.dumps(d).lower()), f"http={s} text={t[:140]!r}")

# J8: token budget stop is honest (B1 contract "J6"). Soma's turns run as agent `admin`. A 1,024-token
# override on that agent is spent by one turn; the next turn must stop with token_budget_exhausted
# and claim no completion. The override is always deleted, and chat must work again afterwards.
# This proves the day/run caps: the first turn's prompt alone exceeds 1,024 tokens, so the second
# turn is refused at preflight by per_agent_day (and per_run/per_team_day). The per-execution stop needs a multi-call turn (D4 lets one call's prompt
# overshoot), so it is proven by the cognitive unit tests, not here.
OVR = "/api/v1/cognitive/budgets/overrides/agent/admin"
lim = {"per_execution": 1024, "per_run": 1024, "per_team_day": 1024, "per_agent_day": 1024}
s0, _, raw0 = call("PUT", OVR, lim)
if s0 != 200:
    record("J8 budget stop is honest", False, f"override PUT http={s0} body={raw0[:160]!r}")
else:
    try:
        chat("Say hello in one short sentence.")
        s1, d1, p1 = chat("Say hello again in one short sentence.")
        blob = json.dumps(d1).lower()
        stopped = s1 == 429 or "token_budget_exhausted" in blob
        claims_done = "result verified" in blob or (p1.get("execution_state") or "") in ("completed", "verified")
        su, _, rawu = call("GET", "/api/v1/cognitive/budgets/usage?agent_id=admin")
        u = (js(rawu).get("data") or {})
        record("J8a over-budget turn stops honestly (day cap)", stopped and not claims_done, f"http={s1} stopped={stopped} claims_done={claims_done} body={blob[:120]!r}")
        record("J8b usage shows the cap reached", su == 200 and u.get("limit") == 1024 and (u.get("used") or 0) >= 1024, f"http={su} used={u.get('used')} limit={u.get('limit')} period={u.get('period')}")
    finally:
        s2, _, _ = call("DELETE", OVR)
    s3, _, p3 = chat("Say hello in one short sentence.")
    record("J8c override removed, chat works again", s2 == 200 and s3 == 200 and len((p3.get("text") or "").strip()) > 0, f"delete http={s2} chat http={s3}")

# J9: MCP install refuses undeclared env keys (MCPA-H). A NODE_OPTIONS env on the curated `fetch`
# entry would reach the npx child inside Core; it must be refused with 400 mcp_env_rejected (even
# for this admin key), the value never echoed, and the installed `fetch` server left untouched.
def fetch_row():
    s, _, raw = call("GET", "/api/v1/mcp/servers")
    rows = js(raw) if isinstance(js(raw), list) else (js(raw).get("servers") or js(raw).get("data") or [])
    return next((json.dumps(r, sort_keys=True) for r in rows if isinstance(r, dict) and r.get("name") == "fetch"), None)
before = fetch_row()
s9, _, raw9 = call("POST", "/api/v1/mcp/library/install", {"name": "fetch", "env": {"NODE_OPTIONS": "--require /tmp/probe-j9-never.js"}})
b9 = raw9.decode("utf-8", "replace")
record("J9 MCP install refuses undeclared env", s9 == 400 and "mcp_env_rejected" in b9 and "probe-j9-never" not in b9 and fetch_row() == before,
       f"http={s9} body={b9[:140]!r} fetch_unchanged={fetch_row() == before}")

# J10: Core confines filesystem MCP paths (MCPL/MCPL2). The curated `filesystem` server is recognized by its
# package (not its name): an in-workspace listing works, while an escape and an upper-case `Path` key are
# refused before any tool call. Compose disables default MCP auto-install, so J10 installs the curated
# server through the admin library route when it is absent and deletes it again afterwards (audited).
def fs_server_id():
    s, _, raw = call("GET", "/api/v1/mcp/servers")
    d = js(raw); rows = d if isinstance(d, list) else (d.get("servers") or d.get("data") or [])
    return next((r.get("id") for r in rows if isinstance(r, dict) and r.get("name") == "filesystem" and r.get("status") == "connected"), None)
fsid, installed_by_probe, si = fs_server_id(), False, None
if not fsid:
    si, _, rawi = call("POST", "/api/v1/mcp/library/install", {"name": "filesystem"}, timeout=300)
    fsid, installed_by_probe = fs_server_id(), si == 200
if not fsid:
    record("J10 filesystem MCP paths confined in Core", False, f"no connected `filesystem` server; probe install http={si}")
else:
    base = f"/api/v1/mcp/servers/{fsid}/tools"
    sa, _, _ = call("POST", f"{base}/list_directory/call", {"arguments": {"path": "."}})
    sb, _, rb = call("POST", f"{base}/read_text_file/call", {"arguments": {"path": "../../etc/passwd"}})
    sc, _, rc = call("POST", f"{base}/read_text_file/call", {"arguments": {"Path": "/etc/passwd"}})
    ok_b = sb == 403 and b"mcp_path_outside_workspace" in rb and b"root:" not in rb
    ok_c = sc in (400, 403) and b"root:" not in rc
    record("J10 filesystem MCP paths confined in Core", sa == 200 and ok_b and ok_c,
           f"list .={sa} escape={sb} ({rb[:60]!r}) upper-case Path={sc} installed_by_probe={installed_by_probe}")
    if installed_by_probe:
        call("DELETE", f"/api/v1/mcp/servers/{fsid}")

w = max(len(r[0]) for r in results)
for n, st, det in results: print(f"{st:4}  {n:<{w}}  {det}")
print(f"\n{sum(r[1]=='PASS' for r in results)}/{len(results)} passed")
