interface TabPillProps {
  label: string;
  active: boolean;
  onClick?: () => void;
  size?: "sm" | "md";
}

export function TabPill({ label, active, onClick, size = "md" }: TabPillProps) {
  const padding = size === "sm" ? "px-4 py-2" : "px-5 py-2";
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-full text-[13px] font-medium ${padding} ${
        active
          ? "bg-teal-600 text-white"
          : "border border-border bg-card text-muted-foreground"
      }`}
    >
      {label}
    </button>
  );
}
