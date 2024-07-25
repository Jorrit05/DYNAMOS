import grpc
from concurrent import futures
from dynamos.logger import InitLogger
from dynamos.grpc_server import GRPCServer
import time

# Configure logging
logger = InitLogger()


def main():
    server = GRPCServer("localhost:50051")

    server.start()


    # Example of other tasks that can run while the server is running
    try:
        while True:
            print("Running other tasks...")
            time.sleep(5)
    except KeyboardInterrupt:
        print("KeyboardInterrupt received, stopping server...")
        server.stop()


if __name__ == '__main__':
    main()
