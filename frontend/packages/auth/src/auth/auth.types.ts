export interface AuthUser {
  id: string;
  username: string;
  email?: string;
  firstName?: string;
  lastName?: string;
  fullName?: string;
  isAdmin: boolean;
  isUser: boolean;
  realmRoles: string[];
  clientRoles: Record<string, string[]>;
  enabled: boolean;
  emailVerified: boolean;
  requirePwdChange: boolean;
  lastLoginAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

export type Role = "admin" | "user" | "manager";
