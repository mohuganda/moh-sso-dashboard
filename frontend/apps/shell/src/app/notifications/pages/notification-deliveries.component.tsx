import {
  Button,
  DataTable,
  InlineLoading,
  Search,
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
  TextArea,
  TextInput,
} from "@carbon/react";
import { Close, Renew, Send, View } from "@carbon/react/icons";
import { useMemo, useState } from "react";

import {
  useCancelNotificationDeliveryMutation,
  useGetAllNotificationDeliveriesQuery,
  useGetNotificationDeliveryMetricsQuery,
  useRetryNotificationDeliveryMutation,
  useSendTestSMSMutation,
} from "../../api";
import type { NotificationDelivery } from "@moh-sso/types";
import {
  DataTablePagination,
  DataTableShell,
  ErrorState,
  RowActionsCell,
  TableStatusTag,
  useHeaderPanel,
  useToast,
} from "@moh-sso/ui";

import "./notification-deliveries.scss";

const CHANNEL_OPTIONS = ["", "in_app", "email", "sms", "webhook"];
const STATUS_OPTIONS = ["", "PENDING", "PROCESSING", "SENT", "FAILED", "RETRY", "CANCELLED"];

const headers = [
  { key: "channel", header: "Channel" },
  { key: "recipient", header: "Recipient" },
  { key: "status", header: "Status" },
  { key: "provider", header: "Provider" },
  { key: "attempts", header: "Attempts" },
  { key: "scheduled", header: "Scheduled" },
  { key: "sent", header: "Sent" },
  { key: "updated", header: "Updated" },
  { key: "actions", header: "" },
  { key: "raw", header: "" },
];

function formatDate(value?: string) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return date.toLocaleString();
}

function displayRecipient(delivery: NotificationDelivery) {
  const recipient = delivery.recipient ?? {};
  return String(
    recipient.email ??
      recipient.phone ??
      recipient.phone_number ??
      recipient.target_role ??
      recipient.user_id ??
      delivery.notification_id,
  );
}

function jsonPreview(value?: Record<string, unknown>) {
  if (!value || Object.keys(value).length === 0) return "-";
  return JSON.stringify(value, null, 2);
}

function DeliveryDetailPanel({ delivery }: { delivery: NotificationDelivery }) {
  return (
    <div className="notification-delivery-panel">
      <Stack gap={5}>
        <section className="notification-delivery-panel__grid">
          <div>
            <span>Delivery ID</span>
            <strong>{delivery.id}</strong>
          </div>
          <div>
            <span>Notification ID</span>
            <strong>{delivery.notification_id}</strong>
          </div>
          <div>
            <span>Channel</span>
            <strong>{delivery.channel}</strong>
          </div>
          <div>
            <span>Status</span>
            <strong>{delivery.status}</strong>
          </div>
          <div>
            <span>Provider</span>
            <strong>{delivery.provider || "-"}</strong>
          </div>
          <div>
            <span>Provider message</span>
            <strong>{delivery.provider_message_id || "-"}</strong>
          </div>
        </section>

        {delivery.last_error && (
          <section>
            <h5>Last error</h5>
            <p className="notification-delivery-panel__error">{delivery.last_error}</p>
          </section>
        )}

        <section>
          <h5>Recipient</h5>
          <pre>{jsonPreview(delivery.recipient)}</pre>
        </section>

        <section>
          <h5>Provider metadata</h5>
          <pre>{jsonPreview(delivery.provider_metadata)}</pre>
        </section>
      </Stack>
    </div>
  );
}

function TestSMSPanel({ onSent }: { onSent: () => void }) {
  const toast = useToast();
  const [to, setTo] = useState("");
  const [message, setMessage] = useState("This is a test SMS from MOH Integrated Health Portal.");
  const [sendTestSMS, { isLoading }] = useSendTestSMSMutation();

  return (
    <div className="notification-delivery-panel">
      <Stack gap={5}>
        <TextInput
          id="test-sms-recipient"
          labelText="Phone number"
          placeholder="+256..."
          value={to}
          onChange={(event) => setTo(event.target.value)}
        />
        <TextArea
          id="test-sms-message"
          labelText="Message"
          rows={5}
          maxCount={320}
          enableCounter
          value={message}
          onChange={(event) => setMessage(event.target.value)}
        />
        <Button
          renderIcon={Send}
          disabled={isLoading || !to.trim() || !message.trim()}
          onClick={async () => {
            try {
              await sendTestSMS({ to, message }).unwrap();
              toast.success({ title: "Test SMS queued", subtitle: "The delivery worker will send it." });
              onSent();
            } catch {
              toast.error({ title: "Unable to queue test SMS" });
            }
          }}
        >
          Queue test SMS
        </Button>
      </Stack>
    </div>
  );
}

