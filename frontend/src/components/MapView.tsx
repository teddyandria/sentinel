import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Map, { Source, Layer, Popup, NavigationControl } from "react-map-gl";
import type { MapRef, MapLayerMouseEvent } from "react-map-gl";
import type { GeoJSONSource } from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { Article, CardArticle } from "../api";
import { ArticleCard } from "./ArticleCard";
import { ArticlePanel } from "./ArticlePanel";
import { clusterLayer, clusterCountLayer, unclusteredPointLayer } from "./mapLayers";

interface Props {
  token: string;
  articles: Article[];
  focus?: Article[]; // articles à surligner (sources d'une réponse RAG) : la carte zoome dessus
}

// Couches cliquables/survolables.
const INTERACTIVE = ["clusters", "cluster-count", "unclustered-point"];

// Au-delà de ce zoom, Mapbox ne cluster plus. Sert aussi de seuil pour décider
// si un cluster peut encore s'éclater (zoom) ou non (points empilés -> panneau).
const CLUSTER_MAX_ZOOM = 14;

export function MapView({ token, articles, focus }: Props) {
  const mapRef = useRef<MapRef>(null);
  const [selected, setSelected] = useState<CardArticle[] | null>(null);
  const [hovered, setHovered] = useState<{ a: CardArticle; lon: number; lat: number } | null>(null);
  const [cursor, setCursor] = useState("");

  // On réinitialise les sélections quand la liste change (ex: changement de topic).
  useEffect(() => {
    setSelected(null);
    setHovered(null);
  }, [articles]);

  // Quand une réponse arrive, on cadre la carte sur ses sources géolocalisées.
  useEffect(() => {
    const map = mapRef.current?.getMap();
    if (!map || !focus || focus.length === 0) return;

    const pts = focus.filter((a) => a.location).map((a) => [a.location!.lon, a.location!.lat] as [number, number]);
    if (pts.length === 0) return;

    if (pts.length === 1) {
      map.flyTo({ center: pts[0], zoom: 5, duration: 800 });
      return;
    }
    const lons = pts.map((p) => p[0]);
    const lats = pts.map((p) => p[1]);
    map.fitBounds(
      [
        [Math.min(...lons), Math.min(...lats)],
        [Math.max(...lons), Math.max(...lats)],
      ],
      { padding: 80, maxZoom: 6, duration: 800 }
    );
  }, [focus]);

  // Les articles géolocalisés deviennent une FeatureCollection que Mapbox cluster lui-même.
  const geojson = useMemo(
    () => ({
      type: "FeatureCollection" as const,
      features: articles
        .filter((a) => a.location)
        .map((a) => ({
          type: "Feature" as const,
          geometry: { type: "Point" as const, coordinates: [a.location!.lon, a.location!.lat] },
          properties: {
            id: a.id,
            title: a.title,
            url: a.url,
            image_url: a.image_url,
            source: a.source,
            topic: a.topic,
          },
        })),
    }),
    [articles]
  );

  const onClick = useCallback((e: MapLayerMouseEvent) => {
    const feature = e.features?.[0];
    if (!feature) {
      setSelected(null); // clic dans le vide = fermeture du panneau
      return;
    }
    const props = feature.properties as Record<string, unknown>;

    if (props.cluster) {
      const map = mapRef.current?.getMap();
      const source = map?.getSource("articles") as GeoJSONSource | undefined;
      if (!map || !source) return;

      const clusterId = props.cluster_id as number;
      const coords =
        feature.geometry.type === "Point" ? (feature.geometry.coordinates as [number, number]) : null;

      source.getClusterExpansionZoom(clusterId, (err, zoom) => {
        if (err || zoom == null) return;
        if (zoom <= CLUSTER_MAX_ZOOM && coords) {
          // Le cluster peut encore s'éclater : on zoome dessus pour le séparer.
          map.easeTo({ center: coords, zoom, duration: 500 });
        } else {
          // Points empilés (mêmes coordonnées) : le zoom ne sépare rien -> on liste.
          source.getClusterLeaves(clusterId, 1000, 0, (e2, leaves) => {
            if (e2 || !leaves) return;
            setSelected(leaves.map((l) => l.properties as unknown as CardArticle));
          });
        }
      });
    } else {
      setSelected([props as unknown as CardArticle]);
    }
  }, []);

  const onMouseMove = useCallback((e: MapLayerMouseEvent) => {
    const feature = e.features?.[0];
    setCursor(feature ? "pointer" : "");

    if (feature && !feature.properties?.cluster && feature.geometry.type === "Point") {
      const [lon, lat] = feature.geometry.coordinates as [number, number];
      setHovered({ a: feature.properties as unknown as CardArticle, lon, lat });
    } else {
      setHovered(null);
    }
  }, []);

  return (
    <div className="map-wrap">
      <Map
        ref={mapRef}
        mapboxAccessToken={token}
        initialViewState={{ longitude: 0, latitude: 20, zoom: 1.4 }}
        mapStyle="mapbox://styles/mapbox/dark-v11"
        style={{ width: "100%", height: "100%" }}
        interactiveLayerIds={INTERACTIVE}
        cursor={cursor}
        onClick={onClick}
        onMouseMove={onMouseMove}
      >
        <NavigationControl position="bottom-right" />

        <Source
          id="articles"
          type="geojson"
          data={geojson}
          cluster
          clusterMaxZoom={CLUSTER_MAX_ZOOM}
          clusterRadius={50}
        >
          <Layer {...clusterLayer} />
          <Layer {...clusterCountLayer} />
          <Layer {...unclusteredPointLayer} />
        </Source>

        {hovered && (
          <Popup
            longitude={hovered.lon}
            latitude={hovered.lat}
            anchor="bottom"
            offset={14}
            closeButton={false}
            closeOnClick={false}
            className="card-popup"
            maxWidth="300px"
          >
            <div style={{ width: 280 }}>
              <ArticleCard article={hovered.a} />
            </div>
          </Popup>
        )}
      </Map>

      <ArticlePanel articles={selected} onClose={() => setSelected(null)} />
    </div>
  );
}
