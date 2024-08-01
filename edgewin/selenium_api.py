import base64
import json
import os
import re
import time
from logging import Logger
from urllib.parse import urlparse

from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.edge.options import Options
from selenium.webdriver.edge.service import Service
from selenium.webdriver.edge.webdriver import WebDriver
from selenium.webdriver.remote.webelement import WebElement
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.wait import WebDriverWait
from webdriver_manager.microsoft import EdgeChromiumDriverManager


def wait_page_loaded(driver: WebDriver, max_seconds: int) -> bool:
    count = 0
    while driver.execute_script("return document.readyState;") != "complete":
        if count >= max_seconds:
            return False
        count = count + 1
        time.sleep(1)
    return True


def connect_local_edge(port: int) -> WebDriver:
    options = Options()
    options.debugger_address = f"127.0.0.1:{port}"
    service = Service(EdgeChromiumDriverManager().install())
    browser = webdriver.Edge(service=service, options=options)
    browser.execute_script("Object.defineProperty(navigator, 'webdriver', {get: () => undefined})")
    return browser


def waitBy(driver: WebDriver, log: Logger, by: By, value: str, seconds: float = 5) -> WebElement | None:
    try:
        element = WebDriverWait(driver, seconds).until(
            EC.presence_of_element_located((by, value))
        )
        if element is None:
            return None
        return element if not isinstance(element, list) else element[0]
    except Exception as e:
        msg = f"wait dom by {by}: {value} failed"
        log.exception(msg=msg, exc_info=e)
        return None


def waitListBy(driver: WebDriver, log: Logger, by: By, value: str, seconds: float = 5) -> list[WebElement]:
    try:
        element = WebDriverWait(driver, seconds).until(
            EC.presence_of_element_located((by, value))
        )
        if element is None:
            return []
        return element if isinstance(element, list) else [element]
    except Exception as e:
        msg = f"wait dom by {by}: {value} failed"
        log.exception(msg=msg, exc_info=e)
        return []


def findBy(driver: WebDriver, log: Logger, by: By, value: str) -> WebElement | None:
    try:
        return driver.find_element(by=by, value=value)
    except Exception as e:
        msg = f"find dom by {by}: {value} failed"
        log.exception(msg=msg, exc_info=e)
        return None


def findListBy(driver: WebDriver, log: Logger, by: By, value: str) -> list[WebElement]:
    try:
        return driver.find_elements(by=by, value=value)
    except Exception as e:
        msg = f"find dom by {by}: {value} failed"
        log.exception(msg=msg, exc_info=e)
        return []


def execScript(driver, log, script, *args) -> tuple[bool, any]:
    # 执行脚本
    real_script = f"""
    try{{
        {script}
    }}
    catch(error){{
        const message = `Error: ${{error.message}}\nStack: ${{error.stack}}`
        return message
    }}
    """
    succeed = True
    result = driver.execute_script(real_script, *args)
    if result and isinstance(result, str):
        if "Error:" in result:
            succeed = False
            log.error(f"ExecScript Failed, {result}")
            # errs = loadPerformanceLogs(driver, log)
            # for err in errs:
            #     log.error(err)

    # # 获取浏览器控制台日志
    # pattern = re.compile(r'CONSOLE\(\d+\)|stun_port\.cc|socket_manager\.cc')
    # for entry in driver.get_log('browser'):
    #     print(json.dumps(entry))
    return succeed, result


def loadPerformanceLogs(driver: WebDriver, log: Logger) -> list[str]:
    results = []
    logs = driver.get_log('performance')
    for log in logs:
        log_json = json.loads(log['message'])
        message = log_json['message']
        if 'method' not in message or 'params' not in message:
            continue
        if 'Network.responseReceived' not in message['method']:
            continue
        params = message['params']
        if 'requestId' not in params or 'response' not in params:
            continue
        requestId = str(params['requestId'])
        response = params['response']
        if 'status' not in response or 'headers' not in response or 'url' not in response:
            continue
        status = int(response['status'])
        statusText = str(response['statusText']) if 'statusText' in response else ''
        if status == 200:
            continue
        url = str(response['url'])
        headers = response['headers']
        if 'content-type' not in headers or 'application/json' not in headers['content-type']:
            continue
        body = ""
        try:
            response_body = driver.execute_cdp_cmd('Network.getResponseBody',
                                                   {'requestId': requestId})
            body = response_body['body']
            if 'base64Encoded' in response_body and response_body['base64Encoded']:
                body = base64.b64decode(body).decode('utf-8')
        except Exception as e:
            log.exception(msg="Error getting response body", exc_info=e)
        results.append(f"{url} failed {status}:{statusText}, response: {body}")
    return results


def saveCookies(driver: WebDriver, root_path: str):
    parsed_url = urlparse(driver.current_url)
    domain = parsed_url.hostname
    cookie_path = os.path.join(root_path, "data", f"cookie_{domain}.json")
    cookies = driver.get_cookies()
    with open(cookie_path, 'w') as file:
        json.dump(cookies, file)


def delCookies(driver: WebDriver, root_path: str):
    parsed_url = urlparse(driver.current_url)
    domain = parsed_url.hostname
    cookie_path = os.path.join(root_path, "data", f"cookie_{domain}.json")
    if os.path.exists(cookie_path):
        os.remove(cookie_path)


def loadCookies(driver: WebDriver, root_path: str):
    parsed_url = urlparse(driver.current_url)
    domain = parsed_url.hostname
    cookie_path = os.path.join(root_path, "data", f"cookie_{domain}.json")
    if os.path.exists(cookie_path):
        # 从文件加载 cookie
        with open(cookie_path, 'r') as file:
            cookies = json.load(file)
        for cookie in cookies:
            driver.add_cookie(cookie)


def switchToTab(driver: WebDriver, index: int, close_current_tab: bool = False):
    all_tabs = driver.window_handles
    if close_current_tab:
        driver.close()
    if 0 <= index < len(all_tabs):
        driver.switch_to.window(all_tabs[index])

