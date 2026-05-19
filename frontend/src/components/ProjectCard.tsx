import { Project } from "../types";
import { Badge } from "./Badge";

function host() {
  return location.hostname;
}

export function ProjectCard({
  project,
  onStart,
  onStop,
  onDelete,
}: {
  project: Project;
  onStart: (id: string) => Promise<void>;
  onStop: (id: string) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
}) {
  const h = host();
  const sshCmd = project.ssh_port ? `ssh -p ${project.ssh_port} coder@${h}` : "";
  const webUrl = project.web_port ? `http://${h}:${project.web_port}` : "";

  return (
    <div className="bg-oc-surface border border-oc-border rounded-xl p-4 flex flex-col gap-3">
      <div>
        <h3 className="text-base font-semibold text-oc-text">
          {project.name || "Untitled"}
        </h3>
        <div className="text-xs text-oc-text-muted mt-1 flex items-center gap-2">
          <span>{project.id}</span>
          <span>•</span>
          <span>{project.image || ""}</span>
          <Badge status={project.status} />
        </div>
      </div>

      {project.git_repo && (
        <div className="text-xs">
          <a
            href={project.git_repo}
            target="_blank"
            rel="noopener noreferrer"
            className="text-oc-accent hover:underline"
          >
            repo
          </a>
        </div>
      )}

      <div className="text-sm space-y-1">
        {sshCmd && (
          <div className="text-oc-text-secondary">
            SSH: <code className="text-oc-text bg-oc-code-bg px-1.5 py-0.5 rounded text-xs">{sshCmd}</code>
          </div>
        )}
        {webUrl && (
          <div className="text-oc-text-secondary">
            Web:{" "}
            <a
              href={webUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-oc-accent hover:underline"
            >
              {webUrl}
            </a>
          </div>
        )}
      </div>

      <div className="flex flex-wrap gap-2 mt-auto pt-2">
        {project.status !== "running" ? (
          <button
            onClick={() => onStart(project.id)}
            className="px-3 py-1.5 bg-oc-green hover:bg-oc-green-hover text-white text-xs font-medium rounded-md transition-colors"
          >
            Start
          </button>
        ) : (
          <button
            onClick={() => onStop(project.id)}
            className="px-3 py-1.5 bg-oc-accent hover:bg-oc-accent-hover text-white text-xs font-medium rounded-md transition-colors"
          >
            Stop
          </button>
        )}
        <button
          onClick={() => onDelete(project.id)}
          className="px-3 py-1.5 bg-oc-red hover:bg-oc-red-hover text-white text-xs font-medium rounded-md transition-colors"
        >
          Delete
        </button>
      </div>
    </div>
  );
}
