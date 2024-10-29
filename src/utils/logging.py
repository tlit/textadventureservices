"""Logging configuration."""
from loguru import logger
import sys

# Configure loguru 
logger.remove()
logger.add(
    sys.stdout,
    format="{time:YYYY-MM-DD HH:mm:ss} | {level} | {message}",
    level="INFO"
)
logger.add(
    "logs/master.log",
    rotation="500 MB",
    retention="10 days",
    compression="zip",
    level="DEBUG"
)

logger.info("Logging configured")
