export function Badge({ status }: { status: string }) {
  const s = (status || "").toLowerCase();

  if (s === "running") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-green-950 text-green-400 border border-green-700">
        Running
      </span>
    );
  }

  if (s === "exited") {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-red-950 text-red-400 border border-red-700">
        Stopped
      </span>
    );
  }

  return (
    <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold uppercase bg-indigo-950 text-indigo-400 border border-indigo-700">
      {status}
    </span>
  );
}
