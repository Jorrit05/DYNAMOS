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
	"github.com/google/uuid"
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

		// Get all jobnames registered for this user of the data steward of this agreement
		jobNames, err := etcd.GetKeysFromPrefix(etcdClient, key, etcd.WithMaxElapsedTime(2*time.Second))
		if err != nil {
			logger.Sugar().Warnf("error get agents: %v", err)
		}
		if len(jobNames) == 0 {
			logger.Debug("no active jobs for this user")
			return
		}
		for _, v := range jobNames {
			logger.Sugar().Warnf(v)
		}

		// New agreement has no allowed archetypes
		if len(relationDetails.AllowedArchetypes) == 0 || (relationDetails.AllowedArchetypes[0] == "" && len(relationDetails.AllowedArchetypes) == 1) {
			logger.Debug("This user no has no allowed archetypes")
			if len(jobNames) > 0 {
				deleteJobInfo(jobNames, relationName, agreement.Name)
			}
			continue
		}
		evaluateArchetypeInActiveJobs(jobNames, agreement, relationName, relationDetails, c)

		// key := fmt.Sprintf("/agents/jobs/%s/%s/", agreement.Name, relationName)
		// activeJobCompositionRequests, err := etcd.GetPrefixListEtcd(etcdClient, key, compositionRequest)
		// if err != nil {
		// 	logger.Sugar().Warnf("error get jobs: %v", err)
		// }

	}
}

func evaluateArchetypeInActiveJobs(jobNames []string, agreement *api.Agreement, relationName string, relationDetails api.Relation, c pb.RabbitMQClient) {
	logger.Debug("starting evaluateArchetypeInActiveJobs")
	ctx := context.Background()
	// alue.ArchetypeId == archetype in current active job from the agreement name.

	// for each job. Check current archetype. versus new archetypes.
	for _, job := range jobNames {

		jobInfoKey := fmt.Sprintf("/agents/jobs/%s/%s/%s", agreement.Name, relationName, job)

		resp, err := etcdClient.Get(ctx, jobInfoKey)
		if err != nil {
			logger.Sugar().Errorf("error getting value from etcd: %v", err)
			continue
		}

		if len(resp.Kvs) == 0 {
			logger.Warn("this should not happen")
			continue
		}

		currentRegisteredJob := &pb.CompositionRequest{}
		err = json.Unmarshal(resp.Kvs[0].Value, currentRegisteredJob)
		if err != nil {
			logger.Sugar().Errorf("error unmarshalling jobinfo: %v", err)
		}

		policyUpdate := &pb.PolicyUpdate{
			Type:            "policyUpdate",
			User:            &pb.User{Id: relationDetails.ID, UserName: relationName},
			RequestMetadata: &pb.RequestMetadata{DestinationQueue: "policyEnforcer-in"},
		}

		correlationId := uuid.New().String()
		policyUpdate.RequestMetadata.CorrelationId = correlationId

		agentsWithThisJob := make(map[string]*pb.CompositionRequest)

		ctx = getJobAcrossAgents(ctx, agentsWithThisJob, job, relationName)

		for k, v := range agentsWithThisJob {
			if v.Role == "all" || v.Role == "dataProvider" {
				policyUpdate.DataProviders = append(policyUpdate.DataProviders, k)
			}
		}

		policyUpdateMutex.Lock()
		policyUpdateMap[policyUpdate.RequestMetadata.CorrelationId] = agentsWithThisJob
		policyUpdateMutex.Unlock()
		c.SendPolicyUpdate(ctx, policyUpdate)
	}
}

