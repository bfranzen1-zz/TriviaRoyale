import os
mysql_pw = os.environ["pw"]
addr = os.environ["ADDR"]
db = os.environ["DBADDR"]
usr = os.environ["usr"]
mqHOST = os.environ["mqHOST"]
mqPORT = os.environ["mqPORT"]
rUSER = os.environ["rUSER"] 
rPW = os.environ["rPW"]
rmQueue = os.environ["rmQueue"]
DATABASE_CONNECTION_URI = "mysql+pymysql://" + usr + ":" + mysql_pw + "@" + db + "/api"