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
      className={`relative overflow-hidden rounded-md transition-colors ${className}`}
    >
      <span
        className={`inline-block transition-opacity duration-200 ${
          loading ? "opacity-0" : "opacity-100"
        }`}
      >
        {children}
      </span>
      {loading && (
        <span className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <span
            className="absolute inset-0 shimmer-sweep"
            style={{
              background:
                "linear-gradient(90deg, transparent 0%, rgba(255,255,255,0.12) 50%, transparent 100%)",
            }}
          />
        </span>
      )}
    </button>
  );
}
