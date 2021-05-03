import "@mapbox/mapbox-gl-draw/dist/mapbox-gl-draw.css";
import {DirectMode, SimpleSelectMode,} from "mapbox-gl-draw-circle";
const Constants = require('@mapbox/mapbox-gl-draw/src/constants');
const CommonSelectors = require('@mapbox/mapbox-gl-draw/src/lib/common_selectors');
const turfHelpers = require('@turf/helpers');
const length = require('@turf/length').default;
const along = require('@turf/along').default;


const MAX_RADIUS = 2.5;  // km

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

SimpleSelectModeOverride.clickOnFeature = function(state, e) {
  // Stop everything
  //MapboxDraw.doubleClickZoom.disable(this);
  this.stopExtendedInteractions(state);

  const featureId = e.featureTarget.properties.id;
  const selectedFeatureIds = this.getSelectedIds();

  // Clear any hover state.
  if (this.hoveredFeatureId) {
    this.getFeature(this.hoveredFeatureId).setProperty("hover", false);
    this.hoveredFeatureId = null;
  }

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

SimpleSelectModeOverride.hoveredFeatureId = null;

SimpleSelectModeOverride.onMouseMove = function(state, e) {
  // On mousemove that is not a drag, stop extended interactions.
  // This is useful if you drag off the canvas, release the button,
  // then move the mouse back over the canvas --- we don't allow the
  // interaction to continue then, but we do let it continue if you held
  // the mouse button that whole time
  this.stopExtendedInteractions(state);

  if (CommonSelectors.isFeature(e)) {
    if (!this.hoveredFeatureId) {
      this.hoveredFeatureId = e.featureTarget.properties.id;
      this.getFeature(this.hoveredFeatureId).setProperty("hover", true);
      this.doRender(this.hoveredFeatureId);
      console.log("Hover->true: " + this.hoveredFeatureId)
    }
  } else if (this.hoveredFeatureId) {
    this.getFeature(this.hoveredFeatureId).setProperty("hover", false);
    this.doRender(this.hoveredFeatureId);
    console.log("Hover->false: " + this.hoveredFeatureId)
    this.hoveredFeatureId = null;
  }

  // Skip render
  return true;
};


SimpleSelectModeOverride.onMouseOutBase = SimpleSelectMode.onMouseOut;
SimpleSelectModeOverride.onMouseOut = function(state, e) {
  if (this.hoveredFeatureId) {
    this.getFeature(this.hoveredFeatureId).setProperty("hover", false);
    this.doRender(this.hoveredFeatureId);
    this.hoveredFeatureId = null;
  }
  return this.onMouseOutBase(state, e);
};