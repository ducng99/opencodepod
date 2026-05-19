export function Badge({ status }: { status: string }) {
  const s = (status || "").toLowerCase();

  if (s === "running") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-green-950 text-oc-green border border-green-800">
        Running
      </span>
    );
  }

  if (s === "exited") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-red-950 text-oc-red border border-red-800">
        Stopped
      </span>
    );
  }

  return (
    <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-oc-surface text-oc-text-secondary border border-oc-border">
      {status}
    </span>
  );
}
