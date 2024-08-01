import os
import sys

from uvicorn import run
from dotenv import load_dotenv
import server
import comtypes.stream

if sys.platform == "win32":
    os.system('chcp 65001')

if __name__ == "__main__":
    load_dotenv()
    port_str = os.environ.get('HTTP_SERVER_PORT')
    port = int(port_str) if port_str else 80
    run(app="server:app", host="0.0.0.0", port=port, workers=1)
    input("...exit with enter")
