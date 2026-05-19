export function Header({ status, error }: { status: string; error: string | null }) {
  return (
    <header className="px-5 py-4 border-b border-oc-border flex items-center justify-between">
      <h1 className="text-xl font-semibold text-oc-text">OpenCodePod</h1>
      <div className="text-sm text-oc-text-muted">
        {error ? <span className="text-oc-red">{error}</span> : status}
      </div>
    </header>
  );
}
