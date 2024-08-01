import logging
import os
import random
import threading
import time

from selenium.webdriver.common.by import By
from selenium.webdriver.remote.webelement import WebElement
from pywinauto import WindowSpecification, ElementNotFoundError
from selenium.webdriver.edge.webdriver import WebDriver

from base import LogLevel, load_config, get_executable_directory, get_logger
from .devtools_api import listen_tab_console_logs, get_navigation_entries
from .pywinauto_api import close_edge_window, kill_edge_process, get_edge_window, start_edge_debug
from .selenium_api import connect_local_edge, findBy, waitBy, execScript, saveCookies, loadCookies, delCookies, \
    switchToTab, findListBy, waitListBy, wait_page_loaded


def _random_sleep(min_seconds: float, max_seconds: float):
    time.sleep(random.uniform(min_seconds, max_seconds))


class EdgeRobot:
    log: logging.Logger
    debugPort: int = 9222
    rootPath: str
    execPath: str
    tabIndex: int = 0
    window: WindowSpecification | None = None
    driver: WebDriver | None = None
    consoleLogEvent: threading.Event | None
    ready: bool = False

    def __init__(self, log_level: LogLevel = LogLevel.INFO):
        load_config('robot')
        self.rootPath = get_executable_directory()
        self.log = get_logger(name="robot", level=log_level)
        exec_path = os.environ.get("EDGE_EXEC_PATH")
        if exec_path:
            self.execPath = exec_path
        port_str = os.environ.get("EDGE_DEBUG_PORT")
        if port_str and int(port_str) > 0:
            self.debugPort = int(port_str)
        self.consoleLogEvent = None

    @staticmethod
    def randomSleep(seconds: float):
        _random_sleep(min_seconds=seconds, max_seconds=seconds * 1.5)

    @staticmethod
    def sleep(seconds: float):
        time.sleep(seconds)

    def start(self) -> bool:
        close_edge_window()
        kill_edge_process()
        start_edge_debug(self.execPath, self.debugPort)
        time.sleep(3)
        self.window = get_edge_window()
        if self.window is None:
            return False
        self.driver = connect_local_edge(self.debugPort)
        self.ready = True
        return self.ready

    def connectDriver(self):
        self.driver = connect_local_edge(self.debugPort)

    def close(self):
        self.ready = False
        if self.consoleLogEvent and not self.consoleLogEvent.is_set():
            self.consoleLogEvent.set()
        if self.isWindowExist():
            if self.driver:
                # self.driver.close()
                self.driver.quit()
                self.driver = None
            if self.window:
                self.window = None
        else:
            self.driver = None
            self.window = None
        kill_edge_process()

    def isWindowExist(self):
        if self.window is None:
            return False
        try:
            return self.window.exists()
        except Exception as e:
            self.log.error(msg='edge window not exist', exc_info=e)
            return False

    def find(self, selector: str) -> WebElement | None:
        return findBy(self.driver, self.log, By.CSS_SELECTOR, selector)

    def finds(self, selector: str) -> list[WebElement]:
        return findListBy(self.driver, self.log, By.CSS_SELECTOR, selector)

    def wait(self, selector: str, seconds: float = 5) -> WebElement | None:
        return waitBy(self.driver, self.log, By.CSS_SELECTOR, selector, seconds)

    def waits(self, selector: str, seconds: float = 5) -> list[WebElement]:
        return waitListBy(self.driver, self.log, By.CSS_SELECTOR, selector, seconds)

    def exec(self, script: str, *args) -> tuple[bool, any]:
        return execScript(self.driver, self.log, script, args)

    def saveCookie(self):
        saveCookies(self.driver, self.rootPath)

    def loadCookie(self):
        loadCookies(self.driver, self.rootPath)

    def delCookie(self):
        delCookies(self.driver, self.rootPath)

    def clearAll(self):
        delCookies(self.driver, self.rootPath)
        script = """
        window.localStorage.clear();
        indexedDB.databases().then((dbs) => {
            for (let db of dbs) {
                indexedDB.deleteDatabase(db.name);
            }
        });        
        """
        execScript(self.driver, self.log, script)

    def switchNext(self):
        self.tabIndex = self.tabIndex + 1
        switchToTab(self.driver, self.tabIndex)

    def switchBack(self):
        if self.tabIndex > 0:
            self.tabIndex = self.tabIndex - 1
            switchToTab(self.driver, self.tabIndex, True)

    def startListenConsoleLog(self):
        self.consoleLogEvent = listen_tab_console_logs(self.debugPort)

    def stopListenConsoleLog(self):
        self.consoleLogEvent.set()
        self.consoleLogEvent = None

    def getTabInfos(self):
        return get_navigation_entries(self.debugPort)

    def waitPageLoaded(self, max_seconds: int = 10) -> bool:
        return wait_page_loaded(self.driver, max_seconds)