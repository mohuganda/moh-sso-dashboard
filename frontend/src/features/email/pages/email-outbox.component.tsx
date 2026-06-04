import { useMemo, useState } from "react";
import {
  Button,
  Column,
  DataTable,
  Grid,
  InlineLoading,
  Pagination,
  Select,
  SelectItem,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tag,
  Tile,
} from "@carbon/react";
import { Add } from "@carbon/react/icons";

import "./email-outbox.scss";
import {
  useListEmailsQuery,
  useListEmailsByStatusQuery,
  useRetryEmailMutation,
  useDeleteEmailMutation,
} from "@/store/api/email.api";
import type { EmailStatus, EmailOutboxItem } from "@/store/types/email.types";
import { EmailOutboxActionsMenu } from "./email-outbox-actions-menu.component";
import { useHeaderPanel } from "@/shared/components/header-panel/header-panel.context";
import { useToast } from "@/shared/components/notifications/toast/useToast";
import EmailPanelComponent from "./email-panel.component";
import { EmailDetailsModal } from "./email-outbox-details-modal.component";

type StatusFilter = "ALL" | EmailStatus | string;

const EMAIL_STATUS_OPTIONS: StatusFilter[] = [
  "ALL",
  "PENDING",
  "PROCESSING",
  "SENT",
  "FAILED",
  "RETRY",
];

const headers = [
  { key: "subject", header: "Subject" },
  { key: "to", header: "Recipient" },
  { key: "status", header: "Status" },
  { key: "attempts", header: "Attempts" },
  { key: "scheduled_at", header: "Scheduled" },
  { key: "sent_at", header: "Sent" },
  { key: "created_at", header: "Created" },
  { key: "actions", header: "" },
];

function formatDate(value?: string | null) {
  if (!value) return "—";

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";

  return date.toLocaleString();
}

function getStatusTagType(status?: string) {
  switch (status) {
    case "SENT":
      return "green";
    case "FAILED":
      return "red";
    case "PROCESSING":
      return "blue";
    case "RETRY":
      return "magenta";
    case "PENDING":
      return "gray";
    default:
      return "cool-gray";
  }
}

function getRecipients(item?: EmailOutboxItem) {
  return item?.message?.to?.map((recipient) => recipient.email).join(", ") || "—";
}

function getSubject(item?: EmailOutboxItem) {
  return item?.message?.subject || "—";
}

