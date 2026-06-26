import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import AdminLayout from "@/components/layout/AdminLayout";
import ConfigPage from "@/pages/admin/ConfigPage";
import TilesPage from "@/pages/admin/TilesPage";
import PropertiesPage from "@/pages/admin/PropertiesPage";
import CardsPage from "@/pages/admin/CardsPage";
import PlayersPage from "@/pages/admin/PlayersPage";
import { Toaster } from "@/components/ui/toaster";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<Navigate to="/admin/config" replace />} />
          <Route path="config" element={<ConfigPage />} />
          <Route path="tiles" element={<TilesPage />} />
          <Route path="properties" element={<PropertiesPage />} />
          <Route path="cards" element={<CardsPage />} />
          <Route path="players" element={<PlayersPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/admin" replace />} />
      </Routes>
      <Toaster />
    </BrowserRouter>
  );
}
