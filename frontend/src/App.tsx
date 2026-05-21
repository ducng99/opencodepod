import { useCallback, useEffect, useState } from "react";
import { Project } from "./types";
import {
  listProjects,
  createProject,
  startProject,
  stopProject,
  deleteProject,
  upgradeProject,
} from "./api";
import { Header } from "./components/Header";
import { Toolbar } from "./components/Toolbar";
import { ProjectCard } from "./components/ProjectCard";

export function App() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [status, setStatus] = useState("Connecting…");
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const data = await listProjects();
      setProjects(data);
      setStatus(`Last updated ${new Date().toLocaleTimeString()}`);
      setError(null);
    } catch (e) {
      setStatus("Error");
      setError(e instanceof Error ? e.message : String(e));
    }
  }, []);

  useEffect(() => {
    refresh();
    const id = setInterval(refresh, 5000);
    return () => clearInterval(id);
  }, [refresh]);

  const handleCreate = async (name: string, gitRepo: string, image: string) => {
    await createProject({
      name,
      git_repo: gitRepo || undefined,
      image: image || undefined,
    });
    await refresh();
  };

  const handleStart = async (id: string) => {
    await startProject(id);
    await refresh();
  };

  const handleStop = async (id: string) => {
    await stopProject(id);
    await refresh();
  };

  const handleUpgrade = async (id: string) => {
    if (!confirm("Upgrade will pull the latest image and recreate the container. Only /workspaces will be kept; all other data will be removed. Continue?")) return;
    await upgradeProject(id);
    await refresh();
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete project and its volume? This cannot be undone.")) return;
    await deleteProject(id);
    await refresh();
  };

  return (
    <div className="min-h-screen bg-oc-bg text-oc-text-secondary">
      <Header status={status} error={error} />
      <main className="max-w-6xl mx-auto px-5 py-5">
        <Toolbar onCreate={handleCreate} />
        {projects.length === 0 ? (
          <p className="text-oc-text-muted">No projects yet.</p>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {projects.map((p) => (
              <ProjectCard
                key={p.id}
                project={p}
                onStart={handleStart}
                onStop={handleStop}
                onUpgrade={handleUpgrade}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