export default function NotificationDeliveriesPage() {
  const toast = useToast();
  const { openPanel, closePanel } = useHeaderPanel();
  const [channel, setChannel] = useState("");
  const [status, setStatus] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const query = {
    channel,
    status,
    search,
    limit: pageSize,
    offset: (page - 1) * pageSize,
  };
  const { data, isLoading, isFetching, isError, error, refetch } =
    useGetAllNotificationDeliveriesQuery(query);
  const { data: metricsData, isLoading: isLoadingMetrics } = useGetNotificationDeliveryMetricsQuery();
  const [retryDelivery] = useRetryNotificationDeliveryMutation();
  const [cancelDelivery] = useCancelNotificationDeliveryMutation();

  const rows = useMemo(
    () =>
      (data?.items ?? []).map((delivery) => ({
        id: delivery.id,
        channel: delivery.channel,
        recipient: displayRecipient(delivery),
        status: delivery.status,
        provider: delivery.provider || "-",
        attempts: `${delivery.attempts}/${delivery.max_attempts}`,
        scheduled: formatDate(delivery.scheduled_at),
        sent: formatDate(delivery.sent_at),
        updated: formatDate(delivery.updated_at),
        actions: "",
        raw: delivery,
      })),
    [data?.items],
  );

  const metrics = metricsData?.items ?? [];
  const total = data?.total ?? 0;

  const openDelivery = (delivery: NotificationDelivery) => {
    openPanel({
      title: `Delivery: ${delivery.channel}`,
      content: <DeliveryDetailPanel delivery={delivery} />,
    });
  };

  const canRetry = (delivery: NotificationDelivery) =>
    ["FAILED", "RETRY", "PENDING", "CANCELLED"].includes(String(delivery.status).toUpperCase());
  const canCancel = (delivery: NotificationDelivery) =>
    ["PENDING", "RETRY", "PROCESSING"].includes(String(delivery.status).toUpperCase());

  return (
    <DataTableShell
      title="Notification Deliveries"
      description="Inspect notification delivery attempts across in-app, email, SMS, and webhook channels."
      rows={rows}
      headers={headers}
      getRowId={(row) => row.id}
      isLoading={isLoading}
      loadingDescription="Loading notification deliveries..."
      emptyTitle="No deliveries found"
      emptyDescription="No delivery records match the selected filters."
      tableState={
        isError ? (
          <ErrorState
            title="Failed to load notification deliveries"
            description={String((error as { message?: string })?.message ?? "Please try again.")}
            primaryAction={{ label: "Retry", onClick: refetch }}
          />
        ) : undefined
      }
      topContent={
        <section className="notification-delivery-metrics">
          {isLoadingMetrics ? (
            <InlineLoading description="Loading delivery metrics..." />
          ) : metrics.length === 0 ? (
            <div className="notification-delivery-metrics__item">
              <span>No delivery metrics yet</span>
              <strong>0</strong>
            </div>
          ) : (
            metrics.map((metric) => (
              <div
                className="notification-delivery-metrics__item"
                key={`${metric.channel}-${metric.provider || "default"}`}
              >
                <span>
                  {metric.channel}
                  {metric.provider ? ` / ${metric.provider}` : ""}
                </span>
                <strong>{metric.total}</strong>
                <small>
                  Sent {metric.sent} · Failed {metric.failed} · Retry {metric.retry}
                </small>
              </div>
            ))
          )}
        </section>
      }
      filters={
        <div className="notification-delivery-filters">
          <Search
            id="notification-delivery-search"
            labelText="Search deliveries"
            placeholder="Search by recipient, notification ID, delivery ID, provider message ID..."
            value={search}
            onChange={(event) => {
              setSearch(event.target.value);
              setPage(1);
            }}
          />
          <Select
            id="notification-delivery-channel"
            labelText="Channel"
            value={channel}
            onChange={(event) => {
              setChannel(event.target.value);
              setPage(1);
            }}
          >
            {CHANNEL_OPTIONS.map((option) => (
              <SelectItem key={option || "all"} value={option} text={option || "All channels"} />
            ))}
          </Select>
          <Select
            id="notification-delivery-status"
            labelText="Status"
            value={status}
            onChange={(event) => {
              setStatus(event.target.value);
              setPage(1);
            }}
          >
            {STATUS_OPTIONS.map((option) => (
              <SelectItem key={option || "all"} value={option} text={option || "All statuses"} />
            ))}
          </Select>
          <Button
            kind="tertiary"
            onClick={() =>
              openPanel({
                title: "Send test SMS",
                content: <TestSMSPanel onSent={closePanel} />,
              })
            }
          >
            Send test SMS
          </Button>
        </div>
      }
    >
      {({ rows, headers }) => (
        <>
          {isFetching && !isLoading ? <InlineLoading description="Refreshing..." /> : null}
          <DataTable rows={rows} headers={headers}>
            {({ rows, headers, getHeaderProps, getRowProps }) => (
              <Table size="lg">
                <TableHead>
                  <TableRow>
                    {headers
                      .filter((header) => header.key !== "raw")
                      .map((header) => {
                        const { key, ...headerProps } = getHeaderProps({ header });
                        return (
                          <TableHeader key={key} {...headerProps}>
                            {header.header}
                          </TableHeader>
                        );
                      })}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {rows.map((row) => {
                    const delivery = row.cells.find((cell) => cell.info.header === "raw")
                      ?.value as NotificationDelivery;
                    const { key, ...rowProps } = getRowProps({ row });

                    return (
                      <TableRow key={key} {...rowProps}>
                        {row.cells.map((cell) => {
                          if (cell.info.header === "raw") return null;
                          if (cell.info.header === "status") {
                            return (
                              <TableCell key={cell.id}>
                                <TableStatusTag status={String(cell.value)} />
                              </TableCell>
                            );
                          }
                          if (cell.info.header === "channel") {
                            return (
                              <TableCell key={cell.id}>
                                <Tag type={cell.value === "sms" ? "purple" : "blue"}>{cell.value}</Tag>
                              </TableCell>
                            );
                          }
                          if (cell.info.header === "actions") {
                            return (
                              <RowActionsCell key={cell.id}>
                                <Button
                                  kind="ghost"
                                  size="sm"
                                  hasIconOnly
                                  iconDescription="View delivery"
                                  renderIcon={View}
                                  onClick={() => openDelivery(delivery)}
                                />
                                <Button
                                  kind="ghost"
                                  size="sm"
                                  hasIconOnly
                                  iconDescription="Retry delivery"
                                  renderIcon={Renew}
                                  disabled={!canRetry(delivery)}
                                  onClick={async () => {
                                    try {
                                      await retryDelivery({
                                        deliveryId: delivery.id,
                                        notificationId: delivery.notification_id,
                                      }).unwrap();
                                      toast.success({ title: "Delivery queued for retry" });
                                    } catch {
                                      toast.error({ title: "Unable to retry delivery" });
                                    }
                                  }}
                                />
                                <Button
                                  kind="ghost"
                                  size="sm"
                                  hasIconOnly
                                  iconDescription="Cancel delivery"
                                  renderIcon={Close}
                                  disabled={!canCancel(delivery)}
                                  onClick={async () => {
                                    try {
                                      await cancelDelivery({
                                        deliveryId: delivery.id,
                                        notificationId: delivery.notification_id,
                                      }).unwrap();
                                      toast.warning({ title: "Delivery cancelled" });
                                    } catch {
                                      toast.error({ title: "Unable to cancel delivery" });
                                    }
                                  }}
                                />
                              </RowActionsCell>
                            );
                          }
                          return <TableCell key={cell.id}>{cell.value}</TableCell>;
                        })}
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            )}
          </DataTable>
          <DataTablePagination
            page={page}
            pageSize={pageSize}
            totalItems={total}
            onChange={({ page, pageSize }) => {
              setPage(page);
              setPageSize(pageSize);
            }}
          />
        </>
      )}
    </DataTableShell>
  );
}
