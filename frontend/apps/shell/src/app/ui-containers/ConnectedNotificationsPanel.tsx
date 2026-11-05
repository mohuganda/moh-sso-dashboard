import { Button } from "@carbon/react";
import { useEffect, useId, useRef, useState } from "react";
import { useAuthorization, PERMISSIONS } from "@moh-sso/auth";
import type { Notification } from "@moh-sso/types";
import { NotificationsPanel, useHeaderPanel } from "@moh-sso/ui";
import {
  useGetNotificationsQuery,
  useGetUnreadNotificationsCountQuery,
  useMarkNotificationAsReadMutation,
  useDeleteNotificationMutation,
} from "../api";
import { NotificationDetailPanel } from "../layouts/admin/NotificationDetailPanel";

export function ConnectedNotificationsPanel({
  summary = false,
  initialNotification,
}: {
  summary?: boolean;
  initialNotification?: Notification;
}) {
  const { can } = useAuthorization();
  const openNotifications = useOpenNotifications();
  const [selected, setSelected] = useState<Notification | null>(initialNotification ?? null);
  const root = useRef<HTMLDivElement>(null);
  const origin = useRef<HTMLElement | null>(null);
  const savedScroll = useRef(0);
  const unread = useGetNotificationsQuery(
    { unread: true, limit: 5, offset: 0 },
    { skip: !summary, refetchOnMountOrArgChange: true },
  );
  const recent = useGetNotificationsQuery(
    { limit: 10, offset: 0 },
    { skip: summary, refetchOnMountOrArgChange: true },
  );
  const { data: count } = useGetUnreadNotificationsCountQuery(undefined, {
    refetchOnMountOrArgChange: true,
  });
  const [markRead] = useMarkNotificationAsReadMutation();
  const [remove] = useDeleteNotificationMutation();

  useEffect(() => {
    const scroller = root.current?.querySelector(".notifications-panel__list");
    if (!scroller) return;
    if (selected) {
      savedScroll.current = scroller.scrollTop;
      scroller.scrollTop = 0;
    } else {
      scroller.scrollTop = savedScroll.current;
    }
  }, [selected]);

  const open = (notification?: Notification) => {
    if (summary) {
      openNotifications(notification);
    } else {
      origin.current =
        document.activeElement instanceof HTMLElement ? document.activeElement : null;
      setSelected(notification ?? null);
    }
  };
  const back = () => {
    setSelected(null);
    requestAnimationFrame(() => {
      if (origin.current?.isConnected) origin.current.focus();
      else
        root.current
          ?.querySelector<HTMLButtonElement>(".notification-card__open")
          ?.focus();
    });
  };

  return (
    <div ref={root} className={`notification-center ${selected === null ? "notification-center--feed" : ""}`}>
      <div hidden={selected !== null}>
        <NotificationsPanel
          notifications={unread.data ?? []}
          recentNotifications={recent.data}
          initialView={summary ? "unread" : "recent"}
          unreadCount={count}
          isLoading={unread.isLoading}
          isLoadingRecent={recent.isLoading}
          isError={unread.isError}
          isErrorRecent={recent.isError}
          showHeading={summary}
          showSwitcher={false}
          onRetry={() => {
            if (summary) void unread.refetch();
            if (!summary) void recent.refetch();
          }}
          onMarkRead={async (id) => {
            await markRead(id).unwrap();
          }}
          onDelete={
            summary
              ? undefined
              : async (id) => {
                  await remove(id).unwrap();
                }
          }
          onView={open}
        />
        {summary && (
          <Button kind="ghost" size="sm" onClick={() => open()}>
            View all notifications
          </Button>
        )}
      </div>
      {selected && (
        <NotificationDetailPanel
          notification={selected}
          onBack={back}
          canRetryDelivery={can(PERMISSIONS.notificationsWrite)}
        />
      )}
    </div>
  );
}

export function useOpenNotifications() {
  const { openPanel } = useHeaderPanel();
  const instance = useId();
  const sequence = useRef(0);
  return (notification?: Notification) =>
    openPanel({
      title: "Notifications",
      size: "lg",
      content: (
        <ConnectedNotificationsPanel
          key={`${instance}-${++sequence.current}`}
          initialNotification={notification}
        />
      ),
    });
}
