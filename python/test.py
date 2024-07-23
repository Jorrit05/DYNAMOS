from dynamos.logger import InitLogger


logger = InitLogger()

logger.info("Hello, World!")

from dynamos import msServerTypes as msCommTypes

msComm = msCommTypes.MicroserviceCommunication()

