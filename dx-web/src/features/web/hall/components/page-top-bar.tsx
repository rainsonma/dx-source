import { TopActions } from "@/features/web/hall/components/top-actions";

/** Truncate text to maxChars, appending ellipsis if exceeded */
function truncate(text: string, maxChars: number) {
  return text.length > maxChars ? text.slice(0, maxChars) + "…" : text;
}

export function PageTopBar({
  title,
  subtitle,
  searchPlaceholder,
}: {
  title: string;
  subtitle: string;
  searchPlaceholder?: string;
}) {
  return (
    <div className="flex w-full items-center justify-between">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold text-foreground" title={title}>
          {truncate(title, 15)}
        </h1>
        <p className="text-sm text-muted-foreground" title={subtitle}>
          {truncate(subtitle, 20)}
        </p>
      </div>
      <TopActions searchPlaceholder={searchPlaceholder} />
    </div>
  );
}
