import React from 'react';
import Map from "./Map";
import JourneyPaper from "./JourneyPaper";
import Vitals from "./Vitals";
import {withStyles} from '@material-ui/core';
import AppBar from '@material-ui/core/AppBar';
import IconButton from '@material-ui/core/IconButton';
import CssBaseline from '@material-ui/core/CssBaseline';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Box from '@material-ui/core/Box';
import '@mapbox/mapbox-gl-draw/dist/mapbox-gl-draw.css';
import GitHubIcon from '@material-ui/icons/GitHub';

const styles = theme => ({
  '@global': {
    ul: {
      margin: 0,
      padding: 0,
      listStyle: 'none',
    },
  },
  appBar: {
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
    fetch(`${process.env.REACT_APP_BACKEND_URL}stations`)
        .then(res => res.json())
        .then((data) => {
          this.setState({ stations: data })
        })
        .catch(console.log)
  }

  journeyQueryUrl(src, dst) {
    return `${process.env.REACT_APP_BACKEND_URL}journey_query?src_lat=${src.center[1]}&src_long=${src.center[0]}&src_radius=${src.radiusInKm}&dst_lat=${dst.center[1]}&dst_long=${dst.center[0]}&dst_radius=${dst.radiusInKm}`
  }

  onFeaturesUpdated(features) {
    fetch(this.journeyQueryUrl(features.src, features.dst))
        .then(res => res.json())
        .then((data) => {
          this.setState({ journeys: data })
        })
        .catch(console.log)
  }

  render() {
    const { classes } = this.props;
    return (
      <Box height="100vh" display="flex" flexDirection="column">
        <CssBaseline />
        <AppBar position="static" color="default" elevation={4} className={classes.appBar}>
          <Toolbar variant="dense" className={classes.toolbar}>
            <Typography variant="h6" color="inherit" noWrap className={classes.toolbarTitle}>
              Citibike Journeys
            </Typography>
            {this.state.vitals ? <Vitals  {...this.state.vitals}/> : null}
            <IconButton href="https://github.com/mitchsw/citibike-journeys" target="_blank" variant="outlined" style={{ color: 'white', marginLeft: '10px' }}>
              <GitHubIcon />
            </IconButton>
          </Toolbar>
        </AppBar>

        <Box flex={1} overflow="auto">
          <Map onFeaturesUpdated={features => this.onFeaturesUpdated(features)} stations={this.state.stations} />
        </Box>
        { this.state.journeys && 
          <JourneyPaper journeys={this.state.journeys} />
        }
      </Box>
    );
  }
}

export default withStyles(styles)(App);
