helm upgrade -i -f values.yaml core /Users/jorrit/Documents/master-software-engineering/thesis/micro-recomposer/charts/core
helm upgrade -i -f values.yaml orchestrator /Users/jorrit/Documents/master-software-engineering/thesis/micro-recomposer/charts/orchestrator
helm upgrade -i -f values.yaml unl1 /Users/jorrit/Documents/master-software-engineering/thesis/micro-recomposer/charts/agents -n unl1