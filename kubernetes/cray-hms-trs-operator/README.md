# Cray HMS TRS Operator

This is the operator for the Task Runner Service (TRS). 

Services that want to use TRS workers to parallelize their requests can do so by creating a custom resouce that this operator will then use to broker the connection between client application and worker.