func processPolicyUpdate(ctx context.Context, agentsWithThisJob map[string]*pb.CompositionRequest, policyUpdate *pb.PolicyUpdate) {
	logger.Sugar().Debugf("processPolicyUpdate")

	// TODO: Kinda threw this in without testing..
	authorizedProviders, err := getAuthorizedProviders(policyUpdate.ValidationResponse)
	if err != nil {
		logger.Sugar().Errorf("error getAuthorizedProviders : %v", err)
	}

	archetype, err := chooseArchetype(policyUpdate.ValidationResponse, authorizedProviders)
	if err != nil {
		logger.Sugar().Errorf("error choosing archetype: %v", err)
	}

	logger.Sugar().Debugf("New archetype: %v", archetype)

	var archetypeConfig api.Archetype
	_, err = etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/archetypes/%s", archetype), &archetypeConfig)
	if err != nil {
		logger.Sugar().Errorf("error choosing archetype: %v", err)
		return
	}

	// technically now, this shouldn't be necessary
	computeProviderAlready := false
	var ttp lib.AgentDetails
	for agent, currentData := range agentsWithThisJob {
		if currentData.ArchetypeId == archetype {
			logger.Sugar().Debug("same archetype, do nothing")
			return
		}
		key := fmt.Sprintf("/agents/jobs/%s/%s/%s", agent, policyUpdate.User.UserName, currentData.JobName)

		if archetypeConfig.ComputeProvider != "other" {
			if currentData.Role == "computeProvider" {
				// Delete this job info
				_, err := etcdClient.Delete(ctx, key)
				if err != nil {
					logger.Sugar().Warnf("error deleting key from etcd: %v", err)
				}
				continue
			}

			// New archetype is computeToData
			newData := currentData
			newData.ArchetypeId = archetype
			newData.Role = "all"
			newData.DataProviders = []string{}
			err := etcd.SaveStructToEtcd[*pb.CompositionRequest](etcdClient, key, newData)
			if err != nil {
				logger.Sugar().Errorf("Error saving struct to etcd: %v", err)
				return
			}
			computeProviderAlready = true
		} else {
			var err error
			ttp, err = chooseThirdParty(policyUpdate.ValidationResponse)
			if err != nil {
				logger.Sugar().Errorf("Error choosing third party: %v", err)
				return
			}

			if currentData.Role == "computeProvider" && agent == ttp.Name {
				computeProviderAlready = true
				continue
			} else if currentData.Role == "computeProvider" && agent != ttp.Name {
				// Delete this job info
				_, err := etcdClient.Delete(ctx, key)
				if err != nil {
					logger.Sugar().Warnf("error deleting key from etcd: %v", err)
				}
				continue
			}

			if currentData.Role == "all" {
				_, ok := policyUpdate.ValidationResponse.ValidDataproviders[agent]
				if !ok {
					// Delete this job info
					_, err := etcdClient.Delete(ctx, key)
					if err != nil {
						logger.Sugar().Warnf("error deleting key from etcd: %v", err)
					}
				}

				// New archetype is dataThroughTtp
				newData := currentData
				newData.ArchetypeId = archetype
				newData.Role = "dataProvider"
				newData.DataProviders = []string{}

				err = etcd.SaveStructToEtcd[*pb.CompositionRequest](etcdClient, key, newData)
				if err != nil {
					logger.Sugar().Errorf("Error saving struct to etcd: %v", err)
					return
				}

			}

		}
	}

	if !computeProviderAlready {
		compositionRequest := &pb.CompositionRequest{}
		compositionRequest.User = policyUpdate.User
		tmpDataProvider := []string{}

		for key := range policyUpdate.ValidationResponse.ValidDataproviders {
			tmpDataProvider = append(tmpDataProvider, key)
		}
		compositionRequest.Role = "computeProvider"
		compositionRequest.DataProviders = tmpDataProvider
		compositionRequest.ArchetypeId = archetype
		for _, v := range agentsWithThisJob {
			compositionRequest.RequestType = v.RequestType
			compositionRequest.JobName = v.JobName
			break
		}

		compositionRequest.DestinationQueue = ttp.RoutingKey

		c.SendCompositionRequest(ctx, compositionRequest)
	}
}

func getJobAcrossAgents(ctx context.Context, targetMap map[string]*pb.CompositionRequest, jobName string, userName string) context.Context {

	var agents *lib.AgentDetails
	key := "/agents/online/"
	activeAgents, err := etcd.GetPrefixListEtcd(etcdClient, key, agents)
	if err != nil {
		logger.Sugar().Warnf("error get agents: %v", err)
	}

	for _, agent := range activeAgents {

		key := fmt.Sprintf("/agents/jobs/%s/%s/%s", agent.Name, userName, jobName)

		resp, err := etcdClient.Get(ctx, key)
		if err != nil {
			logger.Sugar().Errorf("error getting value from etcd: %v", err)
			continue
		}

		if len(resp.Kvs) == 0 {
			logger.Sugar().Debugw("no value found for", "key", key)
			continue
		}

		agentsConfiguration := &pb.CompositionRequest{}
		err = json.Unmarshal(resp.Kvs[0].Value, agentsConfiguration)
		if err != nil {
			logger.Sugar().Errorf("error unmarshalling jobinfo: %v", err)
			continue
		}

		targetMap[agent.Name] = agentsConfiguration
	}

	return ctx
}

func handleRequestApproval(ctx context.Context, validationResponse *pb.ValidationResponse) {
	result := &pb.RequestApprovalResponse{Type: "requestApprovalResponse", RequestMetadata: &pb.RequestMetadata{DestinationQueue: "api-gateway-in"}}

	authorizedProviders, err := getAuthorizedProviders(validationResponse)
	if err != nil {
		result.Error = err.Error()
		c.SendRequestApprovalResponse(ctx, result)
		return
	}

	if len(authorizedProviders) == 0 {
		// TODO Respond with the following to the rabbitmq queue
		// []byte("Request was processed, but no agreements or available dataproviders have been found")
		result.Error = "Request was processed, but no agreements or available dataproviders have been found"
		c.SendRequestApprovalResponse(ctx, result)
		return
	}

	// TODO: Might be able to improve processing by converting functions to go routines
	// Seems a bit tricky though due to the response writer.

	compositionRequest := &pb.CompositionRequest{}
	compositionRequest.User = &pb.User{}
	userTargets, ctx, err := startCompositionRequest(ctx, validationResponse, authorizedProviders, compositionRequest)
	if err != nil {
		switch e := err.(type) {
		case *UnauthorizedProviderError:
			logger.Sugar().Warn("Unauthorized provider error: %v", e)
			return
		default:
			logger.Sugar().Errorf("Error starting composition request: %v", err)
			return
		}
	}

	result.Auth = &pb.Auth{}
	result.User = &pb.User{}

	result.Auth = validationResponse.Auth
	result.User = validationResponse.User

	result.AuthorizedProviders = make(map[string]string)
	result.AuthorizedProviders = userTargets
	result.JobId = compositionRequest.JobName

	c.SendRequestApprovalResponse(ctx, result)
}
