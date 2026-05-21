import { useState, useRef, useEffect } from "react";
import { LoadingButton } from "./LoadingButton";

interface Props {
  onSubmit: (name: string, gitRepo: string, image: string) => Promise<void>;
  onCancel: () => void;
}

export function CreateProjectForm({ onSubmit, onCancel }: Props) {
  const [name, setName] = useState("");
  const [gitRepo, setGitRepo] = useState("");
  const [image, setImage] = useState("");
  const [loading, setLoading] = useState(false);
  const nameRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const id = setTimeout(() => nameRef.current?.focus(), 0);
    return () => clearTimeout(id);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) {
      alert("Name is required");
      return;
    }
    setLoading(true);
    try {
      await onSubmit(trimmed, gitRepo.trim(), image.trim());
      setName("");
      setGitRepo("");
      setImage("");
    } catch (err) {
      alert(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-oc-text-secondary mb-1">
          Project name
        </label>
        <input
          ref={nameRef}
          type="text"
          placeholder="my-project"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-3 py-2 bg-oc-bg border border-oc-border rounded-md text-sm text-oc-text placeholder:text-oc-text-muted focus:outline-none focus:ring-2 focus:ring-oc-accent"
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-oc-text-secondary mb-1">
          Git repo
        </label>
        <input
          type="text"
          placeholder="user/repo (optional)"
          value={gitRepo}
          onChange={(e) => setGitRepo(e.target.value)}
          className="w-full px-3 py-2 bg-oc-bg border border-oc-border rounded-md text-sm text-oc-text placeholder:text-oc-text-muted focus:outline-none focus:ring-2 focus:ring-oc-accent"
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-oc-text-secondary mb-1">
          Image
        </label>
        <input
          type="text"
          placeholder="default image (optional)"
          value={image}
          onChange={(e) => setImage(e.target.value)}
          className="w-full px-3 py-2 bg-oc-bg border border-oc-border rounded-md text-sm text-oc-text placeholder:text-oc-text-muted focus:outline-none focus:ring-2 focus:ring-oc-accent"
        />
      </div>
      <div className="flex justify-end gap-3 pt-2">
        <button
          type="button"
          onClick={onCancel}
          disabled={loading}
          className="px-4 py-2 text-sm font-medium text-oc-text-secondary bg-oc-bg border border-oc-border rounded-md hover:bg-oc-surface-hover transition-colors disabled:opacity-50"
        >
          Cancel
        </button>
        <LoadingButton
          type="submit"
          loading={loading}
          className="px-4 py-2 bg-oc-accent hover:bg-oc-accent-hover disabled:opacity-50 text-white text-sm font-medium rounded-md transition-colors"
        >
          Create Project
        </LoadingButton>
      </div>
    </form>
  );
}
