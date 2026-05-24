export function Badge({ status }: { status: string }) {
    const s = (status || "").toLowerCase();

    const variants: Record<string, { bg: string; text: string; border: string; glow?: string; pulse?: boolean }> = {
        running: {
            bg: "rgba(34, 197, 94, 0.08)",
            text: "#22c55e",
            border: "rgba(34, 197, 94, 0.2)",
            glow: "rgba(34, 197, 94, 0.3)",
            pulse: true,
        },
        starting: {
            bg: "rgba(234, 179, 8, 0.08)",
            text: "#eab308",
            border: "rgba(234, 179, 8, 0.2)",
            pulse: true,
        },
        unhealthy: {
            bg: "rgba(249, 115, 22, 0.08)",
            text: "#f97316",
            border: "rgba(249, 115, 22, 0.2)",
        },
        stopped: {
            bg: "rgba(239, 68, 68, 0.08)",
            text: "#ef4444",
            border: "rgba(239, 68, 68, 0.2)",
        },
    };

    const v = variants[s] || {
        bg: "rgba(255, 255, 255, 0.04)",
        text: "#a1a1aa",
        border: "rgba(255, 255, 255, 0.08)",
    };

    return (
        <span
            className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold uppercase tracking-wider border"
            style={{
                background: v.bg,
                color: v.text,
                borderColor: v.border,
                boxShadow: v.glow ? `0 0 12px -4px ${v.glow}` : undefined,
            }}
        >
            {v.pulse && (
                <span
                    className="relative flex h-1.5 w-1.5"
                    style={{ color: v.text }}
                >
                    <span
                        className="absolute inline-flex h-full w-full rounded-full opacity-60 status-pulse"
                        style={{ backgroundColor: v.text }}
                    />
                    <span
                        className="relative inline-flex rounded-full h-1.5 w-1.5"
                        style={{ backgroundColor: v.text }}
                    />
                </span>
            )}
            {!v.pulse && (
                <span
                    className="w-1.5 h-1.5 rounded-full"
                    style={{ backgroundColor: v.text }}
                />
            )}
            {status}
        </span>
    );
}
