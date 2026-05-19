export function Header({ status, error }: { status: string; error: string | null }) {
  return (
    <header className="px-5 py-4 border-b border-slate-800 flex items-center justify-between">
      <h1 className="text-xl font-semibold text-slate-100">CodePod</h1>
      <div className="text-sm text-slate-500">
        {error ? <span className="text-red-400">{error}</span> : status}
      </div>
    </header>
  );
}
