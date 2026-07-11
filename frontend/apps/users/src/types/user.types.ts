export interface User {
  id: string;
  username: string;
  email?: string;
  firstName?: string;
  lastName?: string;
  enabled: boolean;
  emailVerified: boolean;
  isAdmin: boolean;
  isActive?: boolean;
  realmRoles: string[];
  clientRoles: Record<string, string[]>;
  lastLoginAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateUserPayload {
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
  enabled?: boolean;
  emailVerified?: boolean;
  realmRoles?: string[];
  clientRoles?: Record<string, string[]>;
}
