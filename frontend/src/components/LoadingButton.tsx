import React from "react";

interface LoadingButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    loading?: boolean;
    children: React.ReactNode;
}

export function LoadingButton({ loading = false, children, className = "", ...props }: LoadingButtonProps) {
    return (
        <button
            {...props}
            disabled={loading || props.disabled}
            className={`relative overflow-hidden rounded-lg transition-all duration-200 btn-glow disabled:opacity-50 disabled:cursor-not-allowed ${className}`}
        >
            <span
                className={`inline-flex items-center gap-2 transition-opacity duration-200 ${
                    loading ? "opacity-0" : "opacity-100"
                }`}
            >
                {children}
            </span>
            {loading && (
                <span className="absolute inset-0 flex items-center justify-center pointer-events-none">
                    <svg
                        className="spinner w-4 h-4"
                        viewBox="0 0 24 24"
                        fill="none"
                    >
                        <circle
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            strokeWidth="3"
                            strokeLinecap="round"
                            strokeDasharray="60"
                            strokeDashoffset="20"
                            opacity="0.3"
                        />
                        <path
                            d="M12 2a10 10 0 0 1 10 10"
                            stroke="currentColor"
                            strokeWidth="3"
                            strokeLinecap="round"
                        />
                    </svg>
                </span>
            )}
        </button>
    );
}
