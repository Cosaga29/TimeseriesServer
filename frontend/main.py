import time
import json

from websocket import create_connection

ws = create_connection("ws://localhost:9000/wsclient")

# Request collections
ws.send_text(json.dumps({"type": 1}))
print(json.loads(ws.recv()))

# Select from database
ws.send_text(json.dumps(
        {"type": 2, "data": {"database": "test_db", "collection": "test_collection", "query": {
            "somedata": "test"
        }}}
    )
)
try:
    while True:
        print(json.loads(ws.recv()))
except KeyboardInterrupt:
    pass     

ws.close()