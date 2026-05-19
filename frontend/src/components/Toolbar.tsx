import { useState } from "react";

export function Toolbar({ onCreate }: { onCreate: (name: string, gitRepo: string, image: string) => Promise<void> }) {
  const [name, setName] = useState("");
  const [gitRepo, setGitRepo] = useState("");
  const [image, setImage] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) {
      alert("Name is required");
      return;
    }
    setLoading(true);
    try {
      await onCreate(trimmed, gitRepo.trim(), image.trim());
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
    <form onSubmit={handleSubmit} className="flex flex-wrap gap-3 mb-5">
      <input
        type="text"
        placeholder="Project name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        className="px-3 py-2 bg-oc-surface border border-oc-border rounded-md text-sm text-oc-text placeholder:text-oc-text-muted focus:outline-none focus:ring-2 focus:ring-oc-accent min-w-[140px]"
      />
      <input
        type="text"
        placeholder="Git repo (optional)"
        value={gitRepo}
        onChange={(e) => setGitRepo(e.target.value)}
        className="px-3 py-2 bg-oc-surface border border-oc-border rounded-md text-sm text-oc-text placeholder:text-oc-text-muted focus:outline-none focus:ring-2 focus:ring-oc-accent min-w-[220px]"
      />
      <input
        type="text"
        placeholder="Image (optional)"
        value={image}
        onChange={(e) => setImage(e.target.value)}
        className="px-3 py-2 bg-oc-surface border border-oc-border rounded-md text-sm text-oc-text placeholder:text-oc-text-muted focus:outline-none focus:ring-2 focus:ring-oc-accent min-w-[180px]"
      />
      <button
        type="submit"
        disabled={loading}
        className="px-4 py-2 bg-oc-accent hover:bg-oc-accent-hover disabled:opacity-50 text-white text-sm font-medium rounded-md transition-colors"
      >
        + New Project
      </button>
    </form>
  );
}
