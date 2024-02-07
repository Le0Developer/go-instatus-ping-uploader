# instatus-ping-uploader

Simple script that pings 1.1.1.1 every second and pushes the max latency + packet loss to a instatus metric.

Currently running this on my Raspberry Pi to monitor the connection quality of my home network.

## Usage

Just build & run.

## Environment variables

- `API_TOKEN`: Your instatus API token
- `PAGE_ID`: Your page id
- `PING_METRIC_ID`: The metric ID of the ping/latency instatus component
- `LOSS_METRIC_ID`: The metric ID of the packet loss instatus component

