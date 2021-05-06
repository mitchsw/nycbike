import React from 'react';
import Map from "./Map";
import Vitals from "./Vitals";
import {withStyles} from '@material-ui/core';

import AppBar from '@material-ui/core/AppBar';
import Button from '@material-ui/core/Button';
import CssBaseline from '@material-ui/core/CssBaseline';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Link from '@material-ui/core/Link';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';

// Don't forget to import the CSS
import '@mapbox/mapbox-gl-draw/dist/mapbox-gl-draw.css';

const styles = theme => ({
  '@global': {
    ul: {
      margin: 0,
      padding: 0,
      listStyle: 'none',
    },
  },
  appBar: {
    borderBottom: `1px solid ${theme.palette.divider}`,
    background: `#0098e4`,
    color: `white`,
  },
  toolbar: {
    flexWrap: 'wrap',
  },
  toolbarTitle: {
    flexGrow: 1,
    fontFamily: `'Overpass', sans-serif`,
    textTransform: 'uppercase',
    fontWeight: 700,
  },
  link: {
    margin: theme.spacing(1, 1.5),
  },

  textContainer: {
    position: 'absolute',
    bottom: 20,
    left: 20,
    paddingLeft: 16,
    paddingRight: 16,
    opacity: 0.9,
  }
});

class App extends React.Component {

  constructor(props) {
    super(props);
    this.initialState = {
      src: {center: [0,0], radiusInKm: 0},
      dst: {center: [0,0], radiusInKm: 0},
      journeys: null,
      vitals: null,
    };
    this.state = { ...this.initialState };
  }

  componentDidMount() {
    fetch(`${process.env.REACT_APP_BACKEND_URL}vitals`)
        .then(res => res.json())
        .then((data) => {
          this.setState({ vitals: data })
        })
        .catch(console.log)
  }

  journeyQueryUrl(src, dst) {
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
      <Box height="100vh" display="flex" flexDirection="column">
        <CssBaseline />
        <AppBar position="static" color="default" elevation={4} className={classes.appBar}>
          <Toolbar className={classes.toolbar}>
            <Typography variant="h6" color="inherit" noWrap className={classes.toolbarTitle}>
              Citibike Journeys
            </Typography>
            {this.state.vitals ? <Vitals  {...this.state.vitals}/> : null}
          </Toolbar>
        </AppBar>

        <Box flex={1} overflow="auto">
          <Map onFeaturesUpdated={features => this.onFeaturesUpdated(features)} />
        </Box>
        <Paper elevation={4} className={classes.textContainer}>
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
        </Paper>
    </Box>
    );
  }
}

export default withStyles(styles)(App);
