import React from 'react';
import {
  makeStyles
} from '@material-ui/core';

const useStyles = makeStyles((theme) => ({
  root: {
    backgroundColor: theme.palette.background.dark,
    minHeight: '100%',
    paddingBottom: theme.spacing(3),
    paddingTop: theme.spacing(3)
  },
  mediaCard: {
    position: 'relative',
    border: 'none',
    height: '100vh',
    width: '100%',
  },
  content: {
    display: 'flex',
    flexDirection: 'column',
    alignContent: 'stretch',
    height: `calc(100vh - ${64}px)`, // theme.spacing(8)
    width: '100%',
  },
}));

const Dashboard = () => {
  const classes = useStyles();
  const dashboardUrl = 'http://35.222.5.135:5000/public/dashboards/tz8jqkffPRve7aMwP5ErDjMvWhHwpauXw6Vdi2el?org_slug=default';
  return (

    <div
      className={classes.content}
    >
      <iframe title="Dashboard" className={classes.mediaCard} src={dashboardUrl} />

    </div>
  );
};

export default Dashboard;
