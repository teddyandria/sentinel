// Initialise la carte Mapbox et place un marqueur cliquable par article géolocalisé.

async function main() {
    const mapEl = document.getElementById("map");

    // Le token Mapbox vient du serveur (pas de secret codé en dur dans le JS).
    const cfg = await fetch("/api/config").then((r) => r.json());
    if (!cfg.mapboxToken) {
        mapEl.innerHTML =
            "<p class='map-error'>MAPBOX_TOKEN manquant côté serveur — renseigne-le dans le .env.</p>";
        return;
    }

    mapboxgl.accessToken = cfg.mapboxToken;
    const map = new mapboxgl.Map({
        container: "map",
        style: "mapbox://styles/mapbox/dark-v11",
        center: [0, 20],
        zoom: 1.5,
    });
    map.addControl(new mapboxgl.NavigationControl());

    const articles = await fetch("/api/articles").then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json();
    });

    for (const article of articles ?? []) {
        if (!article.location) continue;

        const popup = new mapboxgl.Popup({ offset: 24 }).setHTML(
            `<a href="${article.url}" target="_blank" rel="noopener">${article.title}</a>`
        );

        // Mapbox attend l'ordre [longitude, latitude].
        new mapboxgl.Marker({ color: "#38bdf8" })
            .setLngLat([article.location.lon, article.location.lat])
            .setPopup(popup)
            .addTo(map);
    }
}

main().catch((err) => console.error("Impossible d'initialiser la carte:", err));
