helm upgrade -i -f core/values.yaml core /Users/jorrit/Documents/master-software-engineering/thesis/DYNAMOS/charts/core
helm upgrade -i -f orchestrator/values.yaml orchestrator /Users/jorrit/Documents/master-software-engineering/thesis/DYNAMOS/charts/orchestrator
helm upgrade -i -f agents/values.yaml unl1 /Users/jorrit/Documents/master-software-engineering/thesis/DYNAMOS/charts/agents -n unl1