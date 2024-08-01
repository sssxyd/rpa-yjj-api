import os

from selenium.webdriver import Keys
from selenium.webdriver.common.by import By

from edgewin import EdgeRobot


def _get_result_count(edge: EdgeRobot) -> int:
    rows = edge.finds("tr.el-table__row")
    return len(rows)


def _get_dict_info(edge: EdgeRobot) -> dict[str, str]:
    data = dict()
    rows = edge.finds('tr.el-table__row')
    for row in rows:
        keyDiv = row.find_element(by=By.CSS_SELECTOR, value="td:nth-child(1) > div > div")
        valDiv = row.find_element(by=By.CSS_SELECTOR, value="td:nth-child(2) > div > div > div")
        data[keyDiv.text] = valDiv.text
    return data


def _advance_search(edge: EdgeRobot, code: str, name: str, wait_seconds: float):
    advanceSearchLink = edge.find("div.gaojisousuo > a")
    if not advanceSearchLink:
        return
    advanceSearchLink.click()
    edge.sleep(0.5)

    nameInput = edge.find("div.gaoji-box > form > div.el-form-item:nth-child(3) > div > div > input")
    if nameInput is None:
        edge.log.error("advance search name input not found!")
        return
    nameInput.send_keys(name)

    codeInput = edge.find("div.gaoji-box > form > div.el-form-item:nth-child(4) > div > div > input")
    if codeInput is None:
        edge.log.error("advance search code input not found!")
        return
    codeInput.send_keys(code)

    confirmBtn = edge.find("div.gaoji-box > div > button:nth-child(2)")
    if confirmBtn is None:
        return
    confirmBtn.click()
    edge.sleep(wait_seconds * 0.5)


def get_drugstore_info(edge: EdgeRobot, code: str, name: str) -> dict[str, str]:
    wait_seconds = float(os.environ.get("PAGE_WAIT_SECONDS"))
    selector = ".el-col.el-col-8:nth-child(7) .el-link--inner"
    item_link = edge.find(selector)
    if item_link is None:
        edge.log.error("link of 药品经营企业 not found!")
        return {}
    item_link.click()

    selector = "div.search-input > input.el-input__inner"
    keyInput = edge.find(selector)
    if keyInput is None:
        edge.log.error("search keyword input not found!")
        return {}
    keyInput.send_keys(Keys.CLEAR)
    keyInput.send_keys(code)
    keyInput.send_keys(Keys.ENTER)
    edge.sleep(wait_seconds)
    edge.switchNext()

    count = _get_result_count(edge)
    if count == 0:
        edge.log.info(f"search drugstore by code: {code} response empty!")
        edge.switchBack()
        return {}
    if count > 1:
        edge.log.info(f"search drugstore by code: {code} response multiple results, advance search by name: {name}")
        _advance_search(edge, code, name, wait_seconds)
        count = _get_result_count(edge)
        if count == 0:
            edge.log.warning(f"search drugstore by code: {code}, name: {name} response empty!")
            edge.switchBack()
            return {}
        elif count > 1:
            edge.log.warning(f"search drugstore by code: {code}, name: {name} response multiple results!")
            edge.switchBack()
            return {}

    selector = "tr.el-table__row > td:nth-child(4) > div > button"
    button = edge.find(selector)
    if button is None:
        edge.log.error("button of 详情 not found")
        edge.switchBack()
        return {}
    button.click()
    edge.sleep(wait_seconds)

    edge.switchNext()
    print(edge.driver.current_url)
    data = _get_dict_info(edge)
    edge.switchBack()
    edge.switchBack()
    print(edge.driver.current_url)
    return data


def get_pharmacist_info(edge: EdgeRobot, code: str, name: str) -> dict[str, str]:
    wait_seconds = int(os.environ.get("PAGE_WAIT_SECONDS"))
    selector = ".el-col.el-col-8:nth-child(12) .el-link--inner"
    item_link = edge.find(selector)
    if item_link is None:
        edge.log.error("link of 注册药师 not found!")
        return {}
    item_link.click()

    selector = "div.search-input > input.el-input__inner"
    keyInput = edge.find(selector)
    if keyInput is None:
        edge.log.error("search keyword input not found!")
        return {}

    keyInput.send_keys(Keys.CLEAR)
    keyInput.send_keys(code)
    keyInput.send_keys(Keys.ENTER)
    edge.sleep(wait_seconds)
    edge.switchNext()
    count = _get_result_count(edge)
    if count == 0:
        edge.log.warning(f"search pharmacist by code: {code}, response empty!")
        edge.switchBack()
        return {}
    if count > 1:
        edge.log.info(f"search pharmacist by code: {code} response multiple results, advance search by name: {name}")
        _advance_search(edge, code, name, wait_seconds)
        count = _get_result_count(edge)
        if count == 0:
            edge.log.warning(f"search pharmacist by code: {code}, name: {name} response empty!")
            edge.switchBack()
            return {}
        elif count > 1:
            edge.log.warning(f"search pharmacist by code: {code}, name: {name} response multiple results!")
            edge.switchBack()
            return {}

    selector = "tr.el-table__row > td:nth-child(5) > div > button"
    button = edge.find(selector)
    if button is None:
        print("there is no button")
        edge.switchBack()
        return {}
    button.click()
    edge.sleep(wait_seconds)
    edge.switchNext()
    print(edge.driver.current_url)
    data = _get_dict_info(edge)
    edge.switchBack()
    edge.switchBack()
    print(edge.driver.current_url)
    return data
