// Types et accès à l'API Sentinel. Le frontend ne connaît rien en dur :
// le token et la liste des topics viennent du backend.

export interface Location {
  name: string;
  lat: number;
  lon: number;
}

// CardArticle = le sous-ensemble de champs affichés dans la card / le panneau.
// Article (complet) le satisfait structurellement, tout comme les propriétés
// renvoyées par les features GeoJSON de la carte.
export interface CardArticle {
  id: number;
  title: string;
  url: string;
  image_url: string;
  source: string;
  topic: string;
}

export interface Article extends CardArticle {
  description: string;
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

// La réponse du RAG : le texte rédigé + les articles qui ont servi de sources.
export interface Answer {
  answer: string;
  sources: Article[];
}

export async function ask(question: string): Promise<Answer> {
  const res = await fetch(`/api/ask?q=${encodeURIComponent(question)}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}
