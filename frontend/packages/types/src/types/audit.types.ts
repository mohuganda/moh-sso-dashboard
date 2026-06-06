export type Cursor = {
  cursor_created_at?: string | null;
  cursor_id?: string | null;
};

export type AuditListResponse = {
  items: AuditLog[];
  next_cursor: Cursor;
  has_more: boolean;
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
  cursor?: Cursor;
}

export interface AuditLog {
  id: string;

  action: string;

  username?: string | null;
  user_id?: string | null;

  /* sql.NullTime */
  created_at: {
    Time: string;
    Valid: boolean;
  };

  /* json.RawMessage wrapped in sql.Null */
  metadata?: {
    RawMessage: {
      ip?: string;
      method?: string;
      status?: number;
      latency_ms?: number;
      user_agent?: string;
      client_id?: string;
      success?: boolean;
      country?: string;
      city?: string;
      [key: string]: any;
    };
    Valid: boolean;
  };
}
