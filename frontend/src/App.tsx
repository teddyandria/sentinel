import { useEffect, useState } from "react";
import { Answer, Article, ask, getArticles, getMapboxToken, getTopics } from "./api";
import { FilterBar } from "./components/FilterBar";
import { AskBar } from "./components/AskBar";
import { AnswerPanel } from "./components/AnswerPanel";
import { MapView } from "./components/MapView";

export default function App() {
  const [token, setToken] = useState("");
  const [topics, setTopics] = useState<string[]>([]);
  const [topic, setTopic] = useState("");
  const [articles, setArticles] = useState<Article[]>([]);
  const [error, setError] = useState("");

  // État de la recherche RAG.
  const [answer, setAnswer] = useState<Answer | null>(null);
  const [asking, setAsking] = useState(false);

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

  // Question -> /api/ask -> réponse + sources (qui seront surlignées sur la carte).
  function handleAsk(question: string) {
    setAsking(true);
    setAnswer({ answer: "", sources: [] }); // ouvre le panneau en mode "chargement"
    ask(question)
      .then(setAnswer)
      .catch(() => setAnswer({ answer: "Erreur lors de la recherche.", sources: [] }))
      .finally(() => setAsking(false));
  }

  return (
    <div className="app">
      <header className="header">
        <div className="brand">
          <h1>🛰 SENTINEL</h1>
          <p>Veille tech géolocalisée</p>
        </div>
        <AskBar onAsk={handleAsk} loading={asking} />
        <FilterBar topics={topics} selected={topic} onSelect={setTopic} />
      </header>

      <main className="content">
        {token ? (
          <MapView token={token} articles={articles} focus={answer?.sources ?? []} />
        ) : (
          <div className="map-error">{error || "Chargement de la carte…"}</div>
        )}

        <AnswerPanel
          open={answer !== null}
          loading={asking}
          answer={answer?.answer ?? ""}
          sources={answer?.sources ?? []}
          onClose={() => setAnswer(null)}
        />
      </main>
    </div>
  );
}
