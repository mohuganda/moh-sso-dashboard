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

export interface AuditFailedLoginsByDayPoint {
  day: number;
  count: number;
}

export interface AuditFailedLoginsByDayResponse {
  series: AuditFailedLoginsByDayPoint[];
}

export interface AuditTopFailureIpPoint {
  ip: string;
  count: number;
}

export interface AuditTopFailureIpsResponse {
  items: AuditTopFailureIpPoint[];
}

export interface AuditTopFailureIpsFilters extends AuditMetricsFilters {
  limit?: number;
}
