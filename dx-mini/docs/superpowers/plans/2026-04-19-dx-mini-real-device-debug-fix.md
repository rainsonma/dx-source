# dx-mini: Fix `真机调试` "Cannot convert undefined or null to object" — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove three project-shape landmines in dx-mini that make WeChat DevTools 2.02.2604152 fail `真机调试` with "Cannot convert undefined or null to object" before any QR code is shown.

**Architecture:** Config-only change across three files + one new file. No runtime code (`.ts`, `.wxml`, `.wxss`) is touched. Every edit is independently revertible.

**Tech Stack:** WeChat Mini Program (native + TypeScript), DevTools 2.02.2604152, glass-easel framework, Vant Weapp 1.11.x.

**Spec:** `dx-mini/docs/superpowers/specs/2026-04-19-dx-mini-real-device-debug-fix-design.md`

**Working directory for all commands:** `/Users/rainsen/Programs/Projects/douxue/dx-source`

---

## File Structure

| File | Change | Responsibility |
|---|---|---|
| `dx-mini/miniprogram/app.json` | Modify (remove 1 field + fix trailing comma) | App shell config — page list, tabBar, window, componentFramework |
| `dx-mini/project.config.json` | Modify (flip 1 bool) | DevTools-visible project settings |
| `dx-mini/project.private.config.json` | Modify (flip 1 bool) | Developer-local project settings overlay |
| `dx-mini/miniprogram/sitemap.json` | Create | WeChat search indexing rules (allow-all) |

---

## Task 1: Remove `lazyCodeLoading` from `app.json`

**Files:**
- Modify: `dx-mini/miniprogram/app.json:39-41`

**Why this task:** `lazyCodeLoading: "requiredComponents"` combined with `componentFramework: "glass-easel"` is the top-ranked landmine in the spec (Section 2.A). Removing it is sufficient on its own in many reports — we do it first so if the fix lands here, the rest is pure hygiene.

- [ ] **Step 1: Verify current state**

Run: `sed -n '38,41p' dx-mini/miniprogram/app.json`
Expected output (exactly):
```
  "style": "v2",
  "componentFramework": "glass-easel",
  "lazyCodeLoading": "requiredComponents"
}
```

If the output differs, STOP — another change has been made and this plan needs an update.

- [ ] **Step 2: Apply the edit**

Use the Edit tool with these parameters:

```
file_path: /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/app.json
old_string:
  "componentFramework": "glass-easel",
  "lazyCodeLoading": "requiredComponents"
}
new_string:
  "componentFramework": "glass-easel"
}
```

Note: both the comma after `"glass-easel"` AND the entire `"lazyCodeLoading"` line go together. Leaving the comma makes the JSON invalid.

- [ ] **Step 3: Verify JSON is still valid**

Run: `python3 -c "import json; json.load(open('dx-mini/miniprogram/app.json')); print('OK')"`
Expected output: `OK`

Run: `sed -n '38,40p' dx-mini/miniprogram/app.json`
Expected output (exactly):
```
  "style": "v2",
  "componentFramework": "glass-easel"
}
```

- [ ] **Step 4: Verify lazyCodeLoading is fully gone**

Run: `grep -r lazyCodeLoading dx-mini/miniprogram/ dx-mini/*.json || echo NONE`
Expected output: `NONE`

No commit yet — we batch the four file changes into one commit at the end of Task 4.

---

## Task 2: Flip `skylineRenderEnable` to `false` in `project.config.json`

**Files:**
- Modify: `dx-mini/project.config.json:27`

**Why this task:** `skylineRenderEnable: true` with zero `"renderer": "skyline"` declarations in page.json files is the config-vs-reality mismatch from Section 2.B of the spec. DevTools' Skyline packager asks each page for a renderer config, and on pages that don't declare one, `Object.keys()` on the missing entry throws the exact error text we're seeing. No runtime behavior changes (Skyline is never actually selected at render time).

- [ ] **Step 1: Verify current state**

