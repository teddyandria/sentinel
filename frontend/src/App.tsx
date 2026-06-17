import { useEffect, useState } from "react";
import { Article, getArticles, getMapboxToken, getTopics } from "./api";
import { FilterBar } from "./components/FilterBar";
import { MapView } from "./components/MapView";

export default function App() {
  const [token, setToken] = useState("");
  const [topics, setTopics] = useState<string[]>([]);
  const [topic, setTopic] = useState("");
  const [articles, setArticles] = useState<Article[]>([]);
  const [error, setError] = useState("");

  // Au montage : token Mapbox + liste des topics.
  useEffect(() => {
    getMapboxToken()
      .then((t) => (t ? setToken(t) : setError("MAPBOX_TOKEN manquant côté serveur (.env).")))
      .catch(() => setError("Token Mapbox indisponible."));
    getTopics()
      .then(setTopics)
      .catch(() => setTopics([]));
  }, []);

  // À chaque changement de topic : on recharge les articles.
  useEffect(() => {
    getArticles(topic)
      .then(setArticles)
      .catch(() => setArticles([]));
  }, [topic]);

  return (
    <div className="app">
      <header className="header">
        <div className="brand">
          <h1>🛰 SENTINEL</h1>
          <p>Veille tech géolocalisée</p>
        </div>
        <FilterBar topics={topics} selected={topic} onSelect={setTopic} />
      </header>

      {token ? (
        <MapView token={token} articles={articles} />
      ) : (
        <div className="map-error">{error || "Chargement de la carte…"}</div>
      )}
    </div>
  );
}
