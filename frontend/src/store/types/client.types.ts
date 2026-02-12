import type { ClientRole } from "./client-role.types";

export interface Client {
  id: string;
  clientId: string;
  name: string;
  description?: string;
  icon?: string;
  enabled: boolean;
  publicClient: boolean;
  rootUrl?: string;
  baseUrl?: string;
  redirectUris: string[];
  webOrigins?: string[];
  roles?: ClientRole[];
  attributes?: ClientAttributes;
  createdAt?: string;
  updatedAt?: string;
}

export type ClientAttributes = Record<string, string | undefined>;

export interface CreateClientPayload {
  clientId: string;
  name: string;
  description?: string;
  icon?: string;
  publicClient?: boolean;
  enabled?: boolean;
  rootUrl?: string;
  baseUrl?: string;
  redirectUris?: string[];
  webOrigins?: string[];
}

export interface UpdateClientPayload {
  name?: string;
  description?: string;
  icon?: string;
  publicClient?: boolean;
  enabled?: boolean;
  rootUrl?: string;
  baseUrl?: string;
  redirectUris?: string[];
  webOrigins?: string[];
}

export type ClientsState = {
  items: Client[];
  activeClientId: string | null;
};
