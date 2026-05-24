import { Project, CreateRequest, UpdateRequest } from "./types";

const API_BASE = "/api";

async function api<T>(path: string, init?: RequestInit): Promise<T> {
    const r = await fetch(API_BASE + path, init);
    if (!r.ok) {
        throw new Error(await r.text());
    }
    const text = await r.text();
    return text ? JSON.parse(text) : undefined as T;
}

export function listProjects() {
    return api<Project[]>("/projects");
}

export function createProject(body: CreateRequest) {
    return api<Project>("/projects", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    });
}

export function updateProject(id: string, body: UpdateRequest) {
    return api<Project>(`/projects/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    });
}

export function startProject(id: string) {
    return api<Project>(`/projects/${id}/start`, { method: "POST" });
}

export function stopProject(id: string) {
    return api<Project>(`/projects/${id}/stop`, { method: "POST" });
}

export function upgradeProject(id: string) {
    return api<Project>(`/projects/${id}/upgrade`, { method: "POST" });
}

export function deleteProject(id: string) {
    return api<void>(`/projects/${id}`, { method: "DELETE" });
}