Run: `grep -n skylineRenderEnable dx-mini/project.config.json`
Expected output:
```
27:    "skylineRenderEnable": true,
```

If the value is already `false`, skip to Step 3.

- [ ] **Step 2: Apply the edit**

Use the Edit tool:

```
file_path: /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/project.config.json
old_string:     "skylineRenderEnable": true,
new_string:     "skylineRenderEnable": false,
```

- [ ] **Step 3: Verify JSON is still valid**

Run: `python3 -c "import json; json.load(open('dx-mini/project.config.json')); print('OK')"`
Expected output: `OK`

Run: `grep -n skylineRenderEnable dx-mini/project.config.json`
Expected output:
```
27:    "skylineRenderEnable": false,
```

---

## Task 3: Flip `skylineRenderEnable` to `false` in `project.private.config.json`

**Files:**
- Modify: `dx-mini/project.private.config.json:9`

**Why this task:** Private config overlays public config — if we flipped it only in `project.config.json`, the private one would still say `true` and override. Both must be in sync.

- [ ] **Step 1: Verify current state**

Run: `grep -n skylineRenderEnable dx-mini/project.private.config.json`
Expected output:
```
9:    "skylineRenderEnable": true,
```

- [ ] **Step 2: Apply the edit**

Use the Edit tool:

```
file_path: /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/project.private.config.json
old_string:     "skylineRenderEnable": true,
new_string:     "skylineRenderEnable": false,
```

- [ ] **Step 3: Verify JSON is still valid**

Run: `python3 -c "import json; json.load(open('dx-mini/project.private.config.json')); print('OK')"`
Expected output: `OK`

Run: `grep -n skylineRenderEnable dx-mini/project.private.config.json`
Expected output:
```
9:    "skylineRenderEnable": false,
```

- [ ] **Step 4: Sanity-check nothing else in the repo sets Skyline on**

Run: `grep -rn skylineRenderEnable dx-mini/`
Expected output: both files now show `false`. No other matches.

```
dx-mini/project.config.json:27:    "skylineRenderEnable": false,
dx-mini/project.private.config.json:9:    "skylineRenderEnable": false,
```

---

## Task 4: Create `miniprogram/sitemap.json`

**Files:**
- Create: `dx-mini/miniprogram/sitemap.json`

**Why this task:** Section 2.C of the spec — missing `sitemap.json` has become a hard-fail on recent DevTools builds. Allow-all is WeChat's documented default and matches the project's intent (public mini, all pages indexable).

- [ ] **Step 1: Verify the file does not already exist**

Run: `ls dx-mini/miniprogram/sitemap.json 2>/dev/null && echo EXISTS || echo MISSING`
Expected output: `MISSING`

If it prints `EXISTS`, STOP — the plan needs an update (someone else added it).

- [ ] **Step 2: Create the file**

Use the Write tool:

```
file_path: /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/sitemap.json
content:
{
  "desc": "关于本文件的更多信息，请参考文档 https://developers.weixin.qq.com/miniprogram/dev/reference/configuration/sitemap.html",
  "rules": [{
    "action": "allow",
    "page": "*"
  }]
}
```

- [ ] **Step 3: Verify it's valid JSON**

Run: `python3 -c "import json; d=json.load(open('dx-mini/miniprogram/sitemap.json')); assert d['rules'][0]['action']=='allow'; assert d['rules'][0]['page']=='*'; print('OK')"`
Expected output: `OK`

- [ ] **Step 4: Verify app.json doesn't point at a different sitemap location**

Run: `grep -n sitemapLocation dx-mini/miniprogram/app.json || echo NO_OVERRIDE`
Expected output: `NO_OVERRIDE`

