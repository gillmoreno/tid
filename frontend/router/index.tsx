import { createBrowserRouter, Navigate } from "react-router-dom";
import { SiteLayout } from "@/components/layout/SiteLayout";
import { factoryEnabled } from "@/lib/features";
import { HomePage } from "@/pages/HomePage";
import { SectionPage } from "@/pages/SectionPage";
import { ItemPage } from "@/pages/ItemPage";
import { FactoryPage } from "@/pages/FactoryPage";
import { IdeasPage } from "@/pages/IdeasPage";
import { ArticleFactoryPage } from "@/pages/ArticleFactoryPage";
import { SourcesPage } from "@/pages/SourcesPage";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <SiteLayout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: "sections/:slug", element: <SectionPage /> },
      { path: "ideas/:slug", element: <ItemPage /> },
      ...(factoryEnabled
        ? [
            { path: "factory", element: <FactoryPage /> },
            { path: "factory/articles", element: <ArticleFactoryPage /> },
            { path: "factory/ideas", element: <IdeasPage /> },
            { path: "factory/sources", element: <SourcesPage /> },
          ]
        : [{ path: "factory", element: <Navigate to="/" replace /> }]),
    ],
  },
]);