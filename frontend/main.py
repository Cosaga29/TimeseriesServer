import time
import json

from websocket import create_connection

ws = create_connection("ws://localhost:9000/wsclient")

time.sleep(1)

ws.send_text(json.dumps({"type": 1, "data": {"msg": "some_data"}}))

# Receive as string
result = ws.recv()
result = json.loads(result)
print(type(result))

print(result)

time.sleep(5)

ws.close()