import { AnimatePresence, motion } from "framer-motion";
import { Article } from "../api";
import { ArticleCard } from "./ArticleCard";

interface Props {
  open: boolean;
  loading: boolean;
  answer: string;
  sources: Article[];
  onClose: () => void;
}

// Panneau de réponse RAG : la réponse rédigée en haut, puis les articles sources.
export function AnswerPanel({ open, loading, answer, sources, onClose }: Props) {
  return (
    <AnimatePresence>
      {open && (
        <motion.aside
          className="panel"
          initial={{ x: "100%" }}
          animate={{ x: 0 }}
          exit={{ x: "100%" }}
          transition={{ type: "tween", duration: 0.2, ease: "easeOut" }}
        >
          <div className="panel-head">
            <span>Réponse</span>
            <button className="panel-close" onClick={onClose} aria-label="Fermer">
              ✕
            </button>
          </div>
          <div className="panel-list">
            <p className="answer-text">{loading ? "Recherche en cours…" : answer}</p>

            {!loading && sources.length > 0 && (
              <>
                <div className="answer-sources-label">Sources</div>
                {sources.map((a) => (
                  <ArticleCard key={a.id} article={a} />
                ))}
              </>
            )}
          </div>
        </motion.aside>
      )}
    </AnimatePresence>
  );
}
