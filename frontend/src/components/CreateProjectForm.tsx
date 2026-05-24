import { useState, useRef, useEffect } from "react";
import { LoadingButton } from "./LoadingButton";

interface Props {
  onSubmit: (
    name: string,
    gitRepo: string,
    image: string,
    branch: string,
    depth: number | undefined
  ) => Promise<void>;
  onCancel: () => void;
}

export function CreateProjectForm({ onSubmit, onCancel }: Props) {
  const [name, setName] = useState("");
  const [gitRepo, setGitRepo] = useState("");
  const [image, setImage] = useState("");
  const [branch, setBranch] = useState("");
  const [depth, setDepth] = useState("");
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const nameRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const id = setTimeout(() => nameRef.current?.focus(), 50);
    return () => clearTimeout(id);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) {
      alert("Name is required");
      return;
    }
    const depthNum = depth.trim() === "" ? undefined : parseInt(depth.trim(), 10);
    if (depth.trim() !== "" && (isNaN(depthNum!) || depthNum! <= 0)) {
      alert("Depth must be a positive integer");
      return;
    }
    setLoading(true);
    try {
      await onSubmit(trimmed, gitRepo.trim(), image.trim(), branch.trim(), depthNum);
      setName("");
      setGitRepo("");
      setImage("");
      setBranch("");
      setDepth("");
      setAdvancedOpen(false);
    } catch (err) {
      alert(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      <div>
        <label className="block text-xs font-semibold text-oc-text-secondary uppercase tracking-wider mb-2">
          Project name
        </label>
        <input
          ref={nameRef}
          type="text"
          placeholder="my-project"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-4 py-2.5 bg-oc-bg border border-oc-border rounded-xl text-sm text-oc-text placeholder:text-oc-text-muted input-glow transition-all duration-200"
        />
      </div>
      <div>
        <label className="block text-xs font-semibold text-oc-text-secondary uppercase tracking-wider mb-2">
          Git repository
        </label>
        <input
          type="text"
          placeholder="user/repo (optional)"
          value={gitRepo}
          onChange={(e) => setGitRepo(e.target.value)}
          className="w-full px-4 py-2.5 bg-oc-bg border border-oc-border rounded-xl text-sm text-oc-text placeholder:text-oc-text-muted input-glow transition-all duration-200 font-mono"
        />
      </div>

      <div className="border border-oc-border rounded-xl overflow-hidden">
        <button
          type="button"
          onClick={() => setAdvancedOpen((v) => !v)}
          className="w-full flex items-center justify-between px-4 py-3 text-sm font-medium text-oc-text-secondary bg-white/[0.02] hover:bg-white/[0.04] transition-colors"
        >
          <span>Advanced clone options</span>
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className={`transition-transform duration-200 ${advancedOpen ? "rotate-180" : ""}`}
          >
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </button>
        <div
          className="overflow-hidden transition-all duration-300 ease-in-out border-t border-oc-border"
          style={{ maxHeight: advancedOpen ? "300px" : "0px" }}
        >
          <div className="px-4 pb-4 pt-2 space-y-4">
            <div>
              <label className="block text-xs font-semibold text-oc-text-secondary uppercase tracking-wider mb-2">
                Branch
              </label>
              <input
                type="text"
                placeholder="main (optional)"
                value={branch}
                onChange={(e) => setBranch(e.target.value)}
                className="w-full px-4 py-2.5 bg-oc-bg border border-oc-border rounded-xl text-sm text-oc-text placeholder:text-oc-text-muted input-glow transition-all duration-200 font-mono"
              />
              <p className="text-xs text-oc-text-muted mt-1.5">
                Clone a specific branch instead of the default.
              </p>
            </div>
            <div>
              <label className="block text-xs font-semibold text-oc-text-secondary uppercase tracking-wider mb-2">
                Depth
              </label>
              <input
                type="number"
                min={1}
                placeholder="1 (optional)"
                value={depth}
                onChange={(e) => setDepth(e.target.value)}
                className="w-full px-4 py-2.5 bg-oc-bg border border-oc-border rounded-xl text-sm text-oc-text placeholder:text-oc-text-muted input-glow transition-all duration-200 font-mono"
              />
              <p className="text-xs text-oc-text-muted mt-1.5">
                Create a shallow clone with limited history.
              </p>
            </div>
          </div>
        </div>
      </div>

      <div>
        <label className="block text-xs font-semibold text-oc-text-secondary uppercase tracking-wider mb-2">
          Docker image
        </label>
        <input
          type="text"
          placeholder="default image (optional)"
          value={image}
          onChange={(e) => setImage(e.target.value)}
          className="w-full px-4 py-2.5 bg-oc-bg border border-oc-border rounded-xl text-sm text-oc-text placeholder:text-oc-text-muted input-glow transition-all duration-200 font-mono"
        />
      </div>

      <div className="flex justify-end gap-3 pt-3">
        <button
          type="button"
          onClick={onCancel}
          disabled={loading}
          className="px-5 py-2.5 text-sm font-medium text-oc-text-secondary bg-transparent border border-oc-border rounded-xl hover:bg-white/5 hover:border-oc-border-strong transition-all duration-200 disabled:opacity-50"
        >
          Cancel
        </button>
        <LoadingButton
          type="submit"
          loading={loading}
          className="px-5 py-2.5 bg-oc-accent hover:bg-oc-accent-hover disabled:opacity-50 text-white text-sm font-semibold rounded-xl transition-all duration-200"
        >
          Create Project
        </LoadingButton>
      </div>
    </form>
  );
}
