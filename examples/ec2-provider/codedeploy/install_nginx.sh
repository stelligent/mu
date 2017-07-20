#!/bin/bash

set -e

yum install -y nginx
chkconfig --level 345 nginx on
