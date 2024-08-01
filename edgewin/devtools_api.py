import json
import threading

import requests
from websocket import create_connection


def listen_tab_console_logs(debug_port: int) -> threading.Event:
    response = requests.get(f'http://127.0.0.1:{debug_port}/json')
    devtools_url = json.loads(response.text)[0]['webSocketDebuggerUrl']
    stop_event = threading.Event()
    # 创建并启动线程
    log_thread = threading.Thread(target=_print_console_logs, args=(devtools_url, stop_event))
    log_thread.start()
    return stop_event


def _print_console_logs(devtools_url: str, stop_event: threading.Event):
    print(f"devtools_url: {devtools_url}")
    ws = create_connection(devtools_url)
    # 启用日志记录
    # 启用日志记录和其他事件
    ws.send(json.dumps({'id': 1, 'method': 'Log.enable'}))
    ws.send(json.dumps({'id': 2, 'method': 'Runtime.enable'}))
    ws.send(json.dumps({'id': 3, 'method': 'Network.enable'}))
    # ws.send(json.dumps({'id': 4, 'method': 'Page.enable'}))
    # ws.send(json.dumps({'id': 5, 'method': 'Debugger.enable'}))
    # ws.send(json.dumps({'id': 6, 'method': 'DOM.enable'}))
    # 获取控制台日志
    while not stop_event.is_set():
        try:
            log_entry = json.loads(ws.recv())
            print(json.dumps(log_entry))
            if log_entry.get('method') == 'Log.entryAdded':
                log = log_entry['params']['entry']
                print(log["text"])
            elif log_entry.get('method') == 'Runtime.consoleAPICalled':
                for arg in log_entry['params']['args']:
                    print(arg['value'])
        except Exception as e:
            if stop_event.is_set():
                break
            print(f'Error: {e}')
    ws.close()


def get_navigation_entries(debug_port: int):
    response = requests.post(
        f'http://127.0.0.1:{debug_port}/json',
        json={'method': 'Page.listNavigationEntries'}
    )
    return response.json()