import { motion } from "framer-motion";
import { CardArticle } from "../api";

// Card affichée au survol d'un point ou dans le panneau latéral. L'image
// n'apparaît que si l'article en fournit une (et se masque si l'URL casse).
export function ArticleCard({ article }: { article: CardArticle }) {
  return (
    <motion.article
      className="card"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.15, ease: "easeOut" }}
    >
      {article.image_url && (
        <div className="card-img">
          <img
            src={article.image_url}
            alt=""
            loading="lazy"
            onError={(e) => {
              const wrap = e.currentTarget.parentElement;
              if (wrap) wrap.style.display = "none";
            }}
          />
        </div>
      )}

      <div className="card-body">
        <div className="card-meta">
          <span className="card-topic">{article.topic}</span>
          {article.source && <span className="card-source">{article.source}</span>}
        </div>
        <h3 className="card-title">{article.title}</h3>
        <a className="card-link" href={article.url} target="_blank" rel="noopener noreferrer">
          Lire la source →
        </a>
      </div>
    </motion.article>
  );
}
