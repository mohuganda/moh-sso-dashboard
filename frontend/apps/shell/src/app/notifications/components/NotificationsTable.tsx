import {
  Button,
  DataTable,
  Dropdown,
  InlineLoading,
  InlineNotification,
  OverflowMenu,
  OverflowMenuItem,
  Pagination,
  Search,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  Tag,
} from "@carbon/react";
import { useId, useMemo, useState } from "react";
import { getSeverityLabel, getSeverityTagType, normalizeSeverity } from "@moh-sso/ui";
import {
  useGetNotificationsQuery,
  useGetUnreadNotificationsCountQuery,
  useMarkNotificationAsReadMutation,
} from "../../api";
import { useOpenNotifications } from "../../ui-containers/ConnectedNotificationsPanel";
import "./notifications-table.scss";

const LIMIT = 100;
const HEADERS = [
  { key: "title", header: "Notification" },
  { key: "severity", header: "Severity" },
  { key: "status", header: "Status" },
  { key: "received", header: "Received" },
  { key: "actions", header: "Actions" },
];
const STATUSES = [
  { id: "all", label: "All notifications" },
  { id: "unread", label: "Unread" },
  { id: "read", label: "Read" },
];
const SEVERITIES = [
  { id: "all", label: "All severities" },
  ...["info", "success", "warning", "critical", "danger"].map((id) => ({
    id,
    label: getSeverityLabel(id),
  })),
];

function receivedAt(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "Date unavailable"
    : new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(date);
}

