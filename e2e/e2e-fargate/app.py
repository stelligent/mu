import os
import requests
import pymysql.cursors 
import socket
from flask import Flask
app = Flask(__name__)

@app.route("/")
def hello():
  connection = pymysql.connect(host=os.environ['DB_HOST'],
                               port=int(os.environ['DB_PORT']),
                               user=os.environ['DB_USERNAME'],
                               password=os.environ['DB_PASSWORD'],
                               db=os.environ['DB_NAME'],
                               charset='utf8mb4',
                               cursorclass=pymysql.cursors.DictCursor)
  try:
      with connection.cursor() as cursor:
          cursor.execute("SELECT * FROM user")
          assert cursor.rowcount > 0
  finally:
      connection.close()

  res = socket.gethostbyname_ex("e2e-fargate.%s" % os.environ['_SERVICE_DISCOVERY_NAME'])[2]
  assert len(res) > 0

  return "ok"
