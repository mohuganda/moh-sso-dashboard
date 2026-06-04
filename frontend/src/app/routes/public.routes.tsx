import { Route } from "react-router-dom";

import PublicLayout from "../layouts/public/PublicLayout";
import NewsFeedPage from "@/features/newsfeed/pages/news_feed.component";

export const publicRoutes = (
  <Route element={<PublicLayout />}>
    <Route index element={<NewsFeedPage />} />
    <Route path="/" element={<NewsFeedPage />} />
  </Route>
);