export default function EmailOutbox() {
  const { openPanel, closePanel } = useHeaderPanel();
  const toast = useToast();

  const [statusFilter, setStatusFilter] = useState<StatusFilter>("ALL");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [selectedEmail, setSelectedEmail] = useState<EmailOutboxItem | null>(null);

  const offset = (page - 1) * pageSize;

  const allEmailsQuery = useListEmailsQuery(
    {
      limit: pageSize,
      offset,
    },
    {
      skip: statusFilter !== "ALL",
    },
  );

  const statusEmailsQuery = useListEmailsByStatusQuery(
    {
      status: statusFilter,
      limit: pageSize,
      offset,
    },
    {
      skip: statusFilter === "ALL",
    },
  );

  const [retryEmail, retryState] = useRetryEmailMutation();
  const [deleteEmail, deleteState] = useDeleteEmailMutation();

  const emails = useMemo(() => {
    if (statusFilter === "ALL") {
      return allEmailsQuery.data ?? [];
    }

    return statusEmailsQuery.data ?? [];
  }, [allEmailsQuery.data, statusEmailsQuery.data, statusFilter]);

  const isLoading = allEmailsQuery.isLoading || statusEmailsQuery.isLoading;
  const isFetching = allEmailsQuery.isFetching || statusEmailsQuery.isFetching;
  const isMutating = retryState.isLoading || deleteState.isLoading;

  const rows = useMemo(() => {
    return emails.map((item) => ({
      id: item.id,
      subject: getSubject(item),
      to: getRecipients(item),
      status: item.status,
      attempts: `${item.attempts ?? 0}/${item.max_attempts ?? "—"}`,
      scheduled_at: formatDate(item.scheduled_at),
      sent_at: formatDate(item.sent_at),
      created_at: formatDate(item.created_at),
      actions: item.id,
    }));
  }, [emails]);

  const emailById = useMemo(() => {
    return new Map(emails.map((item) => [item.id, item]));
  }, [emails]);

  const refetchEmails = () => {
    if (statusFilter === "ALL") {
      allEmailsQuery.refetch();
      return;
    }

    statusEmailsQuery.refetch();
  };

  const handleOpenSendEmailPanel = () => {
    openPanel({
      title: "Send email",
      content: (
        <EmailPanelComponent
          onSuccess={() => {
            closePanel();
            refetchEmails();
          }}
        />
      ),
      size: "md",
    });
  };

  const handleStatusChange = (value: StatusFilter) => {
    setStatusFilter(value);
    setPage(1);
  };

  const handleRetry = async (item: EmailOutboxItem) => {
    try {
      await retryEmail(item.id).unwrap();

      toast.success({
        title: "Email re-queued",
        subtitle: getSubject(item),
      });
    } catch (error) {
      console.error(error);

      toast.error({
        title: "Retry failed",
        subtitle: "Please check the email worker logs and try again.",
      });
    }
  };

  const handleDelete = async (item: EmailOutboxItem) => {
    const confirmed = window.confirm(`Delete email record "${getSubject(item)}"?`);

    if (!confirmed) return;

    try {
      await deleteEmail(item.id).unwrap();

      toast.success({
        title: "Email deleted",
        subtitle: getSubject(item),
      });

      if (selectedEmail?.id === item.id) {
        setSelectedEmail(null);
      }
    } catch (error) {
      console.error(error);

      toast.error({
        title: "Delete failed",
        subtitle: "Please try again.",
      });
    }
  };

  return (
    <div className="email-outbox-page">
      <Grid fullWidth className="email-outbox-page__grid">
        <Column lg={16} md={8} sm={4}>
          <Stack gap={5}>
            <div className="email-outbox-page__header">
              <div>
                <h1 className="email-outbox-page__title">Email Outbox</h1>
                <p className="email-outbox-page__subtitle">
                  View, filter, retry, and delete queued email records.
                </p>
              </div>

              <div className="email-outbox-page__header-actions">
                {isFetching && !isLoading && <InlineLoading description="Refreshing emails..." />}

                <Button
                  kind="primary"
                  size="md"
                  renderIcon={Add}
                  onClick={handleOpenSendEmailPanel}
                >
                  Send email
                </Button>
              </div>
            </div>

            <Tile className="email-outbox-page__filters">
              <Select
                id="email-status-filter"
                labelText="Filter by status"
                value={statusFilter}
                onChange={(event) => handleStatusChange(event.target.value as StatusFilter)}
              >
                {EMAIL_STATUS_OPTIONS.map((status) => (
                  <SelectItem
                    key={status}
                    value={status}
                    text={status === "ALL" ? "All emails" : status}
                  />
                ))}
              </Select>
            </Tile>

            <Tile className="email-outbox-page__table-tile">
              {isLoading ? (
                <InlineLoading description="Loading emails..." />
              ) : (
                <DataTable rows={rows} headers={headers}>
                  {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
                    <Table {...getTableProps()} className="email-outbox-table">
                      <TableHead>
                        <TableRow>
                          {headers.map((header) => {
                            const { key, ...headerProps } = getHeaderProps({
                              header,
                            });

                            return (
                              <TableHeader key={key} {...headerProps}>
                                {header.header}
                              </TableHeader>
                            );
                          })}
                        </TableRow>
                      </TableHead>

                      <TableBody>
                        {rows.length === 0 ? (
                          <TableRow>
                            <TableCell colSpan={headers.length}>
                              <div className="email-outbox-page__empty">No emails found.</div>
                            </TableCell>
                          </TableRow>
                        ) : (
                          rows.map((row) => {
                            const { key, ...rowProps } = getRowProps({ row });
                            const original = emailById.get(row.id);

                            return (
                              <TableRow key={key} {...rowProps}>
                                {row.cells.map((cell) => {
                                  if (cell.info.header === "status") {
                                    return (
                                      <TableCell key={cell.id}>
                                        <Tag type={getStatusTagType(String(cell.value))}>
                                          {String(cell.value)}
                                        </Tag>
                                      </TableCell>
                                    );
                                  }

                                  if (cell.info.header === "actions") {
                                    return (
                                      <TableCell key={cell.id}>
                                        {original && (
                                          <EmailOutboxActionsMenu
                                            email={original}
                                            isMutating={isMutating}
                                            onView={() => setSelectedEmail(original)}
                                            onRetry={() => handleRetry(original)}
                                            onDelete={() => handleDelete(original)}
                                          />
                                        )}
                                      </TableCell>
                                    );
                                  }

                                  return <TableCell key={cell.id}>{cell.value}</TableCell>;
                                })}
                              </TableRow>
                            );
                          })
                        )}
                      </TableBody>
                    </Table>
                  )}
                </DataTable>
              )}

              <Pagination
                page={page}
                pageSize={pageSize}
                pageSizes={[10, 20, 50, 100]}
                totalItems={
                  emails.length < pageSize ? offset + emails.length : offset + pageSize + 1
                }
                onChange={({ page, pageSize }) => {
                  setPage(page);
                  setPageSize(pageSize);
                }}
              />
            </Tile>
          </Stack>
        </Column>
      </Grid>
      <EmailDetailsModal
        open={Boolean(selectedEmail)}
        email={selectedEmail}
        onClose={() => setSelectedEmail(null)}
      />
    </div>
  );
}
