export function Header({ status, error }: { status: string; error: string | null }) {
    const isError = !!error;
    const isConnected = status.includes("updated");

    return (
        <header className="relative z-10 px-6 py-4 border-b border-oc-border backdrop-blur-md bg-oc-bg/80 sticky top-0">
            <div className="max-w-6xl mx-auto flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <div className="relative flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-oc-accent/20 to-oc-green/10 border border-oc-accent/20">
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            className="text-oc-accent"
                        >
                            <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
                            <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
                            <line x1="6" y1="6" x2="6.01" y2="6" />
                            <line x1="6" y1="18" x2="6.01" y2="18" />
                        </svg>
                    </div>
                    <h1 className="text-lg font-semibold text-oc-text tracking-tight">
                        OpenCodePod
                    </h1>
                </div>

                <div className="flex items-center gap-2 text-xs font-medium">
                    {isError
                        ? (
                                <>
                                    <span className="relative flex h-2 w-2">
                                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-oc-red opacity-75" />
                                        <span className="relative inline-flex rounded-full h-2 w-2 bg-oc-red" />
                                    </span>
                                    <span className="text-oc-red font-mono">{error}</span>
                                </>
                            )
                        : (
                                <>
                                    <span className="relative flex h-2 w-2">
                                        {isConnected && (
                                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-oc-green opacity-60" />
                                        )}
                                        <span
                                            className={`relative inline-flex rounded-full h-2 w-2 ${
                                                isConnected ? "bg-oc-green" : "bg-oc-text-muted blink"
                                            }`}
                                        />
                                    </span>
                                    <span
                                        className={`font-mono ${
                                            isConnected ? "text-oc-text-secondary" : "text-oc-text-muted blink"
                                        }`}
                                    >
                                        {status}
                                    </span>
                                </>
                            )}
                </div>
            </div>
        </header>
    );
}
