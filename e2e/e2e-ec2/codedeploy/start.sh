#!/bin/bash

set -e

cd /code
set -a
[[ -f /etc/environment ]] && source /etc/environment
set +a
FLASK_APP=app.py /usr/local/bin/flask run --host=0.0.0.0 &