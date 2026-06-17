// Types et accès à l'API Sentinel. Le frontend ne connaît rien en dur :
// le token et la liste des topics viennent du backend.

export interface Location {
  name: string;
  lat: number;
  lon: number;
}

export interface Article {
  id: number;
  title: string;
  description: string;
  url: string;
  image_url: string;
  source: string;
  topic: string;
  published_at: string;
  location: Location | null;
}

export async function getMapboxToken(): Promise<string> {
  const res = await fetch("/api/config");
  const data = await res.json();
  return data.mapboxToken ?? "";
}

export async function getTopics(): Promise<string[]> {
  const res = await fetch("/api/topics");
  return res.json();
}

export async function getArticles(topic: string): Promise<Article[]> {
  const url = topic ? `/api/articles?topic=${encodeURIComponent(topic)}` : "/api/articles";
  const res = await fetch(url);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  return data ?? [];
}
