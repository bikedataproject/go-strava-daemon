# go-strava-daemon

## About this repository

This repository contains the daemon service to fetch Strava user data.

## Required parametes

This daemon requires some `ENV` variables to be set. Below is an example:

```sh
export CONFIG_POSTGRESHOST="localhost"
export CONFIG_POSTGRESPORT="5432"
export CONFIG_POSTGRESPASSWORD="MyPostgresPassword"
export CONFIG_POSTGRESUSER="postgres"
export CONFIG_POSTGRESDB="bikedata"
export CONFIG_POSTGRESREQUIRESSL="require"
export CONFIG_STRAVACLIENTID="MY_STRAVA_ID"
export CONFIG_STRAVACLIENTSECRET="MY_STRAVA_SECRET"
export CONFIG_CALLBACKURL="https://redirect-to-me.com"
export CONFIG_STRAVAWEBHOOKURL="https://www.strava.com/api/v3/push_subscriptions"
```

## How to run

```sh
docker build -t go-strava-daemon:tag .

docker run -d -p 5000:5000 \
-e CONFIG_POSTGRESHOST="localhost" \
-e CONFIG_POSTGRESPORT="5432" \
-e CONFIG_POSTGRESPASSWORD="MyPostgresPassword" \
-e CONFIG_POSTGRESUSER="postgres" \
-e CONFIG_POSTGRESDB="bikedata" \
-e CONFIG_POSTGRESREQUIRESSL="require" \
-e CONFIG_STRAVACLIENTID="MY_STRAVA_ID" \
-e CONFIG_STRAVACLIENTSECRET="MY_STRAVA_SECRET" \
-e CONFIG_CALLBACKURL="https://redirect-to-me.com" \
-e CONFIG_STRAVAWEBHOOKURL="https://www.strava.com/api/v3/push_subscriptions" \
go-strava-daemon:tag
```
