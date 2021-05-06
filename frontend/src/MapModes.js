import "@mapbox/mapbox-gl-draw/dist/mapbox-gl-draw.css";
import {DirectMode, SimpleSelectMode,} from "mapbox-gl-draw-circle";
const Constants = require('@mapbox/mapbox-gl-draw/src/constants');
const turfHelpers = require('@turf/helpers');
const length = require('@turf/length').default;
const along = require('@turf/along').default;


const MAX_RADIUS = 1.5;  // km

export const DirectModeOverride = DirectMode;
export const SimpleSelectModeOverride = SimpleSelectMode;

DirectModeOverride.dragFeatureBase = DirectMode.dragFeature;
DirectModeOverride.dragVertexBase = DirectMode.dragVertex;
DirectModeOverride.dragFeature = function(state, e, delta) {
  this.dragFeatureBase(state, e, delta);
  this.map.fire("draw.drag", {});
};
DirectModeOverride.dragVertex = function(state, e, delta) {
  if (state.feature.properties.isCircle) {
    const newRadius = turfHelpers.lineString([state.feature.properties.center, [e.lngLat.lng, e.lngLat.lat]]);
    if (length(newRadius) > MAX_RADIUS) {
      const vertex = along(newRadius, MAX_RADIUS).geometry.coordinates;
      e.lngLat.lng = vertex[0];
      e.lngLat.lat = vertex[1];
    }
  }
  this.dragVertexBase(state, e, delta);
  this.map.fire("draw.drag", {});
}

DirectModeOverride.clickInactive = function (state, e) {
  const featureId = e.featureTarget.properties.id;
  return this.changeMode(Constants.modes.DIRECT_SELECT, {
    featureId
  });
};

// The vertex selection/unselected logic is intuitive for a
// line/polygon, but not for a circle in my opinion.
DirectModeOverride.onFeature = function(state, e) {
  state.selectedCoordPaths = [];
  this.startDragging(state, e);
};

SimpleSelectModeOverride.clickOnFeature = function(state, e) {
  // Stop everything
  //MapboxDraw.doubleClickZoom.disable(this);
  this.stopExtendedInteractions(state);

  const featureId = e.featureTarget.properties.id;
  const selectedFeatureIds = this.getSelectedIds();

  if (this.isSelected(featureId)) {
    // Make it the only selected feature
    selectedFeatureIds.forEach(id => this.doRender(id));
    this.setSelected(featureId);
    this.updateUIClasses({ mouse: Constants.cursors.MOVE });
  }
  // Enter direct select mode
  return this.changeMode(Constants.modes.DIRECT_SELECT, {
    featureId
  });
};