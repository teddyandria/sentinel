import { useState } from "react";

interface Props {
  onAsk: (question: string) => void;
  loading: boolean;
}

// Barre de question : l'utilisateur tape une question en langage naturel.
// À la soumission, on remonte la question au parent (qui appelle /api/ask).
export function AskBar({ onAsk, loading }: Props) {
  const [q, setQ] = useState("");

  return (
    <form
      className="askbar"
      onSubmit={(e) => {
        e.preventDefault();
        const question = q.trim();
        if (question) onAsk(question);
      }}
    >
      <input
        className="askbar-input"
        type="text"
        placeholder="Pose une question sur l'actu…"
        value={q}
        onChange={(e) => setQ(e.target.value)}
      />
      <button className="askbar-btn" type="submit" disabled={loading}>
        {loading ? "…" : "Demander"}
      </button>
    </form>
  );
}
