// Builds miniprogram/components/dx-icon/icons.ts from Lucide SVGs.
//
// Adaptation notes:
//   - lucide-static@0.460 renamed "home" -> "house" and "help-circle" -> "circle-help".
//     The logical icon names (used by <dx-icon name="...">) remain "home" and "help-circle".

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '..')

// Canonical inventory. Each entry: [logicalName, lucideFilename].
// Add a row here and re-run `npm run build:icons` to expose a new icon.
const ICONS = [
  ['moon',          'moon'],
  ['sun',           'sun'],
  ['search',        'search'],
  ['bell',          'bell'],
  ['chevron-right', 'chevron-right'],
  ['chevron-left',  'chevron-left'],
  ['star',          'star'],
  ['book-open',     'book-open'],
  ['check',         'check'],
  ['help-circle',   'circle-help'],   // lucide-static renamed help-circle -> circle-help
  ['clock',         'clock'],
  ['crown',         'crown'],
  ['users',         'users'],
  ['gift',          'gift'],
  ['ticket',        'ticket'],
  ['copy',          'copy'],
  ['home',          'house'],         // lucide-static renamed home -> house
  ['notebook-text', 'notebook-text'],
  ['user',          'user'],
  ['book-text',     'book-text'],
  ['trophy',        'trophy'],
  ['chart-pie',      'chart-pie'],
  ['calendar-check', 'calendar-check'],
  ['sticker',        'sticker'],
  ['flag',           'flag'],
  ['keyboard',       'keyboard'],
  ['swords',         'swords'],
  ['shuffle',        'shuffle'],
  ['crosshair',      'crosshair'],
  ['sparkles',       'sparkles'],
  ['coins',          'coins'],
  ['refresh-cw',     'refresh-cw'],
  ['circle-check',   'circle-check'],
  ['message-square', 'message-square'],
  ['flame',          'flame'],
  ['arrow-right',    'arrow-right'],
  ['arrow-left',     'arrow-left'],
  ['x',              'x'],
  ['circle-x',       'circle-x'],
  ['trash-2',        'trash-2'],
  ['search-x',       'search-x'],
  ['heart',           'heart'],
  ['message-circle',  'message-circle'],
  ['bookmark',        'bookmark'],
  ['user-plus',       'user-plus'],
  ['user-check',      'user-check'],
  ['send',            'send'],
  ['image',           'image'],
  ['plus',            'plus'],
  ['more-horizontal', 'ellipsis'],           // lucide-static renamed more-horizontal -> ellipsis
  ['message-circle-more', 'message-circle-more'],
  ['megaphone',           'megaphone'],
  ['rocket',              'rocket'],
  ['shield',              'shield'],
  ['calendar',            'calendar'],
  ['zap',                 'zap'],
  ['party-popper',        'party-popper'],
  ['info',                'info'],
  ['alert-triangle',      'triangle-alert'],   // lucide-static renamed alert-triangle -> triangle-alert
  ['check-circle-2',      'circle-check'],     // lucide-static dropped check-circle-2; circle-check is the closest equivalent
]

const lucideDir = path.join(repoRoot, 'node_modules', 'lucide-static', 'icons')
const wxmlRoot = path.join(repoRoot, 'miniprogram')
const outPath = path.join(repoRoot, 'miniprogram', 'components', 'dx-icon', 'icons.ts')

// Read + assert each SVG has the runtime-injection tokens we rely on.
const svgs = {}
for (const [logicalName, lucideFile] of ICONS) {
  const srcPath = path.join(lucideDir, `${lucideFile}.svg`)
  if (!fs.existsSync(srcPath)) {
    throw new Error(`lucide-static is missing "${lucideFile}.svg" (logical: "${logicalName}"). Check the lucide-static version in package.json.`)
  }
  const svg = fs.readFileSync(srcPath, 'utf8').trim()
  if (!svg.includes('currentColor')) {
    throw new Error(`Lucide SVG "${lucideFile}" has no "currentColor" — runtime color injection would no-op for "${logicalName}".`)
  }
  if (!svg.includes('stroke-width="2"')) {
    throw new Error(`Lucide SVG "${lucideFile}" has no stroke-width="2" — runtime stroke-width injection would no-op for "${logicalName}".`)
  }
  svgs[logicalName] = svg
}

// Static WXML scan: every literal <dx-icon name="foo"> must be declared in ICONS.
// Dynamic {{...}} bindings don't match the regex and are intentionally skipped —
// those cases (e.g. tabbar item.icon) are covered by the curated list itself.
const declared = new Set(ICONS.map(([name]) => name))
const wxmlFiles = []
const walk = (dir) => {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      if (entry.name === 'miniprogram_npm' || entry.name === 'node_modules') continue
      walk(path.join(dir, entry.name))
    } else if (entry.name.endsWith('.wxml')) {
      wxmlFiles.push(path.join(dir, entry.name))
    }
  }
}
walk(wxmlRoot)

const pattern = /<dx-icon[^>]*\sname="([a-z0-9-]+)"/g
for (const file of wxmlFiles) {
  const lines = fs.readFileSync(file, 'utf8').split('\n')
  for (let i = 0; i < lines.length; i++) {
    for (const match of lines[i].matchAll(pattern)) {
      const name = match[1]
      if (!declared.has(name)) {
        const rel = path.relative(repoRoot, file)
        throw new Error(`${rel}:${i + 1}: <dx-icon name="${name}"/> not in ICONS. Add it to scripts/build-icons.mjs and re-run npm run build:icons.`)
      }
    }
  }
}

// Emit icons.ts
const body = ICONS
  .map(([name]) => `  ${JSON.stringify(name)}: ${JSON.stringify(svgs[name])},`)
  .join('\n')

const content =
  `// Auto-generated by scripts/build-icons.mjs — do not edit by hand.\n` +
  `// Run \`npm run build:icons\` to regenerate.\n\n` +
  `export const icons: Record<string, string> = {\n` +
  body + '\n' +
  `}\n`

fs.writeFileSync(outPath, content)

console.log(`Wrote ${ICONS.length} icons to ${path.relative(repoRoot, outPath)}.`)
