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
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 flex flex-col gap-3">
      <div>
        <h3 className="text-base font-semibold text-slate-100">
          {project.name || "Untitled"}
        </h3>
        <div className="text-xs text-slate-500 mt-1 flex items-center gap-2">
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
            className="text-blue-400 hover:text-blue-300"
          >
            repo
          </a>
        </div>
      )}

      <div className="text-sm space-y-1">
        {sshCmd && (
          <div className="text-slate-400">
            SSH: <code className="text-slate-300 bg-slate-950 px-1.5 py-0.5 rounded text-xs">{sshCmd}</code>
          </div>
        )}
        {webUrl && (
          <div className="text-slate-400">
            Web:{" "}
            <a
              href={webUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-400 hover:text-blue-300"
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
            className="px-3 py-1.5 bg-green-600 hover:bg-green-700 text-white text-xs font-medium rounded-md transition-colors"
          >
            Start
          </button>
        ) : (
          <button
            onClick={() => onStop(project.id)}
            className="px-3 py-1.5 bg-blue-500 hover:bg-blue-600 text-white text-xs font-medium rounded-md transition-colors"
          >
            Stop
          </button>
        )}
        <button
          onClick={() => onDelete(project.id)}
          className="px-3 py-1.5 bg-red-600 hover:bg-red-700 text-white text-xs font-medium rounded-md transition-colors"
        >
          Delete
        </button>
      </div>
    </div>
  );
}
