# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2021-09-30
### Changed
- Renamed virtual machine flags with shorter names, removing 'vm-' prefix
- Installer images location/structure
- Unified preseed.cfg file
- Unified pxe configurations
- Prebuilt ldlinux.c32
- Python libs netifaces and numpy are now uploaded from the local package
- Upload module use less network bandwidth discriminating between installer versions

### Added
- Debian installers are now downloadable from online sources using a default mirror (ftp.debian.nl) using --use-default-mirror flag
- Debian installers are now downloadable from online sources using a custor mirror given from command line using --use-custom-mirror flag (refer to --help for further details)
- --vi-preseed allow users to edit package's preseed file (local changes will be overwritten during updates)


## [1.2.10] - 2021-08-14
### Changed
- Updating debian10-64 images in order to match kernel version with repos
### Added
- debian11-64 images

## [1.2.10] - 2021-07-14
### Changed
- Changing default keymap to 'en' in preseed.cfg

## [1.2.9] - 2021-06-14
### Changed
- Updating debian10-64 images in order to match kernel version with repos

## [1.2.8] - 2021-06-11
### Changed
- Setting LVM as default volume manager for both debian10-64 and debian9-64

## [1.2.7] - 2020-10-20
### Changed
- Updating debian10-64 images in order to match kernel version with repos

## [1.2.6] - 2020-10-20
### Changed
- changing repo name and shortening sleep time

## [1.2.5] - 2020-10-20
### Changed
- Updating debian10-64 images in order to match kernel version with repos

## [1.2.1] - 2020-01-29
### Added
- verbose mode

## [1.1.0] - 2019-10-19
### Added
- ESXi disk space check before create .vmdk disk

## [1.0.0] - 2019-06-20
### Added
- CHANGELOG.md

### Changed
- go-ansible lib has been changed from "github.com/apenella/go-ansible" to "github.com/lucabodd/go-ansible" (version hop, not backward compatible)
