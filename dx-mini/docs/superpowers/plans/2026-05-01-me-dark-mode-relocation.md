# 我的 Page Dark-Mode Relocation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the dark-mode sun/moon toggle from the empty top-bar into the menu cell-group as a labeled `深色模式` row, lift the rest of `我的` page flush to the system status bar, and tidy the cell rows so icons sit horizontally aligned with their titles with proper spacing.

**Architecture:** UI-only change limited to one tab page. WXML structural edits, page-scoped WXSS adjustments, and dead-code cleanup in the TS controller. Reuses the existing `toggleTheme` handler, the `dx-icon` component (with its `customClass` prop and `addGlobalClass: true` option), and Vant Weapp's `<van-cell center>` flag. No new dependencies, no new files, no API changes, no JSON changes.

**Tech Stack:** WeChat Mini Program (native pages), TypeScript, Vant Weapp 1.11.x, Lucide SVG icons via the in-house `dx-icon` component.

**Spec:** `dx-mini/docs/superpowers/specs/2026-05-01-me-dark-mode-relocation-design.md`

---

## File Structure

| Path | Role | Action |
|---|---|---|
| `dx-mini/miniprogram/pages/me/me.wxml` | template — top bar removed; profile-header chevron removed; first cell-group rewritten with `center` + `custom-class` on each cell, plus 1 new `深色模式` cell appended | modify |
| `dx-mini/miniprogram/pages/me/me.wxss` | page styles — page-container padding-top reduced to status-bar-height; `.top-bar` deleted; `.cell-icon { margin-right: 12px }` added | modify |
| `dx-mini/miniprogram/pages/me/me.ts` | controller — drop dead `arrowColor` field and its two `setData` lines | modify |
| `dx-mini/miniprogram/pages/me/me.json` | unchanged — `usingComponents` already includes `van-cell`, `van-cell-group`, `dx-icon` | — |
| `dx-mini/miniprogram/components/dx-icon/index.{ts,wxml,wxss}` | unchanged — relies on existing `customClass` prop and `addGlobalClass: true` | — |

No tests are added or modified — this is a WeChat mini-program UI change with no automated UI test framework in this repo. Verification surface is TypeScript compilation (compare against baseline) and a manual smoke test in WeChat Developer Tools.

---

### Task 1: Pre-flight — verify starting state and capture TS baseline

**Files:**
- Read-only: `dx-mini/miniprogram/pages/me/me.wxml`, `me.wxss`, `me.ts`

**Why:** The repo already has known TS noise from the `Component({methods})` typing bug in `miniprogram-api-typings@2.8.3` (documented in CLAUDE.md). We need the baseline error count and identity so we can confirm our changes introduce zero new ones. We also need to confirm the file state matches the spec's assumptions before applying diffs that would break on a different starting state.

