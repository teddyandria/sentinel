import { AnimatePresence, motion } from "framer-motion";
import { CardArticle } from "../api";
import { ArticleCard } from "./ArticleCard";

interface Props {
  articles: CardArticle[] | null;
  onClose: () => void;
}

// Panneau latéral : liste tous les articles d'un point/cluster cliqué.
// C'est lui qui débloque l'accès aux articles empilés sur une même ville.
export function ArticlePanel({ articles, onClose }: Props) {
  return (
    <AnimatePresence>
      {articles && articles.length > 0 && (
        <motion.aside
          className="panel"
          initial={{ x: "100%" }}
          animate={{ x: 0 }}
          exit={{ x: "100%" }}
          transition={{ type: "tween", duration: 0.2, ease: "easeOut" }}
        >
          <div className="panel-head">
            <span>
              {articles.length} article{articles.length > 1 ? "s" : ""}
            </span>
            <button className="panel-close" onClick={onClose} aria-label="Fermer">
              ✕
            </button>
          </div>
          <div className="panel-list">
            {articles.map((a) => (
              <ArticleCard key={a.id} article={a} />
            ))}
          </div>
        </motion.aside>
      )}
    </AnimatePresence>
  );
}
