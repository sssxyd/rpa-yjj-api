import asyncio
import os
import sys
import time

from fastapi import FastAPI
from starlette.responses import FileResponse
from starlette.staticfiles import StaticFiles

from edgewin import EdgeRobot
from yjj import get_drugstore_info, get_pharmacist_info


def edge_visit_url(edge: EdgeRobot, url: str, wait_seconds: float):
    edge.window.set_focus()
    edge.window.type_keys('^l')
    edge.sleep(1)
    edge.window.type_keys('{DELETE}')
    edge.sleep(1)
    edge.window.type_keys(url + "{ENTER}")
    edge.sleep(wait_seconds)


def init_robot(edge: EdgeRobot):
    edge.start()
    if not edge.ready:
        edge.log.error("start edge failed!")
        sys.exit(1)
    main_page_url = os.environ.get("MAIN_PAGE_URL")
    edge_visit_url(edge, main_page_url, 2)
    edge.log.info(f"Edge with page {main_page_url} is ready")


def prepare_robot(edge: EdgeRobot):
    if not edge.isWindowExist():
        robot.log.error("edge window not exist")
        edge.close()
        edge.start()
        return

    if edge.window.is_minimized():
        robot.log.info("edge window minimized, restore edge")
        edge.window.restore()

    main_page_url = os.environ.get("MAIN_PAGE_URL")

    if edge.driver is None:
        edge.connectDriver()

    global last_session_time
    now = time.time()
    if edge.driver.current_url == main_page_url:
        timeout = int(os.environ.get("MAIN_PAGE_TIMEOUT"))
        # cookie/session 有效
        if now - last_session_time < timeout:
            return

    # 重新访问主页
    last_session_time = now
    edge_visit_url(edge, main_page_url, 2)


app = FastAPI()
app.mount("/static", StaticFiles(directory="static"), name="static")
robot = EdgeRobot()
last_session_time = time.time()
robot_lock = asyncio.Lock()


@app.on_event("startup")
async def startup_event():
    robot.log.info("init edge robot...")
    init_robot(robot)
    robot.log.info("init complete")


@app.on_event("shutdown")
async def shutdown_event():
    robot.log.info("dispose edge robot")
    robot.close()
    robot.log.info("edge robot closed")


@app.get("/favicon.ico")
async def favicon_ico():
    return FileResponse('static/favicon.ico')


@app.get("/")
@app.get("/usage")
@app.get("/help")
@app.get("/manual")
async def usage():
    return FileResponse('static/usage.txt')


@app.get("/drugstore")
async def drugstore(code: str = '', name: str = ''):
    start_time = time.time()
    async with robot_lock:
        if not code or not name:
            return {'code': 400, 'msg': f'code[{code}] and name[{name}] invalid', 'result': '{}'}
        exec_time = time.time()
        prepare_robot(robot)
        data = get_drugstore_info(edge=robot, code=code, name=name)
        robot.window.minimize()
        robot.log.info(f"/drugstore?code={code}&name={name}, response: {data}, "
                       f"cost: {round(time.time() - start_time, 3)}, wait: {round(exec_time - start_time, 3)}")
        return {'code': 1, 'msg': '', 'result': data}


@app.get("/pharmacist")
async def pharmacist(code: str = '', name: str = ''):
    start_time = time.time()
    async with robot_lock:
        if not code or not name:
            return {'code': 400, 'msg': f'code[{code}] and name[{name}] invalid', 'result': '{}'}
        exec_time = time.time()
        prepare_robot(robot)
        data = get_pharmacist_info(edge=robot, code=code, name=name)
        robot.window.minimize()
        robot.log.info(f"/pharmacists?code={code}&name={name}, response: {data}, "
                       f"cost: {round(time.time() - start_time, 3)}, wait: {round(exec_time - start_time, 3)}")
        return {'code': 1, 'msg': '', 'result': data}
