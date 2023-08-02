package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

// /agents/jobs/SURF/jorrit.stutterheim@cloudnation.nl/jorrit-stutterheim-43ea82da
// {"archetype_id":"dataThroughTtp","request_type":"sqlDataRequest","role":"computeProvider","user":{"id":"12324","user_name":"jorrit.stutterheim@cloudnation.nl"},"data_providers":["UVA"],"destination_queue":"SURF-in","job_name":"jorrit-stutterheim-43ea82da","local_job_name":"jorrit-stutterheim-43ea82dasurf1"}
// /agents/jobs/SURF/queueInfo/jorrit-stutterheim-43ea82dasurf1
// jorrit-stutterheim-43ea82dasurf1
// /agents/jobs/UVA/jorrit.stutterheim@cloudnation.nl/jorrit-stutterheim-43ea82da
// {"archetype_id":"dataThroughTtp","request_type":"sqlDataRequest","role":"dataProvider","user":{"id":"12324","user_name":"jorrit.stutterheim@cloudnation.nl"},"destination_queue":"UVA-in","job_name":"jorrit-stutterheim-43ea82da","local_job_name":"jorrit-stutterheim-43ea82dauva1"}
// /agents/jobs/UVA/queueInfo/jorrit-stutterheim-43ea82dauva1
// jorrit-stutterheim-43ea82dauva1

func deleteJobInfo(jobNames []string, userName string, changedAgreementName string) {
	ctx := context.Background()
	// get all online agents
	var agents *lib.AgentDetails
	key := "/agents/online/"
	activeAgents, err := etcd.GetPrefixListEtcd(etcdClient, key, agents)
	if err != nil {
		logger.Sugar().Warnf("error get agents: %v", err)
	}

	for _, job := range jobNames {

		for _, agent := range activeAgents {
			jobInfoKey := fmt.Sprintf("/agents/jobs/%s/%s/%s", agent.Name, userName, job)

			resp, err := etcdClient.Get(ctx, jobInfoKey)
			if err != nil {
				logger.Sugar().Errorf("error getting value from etcd: %v", err)
			}

			if len(resp.Kvs) == 0 {
				continue
			}

			compositionRequest := &pb.CompositionRequest{}

			err = json.Unmarshal(resp.Kvs[0].Value, compositionRequest)
			if err != nil {
				logger.Sugar().Errorf("failed to unmarshal JSON: %v", err)
				return
			}

			key := fmt.Sprintf("/agents/jobs/%s/queueInfo/%s", agent.Name, compositionRequest.LocalJobName)
			_, err = etcdClient.Delete(ctx, key)
			if err != nil {
				logger.Sugar().Errorf("failed to delete key: %v", err)
				continue
			}

		}
	}
}
func checkJobs(agreement *api.Agreement) {
	// compositionRequest := &pb.CompositionRequest{}
	for relationName, relationDetails := range agreement.Relations {

		key := fmt.Sprintf("/agents/jobs/%s/%s", agreement.Name, relationName)

		jobNames, err := etcd.GetKeysFromPrefix(etcdClient, key, etcd.WithMaxElapsedTime(2*time.Second))
		if err != nil {
			logger.Sugar().Warnf("error get agents: %v", err)
		}
		if len(jobNames) == 0 {
			logger.Debug("no active jobs for this user")
			return
		}

		if len(relationDetails.AllowedArchetypes) == 0 || relationDetails.AllowedArchetypes[0] == "" {
			logger.Debug("This user no has no allowed archetypes")
			if len(jobNames) > 0 {
				deleteJobInfo(jobNames, relationName, agreement.Name)
			}
			// key := fmt.Sprintf("/agents/jobs/%s/queueInfo/%s", agreement.Name, activeJobCompositionRequests)

			// activeJobCompositionRequests.de
			continue
		}

		// key := fmt.Sprintf("/agents/jobs/%s/%s/", agreement.Name, relationName)
		// activeJobCompositionRequests, err := etcd.GetPrefixListEtcd(etcdClient, key, compositionRequest)
		// if err != nil {
		// 	logger.Sugar().Warnf("error get jobs: %v", err)
		// }

		// evaluateArchetypeChoice(activeJobCompositionRequests, agreement, relationDetails)
	}
}

func evaluateArchetypeChoice(activeJobCompositionRequests []*pb.CompositionRequest, agreement *api.Agreement, relationDetails api.Relation) {
	logger.Debug("starting evaluateArchetypeChoice")
	// alue.ArchetypeId == archetype in current active job from the agreement name.

	for _, value := range activeJobCompositionRequests {
		if lib.NewSet(relationDetails.AllowedArchetypes).Has(value.ArchetypeId) {
			println("still matching archetypes")
		} else {
			println("previous archetype no longer allowed")
			// chooseArchetype()
			// Now  I think I can use existing functions to determine a new  archetype
		}
	}

}
