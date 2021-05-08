import * as React from "react";
import {
  Chart,
  ArgumentAxis,
  ValueAxis,
  AreaSeries,
  Legend
} from "@devexpress/dx-react-chart-material-ui";
import { withStyles } from "@material-ui/core/styles";
import {
  ArgumentScale,
  Animation
} from "@devexpress/dx-react-chart";
import { curveCatmullRom, area } from "d3-shape";
import { scalePoint } from "d3-scale";

const egress = [3842,2687,1726,1181,657,473,1042,2354,4614,7446,10484,13146,14226,14685,14920,15436,15691,14612,13562,11545,8420,6352,4659,3068,1435,780,471,302,315,1527,4755,11256,20653,16210,10938,10180,12590,12960,12599,13362,15221,18452,18689,14621,10521,7007,4987,3193,1596,720,454,252,336,1820,6480,13493,24259,18794,12023,11157,13101,13110,13274,14171,15722,19827,20666,15723,11339,8040,5609,3599,2004,1043,534,313,362,2076,6439,14614,27678,21564,13650,13298,15511,15276,15220,15788,17230,20644,20809,16415,11758,8349,6000,4042,1993,988,560,410,413,1764,6057,13412,23644,19216,12160,11585,14144,14339,14242,14434,16252,19532,19178,14710,10276,7396,5952,4558,2447,1217,668,385,425,1350,4704,11312,20014,17549,12194,11836,13783,14008,14609,14767,16474,18778,18301,14179,9905,7308,5851,5326,3610,2403,1484,1056,591,496,1366,3346,6830,10405,13945,16246,17387,17320,17318,17560,16599,15310,14019,12334,9158,7123,6031,5127];
const ingress = [3440,2442,1418,821,509,342,816,1570,3735,6111,9255,11924,13780,14638,14289,15082,15946,15449,14362,12037,9115,6814,4894,2985,1570,788,461,314,375,1229,3806,7969,12090,10480,8856,9388,12203,12608,12755,13417,15515,24912,27400,19165,12802,8414,6070,3312,1686,841,514,340,439,1577,4873,9143,13760,11692,9403,10198,12966,13126,12793,14435,16451,26747,30222,20957,13670,9356,6779,3987,2290,1064,650,385,505,1843,5092,9915,15858,13473,11207,12336,15124,15415,14902,15976,18007,28584,30967,21655,14336,9759,7004,4340,2223,1033,633,343,392,1409,4437,9493,13918,11937,9632,10423,14006,14053,13864,14857,16790,25934,28263,19750,12719,8633,6973,4784,2785,1318,699,378,379,1157,3736,7877,12248,11093,9638,10957,13568,14213,14511,15852,18443,25347,23880,17428,11561,8157,6254,5107,3357,2111,1347,833,468,359,938,2621,5735,8833,12078,14753,16909,17024,16912,17053,16976,15639,14717,13038,9826,7613,6028,4815];

const makeData = () => {
  let arr = [];
  for (var i = 0; i < 24*7; i++) {
    let h = (i+24)%(24*7);  // Start on Monday
    arr.push({ hour: i, egress: egress[h], ingress: -ingress[h] });
  }
  return arr;
}

const legendStyles = () => ({
  root: {
    display: 'flex',
    margin: 'auto',
    flexDirection: 'row',
  },
});
const legendRootBase = ({ classes, ...restProps }) => (
  <Legend.Root {...restProps} className={classes.root} />
);
const Root = withStyles(legendStyles, { name: 'LegendRoot' })(legendRootBase);
const legendLabelStyles = () => ({
  label: {
    whiteSpace: 'nowrap',
  },
});
const legendLabelBase = ({ classes, ...restProps }) => (
  <Legend.Label className={classes.label} {...restProps} />
);
const Label = withStyles(legendLabelStyles, { name: 'LegendLabel' })(legendLabelBase);

const chartStyles = () => ({
  chart: {
    paddingRight: "20px",
  }
});

const Area = (props) => (
  <AreaSeries.Path
    {...props}
    path={area()
      .x(({ arg }) => arg)
      .y1(({ val }) => val)
      .y0(({ startVal }) => startVal)
      .curve(curveCatmullRom)}
  />
);

function weekScale() {
  let s = scalePoint();
  // A tick every 6 hours.
  s.ticks = function() {
    return Array(4*7).fill().map((_, i) => 6*i)
  }
  // A tick legend every 12 hours.
  s.tickFormat = function() {
    return (hour) => {
      let day = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
      if (hour%24 === 0) {
        return day[hour/24];
      }
      if (hour%12 === 0) {
        return (hour%24) + ":00";
      }
      return "";
    };
  };
  return s;
}

const valueTickFormat = () => (d) => Math.abs(d).toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");

class JourneyChart extends React.PureComponent {
  constructor(props) {
    super(props);

    this.state = {
      data: makeData()
    };
  }

  render() {
    const { data: chartData } = this.state;
    const { classes } = this.props;
    return (
      <Chart data={chartData} className={classes.chart}>
        <ArgumentScale factory={weekScale} />
        <ArgumentAxis/>
        <ValueAxis tickFormat={valueTickFormat} />

        <AreaSeries
          name="Ingress"
          valueField="ingress"
          argumentField="hour"
          seriesComponent={Area}
        />
        <AreaSeries
          name="Egress"
          valueField="egress"
          argumentField="hour"
          seriesComponent={Area}
        />
        <Animation />
        <Legend position="bottom" rootComponent={Root} labelComponent={Label} />
      </Chart>
    );
  }
}

export default withStyles(chartStyles, { name: "JourneyChart" })(JourneyChart);
