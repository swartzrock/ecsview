
![screenshot](docs/logo.jpg)

EcsView (ecsview) is a terminal UI for browsing Amazon Elastic Container Service (ECS) clusters. You can use it to view cluster usage, services, tasks, and individual container instances (ec2's). Uses the excellent [tview](https://github.com/rivo/tview/) library to build and manage the UI.


![screenshot](docs/tasksview.jpg)

## Installation

Run `make install` to install ecsview in your Go binaries directory.

## Usage

Run `AWS_PROFILE=<profile> ecsview` with a configured AWS profile to view your account's ECS clusters in detail.