(If `app.json` had a `sitemapLocation` override, the file we just created would be ignored. Since there isn't one, DevTools picks up `miniprogram/sitemap.json` automatically.)

- [ ] **Step 5: Review all four changes together**

Run: `git -C /Users/rainsen/Programs/Projects/douxue/dx-source status -s dx-mini/`
Expected output (exact order doesn't matter):
```
 M dx-mini/miniprogram/app.json
 M dx-mini/project.config.json
 M dx-mini/project.private.config.json
?? dx-mini/miniprogram/sitemap.json
```

Run: `git -C /Users/rainsen/Programs/Projects/douxue/dx-source diff dx-mini/miniprogram/app.json dx-mini/project.config.json dx-mini/project.private.config.json`
Review the diff — confirm it matches exactly:
- `app.json`: one-line removal of `"lazyCodeLoading": "requiredComponents"` and trailing comma of the previous line.
- `project.config.json:27`: `true` → `false`.
- `project.private.config.json:9`: `true` → `false`.

Nothing else should be in the diff. If there is, STOP and investigate.

- [ ] **Step 6: Stop and report before committing**

Do NOT create the commit yet. Report to the user:

> "Four-file change ready for commit. Please verify `真机调试` works in WeChat DevTools now (instructions in Task 5). Once you confirm, I'll commit."

---

## Task 5: User-side verification in WeChat DevTools

**Files:** none (manual verification by the user)

**Why this task:** The only reliable way to verify this fix is to click the `真机调试` button and see if the error is gone. We cannot automate DevTools from the CLI.

- [ ] **Step 1: Ask the user to reload the project in DevTools**

Report to the user:

> "In WeChat DevTools: click the `编译` button (or Cmd+B) to reload the project. Watch the DevTools console for any new errors — especially JSON parse errors on `app.json`, `project.config.json`, `project.private.config.json`, or `sitemap.json`. Expected: compile succeeds, simulator boots, login page shows the 使用微信登录 button."

Wait for user confirmation. If the user reports a new error, STOP and investigate.

- [ ] **Step 2: Ask the user to click `真机调试`**

Report to the user:

> "Click the `真机调试` button. Expected: a QR code appears within ~2 seconds, **no** error banner. If you want, scan the QR code with your phone's WeChat and confirm the login page renders on the phone. Report back what you see."

Wait for user confirmation.

- [ ] **Step 3: Handle the outcome**

**If the user reports success (QR code appears, no error):** Proceed to Task 6.

**If the user reports the same error:** The fix did not address the root cause. Tell the user:

> "The three-config-landmine hypothesis did not resolve the error. Next diagnostic steps:
> 1. Check DevTools `详情 → 基础库` matches the `libVersion` in `project.config.json` (currently `3.15.1`).
> 2. Try `微信开发者工具 → 设置 → 通用设置 → 清除全部缓存` and reload.
> 3. Check if the error persists in `预览` (preview) — if `预览` works but `真机调试` fails, it's likely a DevTools-to-phone bridge issue independent of project config.
>
> Which of these would you like me to investigate first?"

Do NOT commit the config change if the fix did not work — revert Tasks 1-4 first with:
`git -C /Users/rainsen/Programs/Projects/douxue/dx-source checkout -- dx-mini/miniprogram/app.json dx-mini/project.config.json dx-mini/project.private.config.json && rm dx-mini/miniprogram/sitemap.json`

**If the user reports a DIFFERENT error:** Do not commit. Report the new error, update the diagnosis, return to brainstorming-style debug.

---

## Task 6: Commit the fix (only after user confirms `真机调试` works)

**Files:** none directly (commits files from Tasks 1-4)

**Why this task:** Per global user rules, we commit only after the user has verified the fix resolves the bug. Batching into one commit keeps the history clean and makes rollback trivial.

- [ ] **Step 1: Ask for explicit commit authorization**

Report to the user:

> "Ready to commit. The message will be:
> ```
> fix(mini): remove DevTools 真机调试 config landmines
>
> Three pre-processing issues that caused WeChat DevTools 2.02.2604152
> to throw "Cannot convert undefined or null to object" when clicking
> 真机调试, before any QR code was shown:
>
> 1. app.json had lazyCodeLoading: "requiredComponents" together with
>    componentFramework: "glass-easel". Known interaction that trips
>    DevTools real-device builds; removed.
> 2. project.config.json + project.private.config.json had
>    skylineRenderEnable: true while no page.json declared
>    "renderer": "skyline". Skyline packager Object.keys'd the missing
>    config. Flipped to false to match reality.
> 3. miniprogram/sitemap.json was missing. Recent DevTools surface this
>    as a hard error on some paths. Added with the documented allow-all
>    default rule.
>
> No runtime code changed. Verification: 真机调试 now produces a QR code
> without the error banner.
> ```
>
> Commit this? (yes / edit message / don't commit)"

Wait for user response.

- [ ] **Step 2: Create the commit**

If the user said yes:

Run (from `/Users/rainsen/Programs/Projects/douxue/dx-source`):
```bash
git add dx-mini/miniprogram/app.json dx-mini/project.config.json dx-mini/project.private.config.json dx-mini/miniprogram/sitemap.json
git commit -m "$(cat <<'EOF'
fix(mini): remove DevTools 真机调试 config landmines

Three pre-processing issues that caused WeChat DevTools 2.02.2604152
to throw "Cannot convert undefined or null to object" when clicking
真机调试, before any QR code was shown:

1. app.json had lazyCodeLoading: "requiredComponents" together with
   componentFramework: "glass-easel". Known interaction that trips
   DevTools real-device builds; removed.
2. project.config.json + project.private.config.json had
   skylineRenderEnable: true while no page.json declared
   "renderer": "skyline". Skyline packager Object.keys'd the missing
   config. Flipped to false to match reality.
3. miniprogram/sitemap.json was missing. Recent DevTools surface this
   as a hard error on some paths. Added with the documented allow-all
   default rule.

No runtime code changed. Verification: 真机调试 now produces a QR code
without the error banner.
EOF
)"
```

- [ ] **Step 3: Confirm the commit landed**

Run: `git -C /Users/rainsen/Programs/Projects/douxue/dx-source log --oneline -1`
Expected: top commit message starts with `fix(mini): remove DevTools 真机调试 config landmines`.

Run: `git -C /Users/rainsen/Programs/Projects/douxue/dx-source status -s`
Expected: clean — no uncommitted changes in `dx-mini/`.

- [ ] **Step 4: Optionally commit the spec + plan**

Ask the user:

> "The spec (`dx-mini/docs/superpowers/specs/2026-04-19-dx-mini-real-device-debug-fix-design.md`) and plan (`dx-mini/docs/superpowers/plans/2026-04-19-dx-mini-real-device-debug-fix.md`) are still untracked. Commit them too?"

If yes:
```bash
git add dx-mini/docs/superpowers/specs/2026-04-19-dx-mini-real-device-debug-fix-design.md dx-mini/docs/superpowers/plans/2026-04-19-dx-mini-real-device-debug-fix.md
git commit -m "docs(mini): add spec + plan for 真机调试 config-landmine fix"
```

If no: leave them untracked.

---

## Self-review notes (completed during plan authoring)

- **Spec coverage:** All three spec landmines (A=lazyCodeLoading, B=skylineRenderEnable, C=sitemap.json) have dedicated tasks (Tasks 1, 2+3, 4 respectively). Section 6 (verification) maps to Task 5. Section 7 (rollback) is referenced from Task 5's failure branch.
- **Placeholder scan:** No "TBD"/"handle edge cases"/"similar to previous"/"add validation" — every step has concrete commands and expected output.
- **Type consistency:** Only JSON-level field names used — no TS/function identifiers to cross-check. Field names (`lazyCodeLoading`, `skylineRenderEnable`, `componentFramework`, `sitemapLocation`) spelled identically across all tasks.
- **Scope:** One focused subsystem (dx-mini config). No need to split further.
- **Gates on user confirmation:** Tasks 4 Step 6, 5, and 6 Step 1 all STOP for user input before irreversible action. Aligned with user's global rule "NEVER commit without explicit authorization".
