import React from 'react';
import clsx from 'clsx';
import PropTypes from 'prop-types';
import PerfectScrollbar from 'react-perfect-scrollbar';
import {
  Box,
  Card,
  CardHeader,
  CardContent,
  Divider,
  CircularProgress,
  makeStyles, useTheme, fade
} from '@material-ui/core';
import GenericMoreButton from 'src/components/GenericMoreButton';
import Chart from './Chart';
import { Line, Bar, Pie } from 'react-chartjs-2';

import cubejs from '@cubejs-client/core';
import { QueryRenderer } from '@cubejs-client/react';

const useStyles = makeStyles(() => ({
  root: {},
  chart: {
    height: '100%'
  }
}));

const COLORS_SERIES = ['#FF6492', '#141446', '#7A77FF'];
const commonOptions = {
  maintainAspectRatio: false,
};


const cubejsApi = cubejs(
  '***REMOVED-SECRET***',
  { apiUrl: 'http://localhost:4000/cubejs-api/v1' }
);

const renderChart = ({ resultSet, error, pivotConfig, classes, className, theme, ...rest}) => {
  if (error) {
    return <div>{error.toString()}</div>;
  }

  if (!resultSet) {
    return <CircularProgress />;
  }

  // const data = {
  //   labels: resultSet.categories().map((c) => c.category),
  //   datasets: resultSet.series().map((s, index) => (
  //     s.series.map((r) => r.value)
  //   )),
  // };
  const labels = resultSet.categories().map((c) => c.category)
  const dataProp = resultSet.series().map((s, index) => (
      s.series.map((r) => r.value)
  ))

  const data = (canvas) => {
    const ctx = canvas.getContext('2d');
    const gradient = ctx.createLinearGradient(0, 0, 0, 400);

    gradient.addColorStop(0, fade(theme.palette.secondary.main, 0.2));
    gradient.addColorStop(0.9, 'rgba(255,255,255,0)');
    gradient.addColorStop(1, 'rgba(255,255,255,0)');

    return {
      datasets: resultSet.series().map((s, index) => ({
        label: s.title,
        data: s.series.map((r) => r.value),
        borderColor: theme.palette.secondary.main,
        pointBorderColor: theme.palette.background.default,
        pointBorderWidth: 3,
        pointRadius: 6,
        pointBackgroundColor: theme.palette.secondary.main,
        fill: false,
      })),
      labels
    };
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    animation: false,
    legend: {
      display: false
    },
    layout: {
      padding: 0
    },
    scales: {
      xAxes: [
        {
          gridLines: {
            display: false,
            drawBorder: false
          },
          ticks: {
            padding: 20,
            fontColor: theme.palette.text.secondary
          }
        }
      ],
      yAxes: [
        {
          gridLines: {
            borderDash: [2],
            borderDashOffset: [2],
            color: theme.palette.divider,
            drawBorder: false,
            zeroLineBorderDash: [2],
            zeroLineBorderDashOffset: [2],
            zeroLineColor: theme.palette.divider
          },
          ticks: {
            padding: 20,
            fontColor: theme.palette.text.secondary,
            beginAtZero: true,
            min: 0,
            callback: (value) => (value > 0 ? `${value}K` : value)
          }
        }
      ]
    },
    tooltips: {
      enabled: true,
      mode: 'index',
      intersect: false,
      caretSize: 10,
      yPadding: 20,
      xPadding: 20,
      borderWidth: 1,
      borderColor: theme.palette.divider,
      backgroundColor: theme.palette.background.default,
      titleFontColor: theme.palette.text.primary,
      bodyFontColor: theme.palette.text.secondary,
      footerFontColor: theme.palette.text.secondary,
      callbacks: {
        title: () => {},
        label: (tooltipItem) => {
          let label = `Income: ${tooltipItem.yLabel}`;

          if (tooltipItem.yLabel > 0) {
            label += 'K';
          }

          return label;
        }
      }
    }
  };
  return (
    <Card
      className={clsx(classes.root, className)}
      {...rest}
    >
      <CardHeader
        action={<GenericMoreButton />}
        title="Performance Over Time"
      />
      <Divider />
      <CardContent>
        <PerfectScrollbar>
          <Box
            height={375}
            minWidth={500}
          >
            <Chart
              className={classes.chart}
              data={data}
              options={options}
            />
          </Box>
        </PerfectScrollbar>
      </CardContent>
    </Card>
  );

};

const PerformanceOverTime = ({ className, ...rest }) => {
  const classes = useStyles();
  const theme = useTheme();

  return (
    <QueryRenderer
      query={{
        "measures": [
          "Postlots.soldprice"
        ],
        "timeDimensions": [],
        "order": {
          "Postlots.soldprice": "desc"
        },
        "dimensions": [
          "Postlots.rowid"
        ]
      }}
      cubejsApi={cubejsApi}
      resetResultSetOnChange={false}
      render={(props) => renderChart({
        ...props,
        pivotConfig: {
          "x": [
            "Postlots.rowid"
          ],
          "y": [
            "measures"
          ],
          "fillMissingDates": true,
          "joinDateRange": false
        },
        classes,
        className,
        theme,
        ...rest
      })}
    />
  );
};

PerformanceOverTime.propTypes = {
  className: PropTypes.string
};

export default PerformanceOverTime;
