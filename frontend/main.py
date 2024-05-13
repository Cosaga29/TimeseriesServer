import time
import json
from datetime import datetime, timezone

from websocket import create_connection

ws = create_connection("ws://localhost:9000/wsclient")

req = 2

# Request collections
# First byte = message type, rest = JSON request
payload = req.to_bytes(length=4, byteorder="big", signed=False) + bytes(
    json.dumps({
        "database": "test_db_q",
        "collection": "test_collection",
        "startIsoDate": "2024-05-13T01:17:43.436Z",
        "endIsoDate": "2024-05-13T01:17:43.436Z"
    }), encoding="utf-8"
)
ws.send_bytes(payload)

try:
    while True:
        print(json.loads(ws.recv()))
except KeyboardInterrupt:
    pass  

print(json.loads(ws.recv()))

ws.close()
exit() 

ws.close()