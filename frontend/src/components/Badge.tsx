export function Badge({ status }: { status: string }) {
  const s = (status || "").toLowerCase();

  if (s === "running") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-green-950 text-oc-green border border-green-800">
        Running
      </span>
    );
  }

  if (s === "starting") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-yellow-950 text-yellow-400 border border-yellow-800">
        Starting
      </span>
    );
  }

  if (s === "unhealthy") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-orange-950 text-orange-400 border border-orange-800">
        Unhealthy
      </span>
    );
  }

  if (s === "stopped") {
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
