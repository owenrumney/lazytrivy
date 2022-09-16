[![Go Report Card](https://goreportcard.com/badge/github.com/owenrumney/lazytrivy)](https://goreportcard.com/report/github.com/owenrumney/lazytrivy)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/owenrumney/lazytrivy/blob/master/LICENSE)
[![Github Release](https://img.shields.io/github/release/owenrumney/lazytrivy.svg)](https://github.com/owenrumney/lazytrivy/releases)
[![GitHub All Releases](https://img.shields.io/github/downloads/owenrumney/lazytrivy/total)](https://github.com/owenrumney/lazytrivy/releases)

# lazytrivy

lazytrivy is a wrapper for [Trivy](https://github.com/aquasecurity/trivy) that allows you to run Trivy without
remembering the command arguments.

The idea was very heavily inspired by the superb tools from [Jesse Duffield](https://github.com/jesseduffield) (
lazydocker, lazynpm, lazygit)

![Scan All Images](./.github/images/scan_all.png)

## Features

- Image Scanning
  - [Scan all images on your system](#scanning-all-local-images)
  - [Scan a single image](#scanning-a-specific-image)
  - [Scan a remote image](#scanning-a-remote-image)
- AWS Scanning
  - [Scan your cloud account](#scanning-an-aws-account)
- File System Scanning
  - [Scan a filesystem for vulnerabilities and misconfigurations](#scanning-a-filesystem)


## Installation

### Prerequisites

In order for lazytrivy to be cross-platform, it uses the Trivy docker image. This means that you will need to have Docker running on your machine for lazytrivy to work.

> :warning: Docker Desktop has degraded functionality. Locally built images not in a repository can not be scanned :warning:

#### Install with Go

The quickest way to install if you have `Go` installed is to get the latest with `go install`

```bash
go install github.com/owenrumney/lazytrivy@latest
```

#### Download from Releases

Alternatively, you can get the latest releases from [GitHub](https://github.com/owenrumney/lazytrivy)

### Config

A config file can be added to `~/.config/lazytrivy/config.yml` to set default options.

```yaml
aws:
    accountno: "1234567890981"
    region: eu-west-1
vulnerability:
    ignoreunfixed: false
cachedirectory: /home/owen/.cache/trivy
debug: false
```

By setting `debug` to true, additional logs will be generated in `/tmp/lazytrivy.log`

## Usage

`lazytrivy` is super easy to use, just run it with the following command:

```bash
lazytrivy
```

### Starting in a specific mode

You can start `lazytrivy` in a specific mode using `aws`, `images` or `filesystem`:

For example, to scan a specific filesystem folder, you could run:

```bash
lazytrivy fs /home/owen/code/github/owenrumney/example
```

This will start in that mode.


### Scanning all local images

Pressing `a` will scan all the images that are shown in the left hand pane. On completion, you will be shown a
summary of any vulnerabilities found.

You can then scan individual images to get more details

![Scanning all images](./.github/images/scan_all_images.gif)

### Scanning a specific image

Select an image from the left hand pane and press `s` to scan it. Use the left and right arrow keys to switch between
views and up down arrow keys to select an image.

Press `s` to scan the currently selected image.

![Scanning an image](./.github/images/scan_individual_images.gif)

### Scanning a remote image

To scan an image that is not already locally on the machine, you can use the `r` key to scan a remote image.

![Scanning a remote image](./.github/images/scan_remote_image.gif)

### Scanning an AWS Account

To scan an AWS account, you can use the `w` key to switch to AWS mode, from there you can use the `s` key to scan, it will detect any valid credentials it can.

![Scanning an AWS account](./.github/images/scan_aws_account.gif)

By pressing `r` you can switch region in results you already have.

### Scanning a filesystem

To scan a filessystem, you can use the `w` key to switch to Filesystem mode, from there you will get all the vulnerabilities, misconfigurations and secrets from the current working directory

![Scanning a filesystem](./.github/images/scan_filesystem.gif)
