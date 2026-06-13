import { Route } from "react-router-dom";

import PublicLayout from "../layouts/public/public-layout.component";
import NewsFeedPage from "@/app/newsfeed/pages/news_feed.component";
import { ForbiddenPage } from "./ForbiddenPage";

export const publicRoutes = (
  <Route element={<PublicLayout />}>
    <Route index element={<NewsFeedPage />} />
    <Route path="/" element={<NewsFeedPage />} />
    <Route path="/forbidden" element={<ForbiddenPage />} />
  </Route>
);