export function NotificationsTable() {
  const instance = useId();
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("unread");
  const [severity, setSeverity] = useState("all");
  const [requestedPage, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(5);
  const [pending, setPending] = useState<string | null>(null);
  const [error, setError] = useState("");
  const open = useOpenNotifications();
  const {
    data: items = [],
    isLoading,
    isFetching,
    isError,
    refetch,
  } = useGetNotificationsQuery({ limit: LIMIT, offset: 0 }, { refetchOnMountOrArgChange: true });
  const { data: unreadCount } = useGetUnreadNotificationsCountQuery();
  const [markRead] = useMarkNotificationAsReadMutation();
  const filtered = useMemo(() => {
    const query = search.trim().toLocaleLowerCase();
    return items.filter((item) => {
      if ((status === "read" && !item.read) || (status === "unread" && item.read)) return false;
      if (severity !== "all" && normalizeSeverity(item.severity) !== severity) return false;
      return [item.title, item.message, item.type]
        .filter(Boolean)
        .join(" ")
        .toLocaleLowerCase()
        .includes(query);
    });
  }, [items, search, status, severity]);
  const page = Math.min(requestedPage, Math.max(1, Math.ceil(filtered.length / pageSize)));
  const visible = filtered.slice((page - 1) * pageSize, page * pageSize);
  const byId = new Map(items.map((item) => [item.id, item]));
  const rows = visible.map((item) => ({
    id: item.id,
    title: item.title,
    severity: item.severity,
    status: item.read ? "Read" : "Unread",
    received: item.created_at,
    actions: "",
  }));
  const reset = () => {
    setSearch("");
    setStatus("unread");
    setSeverity("all");
    setPage(1);
  };

  return (
    <section className="notifications-table" aria-labelledby={`${instance}-heading`}>
      <div className="notifications-table__heading">
        <div>
          <h4 id={`${instance}-heading`}>Notifications</h4>
          <p>
            {items.length >= LIMIT
              ? `Latest ${LIMIT} notifications. Search and filters apply to this recent activity.`
              : "Search recent alerts and updates, or open a notification for full details."}
          </p>
        </div>
        {unreadCount !== undefined && (
          <Tag size="sm" type={unreadCount > 0 ? "blue" : "gray"}>
            {unreadCount} unread in total
          </Tag>
        )}
      </div>
      <div className="notifications-table__filters">
        <Search
          id={`${instance}-search`}
          labelText="Search notifications"
          placeholder="Search title, message or event type"
          value={search}
          onChange={(event) => {
            setSearch(event.target.value);
            setPage(1);
          }}
          onClear={() => {
            setSearch("");
            setPage(1);
          }}
        />
        <Dropdown
          id={`${instance}-status`}
          titleText="Status"
          label="Status"
          items={STATUSES}
          itemToString={(item) => item?.label ?? ""}
          selectedItem={STATUSES.find((item) => item.id === status)}
          onChange={({ selectedItem }) => {
            if (selectedItem) {
              setStatus(selectedItem.id);
              setPage(1);
            }
          }}
        />
        <Dropdown
          id={`${instance}-severity`}
          titleText="Severity"
          label="Severity"
          items={SEVERITIES}
          itemToString={(item) => item?.label ?? ""}
          selectedItem={SEVERITIES.find((item) => item.id === severity)}
          onChange={({ selectedItem }) => {
            if (selectedItem) {
              setSeverity(selectedItem.id);
              setPage(1);
            }
          }}
        />
        <Button kind="ghost" onClick={reset}>
          Reset filters
        </Button>
      </div>
      {error && (
        <InlineNotification
          kind="error"
          title="Action failed"
          subtitle={error}
          lowContrast
          hideCloseButton
        />
      )}
      {isLoading ? (
        <InlineLoading description="Loading notifications…" />
      ) : isError ? (
        <div className="notifications-table__state" role="alert">
          <p>Unable to load notifications.</p>
          <Button kind="tertiary" size="sm" onClick={() => void refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <>
          <p className="notifications-table__count" role="status" aria-live="polite">
            {filtered.length} matching notification{filtered.length === 1 ? "" : "s"}
            {isFetching ? " · Refreshing…" : ""}
          </p>
          <DataTable rows={rows} headers={HEADERS}>
            {({ rows, headers, getHeaderProps, getRowProps }) => (
              <TableContainer>
                <div
                  className="notifications-table__scroll"
                  role="region"
                  aria-label="Notification table"
                  tabIndex={0}
                >
                  <Table size="lg" aria-label="Notifications">
                    <TableHead>
                      <TableRow>
                        {headers.map((header) => {
                          const { key, ...props } = getHeaderProps({ header });
                          return (
                            <TableHeader key={key} {...props}>
                              {header.header}
                            </TableHeader>
                          );
                        })}
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {rows.map((row) => {
                        const item = byId.get(row.id);
                        if (!item) return null;
                        const { key, ...props } = getRowProps({ row });
                        return (
                          <TableRow key={key} {...props}>
                            <TableCell>
                              <div className="notifications-table__title">
                                <button
                                  type="button"
                                  aria-label={`View ${item.title}`}
                                  onClick={() => open(item)}
                                >
                                  {item.title}
                                </button>
                                <span className="notifications-table__message" title={item.message}>
                                  {item.message}
                                </span>
                              </div>
                            </TableCell>
                            <TableCell>
                              <Tag size="sm" type={getSeverityTagType(item.severity)}>
                                {getSeverityLabel(item.severity)}
                              </Tag>
                            </TableCell>
                            <TableCell>
                              <Tag size="sm" type={item.read ? "gray" : "blue"}>
                                {item.read ? "Read" : "Unread"}
                              </Tag>
                            </TableCell>
                            <TableCell className="notifications-table__date">
                              {receivedAt(item.created_at)}
                            </TableCell>
                            <TableCell>
                              <OverflowMenu
                                size="sm"
                                flipped
                                iconDescription={`Actions for ${item.title}`}
                                aria-label={`Actions for ${item.title}`}
                              >
                                <OverflowMenuItem
                                  itemText="View details"
                                  onClick={() => open(item)}
                                />
                                {!item.read && (
                                  <OverflowMenuItem
                                    itemText={
                                      pending === item.id ? "Marking read…" : "Mark as read"
                                    }
                                    disabled={pending !== null}
                                    onClick={async () => {
                                      setPending(item.id);
                                      setError("");
                                      try {
                                        await markRead(item.id).unwrap();
                                      } catch {
                                        setError(
                                          "Could not mark the notification as read. Please try again.",
                                        );
                                      } finally {
                                        setPending(null);
                                      }
                                    }}
                                  />
                                )}
                              </OverflowMenu>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </div>
                {filtered.length === 0 && (
                  <div className="notifications-table__state">
                    <p>
                      {items.length
                        ? "No notifications match your filters."
                        : "No notifications yet."}
                    </p>
                    {items.length > 0 && (
                      <Button kind="ghost" size="sm" onClick={reset}>
                        Reset filters
                      </Button>
                    )}
                  </div>
                )}
                <Pagination
                  page={page}
                  pageSize={pageSize}
                  pageSizes={[5, 10, 20]}
                  totalItems={filtered.length}
                  onChange={({ page, pageSize }) => {
                    setPage(page);
                    setPageSize(pageSize);
                  }}
                />
              </TableContainer>
            )}
          </DataTable>
        </>
      )}
    </section>
  );
}
