import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Overview

:::warning[Important]

`display` is known as a daemon service. You do not need to manage this service yourself. For more information, please visit the page about [`roverd`'s daemon services](https://ase.vu.nl/docs/tutorials/knowing-more/roverd-daemons).

:::

## Purpose 

The `display` service shows the Rover identity and resource utilization on the OLED ssd1306 display that is mounted on the front of each Rover.

![The ssd1306 display mounted on the front of Rover 1](https://github.com/user-attachments/assets/038036e2-43dd-47f0-a785-140abefc9612)

## Requirements

- The ssd1306 display should be connected to the Rover

## Inputs

As defined in the [*service.yaml*](https://github.com/VU-ASE/display/blob/main/service.yaml), this service does not depend on any other services.

## Outputs

As defined in the [*service.yaml*](https://github.com/VU-ASE/display/blob/main/service.yaml), this service does not expose any write streams.
