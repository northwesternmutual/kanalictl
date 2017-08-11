#!/bin/bash

set -e

if [[ $TRAVIS_TAG =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  gox
  echo $TRAVIS_TAG >> latest.txt
  aws s3 mv latest.txt $S3_BASE_PATH/latest.txt
  aws s3 mv kanalictl_darwin_386 $S3_BASE_PATH/$TRAVIS_TAG/darwin/386/kanalictl
  aws s3 mv kanalictl_darwin_amd64 $S3_BASE_PATH/$TRAVIS_TAG/darwin/amd64/kanalictl
  aws s3 mv kanalictl_linux_386 $S3_BASE_PATH/$TRAVIS_TAG/linux/386/kanalictl
  aws s3 mv kanalictl_linux_amd64 $S3_BASE_PATH/$TRAVIS_TAG/linux/amd64/kanalictl
  aws s3 mv kanalictl_linux_arm $S3_BASE_PATH/$TRAVIS_TAG/linux/arm/kanalictl
  aws s3 mv kanalictl_freebsd_386 $S3_BASE_PATH/$TRAVIS_TAG/freebsd/386/kanalictl
  aws s3 mv kanalictl_freebsd_amd64 $S3_BASE_PATH/$TRAVIS_TAG/freebsd/amd64/kanalictl
  aws s3 mv kanalictl_freebsd_arm $S3_BASE_PATH/$TRAVIS_TAG/freebsd/arm/kanalictl
  aws s3 mv kanalictl_openbsd_386 $S3_BASE_PATH/$TRAVIS_TAG/openbsd/386/kanalictl
  aws s3 mv kanalictl_openbsd_amd64 $S3_BASE_PATH/$TRAVIS_TAG/openbsd/amd64/kanalictl
  aws s3 mv kanalictl_windows_386.exe $S3_BASE_PATH/$TRAVIS_TAG/windows/386/kanalictl.exe
  aws s3 mv kanalictl_windows_amd64.exe $S3_BASE_PATH/$TRAVIS_TAG/windows/amd64/kanalictl.exe
  aws s3 mv kanalictl_netbsd_386 $S3_BASE_PATH/$TRAVIS_TAG/netbsd/386/kanalictl
  aws s3 mv kanalictl_netbsd_amd64 $S3_BASE_PATH/$TRAVIS_TAG/netbsd/amd64/kanalictl
  aws s3 mv kanalictl_netbsd_arm $S3_BASE_PATH/$TRAVIS_TAG/netbsd/arm/kanalictl
fi

exit 0