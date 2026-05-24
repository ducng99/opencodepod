import { useEffect, useRef } from "react";

interface Props {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
}

export function Modal({ isOpen, onClose, title, children }: Props) {
    const contentRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!isOpen) return;
        const handleKey = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                onClose();
            }
        };
        document.addEventListener("keydown", handleKey);
        return () => document.removeEventListener("keydown", handleKey);
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center backdrop-animate">
            <div
                className="absolute inset-0 bg-black/70 backdrop-blur-md"
                onClick={onClose}
            />
            <div
                ref={contentRef}
                className="relative z-10 w-full max-w-md bg-oc-surface border border-oc-border-strong rounded-2xl shadow-2xl p-6 mx-4 modal-animate"
                style={{
                    boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.7), 0 0 0 1px rgba(255,255,255,0.05)",
                }}
            >
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-lg font-semibold text-oc-text tracking-tight">{title}</h2>
                    <button
                        type="button"
                        onClick={onClose}
                        className="text-oc-text-muted hover:text-oc-text transition-colors w-8 h-8 flex items-center justify-center rounded-lg hover:bg-white/5"
                        aria-label="Close"
                    >
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                            <path d="M3 3l10 10M13 3L3 13" />
                        </svg>
                    </button>
                </div>
                {children}
            </div>
        </div>
    );
}
