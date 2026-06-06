import { useListMyAnnouncementsQuery } from "@/store/api/announcement.api";
import NewsFeedView from "./news-feed-view.component";

export default function UserNewsFeedPage() {
  const queryArgs = { limit: 20, offset: 0 };

  const myQuery = useListMyAnnouncementsQuery(queryArgs, {
    refetchOnMountOrArgChange: true,
  });

  return (
    <NewsFeedView
      announcements={myQuery.data ?? []}
      isLoading={myQuery.isLoading}
      isFetching={myQuery.isFetching}
      hasError={myQuery.isError && !myQuery.data?.length}
      onRetry={myQuery.refetch}
      showStatusTags
    />
  );
}
