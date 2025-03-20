# Intro

The purpose of this document is to list all known bugs that should be addressed in the future.

## Unknown Data Providers make the runtime of the requestApproval much longer

If a data provider is added to the array that is not in ETCD, the request will take approximately 40 seconds until the ETCD check times out. This should either immediately ignore it or raise an error.

## Traces are not properly closed

When checking the Jaeger UI, most requestApproval are shown to take 10 minutes, when in reality they take 7-10 seconds (assuming a correct payload). The source of this is that the traces are probably not being properly closed. From the statistics  of Jaegar UI, it seems that UVA and VU take 99% of the time of the trace, which leads me to believe that they are not being closed there.

## Job name counter

The function `go/cmd/agent/composition_request_handler#generateJobName` has a bug (temporarily fixed). The intention is to generate a job name based on how many jobs there are in the system. This assumes that the rabbit mq queues follow the same rules and have an index of the job number as the suffix of the queue name. This does not seem to be working as intended and would break any request made to the system after the first.

To temporarily fix this I have set the counter to be hardcoded to one, this fixes the behaviour, however I assume this would break concurrent requests, which is a big feature of the system.
