export type Issue = {
  id?: string;
  issue_id: number;
  issue_code: string;
  dataset: string;
  data_element: string;
  org_unit: string;
  region?: string;
  district?: string;
  issue: string;
  date_reported: string;
  reported_by: string;
  status: string;
  issue_type: string;
  updated_by: string;
  updated_date: string;
  priority: string;
  severity: string;
  time_period: string;
  time_Period: string;
};

export type IssuePayload = {
  dataset: string;
  data_element: string;
  org_unit: string;
  region?: string;
  district?: string;
  issue: string;
  issue_type: string;
  reported_by?: string;
  updated_by?: string;
  priority?: string;
  severity?: string;
  time_period: string;
};

export type IssueTransaction = {
  id?: string | number;
  issue_id?: number;
  issue_code?: string;

  resolution_date?: string | null;
  resolved_by?: string | null;
  resolution_action?: string | null;

  comment?: string | null;
  status?: string | null;
  created_by?: string | null;
  updated_by?: string | null;
  created_at?: string | null;
  updated_at?: string | null;
};

export type IssueTransactionPayload = {
  comment?: string;
  status?: string;
  resolution?: string;
  updated_by?: string;
  [key: string]: unknown;
};

export type IssueProgramSummary = {
  program: string;
  issue_count: number;
  open_count: number;
  resolved_count: number;
};

export type GetIssuesParams = {
  limit: number;
  offset: number;
  program?: string;
};

export type GetIssuesResponse = {
  items: Issue[];
  totalCount: number;
};


