const BASE = "/api";

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...init?.headers },
    credentials: "include",
    ...init,
  });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  if (res.status === 204) return undefined as T;
  return res.json();
}

// ---- Config ----
export const getConfig = () => req<Record<string, unknown>>("/admin/config");
export const updateConfig = (body: Record<string, unknown>) =>
  req("/admin/config", { method: "PUT", body: JSON.stringify(body) });

// ---- Tiles ----
export type Tile = {
  position: number;
  type: string;
  name: string;
  description: string;
  color_group: string;
  icon: string;
};
export const getTiles = () => req<Tile[]>("/admin/tiles");
export const updateTile = (position: number, body: Partial<Tile>) =>
  req(`/admin/tiles/${position}`, { method: "PUT", body: JSON.stringify(body) });

// ---- Properties ----
export type Property = {
  id: string;
  tile_position: number;
  name: string;
  color_group: string;
  price: number;
  mortgage_value: number;
  house_cost: number;
  rent: number[];
};
export const getProperties = () => req<Property[]>("/admin/properties");
export const updateProperty = (id: string, body: Partial<Property>) =>
  req(`/admin/properties/${id}`, { method: "PUT", body: JSON.stringify(body) });

// ---- Cards ----
export type Card = {
  id: string;
  deck: "chance" | "community_chest";
  name: string;
  description: string;
  effect_type: string;
  effect_payload: Record<string, unknown>;
  is_active: boolean;
  sort_order: number;
};
export const getCards = (deck?: string) =>
  req<Card[]>(`/admin/cards${deck ? `?deck=${deck}` : ""}`);
export const createCard = (body: Omit<Card, "id">) =>
  req<Card>("/admin/cards", { method: "POST", body: JSON.stringify(body) });
export const updateCard = (id: string, body: Partial<Card>) =>
  req(`/admin/cards/${id}`, { method: "PUT", body: JSON.stringify(body) });
export const deleteCard = (id: string) =>
  req(`/admin/cards/${id}`, { method: "DELETE" });

// ---- Players ----
export type Player = {
  bastion_user_id: string;
  display_name: string;
  avatar: string;
  games_played: number;
  games_won: number;
  last_seen: string;
};
export const getPlayers = () => req<Player[]>("/admin/players");
