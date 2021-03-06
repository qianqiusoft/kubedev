import React, { useState, Suspense } from 'react';
import { Redirect } from '@reach/router';
import { Global, css } from '@emotion/core';
import { ThemeProvider } from 'emotion-theming';
import useAxios from '@use-hooks/axios';

import Sidebar from '../components/Sidebar';
import Header from './Header';
import AppContainer from '../components/AppContainer';
import CustomRouter from '../components/CustomRouter';
import {
  primary,
  primaryDark,
  darkLight,
  darkDark,
  primaryLight,
  fontColor,
  fontColorWhite,
  fontColorDark,
  cardBackgroundLight,
  cardBackgroundDark,
  backgroundLight,
  tableBorderLight,
  tableBorderDark
} from '../util/colors';
import RouterLoading from '../components/RouterLoading';

const Home = React.lazy(() => import('./Home'));
const Logs = React.lazy(() => import('./Logs'));
const PodInfo = React.lazy(() => import('./PodInfo'));
const Deployments = React.lazy(() => import('./Deployments'));
const Services = React.lazy(() => import('./Services'));
const ServiceInfo = React.lazy(() => import('./ServiceInfo'));
const DeploymentInfo = React.lazy(() => import('./DeploymentInfo'));
const Jobs = React.lazy(() => import('./Jobs'));
const CronJobs = React.lazy(() => import('./CronJobs'));
const JobInfo = React.lazy(() => import('./JobInfo'));
const CronJobInfo = React.lazy(() => import('./CronJobInfo'));
const Pods = React.lazy(() => import('./Pods'));

const lightTheme = {
  name: 'light',
  background: backgroundLight,
  header: primaryDark,
  sidebarBackground: primaryLight,
  sidebarFontColor: fontColor,
  containerFont: fontColorDark,
  cardBackground: cardBackgroundLight,
  tableBorder: tableBorderLight,
  controllerBackground: backgroundLight,
  controllerColor: fontColor,
  controllerBorder: fontColor
};

const darkTheme = {
  name: 'dark',
  background: darkDark,
  header: darkLight,
  sidebarBackground: darkDark,
  sidebarFontColor: fontColorWhite,
  containerFont: fontColorWhite,
  cardBackground: cardBackgroundDark,
  tableBorder: tableBorderDark,
  controllerBackground: darkLight,
  controllerColor: fontColorWhite,
  controllerBorder: tableBorderLight
};

function App() {
  const [theme, setTheme] = useState(darkTheme);
  const [links, setLinks] = useState([]);
  const { response } = useAxios({
    url: `${process.env.API}`,
    method: 'GET',
    trigger: null
  });

  const { data } = response || {};
  let namespaces = [];
  if (data) {
    namespaces = data.items.map(({ metadata }) => metadata.name);
  }

  const handleSidebarChange = newLink => {
    if (
      !links.find(
        link => link.type === newLink.type && link.name === newLink.name
      )
    )
      setLinks(links => links.concat(newLink));
  };

  const handleThemeChange = () => {
    setTheme(theme => (theme.name === 'light' ? darkTheme : lightTheme));
  };

  return (
    <ThemeProvider theme={theme}>
      <Global
        styles={css`
          * {
            padding: 0 0;
            margin: 0 0;
            box-sizing: border-box;
            font-family: 'Roboto Mono', monospace;
          }
        `}
      />
      <Header />
      <AppContainer>
        <Sidebar
          namespaces={namespaces}
          links={links}
          onThemeChange={handleThemeChange}
        />

        <Suspense fallback={<RouterLoading />}>
          <CustomRouter>
            <Redirect from="/" to="/default/pods" noThrow />
            {/* <Home path="/:namespace" /> */}
            <Pods path="/:namespace/pods" />
            <PodInfo path="/:namespace/pods/:name/info" />
            <Services path="/:namespace/services" />
            <ServiceInfo path="/:namespace/services/:name/info" />
            <Deployments path="/:namespace/deployments" />
            <DeploymentInfo path="/:namespace/deployments/:name/info" />
            <Jobs path="/:namespace/jobs" />
            <JobInfo path="/:namespace/jobs/:name/info" />
            <CronJobs path="/:namespace/cron-jobs" />
            <CronJobInfo path="/:namespace/cron-jobs/:name/info" />
            <Logs
              path="/:namespace/pods/:name/logs/container/:selectedContainer"
              onLogInit={handleSidebarChange}
            />
          </CustomRouter>
        </Suspense>
      </AppContainer>
    </ThemeProvider>
  );
}

export default App;
