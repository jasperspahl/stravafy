# Stravafy

Stravafy is a webservice that aims to sync your Spotify listen history to Strava

## Setup

Stravafy can be easily deployed using docker compose. It's only one container
that reads all configuration from a file.

To deploy just clone the repo and run `docker compose build` and `docker compose up -d`.

This will create a config file where you have to adjust the values (strava, spotify secrets ...)

```yaml

database:
    source: file:stravafy.db?mode=rwc
listen:
    host: 0.0.0.0
    port: 8888
spotify:
    clientid: <spotify_client_id>
    clientsecret: <spotify_client_secret>
    updateinterval: 30
    showdialog: true
strava:
    clientid: <strava_client_id>
    clientsecret: <strava_client_secret>
    approvalprompt: force
    statestring: stravafy
    webhookhost: https://your.domain.com
```

It's very important to set the strava webhookhost to the exact address that the
service can be accessed from otherwise the service doesn't get the callbacks
from Strava to upload the data.

## Use without self-hosting

I also host an instance of this project.

If you want to use it, feel free to do that but be aware of the fact I can see
you whole spotify listening history form this point onward since I have no
solution yet to disconnect / delete accounts from a user side.

[Hosted instance](https://stravafy.servebeer.com)
