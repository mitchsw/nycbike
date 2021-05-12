import React from 'react';
import Paper from '@material-ui/core/Paper';
import {withStyles} from '@material-ui/core/styles';
import TimelineIcon from '@material-ui/icons/Timeline';
import Box from '@material-ui/core/Box';
import DnsIcon from '@material-ui/icons/Dns';

const styles = {
  vitals: {
    padding: '0px 16px',
    display: 'flex',
    marginLeft: '8px',
    background: '#095780',
    color: 'white',
    fontFamily: `'Overpass', sans-serif`,
  },
};

function commaDelimitedNum(v) {
  return v.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

function Vitals(props) {
  const {classes} = props;
  const tripCount = commaDelimitedNum(props.TripCount);
  const stationCount = commaDelimitedNum(props.StationCount);
  const edgeCount = commaDelimitedNum(props.EdgeCount);
  return (
    <React.Fragment>
      <Paper elevation={4} className={classes.vitals}>
        <Box pt="5px" mr="12px"><TimelineIcon /></Box>
        <Box pt="7px">
          Searching <strong>{ tripCount }</strong> trips
          on graph of <strong>{ stationCount }</strong> stations
          with <strong>{ edgeCount }</strong> edges.
        </Box>
      </Paper>
      <Paper elevation={4} className={classes.vitals}>
        <Box pt="5px" mr="12px"><DnsIcon /></Box>
        <Box pt="7px">
          Redis Memory Usage: <strong>{ props.MemoryUsageHuman }</strong>
        </Box>
      </Paper>
    </React.Fragment>
  );
}

export default withStyles(styles)(Vitals);
