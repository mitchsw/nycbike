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
      features: {
        src: {center: [0,0], radiusInKm: 0},
        dst: {center: [0,0], radiusInKm: 0},
      }
    };
    this.state = { ...this.initialState };
  }

  render() {
    const { classes } = this.props;
    const src = this.state.features.src;
    const dst = this.state.features.dst;
    return (
      <div>
        <div className={classes.map}>
          <Map
            onFeaturesUpdated={
              (features) => this.setState({features: features})
            }>
          </Map>
        </div>
        <div className={classes.textContainer}>
          <p>
            <strong>Src:</strong>
            {` Center: [${src.center.join(', ')}]`}
            {` Radius: ${src.radiusInKm.toFixed(4)} kms`}
          </p>
          <p>
            <strong>Dst:</strong>
            {` Center: [${dst.center.join(', ')}]`}
            {` Radius: ${dst.radiusInKm.toFixed(4)} kms`}
          </p>
        </div>
      </div>
    );
  }
}

export default withStyles(styles)(App);
