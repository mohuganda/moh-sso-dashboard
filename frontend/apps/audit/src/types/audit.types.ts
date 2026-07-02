export type Cursor = {
  cursor_created_at?: string | null;
  cursor_id?: string | null;
};

export type AuditListResponse = {
  items: AuditLog[];
  next_cursor: Cursor;
  has_more: boolean;
};

export type AuditActionsResponse = {
  actions: string[];
};

export interface AuditFilters {
  from: string;
  to: string;
  action?: string;
  user_id?: string;
  client_id?: string;
  ip?: string;
  success?: "true" | "false";
  limit?: number;
  cursor_id?: string | null;
  cursor_created_at?: string | null;
}

export interface AuditLog {
  id: string;
  createdAt: string;
  action: string;
  username?: string | null;
  userId?: string;
  metadata: {
    ip?: string;
    method?: string;
    status?: number;
    latency_ms?: number;
    user_agent?: string;
    client_id?: string;
    success?: boolean;
    country?: string;
    city?: string;
    [key: string]: unknown;
  };
  ip?: string;
  clientId?: string;
  success?: boolean;
}
