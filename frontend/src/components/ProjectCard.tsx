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
}: {
  project: Project;
  onStart: (id: string) => Promise<void>;
  onStop: (id: string) => Promise<void>;
  onUpgrade: (id: string) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
}) {
  const [starting, setStarting] = useState(false);
  const [stopping, setStopping] = useState(false);
  const [upgrading, setUpgrading] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [sshCopied, setSshCopied] = useState(false);

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

  return (
    <div className="card-glow p-5 flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-base font-semibold text-oc-text tracking-tight truncate">
            {project.name || "Untitled"}
          </h3>
          <Badge status={project.status} />
        </div>
        <div className="text-xs text-oc-text-muted flex items-center gap-2 font-mono">
          <span className="px-1.5 py-0.5 rounded bg-white/[0.04] border border-white/[0.06] text-oc-text-secondary shrink-0">
            {project.id}
          </span>
          <span className="text-oc-text-muted/50">/</span>
          <span className="truncate max-w-[180px]" style={{ direction: "rtl" }} title={project.image || ""}>
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
                className="shrink-0 p-1 rounded bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.06] hover:border-white/[0.12] transition-all duration-200 opacity-0 group-hover:opacity-100 focus:opacity-100"
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
            className="px-3.5 py-2 bg-oc-green/10 hover:bg-oc-green/20 text-oc-green text-xs font-semibold rounded-lg border border-oc-green/20 hover:border-oc-green/40 transition-all duration-200"
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
            className="px-3.5 py-2 bg-oc-accent/10 hover:bg-oc-accent/20 text-oc-accent text-xs font-semibold rounded-lg border border-oc-accent/20 hover:border-oc-accent/40 transition-all duration-200"
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
          className="px-3.5 py-2 bg-white/[0.04] hover:bg-white/[0.08] text-oc-text-secondary text-xs font-semibold rounded-lg border border-white/[0.08] hover:border-white/[0.14] transition-all duration-200"
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
          className="px-3.5 py-2 bg-transparent hover:bg-oc-red/10 text-oc-text-muted hover:text-oc-red text-xs font-semibold rounded-lg border border-white/[0.06] hover:border-oc-red/30 transition-all duration-200 ml-auto"
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
