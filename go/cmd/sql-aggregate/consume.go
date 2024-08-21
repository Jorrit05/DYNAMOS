
// Wrapper function to handle incoming messages either from rabbitMQ or a previous microservice
func incomingMessageWrapper(ctx context.Context, msComm *pb.MicroserviceCommunication) {
	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: incomingMessageWrapper, process grpc MS", msComm.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End()

	// Wait till all services and connections have started
	logger.Debug("Wait for all services to start")
	<-COORDINATOR

	c := pb.NewMicroserviceClient(config.NextConnection)

	mscommList = append(mscommList, msComm)
	logger.Sugar().Infof("mscommList: %v", mscommList)
	logger.Sugar().Infof("amount of data providers %v", NR_OF_DATA_PROVIDERS)
	logger.Sugar().Infof("Lenght of msCommList %v", len(mscommList))

	// If NR_OF_DATA_PROVIDERS == 0 aggregate won't actually function and pass on the message.
	// This can happen at this moment if the aggregate flag is set to True, but it is not allowed by policy.
	if len(mscommList) == NR_OF_DATA_PROVIDERS || NR_OF_DATA_PROVIDERS == 0 {
		logger.Sugar().Debugf(mscommList[0].Data.String())
		// All messages have arrived
		logger.Sugar().Infof("All messages have arrived, %v", len(mscommList))

		switch msComm.RequestType {
		case "sqlDataRequest":
			ctx, msComm, err = handleSqlDataRequest(ctx, mscommList)
			if err != nil {
				logger.Sugar().Errorf("Error in handlesqlrequest: %v", err)
			}

		default:
			logger.Sugar().Errorf("Unknown RequestType type: %v", msComm.RequestType)
		}

		logger.Sugar().Debug("Printing merged data")
		c.SendData(ctx, msComm)
		logger.Sugar().Debugf(msComm.Data.String())
		close(config.StopMicroservice)
	}
}