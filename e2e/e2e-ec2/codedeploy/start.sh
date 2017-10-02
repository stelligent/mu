#!/bin/bash

set -e

set -a
source /etc/environment
set +a

cd /code
FLASK_APP=app.py /usr/local/bin/flask run --host=0.0.0.0 2>&1 < /dev/null |logger -t flask > /dev/null 2> /dev/null &

exit 0