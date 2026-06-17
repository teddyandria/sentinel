interface Props {
  topics: string[];
  selected: string;
  onSelect: (topic: string) => void;
}

// Barre de filtres : "" = tous les sujets, puis un bouton par topic de l'API.
export function FilterBar({ topics, selected, onSelect }: Props) {
  return (
    <nav className="filters">
      {["", ...topics].map((topic) => (
        <button
          key={topic || "all"}
          className={"filter-btn" + (topic === selected ? " active" : "")}
          onClick={() => onSelect(topic)}
        >
          {topic === "" ? "Tous" : topic}
        </button>
      ))}
    </nav>
  );
}
