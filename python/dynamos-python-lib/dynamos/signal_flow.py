import threading

def signal_continuation(event: threading.Event, condition: threading.Condition) -> None:
    with condition:
        event.set()
        condition.notify()

def signal_wait(event: threading.Event, condition: threading.Condition) -> None:
    with condition:
        while not event.is_set():
            condition.wait()  # Wait for the signal to stop
