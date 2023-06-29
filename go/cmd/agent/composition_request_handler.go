package main

import (
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func compositionRequestHandler(compositionRequest *pb.CompositionRequest) {
	// get local requiredServices
	// Spin up pod
	// Save session information in etcd
	//

	deployJob(compositionRequest)
}
