# go-strava-daemon

## About this repository

This repository contains the daemon service to fetch Strava user data.

## Required parametes

This daemon requires some `ENV` variables to be set. Below is an example:

```sh
export CONFIG_POSTGRESENDPOINT="localhost"
export CONFIG_STRAVACLIENTID="MY_STRAVA_ID"
export CONFIG_STRAVA_CLIENTSECRET="MY_STRAVA_SECRET"
export CONFIG_STRAVACALLBACKURL="https://redirect-to-me.com"
```
