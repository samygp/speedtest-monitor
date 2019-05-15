# Speedtest monitor and alerts
Code for speedtest forked from https://github.com/showwin/speedtest-monitor
This repository contains code to query speedtest.net from your local network,
and display the ping, download and upload speeds. Optionally,
it can run in server mode, polling speedtest.net repeatedly on a defined interval, with an API accessible through HTTP, displaying the results from
the latest speed test performed. It can also send alerts using a Slack webhook.

## Running locally
Execute `./runLocally.sh` file in the `docker` directory

## Compile code and create local docker image
Execute `./build.sh` file in the `docker` directory

## Configuration
There are a few parameters that can be configured to specify how the program
will execute. Define these parameters in `./docker/configuration.json` file:
- **appName**: Name to identify this service.
- **serverMode**: If set to `true`, it will test periodically for an interval defined in the `interval` parameter.
- **inverval**: Number of seconds between each poll to speedtest.net when in `serverMode`.
- **loggingLevel**: If set to `"debug"`, the program will print additional logs meant for debugging purposes.
- **slackEndpoint**: Webhook that will be used to send messages and alerts using Slack. This parameter can be an empty string if Slack messaging is not required or available.
- **downloadThreshold**: Threshold to use in Slack alerts to notify when download speed is too low.
- **uploadThreshold**: Threshold to use in Slack alerts to notify when upload speed is too low.

## Using local API
The API is configured to run in port `12321` by Default. If you wish to override
this, you can change the mapping in `./docker/docker-compose.yml` file.

### Verify it's running
This will print the `appName` parameter if the service is working
`http://localhost:12321/`

### Get latest results
This will print the latest results in JSON format
`http://localhost:12321/getLatestResult`
```
{
  "DownSpeed": 90.59589004313125,
  "DownSpeedStr": "Download: 90.60 Mbit/s\n",
  "Ping": "Latency: 74.777ms",
  "UpSpeed": 50.683460024,
  "UpSpeedStr": "Upload: 50.68 Mbit/s\n",
  "LastQuery": "2019-04-17 17:57:38.7436458 +0000 UTC m=+84.914258601"
}
```

### Query on demand
To trigger an on-demand speed test, use
`http://localhost:12321/testSpeedNow`