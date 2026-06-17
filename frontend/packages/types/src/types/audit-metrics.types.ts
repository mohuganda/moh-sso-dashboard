export interface AuditOverview {
  total_events: number;
  total_failures: number;
  failed_logins: number;
  successful_logins: number;
}

export interface AuditMetricsFilters {
  from: string;
  to: string;
}
