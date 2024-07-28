import grpc
from concurrent import futures
from dynamos.logger import InitLogger
from dynamos.grpc_server import GRPCServer
import time

# Configure logging
logger = InitLogger()


def msCommHandler(msComm):
    logger.info(f"Start msCommHandler")

    logger.info(f"Received message mscomm: {msComm.request_metadata.destination_queue}")
    logger.info(f"Received message mscomm: {msComm.request_type}")

    # Implement the logic to handle the message here



def main():
    server = GRPCServer("localhost:50052", msCommHandler)
    server.start()


    # Example of other tasks that can run while the server is running
    try:
        while True:
            time.sleep(5)
    except KeyboardInterrupt:
        print("KeyboardInterrupt received, stopping server...")
        server.stop()


if __name__ == '__main__':
    main()
