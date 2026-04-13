type Props = {
  children: string;
};

export function DocCodeBlock({ children }: Props) {
  return (
    <pre className="overflow-x-auto rounded-lg bg-slate-800 px-5 py-4">
      <code className="font-mono text-[13px] text-slate-200">{children}</code>
    </pre>
  );
}