- [ ] **Step 1: Confirm `me.wxml` starting state.**

  Read `dx-mini/miniprogram/pages/me/me.wxml`. Confirm:
  - Lines 3–10 are `<view class="top-bar">…</view>` containing only `<dx-icon name="{{theme === 'dark' ? 'sun' : 'moon'}}" size="22px" color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}" bind:click="toggleTheme" />`.
  - Line 35 (inside `profile-header`, after `profile-info`'s closing `</view>`) is `<dx-icon name="chevron-right" size="16px" color="{{arrowColor}}" />`.
  - The first `<van-cell-group inset custom-style="margin:16px;">` block contains exactly 5 cells in this order: `我的团队` (`users`), `推荐有礼` (`gift`), `兑换码` (`ticket`), `购买会员` (`crown`), `收藏的课程` (`star`).
  - The second `<van-cell-group …>` contains the single `退出登录` cell.

  If anything diverges, stop — re-read the spec and update the diffs in this plan to match actual state before proceeding.

- [ ] **Step 2: Confirm `me.wxss` starting state.**

  Read `dx-mini/miniprogram/pages/me/me.wxss`. Confirm:
  - `.page-container` rule has `padding-top: calc(var(--status-bar-height, 20px) + 88rpx);`.
  - A `.top-bar` rule exists with `display: flex; justify-content: flex-end; align-items: center; gap: 20rpx; padding: 12rpx 32rpx;`.
  - The last two rules are `.dark .vip-bar { background: rgba(217,119,6,0.15); }` and `.dark .grade-badge { background: rgba(217,119,6,0.2); }`.
  - No existing `.cell-icon` rule.

- [ ] **Step 3: Confirm `me.ts` starting state.**

  Read `dx-mini/miniprogram/pages/me/me.ts`. Confirm:
  - `data` object has `arrowColor: '#9ca3af',`.
  - `onShow` calls `setData` with `arrowColor: theme === 'dark' ? '#6b7280' : '#9ca3af',`.
  - `toggleTheme` calls `setData` with `arrowColor: next === 'dark' ? '#6b7280' : '#9ca3af',`.

- [ ] **Step 4: Run baseline `tsc` and record output.**

  ```bash
  cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx --no -- tsc -p tsconfig.json --noEmit > /tmp/dx-mini-tsc-baseline.txt 2>&1; echo "exit=$?"; wc -l /tmp/dx-mini-tsc-baseline.txt
  ```

  Expected: `exit=2` (tsc returns non-zero when errors are present), and a line count somewhere in the 50–200 range, all errors of the documented `Component({methods})` shape (e.g., `Property 'setData' does not exist on type '{ … }'`). Save the line count for later comparison.

- [ ] **Step 5: Confirm baseline contains no `me.ts` errors.**

  ```bash
  grep -n "pages/me/me.ts" /tmp/dx-mini-tsc-baseline.txt; echo "exit=$?"
  ```

  Expected: `exit=1` (grep found no matches). If matches appear, stop — `me.ts` already had errors before our change, which would invalidate the post-change diff.

---

### Task 2: WXML — strip the top-bar wrapper

**Files:**
- Modify: `dx-mini/miniprogram/pages/me/me.wxml`

- [ ] **Step 1: Delete the `<view class="top-bar">…</view>` block.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.wxml`.

  `old_string`:
  ```
      <view class="top-bar">
        <dx-icon
          name="{{theme === 'dark' ? 'sun' : 'moon'}}"
          size="22px"
          color="{{theme === 'dark' ? '#14b8a6' : '#0d9488'}}"
          bind:click="toggleTheme"
        />
      </view>

      <van-loading wx:if="{{loading}}" size="30px" color="{{primaryColor}}" class="center-loader" />
  ```

  `new_string`:
  ```
      <van-loading wx:if="{{loading}}" size="30px" color="{{primaryColor}}" class="center-loader" />
  ```

  After this edit, `<van-loading …>` becomes the first child under `<view class="page-container …">`.

---

### Task 3: WXML — drop the chevron-right from `profile-header`

**Files:**
- Modify: `dx-mini/miniprogram/pages/me/me.wxml`

- [ ] **Step 1: Remove the trailing `<dx-icon name="chevron-right" …/>` inside `profile-header`.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.wxml`.

  `old_string`:
  ```
              <text class="exp-badge">Lv.{{profile.level}}</text>
            </view>
          </view>
          <dx-icon name="chevron-right" size="16px" color="{{arrowColor}}" />
        </view>
  ```

  `new_string`:
  ```
              <text class="exp-badge">Lv.{{profile.level}}</text>
            </view>
          </view>
        </view>
  ```

  After this edit, `profile-header` ends with the closing `</view>` of `profile-info` followed by the closing `</view>` of `profile-header` itself, with no chevron in between. The `bind:tap="goProfileEdit"` on `profile-header` is preserved.

---

### Task 4: WXML — rewrite the first cell-group (5 updated + 1 new)

**Files:**
- Modify: `dx-mini/miniprogram/pages/me/me.wxml`

This single Edit performs all four logical changes inside the first cell-group:
1. Adds `center` to each `<van-cell>` so the icon and title flex children align on the same horizontal line.
2. Adds `custom-class="cell-icon"` to each `<dx-icon slot="icon" …>` so the shared margin-right rule applies.
3. Appends the new `深色模式` cell at the end of the same group, right after `收藏的课程`.
4. Leaves the second cell-group (`退出登录`) untouched — it has no icon, no alignment issue.

- [ ] **Step 1: Replace the entire first cell-group block.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.wxml`.

  `old_string`:
  ```
        <van-cell title="我的团队" is-link bind:click="goGroups">
          <dx-icon slot="icon" name="users" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="推荐有礼" is-link bind:click="goInvite">
          <dx-icon slot="icon" name="gift" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="兑换码" is-link bind:click="goRedeem">
          <dx-icon slot="icon" name="ticket" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="购买会员" is-link bind:click="goPurchase">
          <dx-icon slot="icon" name="crown" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="收藏的课程" is-link bind:click="goFavorites">
          <dx-icon slot="icon" name="star" size="20px" color="{{cellIconColor}}" />
        </van-cell>
      </van-cell-group>
  ```

  `new_string`:
  ```
        <van-cell title="我的团队" is-link center bind:click="goGroups">
          <dx-icon slot="icon" custom-class="cell-icon" name="users" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="推荐有礼" is-link center bind:click="goInvite">
          <dx-icon slot="icon" custom-class="cell-icon" name="gift" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="兑换码" is-link center bind:click="goRedeem">
          <dx-icon slot="icon" custom-class="cell-icon" name="ticket" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="购买会员" is-link center bind:click="goPurchase">
          <dx-icon slot="icon" custom-class="cell-icon" name="crown" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="收藏的课程" is-link center bind:click="goFavorites">
          <dx-icon slot="icon" custom-class="cell-icon" name="star" size="20px" color="{{cellIconColor}}" />
        </van-cell>
        <van-cell title="深色模式" center bind:click="toggleTheme">
          <dx-icon
            slot="icon"
            custom-class="cell-icon"
            name="{{theme === 'dark' ? 'sun' : 'moon'}}"
            size="20px"
            color="{{cellIconColor}}"
          />
        </van-cell>
      </van-cell-group>
  ```

  Notes on the new cell:
  - No `is-link` (toggle, not navigation — no chevron).
  - Sun/moon swap (`theme === 'dark' ? 'sun' : 'moon'`) preserves the prior top-bar's "icon shows the action" convention.
  - No right-side label and no `<van-switch>` — keeps the row visually consistent with siblings and avoids adding a dependency to `usingComponents` (which would require running 构建 npm in WeChat DevTools).

- [ ] **Step 2: Spot-check the WXML.**

  ```bash
  grep -n "top-bar\|chevron-right\|cell-icon\|深色模式\|center bind\|is-link center" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/pages/me/me.wxml
  ```

  Expected output (order may vary slightly by line numbers):
  - **Zero** matches for `top-bar` and `chevron-right` (both removed).
  - **6** matches for `cell-icon` (one per cell with an icon).
  - **1** match for `深色模式`.
  - **5** matches for `is-link center` (the navigation cells).
  - **1** match for `center bind` that is NOT `is-link center bind` (the toggle cell).

---

### Task 5: WXSS — lift padding, delete `.top-bar`, add `.cell-icon`

**Files:**
- Modify: `dx-mini/miniprogram/pages/me/me.wxss`

- [ ] **Step 1: Reduce `.page-container` padding-top to clear only the system status bar.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.wxss`.

  `old_string`:
  ```
  .page-container {
    min-height: 100vh;
    background: var(--bg-page);
    padding-top: calc(var(--status-bar-height, 20px) + 88rpx);
    padding-bottom: 100rpx;
  }
  ```

  `new_string`:
  ```
  .page-container {
    min-height: 100vh;
    background: var(--bg-page);
    padding-top: var(--status-bar-height, 20px);
    padding-bottom: 100rpx;
  }
  ```

- [ ] **Step 2: Delete the `.top-bar` rule.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.wxss`.

  `old_string`:
  ```
  .top-bar {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 20rpx;
    padding: 12rpx 32rpx;
  }
  .center-loader { display: flex; justify-content: center; padding: 40px; }
  ```

  `new_string`:
  ```
  .center-loader { display: flex; justify-content: center; padding: 40px; }
  ```

- [ ] **Step 3: Append the `.cell-icon` rule at the end of the file.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.wxss`.

  `old_string`:
  ```
  .dark .vip-bar { background: rgba(217,119,6,0.15); }
  .dark .grade-badge { background: rgba(217,119,6,0.2); }
  ```

  `new_string`:
  ```
  .dark .vip-bar { background: rgba(217,119,6,0.15); }
  .dark .grade-badge { background: rgba(217,119,6,0.2); }
  .cell-icon { margin-right: 12px; }
  ```

  Why this works: `dx-icon` declares `options: { addGlobalClass: true }` and its WXML applies the `customClass` prop value as a class on the host view (`<view class="dx-icon-host {{customClass}}">`). Page-level styles in `me.wxss` therefore reach the component's host element. 12px is between Vant's default `--padding-base` (4px, too tight against a 20px icon and 14px Chinese text) and overly wide; matches typical iOS Settings-style row spacing.

- [ ] **Step 4: Spot-check the WXSS.**

  ```bash
  grep -n "top-bar\|cell-icon\|padding-top" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/pages/me/me.wxss
  ```

  Expected:
  - **Zero** matches for `top-bar` (both selector and rule body gone).
  - **1** match for `cell-icon` (the new rule).
  - **1** match for `padding-top` (the lifted rule on `.page-container`, with `var(--status-bar-height, 20px)` and **no** `+ 88rpx`).

---

### Task 6: TS — drop dead `arrowColor` field

**Files:**
- Modify: `dx-mini/miniprogram/pages/me/me.ts`

The chevron was the only consumer of `arrowColor`; the data field and the two `setData` writes are now dead.

- [ ] **Step 1: Remove `arrowColor` from `data`.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.ts`.

  `old_string`:
  ```
    data: {
      theme: 'light' as 'light' | 'dark',
      primaryColor: '#0d9488',
      arrowColor: '#9ca3af',
      cellIconColor: '#6b7280',
  ```

  `new_string`:
  ```
    data: {
      theme: 'light' as 'light' | 'dark',
      primaryColor: '#0d9488',
      cellIconColor: '#6b7280',
  ```

- [ ] **Step 2: Remove `arrowColor` from the `onShow` `setData` call.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.ts`.

  `old_string`:
  ```
      this.setData({
        theme,
        primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488',
        arrowColor: theme === 'dark' ? '#6b7280' : '#9ca3af',
        cellIconColor: theme === 'dark' ? '#9ca3af' : '#6b7280',
      });
  ```

  `new_string`:
  ```
      this.setData({
        theme,
        primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488',
        cellIconColor: theme === 'dark' ? '#9ca3af' : '#6b7280',
      });
  ```

- [ ] **Step 3: Remove `arrowColor` from the `toggleTheme` `setData` call.**

  Use `Edit` on `dx-mini/miniprogram/pages/me/me.ts`.

  `old_string`:
  ```
      this.setData({
        theme: next,
        primaryColor: next === 'dark' ? '#14b8a6' : '#0d9488',
        arrowColor: next === 'dark' ? '#6b7280' : '#9ca3af',
        cellIconColor: next === 'dark' ? '#9ca3af' : '#6b7280',
      })
  ```

  `new_string`:
  ```
      this.setData({
        theme: next,
        primaryColor: next === 'dark' ? '#14b8a6' : '#0d9488',
        cellIconColor: next === 'dark' ? '#9ca3af' : '#6b7280',
      })
  ```

- [ ] **Step 4: Spot-check the TS.**

  ```bash
  grep -n "arrowColor" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini/miniprogram/pages/me/me.ts; echo "exit=$?"
  ```

  Expected: `exit=1` (no matches). All `arrowColor` references gone.

---

### Task 7: Verify TypeScript output is unchanged from baseline

**Files:** none modified

- [ ] **Step 1: Re-run `tsc` and diff against baseline.**

  ```bash
  cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-mini && npx --no -- tsc -p tsconfig.json --noEmit > /tmp/dx-mini-tsc-after.txt 2>&1; echo "exit=$?"; diff /tmp/dx-mini-tsc-baseline.txt /tmp/dx-mini-tsc-after.txt; echo "diff_exit=$?"
  ```

  Expected: `diff_exit=0` (identical output, no new errors, no removed errors).

- [ ] **Step 2: If the diff is non-empty, investigate.**

  Specifically check for new entries in `pages/me/me.ts`:

  ```bash
  grep "pages/me/me.ts" /tmp/dx-mini-tsc-after.txt; echo "exit=$?"
  ```

  Expected: `exit=1` (no me.ts errors). If any appear:
  - Confirm all three `arrowColor` references in me.ts are removed.
  - Confirm no syntax breakage (e.g., trailing commas in the wrong spot, unclosed braces).
  - Read me.ts and compare against the diffs in Task 6.

  Do not proceed to commit until baseline matches.

---

### Task 8: Final visual diff and commit

**Files:**
- Commit: all three modified files in one commit

- [ ] **Step 1: Inspect the full unstaged diff.**

  ```bash
  cd /Users/rainsen/Programs/Projects/douxue/dx-source && git --no-pager diff dx-mini/miniprogram/pages/me/
  ```

  Eyeball checks against the spec:
  - **`me.wxml`:** No `top-bar` reference. No `chevron-right` reference. The first cell-group has 6 cells in order: `我的团队`, `推荐有礼`, `兑换码`, `购买会员`, `收藏的课程`, `深色模式`. Each has `center` and each `<dx-icon slot="icon">` has `custom-class="cell-icon"`. The second cell-group (`退出登录`) is unchanged.
  - **`me.wxss`:** `.page-container` `padding-top: var(--status-bar-height, 20px);` (no `+ 88rpx`). No `.top-bar` rule. Last rule is `.cell-icon { margin-right: 12px; }`.
  - **`me.ts`:** No `arrowColor` references. The three `setData` calls are intact otherwise.

- [ ] **Step 2: Stage and commit.**

  ```bash
  cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-mini/miniprogram/pages/me/me.wxml dx-mini/miniprogram/pages/me/me.wxss dx-mini/miniprogram/pages/me/me.ts && git commit -m "$(cat <<'EOF'
  refactor(mini): move dark mode toggle into 我的 favorites cell, lift content to status bar

  - Drop the standalone top-bar; sun/moon now lives as a 深色模式 row
    appended to the first cell-group, right after 收藏的课程.
  - Lift page-container padding-top to var(--status-bar-height) so the
    profile-header sits flush below the system status bar (capsule overlays
    the empty right edge cleanly now that the chevron is gone).
  - Add `center` to each cell-group cell and a shared `cell-icon`
    margin-right rule so icons sit horizontally aligned with titles
    and have proper spacing.
  - Drop the now-dead arrowColor data field and its two setData writes.
  EOF
  )"
  ```

- [ ] **Step 3: Confirm the commit.**

  ```bash
  cd /Users/rainsen/Programs/Projects/douxue/dx-source && git --no-pager log -1 --stat
  ```

  Expected: 1 commit, 3 files changed under `dx-mini/miniprogram/pages/me/`. Summary line should match the commit message above.

---

### Task 9: Manual smoke-test handoff (user-driven)

WeChat Developer Tools cannot be driven from this CLI session. After implementation, the user (or a human reviewer) performs these manual checks. The agent's job is to print this list verbatim so the user knows what to validate.

- [ ] **Step 1: Print the smoke-test checklist for the user.**

  Output (verbatim) to the user:

  ```
  Implementation done. Please open WeChat Developer Tools, navigate to the 我的 tab, and verify:

  Light mode (default):
  - 我的 page profile-header sits flush below the system status bar.
  - WeChat capsule (top-right …/× pill) hovers over the empty right edge of profile-header — no content under it (chevron is gone, name area is short enough).
  - First cell-group rows in order: 我的团队, 推荐有礼, 兑换码, 购买会员, 收藏的课程, 深色模式.
  - Each row: icon on the left, ~12px gap, title text, all on the same horizontal line (vertically centered).
  - Tap profile-header (avatar / name / badges) → opens profile-edit.
  - Tap each navigation cell → routes to its respective page (groups, invite, redeem, purchase, favorites).
  - Tap 深色模式 → switches to dark mode; icon flips moon → sun; all six icons recolor.
  - Tap 退出登录 → confirm dialog; confirm logs out and reLaunches login.

  Dark mode:
  - All checks above pass; dark backgrounds, lighter text, primary color #14b8a6.
  - Tap 深色模式 again → switches back to light mode.
  - Reload the page → theme persists (storage roundtrip via dx_theme key).

  Bottom tab-bar:
  - After theme toggle on 我的, the bottom tab-bar reflects the new theme immediately.
  - Switch to other tabs and return — theme stays consistent everywhere.
  ```

- [ ] **Step 2: Wait for user confirmation before considering the task closed.**

  If the user reports a visual issue (e.g., long-nickname overlap with capsule), follow up with a small targeted fix per the spec's "Edge cases" guidance (e.g., add `padding-right: 100px` to `.profile-header`).

---

## Final Notes

- **Why no automated tests?** WeChat mini-program native pages have no test framework wired into this repo (no jest, no playwright, no Cypress). Manual smoke testing in WeChat Developer Tools is the verification surface for behavior. This is consistent with how recent dx-mini commits ship.
- **Why one commit, not per-file?** All three files implement one cohesive UI change and depend on each other (the cell-icon class needs both the WXML class assignment and the WXSS rule to take effect). A single commit keeps the change atomic in `git log` and makes a potential revert clean.
- **`addGlobalClass` mechanics, recap:** `dx-icon`'s component options include `addGlobalClass: true`. Its WXML host is `<view class="dx-icon-host {{customClass}}">`. When the page renders `<dx-icon custom-class="cell-icon" …>`, the host gets `class="dx-icon-host cell-icon"` and the page's `me.wxss` selectors apply. This is the documented mechanism in the WeChat custom-component style isolation rules.
