export interface Project {
  id: string;
  name: string;
  git_repo?: string;
  status: string;
  ssh_port: number;
  web_port: number;
  volume: string;
  image: string;
}

export interface CreateRequest {
  name: string;
  git_repo?: string;
  image?: string;
}
