import React from 'react';
import JourneyChart from './JourneyChart';
import {withStyles} from '@material-ui/core';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import FlashChange from '@avinlab/react-flash-change';

const styles = (theme) => ({
  textContainer: {
    position: 'absolute',
    bottom: 15,
    left: 15,
    right: 15,
    paddingBottom: 12,
    opacity: 0.95,
  },

  resultBox: {
    textAlign: 'center',
    fontFamily: `'Overpass', sans-serif`,
    padding: 5,
  },
});

const FlashVal = ({value}) =>
  <FlashChange
    value={value}
    outerElementType="strong"
    style={{backgroundColor: 'transparent', transition: 'background-color 100ms'}}
    flashStyle={{backgroundColor: '#a3d2fb', transition: 'background-color 200ms'}}
    flashDuration={200}
    compare={(prevProps, nextProps) => {
      return nextProps.value !== prevProps.value;
    }}
  >
  &nbsp;{value}&nbsp;
  </FlashChange>;

class JourneyPaper extends React.PureComponent {
  render() {
    const {classes} = this.props;
    let tripSum =
      this.props.journeys.Egress.reduce((a, b) => a + b, 0) +
      this.props.journeys.Ingress.reduce((a, b) => a + b, 0);
    tripSum = tripSum.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    return (
      <Paper elevation={4} className={classes.textContainer}>
        <Box className={classes.resultBox}>
          RedisGraph found
          <FlashVal value={tripSum} />
          trips in
          <FlashVal value={this.props.journeys.RunTimeMs.toFixed(0)} />ms.
        </Box>
        <JourneyChart egress={this.props.journeys.Egress} ingress={this.props.journeys.Ingress} />
      </Paper>
    );
  }
}

export default withStyles(styles)(JourneyPaper);
