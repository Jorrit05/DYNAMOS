"""
Package dynamos, implements functionality for handling Microservice chains in Python.

File: signal_flow.py

Description:
This file contains simple functions for multi-thread locking and signalling.

Notes:

Author: Jorrit Stutterheim
"""

import threading

def signal_continuation(event: threading.Event, condition: threading.Condition) -> None:
    """
    Signals the continuation of a thread waiting on a condition.

    Args:
        event (threading.Event): The event object to set.
        condition (threading.Condition): The condition object to notify.

    Returns:
        None
    """
    with condition:
        event.set()
        condition.notify()

def signal_wait(event: threading.Event, condition: threading.Condition) -> None:
    """
    Wait for a signal to stop.

    Args:
        event (threading.Event): The event object representing the signal.
        condition (threading.Condition): The condition object used for synchronization.

    Returns:
        None
    """
    with condition:
        while not event.is_set():
            condition.wait()  # Wait for the signal to stop
