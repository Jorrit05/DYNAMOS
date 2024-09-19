<!-- /agents/unl1_agent
/agents/unl2_agent
/microservices/anonymize_service
/microservices/query_service
/reasoner/archetype_config/ArcheType1
/reasoner/archetype_config/ArcheType2
/reasoner/requestor_config/uva
/reasoner/requestor_config/vu
/unl1_agent/services/anonymize_service
/unl1_agent/services/query_service

Todo?:
/unl1_agent/services/input_queue
/unl1_agent/services/output_queue

/request-types/sql-data-request -> JSON definition -->

/requestTypes/sqlDataRequest -> JSON definition
/archetypes/computeToData -> JSON definition
/archetypes/dataThroughTtp -> JSON definition
/microservices/queryService/chainMetadata -> JSON definition


/policyEnforcer/agreements/VU -> JSON definition
/policyEnforcer/agreements/UVA -> JSON definition
/policyEnforcer/agreements/RUG -> JSON definition
/dataStewards/UVA (agent availability in the distributed system)
<!-- /microservices/queryService/deploymentData -> JSON deployment definition -->
<!--
/activeJobs/<agent>/fullName -> jobName?
/activeJobs/<agent>/Jobname -> Full composition struct? -->

/agents/jobs/UVA/jorrit.stutterheim@cloudnation.nl -> jobname
/agents/jobs/UVA/jorrit.stutterheim@cloudnation.nl/jobname -> compositionRequest with local jobName