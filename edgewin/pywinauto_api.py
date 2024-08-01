import subprocess

import psutil
from pywinauto import Desktop, Application, WindowSpecification
from pywinauto.findwindows import find_windows


def close_edge_window():
    # 获取所有顶级窗口
    all_windows = Desktop(backend="uia").windows()
    for window in all_windows:
        # 检查窗口标题是否包含 "Edge"
        if "Edge" in window.window_text():
            try:
                # 尝试关闭窗口
                window.close()
                print(f"Closed window: {window.window_text()}")
            except Exception as e:
                print(f"Failed to close window: {window.window_text()}. Error: {e}")


def kill_edge_process():
    for proc in psutil.process_iter(['pid', 'name']):
        # 检查进程名是否包含 "msedge"
        if "msedge" in proc.info['name']:
            try:
                proc.kill()
                print(f"Killed process: {proc.info['name']} (PID: {proc.info['pid']})")
            except Exception as e:
                print(f"Failed to kill process: {proc.info['name']} (PID: {proc.info['pid']}). Error: {e}")


def start_edge_debug(exec_path: str, port: int = 9222) -> None:
    # Edge 可执行文件路径和参数分开
    arguments = f"--remote-debugging-port={port} --remote-allow-origins=http://127.0.0.1:{port}"
    # 启动 Edge 浏览器，并指定远程调试端口
    Application(backend="uia").start(f'"{exec_path}" {arguments}')


def get_edge_window() -> WindowSpecification | None:
    all_windows = find_windows(visible_only=True)
    for win_hwnd in all_windows:
        try:
            win_app = Application(backend="uia").connect(handle=win_hwnd)
            win_window = win_app.window(handle=win_hwnd)
            window_title = win_window.window_text()
            if "Microsoft" in window_title and "Edge" in window_title:
                return win_window
        except Exception as e:
            print(f"Error processing window with handle {win_hwnd}: {e}")
    return None
