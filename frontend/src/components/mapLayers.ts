import type { LayerProps } from "react-map-gl";

// Bulle de cluster : un cercle dont le rayon grandit avec le nombre de points.
export const clusterLayer: LayerProps = {
  id: "clusters",
  type: "circle",
  source: "articles",
  filter: ["has", "point_count"],
  paint: {
    "circle-color": "#38bdf8",
    "circle-stroke-width": 2,
    "circle-stroke-color": "#111111",
    "circle-radius": ["step", ["get", "point_count"], 16, 10, 22, 30, 30],
  },
};

// Nombre d'articles affiché au centre de la bulle.
export const clusterCountLayer: LayerProps = {
  id: "cluster-count",
  type: "symbol",
  source: "articles",
  filter: ["has", "point_count"],
  layout: {
    "text-field": ["get", "point_count_abbreviated"],
    "text-font": ["DIN Pro Medium", "Arial Unicode MS Bold"],
    "text-size": 13,
  },
  paint: { "text-color": "#111111" },
};

// Point isolé (une ville avec un seul article).
export const unclusteredPointLayer: LayerProps = {
  id: "unclustered-point",
  type: "circle",
  source: "articles",
  filter: ["!", ["has", "point_count"]],
  paint: {
    "circle-color": "#38bdf8",
    "circle-radius": 7,
    "circle-stroke-width": 2,
    "circle-stroke-color": "#111111",
  },
};
