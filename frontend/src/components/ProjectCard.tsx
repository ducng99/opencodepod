import { useState } from "react";
import { Project } from "../types";
import { Badge } from "./Badge";
import { LoadingButton } from "./LoadingButton";

function host() {
  return location.hostname;
}

export function ProjectCard({
  project,
  onStart,
  onStop,
  onUpgrade,
  onDelete,
  onUpdate,
}: {
  project: Project;
  onStart: (id: string) => Promise<void>;
  onStop: (id: string) => Promise<void>;
  onUpgrade: (id: string) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onUpdate: (id: string, name: string) => Promise<void>;
}) {
  const [starting, setStarting] = useState(false);
  const [stopping, setStopping] = useState(false);
  const [upgrading, setUpgrading] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [sshCopied, setSshCopied] = useState(false);
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState(project.name);
  const [saving, setSaving] = useState(false);

  const copyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setSshCopied(true);
      setTimeout(() => setSshCopied(false), 2000);
    } catch {
      // ignore
    }
  };

  const h = host();
  const sshCmd = project.ssh_port ? `ssh -p ${project.ssh_port} coder@${h}` : "";
  const webUrl = project.web_port ? `http://${h}:${project.web_port}` : "";

  const handleStart = async () => {
    setStarting(true);
    try {
      await onStart(project.id);
    } finally {
      setStarting(false);
    }
  };

  const handleStop = async () => {
    setStopping(true);
    try {
      await onStop(project.id);
    } finally {
      setStopping(false);
    }
  };

  const handleUpgrade = async () => {
    setUpgrading(true);
    try {
      await onUpgrade(project.id);
    } finally {
      setUpgrading(false);
    }
  };

  const handleDelete = async () => {
    setDeleting(true);
    try {
      await onDelete(project.id);
    } finally {
      setDeleting(false);
    }
  };

  const handleSave = async () => {
    const trimmed = editName.trim();
    if (!trimmed || trimmed === project.name) {
      setEditing(false);
      setEditName(project.name);
      return;
    }
    setSaving(true);
    try {
      await onUpdate(project.id, trimmed);
      setEditing(false);
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    setEditing(false);
    setEditName(project.name);
  };

  return (
    <div className="card-glow p-5 flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center justify-between gap-3">
          {editing ? (
            <div className="flex items-center gap-2 flex-1 min-w-0">
              <input
                type="text"
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    handleSave();
                  } else if (e.key === "Escape") {
                    handleCancel();
                  }
                }}
                disabled={saving}
                autoFocus
                className={`input-inline ${saving ? "opacity-50" : ""}`}
              />
              <button
                onClick={handleSave}
                disabled={saving}
                className="btn-icon-green"
                title="Save"
              >
                {saving ? (
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="animate-spin">
                    <path d="M21 12a9 9 0 1 1-6.219-8.56" />
                  </svg>
                ) : (
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                )}
              </button>
              <button
                onClick={handleCancel}
                disabled={saving}
                className="btn-icon-neutral"
                title="Cancel"
              >
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2 flex-1 min-w-0 group/name">
              <h3
                className="text-base font-semibold text-oc-text tracking-tight truncate cursor-pointer hover:text-oc-accent transition-colors"
                onClick={() => setEditing(true)}
                title="Click to rename"
              >
                {project.name || "Untitled"}
              </h3>
              <button
                onClick={() => setEditing(true)}
                className="btn-icon-subtle opacity-0 group-hover/name:opacity-100"
                title="Rename"
              >
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-oc-text-muted">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                </svg>
              </button>
            </div>
          )}
          <Badge status={project.status} />
        </div>
        <div className="text-xs text-oc-text-muted flex items-center gap-2 font-mono">
          <span className="px-1.5 py-0.5 rounded bg-white/4 border border-white/6 text-oc-text-secondary shrink-0">
            {project.id}
          </span>
          <span className="text-oc-text-muted/50">/</span>
          <span className="truncate max-w-45" style={{ direction: "rtl" }} title={project.image || ""}>
            {project.image || "default"}
          </span>
        </div>
      </div>

      <div className="space-y-2.5 text-sm">
        {sshCmd && (
          <div className="flex items-center gap-2 text-oc-text-secondary group">
            <span className="text-xs font-semibold uppercase tracking-wider text-oc-text-muted w-8 shrink-0">SSH</span>
            <div className="flex items-center gap-1.5 min-w-0 flex-1">
              <code className="code-block truncate flex-1" title={sshCmd}>{sshCmd}</code>
              <button
                onClick={() => copyText(sshCmd)}
                className="btn-icon-subtle opacity-0 group-hover:opacity-100 focus:opacity-100"
                title="Copy SSH command"
              >
                {sshCopied ? (
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-oc-green">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                ) : (
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                  </svg>
                )}
              </button>
            </div>
          </div>
        )}
        {webUrl && (
          <div className="flex items-center gap-2 text-oc-text-secondary">
            <span className="text-xs font-semibold uppercase tracking-wider text-oc-text-muted w-8 shrink-0">Web</span>
            <a
              href={webUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="code-block truncate text-oc-accent hover:text-oc-accent-hover transition-colors"
            >
              {webUrl}
            </a>
          </div>
        )}
      </div>

      <div className="flex flex-wrap gap-2 mt-auto pt-2">
        {project.status !== "running" ? (
          <LoadingButton
            onClick={handleStart}
            loading={starting}
            className="btn-action-green"
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <polygon points="5 3 19 12 5 21 5 3" />
            </svg>
            Start
          </LoadingButton>
        ) : (
          <LoadingButton
            onClick={handleStop}
            loading={stopping}
            className="btn-action-accent"
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <rect x="6" y="6" width="12" height="12" rx="1" />
            </svg>
            Stop
          </LoadingButton>
        )}
        <LoadingButton
          onClick={handleUpgrade}
          loading={upgrading}
          className="btn-action-neutral"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
            <path d="M3 3v5h5" />
            <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
            <path d="M16 21h5v-5" />
          </svg>
          Upgrade
        </LoadingButton>
        <LoadingButton
          onClick={handleDelete}
          loading={deleting}
          className="btn-action-danger ml-auto"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M3 6h18" />
            <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
            <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
          </svg>
          Delete
        </LoadingButton>
      </div>
    </div>
  );
}
