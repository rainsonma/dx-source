/** Colored badge for status/category labels */
export function Badge({ label, bg, text }: { label: string; bg: string; text: string }) {
  return (
    <span className={`rounded-md px-2.5 py-1 text-[11px] font-semibold ${bg} ${text}`}>
      {label}
    </span>
  );
}
