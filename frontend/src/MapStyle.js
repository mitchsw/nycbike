function FeatureStyle(id, color) {
  return [
    // INACTIVE
    // polygon fill
    {
      "id": id + "-inactive-poly-fill",
      "type": "fill",
      "filter": ["all", ["==", "$type", "Polygon"], ["==", "id", id], ["==", "active", "false"]],
      "layout": {
        "fill-sort-key": 1,
      },
      "paint": {
        "fill-color": color,
        "fill-opacity": 0.1
      }
    },
    // polygon outline
    {
      "id": id + "-inactive-poly-stroke",
      "type": "line",
      "filter": ["all", ["==", "$type", "Polygon"], ["==", "id", id], ["==", "active", "false"]],
      "layout": {
        "line-cap": "round",
        "line-join": "round",
        "line-sort-key": 2,
      },
      "paint": {
        "line-color": color,
        'line-width': [
          "case", ["get", "user_hover"], 3, 1
        ],
        "line-opacity": 0.6,
      }
    },

    // ACTIVE (selected)
    // polygon fill
    {
      "id": id + "-actvev-poly-fill",
      "type": "fill",
      "filter": ["all", ["==", "$type", "Polygon"], ["==", "id", id], ["==", "active", "true"]],
      "layout": {
        "fill-sort-key": 1,
      },
      "paint": {
        "fill-color": color,
        "fill-opacity": 0.4
      }
    },
    // polygon outline stroke
    // This doesn't style the first edge of the polygon, which uses the line stroke styling instead
    {
      "id": id + "-active-poly-stroke",
      "type": "line",
      "filter": ["all", ["==", "$type", "Polygon"], ["==", "id", id], ["==", "active", "true"]],
      "layout": {
        "line-cap": "round",
        "line-join": "round",
        "line-sort-key": 2,
      },
      "paint": {
        "line-color": color,
        "line-dasharray": [0.2, 2],
        "line-width": 3
      }
    },
    // vertex point halos
    {
      "id": id + "-vertex-halo",
      "type": "circle",
      "filter": ["all", ["==", "meta", "vertex"], ["==", "parent", id], ["==", "$type", "Point"]],
      "layout": {
        "circle-sort-key": 3,
      },
      "paint": {
        "circle-radius": 9,
        "circle-color": "#FFF"
      }
    },
    // vertex points
    {
      "id": id + "-vertex-point",
      "type": "circle",
      "filter": ["all", ["==", "meta", "vertex"], ["==", "parent", id], ["==", "$type", "Point"]],
      "layout": {
        "circle-sort-key": 4,
      },
      "paint": {
        "circle-radius": 5,
        "circle-color": color,
      }
    },
  ]
}

const LINE_STYLE = [
  {
    "id": "line",
    "type": "line",
    "filter": ["all", ["==", "$type", "LineString"]],
    "layout": {
      "line-cap": "square",
      "line-join": "round",
      "line-sort-key": 1,  // TODO: why not working!?
    },
    "paint": {
      "line-color": "#444",
      "line-opacity": 0.7,
      "line-dasharray": [2, 1],
      "line-width": 6
    }
}];

export const STYLE = FeatureStyle("src", "green").concat(FeatureStyle("dst", "red")).concat(LINE_STYLE);