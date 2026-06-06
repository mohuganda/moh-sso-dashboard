import { useListPublicAnnouncementsQuery } from "@/store/api/announcement.api";
import NewsFeedView from "./news-feed-view.component";

export default function PublicNewsFeedPage() {
  const queryArgs = { limit: 20, offset: 0 };

  const publicQuery = useListPublicAnnouncementsQuery(queryArgs, {
    refetchOnMountOrArgChange: true,
  });

  return (
    <NewsFeedView
      announcements={publicQuery.data ?? []}
      isLoading={publicQuery.isLoading}
      isFetching={publicQuery.isFetching}
      hasError={publicQuery.isError && !publicQuery.data?.length}
      onRetry={publicQuery.refetch}
      showStatusTags={false}
    />
  );
}
