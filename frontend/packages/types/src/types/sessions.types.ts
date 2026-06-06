export interface UserSession {
  id: string;
  ipAddress: string;
  start: number;
  lastAccess: number;
  clients: Record<string, string>;
}
