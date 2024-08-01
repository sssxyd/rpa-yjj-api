import os
import smtplib
import sys
from datetime import datetime
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText

from dotenv import load_dotenv

_loaded_configs: set[str] = set()


def load_config(name: str):
    if name in _loaded_configs:
        return
    conf_path = os.path.join(os.getcwd(), 'conf', f"{name}.env")
    if not os.path.exists(conf_path):
        return
    if os.path.exists(conf_path):
        load_dotenv(conf_path)
        _loaded_configs.add(name)


def load_args() -> dict[str, any]:
    params = dict()
    for i in range(1, len(sys.argv)):
        arg = sys.argv[i]
        if arg.startswith('--'):
            params[arg[2:]] = True
        elif arg.startswith('-'):
            idx = arg.find('=')
            if idx > 0:
                params[arg[1:idx].strip()] = arg[idx + 1:].strip()
    return params


def get_duration(start_time: datetime) -> str:
    # 获取当前时间
    now = datetime.now()

    # 计算时间差
    duration = now - start_time

    # 提取天、秒、小时、分钟
    days = duration.days
    seconds = duration.seconds
    hours = seconds // 3600
    minutes = (seconds % 3600) // 60
    seconds = seconds % 60

    # 构建人类友好的字符串
    parts = []
    if days > 0:
        parts.append(f"{days} 天")
    if hours > 0:
        parts.append(f"{hours} 小时")
    if minutes > 0:
        parts.append(f"{minutes} 分钟")
    if seconds > 0:
        parts.append(f"{seconds} 秒")

    return ", ".join(parts)


