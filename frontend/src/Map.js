import React from "react";
import ReactMapboxGl from "react-mapbox-gl";
import { ZoomControl } from "react-mapbox-gl";
import MapboxDraw from "@mapbox/mapbox-gl-draw";
import "@mapbox/mapbox-gl-draw/dist/mapbox-gl-draw.css";
import { STYLE } from "./MapStyle";
import { DirectModeOverride, SimpleSelectModeOverride } from "./MapModes";
const Constants = require('@mapbox/mapbox-gl-draw/src/constants');
const turfHelpers = require('@turf/helpers');
const circle = require('@turf/circle').default;
const distance = require('@turf/distance').default;
const along = require('@turf/along').default;

const ReactMap = ReactMapboxGl({ accessToken: process.env.REACT_APP_MAPBOX_ACCESS_TOKEN });

const initalCircles = {
  src: {center: [-74.00811876441851, 40.71602161602726], radiusInKm: 0.7},
  dst: {center: [-73.98468539999953, 40.75472153232781], radiusInKm: 1.2},
};

const mapboxProps = {
  style: "mapbox://styles/mapbox/streets-v11",
  center: [-73.981865, 40.7263966],
  zoom: [12],
  containerStyle: {
    height: "100vh",
    width: "100%",
  }
};

class Map extends React.Component {
  constructor (props) {
    super(props);
    this.draw = new MapboxDraw({
      displayControlsDefault: false,
      userProperties: true,
      defaultMode: "simple_select",
      clickBuffer: 12,
      touchBuffer: 20,
      boxSelect: true,
      modes: {
        direct_select: DirectModeOverride,
        simple_select: SimpleSelectModeOverride,
      },
      styles: STYLE,
    });
  }

  updateLine() {
    this.draw.delete("trip_line");
    const src = this.draw.get("src");
    const dst = this.draw.get("dst");

    const lineLen = distance(
      turfHelpers.point(src.properties.center),
      turfHelpers.point(dst.properties.center)) - src.properties.radiusInKm - dst.properties.radiusInKm;
    if (lineLen < 0) return;

    const start = along(turfHelpers.lineString([src.properties.center, dst.properties.center]), src.properties.radiusInKm).geometry.coordinates;
    const end = along(turfHelpers.lineString([dst.properties.center, src.properties.center]), dst.properties.radiusInKm).geometry.coordinates;
    this.draw.add({
      id: "trip_line",
      type: "Feature",
      properties: {},
      geometry: {
          coordinates: [start, end],
          type: "LineString"
      }
  });
  }

  notifyParent() {
    const src = this.draw.get("src");
    const dst = this.draw.get("dst");
    this.props.onFeaturesUpdated(
      {
        src: {center: src.properties.center, radiusInKm: src.properties.radiusInKm},
        dst: {center: dst.properties.center, radiusInKm: dst.properties.radiusInKm},
      }
    );
  }

  drawCircle(id, center, radiusInKm) {
    const circleFeature = circle(center, radiusInKm);
    this.draw.add({
      id: id,
      type: "Feature",
      properties: {isCircle: true, center: center, radiusInKm: radiusInKm},
      geometry: {
        coordinates: circleFeature.geometry.coordinates,
        type: "Polygon",
      },
    });
  }

  onMapLoaded = map => {
    map.addControl(this.draw);

    this.drawCircle("src", initalCircles.src.center, initalCircles.src.radiusInKm);
    this.drawCircle("dst", initalCircles.dst.center, initalCircles.dst.radiusInKm);
    this.draw.changeMode(Constants.modes.DIRECT_SELECT, {featureId: "dst"});
    this.updateLine();
    this.notifyParent();

    map.on("draw.drag", e => this.updateLine());
    map.on("draw.update", e => this.notifyParent());
    map.on("draw.modechange", e => {console.log(e.mode); })
  };

  render() {
    return (
      <div>
        <ReactMap {...mapboxProps} onStyleLoad={map => this.onMapLoaded(map)}>
          <ZoomControl position="top-left" />
        </ReactMap>
      </div>
    );
  }
}

export default Map;
