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
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div
        ref={contentRef}
        className="relative z-10 w-full max-w-md bg-oc-surface border border-oc-border rounded-lg shadow-xl p-6 mx-4"
      >
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-semibold text-oc-text">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            className="text-oc-text-muted hover:text-oc-text transition-colors text-xl leading-none"
            aria-label="Close"
          >
            &times;
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}
