export interface Git {
    repo?: string;
    branch?: string;
    depth?: number;
}

export interface Project {
    id: string;
    name: string;
    git?: Git;
    status: string;
    ssh_port: number;
    web_port: number;
    volumes: string[];
    image: string;
}

export interface CreateRequest {
    name: string;
    git?: Git;
    image?: string;
}

export interface UpdateRequest {
    name: string;
}
