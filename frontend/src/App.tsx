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
import { Modal } from "./components/Modal";
import { CreateProjectForm } from "./components/CreateProjectForm";
import { ProjectCard } from "./components/ProjectCard";

export function App() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [status, setStatus] = useState("Connecting…");
  const [error, setError] = useState<string | null>(null);
  const [modalOpen, setModalOpen] = useState(false);

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
    <div className="min-h-screen bg-oc-bg text-oc-text-secondary relative">
      <div className="ambient-grid" />

      <Header status={status} error={error} />

      <main className="relative z-10 max-w-6xl mx-auto px-6 py-8">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h2 className="text-2xl font-bold text-oc-text tracking-tight">Projects</h2>
            <p className="text-sm text-oc-text-muted mt-1">
              {projects.length} {projects.length === 1 ? "workspace" : "workspaces"} managed
            </p>
          </div>
          <button
            onClick={() => setModalOpen(true)}
            className="group inline-flex items-center gap-2 px-5 py-2.5 bg-oc-accent hover:bg-oc-accent-hover text-white text-sm font-semibold rounded-xl transition-all duration-200 btn-glow"
            style={{
              boxShadow: "0 0 0 1px rgba(59,130,246,0.2), 0 8px 24px -6px rgba(59,130,246,0.35)",
            }}
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              className="transition-transform duration-200 group-hover:rotate-90"
            >
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            New Project
          </button>
        </div>

        <Modal
          isOpen={modalOpen}
          onClose={() => setModalOpen(false)}
          title="New Project"
        >
          <CreateProjectForm
            onSubmit={async (name, gitRepo, image) => {
              await handleCreate(name, gitRepo, image);
              setModalOpen(false);
            }}
            onCancel={() => setModalOpen(false)}
          />
        </Modal>

        {projects.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-24 text-center">
            <div className="w-16 h-16 rounded-2xl bg-white/[0.03] border border-white/[0.06] flex items-center justify-center mb-5">
              <svg
                width="28"
                height="28"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
                className="text-oc-text-muted"
              >
                <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
                <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
                <line x1="6" y1="6" x2="6.01" y2="6" />
                <line x1="6" y1="18" x2="6.01" y2="18" />
              </svg>
            </div>
            <p className="text-oc-text-secondary font-medium">No projects yet</p>
            <p className="text-sm text-oc-text-muted mt-1 max-w-xs">
              Create your first workspace to get started with isolated development environments.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
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
