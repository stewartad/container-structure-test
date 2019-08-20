# Container Structure Tests for Singularity

## Overview

This repository was forked from [GoogleContainerTools/container-structure-test](https://github.com/GoogleContainerTools/container-structure-test) to extend the framework to work with Singularity.

### Files that were modified

* package/drivers/driver.go
  * Modified the constant variables so the framework can find the singularity driver

### Files Added

* package/drivers/singularity.go
  * Used go library [singolang](https://github.com/stewartad/singolang) for Singularity to implement the already defined driver interface

## Results

  Tests can be run on Singularity containers using the Google framework.

### Installation

  Download the binary file from the [latest release](https://github.com/stewartad/container-structure-test/releases/tag/v1.8.2) and add it to your PATH, or run it directly. An example command is:

  ```bash
  container-structure-test test --driver singularity --image lolcow_latest.sif --config config.yaml
  ```

### Examples and Documentation

  For more examples and documentation, check out the [singularity-container-test](https://github.com/stewartad/singularity-container-test) and the original [GoogleContainerTools/container-structure-test](https://github.com/GoogleContainerTools/container-structure-test) repositories.
