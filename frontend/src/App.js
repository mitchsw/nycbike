import React from 'react';
import './App.css';
import Map from "./Map";
import {withStyles} from '@material-ui/core';

// Don't forget to import the CSS
import '@mapbox/mapbox-gl-draw/dist/mapbox-gl-draw.css';

const styles = theme => ({
  map: {
    zIndex: -1,
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%'
  },
  textContainer: {
    position: 'absolute',
    bottom: 32,
    left: 16,
    background: '#eee',
    paddingLeft: 16,
    paddingRight: 16,
  }
});

class App extends React.Component {

  constructor(props) {
    super(props);
    this.initialState = {
      src: {center: [0,0], radiusInKm: 0},
      dst: {center: [0,0], radiusInKm: 0},
      journeys: null,
    };
    this.state = { ...this.initialState };
  }

  journeyQueryUrl(src, dst) {
    console.log(process.env);
    return `${process.env.REACT_APP_BACKEND_URL}journey_query?src_lat=${src.center[1]}&src_long=${src.center[0]}&src_radius=${src.radiusInKm}&dst_lat=${dst.center[1]}&dst_long=${dst.center[0]}&dst_radius=${dst.radiusInKm}`
  }

  onFeaturesUpdated(features) {
    this.setState({src: features.src, dst: features.dst, journeys: null})
    fetch(this.journeyQueryUrl(features.src, features.dst))
        .then(res => res.json())
        .then((data) => {
          this.setState({ journeys: data })
        })
        .catch(console.log)
  }

  render() {
    const { classes } = this.props;
    let journeys_string;
    if (this.state.journeys != null) {
      let egress_sum = this.state.journeys.Egress.reduce((a, b) => a + b, 0).toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",")
      let ingress_sum = this.state.journeys.Ingress.reduce((a, b) => a + b, 0).toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",")
      journeys_string = `Egress: ${egress_sum} Ingress: ${ingress_sum} Runtime: ${this.state.journeys.RunTimeMs}ms`
    } else {
      journeys_string = "-"
    }
    return (
      <div>
        <div className={classes.map}>
          <Map onFeaturesUpdated={features => this.onFeaturesUpdated(features)} />
        </div>
        <div className={classes.textContainer}>
          <p>
            <strong>Src:</strong>
            {` Center: [${this.state.src.center.join(', ')}]`}
            {` Radius: ${this.state.src.radiusInKm.toFixed(4)} kms`}
          </p>
          <p>
            <strong>Dst:</strong>
            {` Center: [${this.state.dst.center.join(', ')}]`}
            {` Radius: ${this.state.dst.radiusInKm.toFixed(4)} kms`}
          </p>
          <p>
            { journeys_string }
          </p>
        </div>
      </div>
    );
  }
}

export default withStyles(styles)(App);
