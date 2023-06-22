package etcd

// func UnmarshalStackFile(fileLocation string) MicroServiceData {

// 	yamlFile, err := os.ReadFile(fileLocation)
// 	if err != nil {
// 		logger.Sugar().Errorw("Failed to read the YAML file: %v", err)
// 	}

// 	service := MicroServiceData{}
// 	err = yaml.Unmarshal(yamlFile, &service)
// 	if err != nil {
// 		logger.Sugar().Errorw("Failed to unmarshal the YAML file: %v", err)
// 	}
// 	return service
// }

// // Take a given docker stack yaml file, and save all pertinent info (struct MicroServiceData), like the
// // required env variable and volumes etc. Into etcd.
// func SetMicroservicesEtcd(etcdClient EtcdClient, fileLocation string, etcdPath string) (map[string]MicroService, error) {
// 	if etcdPath == "" {
// 		etcdPath = "/microservices"
// 	}

// 	var service MicroServiceData = UnmarshalStackFile(fileLocation)

// 	processedServices := make(map[string]MicroService)

// 	for serviceName, payload := range service.Services {

// 		jsonPayload, err := json.Marshal(payload)
// 		if err != nil {
// 			logger.Sugar().Errorw("Failed to marshal the payload to JSON: %v", err)
// 			return nil, err
// 		}
// 		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()
// 		_, err = etcdClient.Put(ctx, fmt.Sprintf("%s/%s", etcdPath, serviceName), string(jsonPayload))
// 		if err != nil {
// 			logger.Sugar().Errorw("Failed creating service config in etcd: %s", err)
// 			return nil, err
// 		}
// 		processedServices[serviceName] = payload

// 	}
// 	return processedServices, nil
// }
