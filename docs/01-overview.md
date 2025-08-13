import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Overview

## Purpose 

The `display` service shows the Rover identity and resource utilization on the OLED ssd1306 display that is mounted on the front of each Rover.

:::tip

We do not recommend installing this service manually, as `roverd` will automatically check for updates and install it as a daemon service every time the Rover boots.

:::

## Installation

To install this service, the latest release of [`roverctl`](https://ase.vu.nl/docs/framework/Software/rover/roverctl/installation) should be installed for your system and your Rover should be powered on.

<Tabs groupId="installation-method">
<TabItem value="roverctl" label="Using roverctl" default>

1. Install the service from your terminal
```bash
# Replace ROVER_NUMBER with your the number label on your Rover (e.g. 7)
roverctl service install -r <ROVER_NUMBER> https://github.com/VU-ASE/display/releases/latest/download/display.zip 
```

</TabItem>
<TabItem value="roverctl-web" label="Using roverctl-web">

1. Open `roverctl-web` for your Rover
```bash
# Replace ROVER_NUMBER with your the number label on your Rover (e.g. 7)
roverctl -r <ROVER_NUMBER>
```
2. Click on "install a service" button on the bottom left, and click "install from URL"
3. Enter the URL of the latest release:
```
https://github.com/VU-ASE/display/releases/latest/download/display.zip 
```

</TabItem>
</Tabs>

Follow [this tutorial](https://ase.vu.nl/docs/tutorials/write-a-service/upload) to understand how to use an ASE service. You can find more useful `roverctl` commands [here](/docs/framework/Software/rover/roverctl/usage)

## Requirements

- The ssd1306 display should be connected to the Rover

## Inputs

As defined in the [*service.yaml*](https://github.com/VU-ASE/display/blob/main/service.yaml), this service does not depend on any other services.

## Outputs

As defined in the [*service.yaml*](https://github.com/VU-ASE/display/blob/main/service.yaml), this service does not expose any write streams.