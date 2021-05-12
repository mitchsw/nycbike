import * as React from 'react';
import {
  Chart,
  ArgumentAxis,
  ValueAxis,
  AreaSeries,
} from '@devexpress/dx-react-chart-material-ui';
import {withStyles} from '@material-ui/core/styles';
import {
  ArgumentScale,
  Animation,
} from '@devexpress/dx-react-chart';
import {curveCatmullRom, area} from 'd3-shape';
import {scalePoint} from 'd3-scale';
import Box from '@material-ui/core/Box';

const makeChartData = (egress, ingress) => {
  const arr = [];
  for (let i = 0; i < 24*7; i++) {
    const h = (i+24)%(24*7); // Start on Monday
    arr.push({hour: i, egress: egress[h], ingress: -ingress[h]});
  }
  return arr;
};

const chartStyles = () => ({
  chart: {
    padding: 0,
    paddingRight: '12px',
  },
  label: {
    width: '20px',
    transform: 'rotate(-90deg)',
    float: 'left',
    marginTop: '100px',
    fontFamily: `'Overpass', sans-serif`,
    textTransform: 'uppercase',
    fontSize: 'x-small',
  },
});

const Area = (props) => (
  <AreaSeries.Path
    {...props}
    path={area()
        .x(({arg}) => arg)
        .y1(({val}) => val)
        .y0(({startVal}) => startVal)
        .curve(curveCatmullRom)}
  />
);

function weekScale() {
  const s = scalePoint();
  // A tick every 6 hours.
  s.ticks = function() {
    return Array(4*7).fill().map((_, i) => 6*i);
  };
  // A tick legend every 12 hours.
  s.tickFormat = function() {
    return (hour) => {
      const day = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
      if (hour%24 === 0) {
        return day[hour/24];
      }
      if (hour%12 === 0) {
        return (hour%24) + ':00';
      }
      return '';
    };
  };
  return s;
}

const valueTickFormat = () =>
  (d) => Math.abs(d).toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');

class JourneyChart extends React.PureComponent {
  constructor(props) {
    super(props);

    this.state = {
      data: makeChartData(this.props.egress, this.props.ingress),
    };
  }

  static getDerivedStateFromProps(props, state) {
    return {
      data: makeChartData(props.egress, props.ingress),
    };
  }

  render() {
    const {data: chartData} = this.state;
    const {classes} = this.props;
    return (
      <React.Fragment>
        <Box className={classes.label}>Trips&nbsp;Count</Box>
        <Chart data={chartData} className={classes.chart} height={200}>
          <ArgumentScale factory={weekScale} />
          <ArgumentAxis showGrid={true} />
          <ValueAxis
            showGrid={false}
            showTicks={true}
            tickFormat={valueTickFormat}
          />

          <AreaSeries
            name="Egress"
            valueField="egress"
            argumentField="hour"
            seriesComponent={Area}
          />
          <AreaSeries
            name="Ingress"
            valueField="ingress"
            argumentField="hour"
            seriesComponent={Area}
          />

          <Animation />
        </Chart>
      </React.Fragment>
    );
  }
}

export default withStyles(chartStyles)(JourneyChart);
