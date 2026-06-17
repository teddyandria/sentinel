import { useRef, useState } from "react";
import Map, { Marker, Popup, NavigationControl } from "react-map-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { Article } from "../api";
import { ArticleCard } from "./ArticleCard";

interface Props {
  token: string;
  articles: Article[];
}

export function MapView({ token, articles }: Props) {
  const [hovered, setHovered] = useState<Article | null>(null);
  // Petit délai à la fermeture : laisse le temps de passer du marqueur à la card
  // sans que celle-ci ne disparaisse (sinon le lien ne serait pas cliquable).
  const closeTimer = useRef<number | null>(null);

  const open = (a: Article) => {
    if (closeTimer.current) window.clearTimeout(closeTimer.current);
    setHovered(a);
  };
  const scheduleClose = () => {
    closeTimer.current = window.setTimeout(() => setHovered(null), 150);
  };

  const geolocated = articles.filter((a) => a.location);

  return (
    <div className="map-wrap">
      <Map
        mapboxAccessToken={token}
        initialViewState={{ longitude: 0, latitude: 20, zoom: 1.4 }}
        mapStyle="mapbox://styles/mapbox/dark-v11"
        style={{ width: "100%", height: "100%" }}
      >
        <NavigationControl position="bottom-right" />

        {geolocated.map((a) => (
          <Marker key={a.id} longitude={a.location!.lon} latitude={a.location!.lat} anchor="bottom">
            <div className="pin" onMouseEnter={() => open(a)} onMouseLeave={scheduleClose} />
          </Marker>
        ))}

        {hovered && hovered.location && (
          <Popup
            longitude={hovered.location.lon}
            latitude={hovered.location.lat}
            anchor="bottom"
            offset={18}
            closeButton={false}
            closeOnClick={false}
            className="card-popup"
            maxWidth="300px"
          >
            <div onMouseEnter={() => open(hovered)} onMouseLeave={scheduleClose}>
              <ArticleCard article={hovered} />
            </div>
          </Popup>
        )}
      </Map>
    </div>
  );
}
