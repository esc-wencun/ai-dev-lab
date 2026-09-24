import os
import sys
from loguru import logger

# 日志存放路径（Java版logback.xml的log.path；Windows下放项目目录，Linux可用/home/ruoyi/logs）
LOG_PATH = os.environ.get('APP_LOG_PATH') or os.path.join(os.getcwd(), 'logs')
if not os.path.exists(LOG_PATH):
    os.makedirs(LOG_PATH, exist_ok=True)

# 移除loguru默认handler
logger.remove()

# 输出格式：对齐Java logback.xml的pattern
# java: %d{HH:mm:ss.SSS} [%thread] %-5level %logger{20} - [%method,%line] - %msg%n
LOG_PATTERN = (
    "<green>{time:HH:mm:ss.SSS}</green> "
    "[{thread.name}] "
    "<level>{level: <5}</level> "
    "{name} - [{function},{line}] - "
    "<level>{message}</level>"
)

# 控制台输出（对应logback console appender，root level=info）
logger.add(sys.stdout, level="INFO", format=LOG_PATTERN, enqueue=True)


def _make_file_sink(filename: str, filter_func=None) -> str:
    """
    构建按天滚动、保留60天的文件sink路径（对应TimeBasedRollingPolicy + maxHistory 60）
    """
    return os.path.join(LOG_PATH, filename)


# 系统日志输出 sys-info.log（精确接收INFO级别，对应LevelFilter INFO/ACCEPT/DENY）
logger.add(
    _make_file_sink("sys-info.log"),
    level="INFO",
    format=LOG_PATTERN,
    filter=lambda record: record["level"].name == "INFO",
    rotation="00:00",
    retention=60,
    encoding="utf-8",
    enqueue=True,
    backtrace=False,
    diagnose=False,
)

# 系统日志输出 sys-error.log（精确接收ERROR及以上级别，对应LevelFilter ERROR/ACCEPT/DENY）
logger.add(
    _make_file_sink("sys-error.log"),
    level="ERROR",
    format=LOG_PATTERN,
    filter=lambda record: record["level"].name == "ERROR",
    rotation="00:00",
    retention=60,
    encoding="utf-8",
    enqueue=True,
    backtrace=False,
    diagnose=False,
)

# 用户访问日志 sys-user.log（对应sys-user logger + appender；不区分级别全部接收）
logger.add(
    _make_file_sink("sys-user.log"),
    level="INFO",
    format=LOG_PATTERN,
    filter=lambda record: record["extra"].get("name") == "sys-user",
    rotation="00:00",
    retention=60,
    encoding="utf-8",
    enqueue=True,
    backtrace=False,
    diagnose=False,
)

# sys-user 命名logger（对应Java LoggerFactory.getLogger("sys-user")）
# bind的值存入record["extra"]["name"]，由上面的filter匹配
sys_user_logger = logger.bind(name="sys-user")


def get_block(msg) -> str:
    """
    日志分块（对应Java LogUtils.getBlock）：[msg]
    """
    return f"[{'' if msg is None else msg}]"


def record_logininfor_log(ip: str, address: str, username: str, status: str, message: str):
    """
    记录登录信息到sys-user.log（对应Java AsyncFactory.recordLogininfor中的
    sys_user_logger.info([ip][地点][用户名][状态][消息])）
    """
    line = f"{get_block(ip)}{get_block(address)}{get_block(username)}{get_block(status)}{get_block(message)}"
    sys_user_logger.info(line)
