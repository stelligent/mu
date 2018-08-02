import os
import requests
import pymysql.cursors 
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

  return "ok"
