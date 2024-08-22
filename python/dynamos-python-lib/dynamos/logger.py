"""
Package dynamos, implements functionality for handling Microservice chains in Python.

File: logger.py

Description:
This file contains the logger. Simple generic logger initiation.

Notes:

Author: Jorrit Stutterheim
"""

import logging
import os
import sys


def InitLogger():
# Set up the logger
    logger = logging.getLogger(os.path.basename(sys.argv[0])) # use program name as logger name
    logger.setLevel(logging.DEBUG)

    # Create a console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(logging.DEBUG)

    # Create a formatter
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(filename)s:%(lineno)d - %(levelname)s - %(message)s')

    # Add the formatter to the handler and the handler to the logger
    console_handler.setFormatter(formatter)

    # Add the handler to the logger only if it has no handlers yet
    if not logger.handlers:
        logger.addHandler(console_handler)

    return logger
