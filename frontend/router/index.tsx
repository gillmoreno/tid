import { createBrowserRouter } from "react-router-dom";
import { SiteLayout } from "@/components/layout/SiteLayout";
import { HomePage } from "@/pages/HomePage";
import { SectionPage } from "@/pages/SectionPage";
import { ItemPage } from "@/pages/ItemPage";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <SiteLayout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: "sections/:slug", element: <SectionPage /> },
      { path: "ideas/:slug", element: <ItemPage /> },
    ],
  },
]);