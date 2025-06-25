# How to create a new python microservice? 

The easiest way would be to copy the source code from one of the existing microservices. E.g. duplicate the whole ml-anonymize-dataset folder (it is in this branch only) and then rename the folder. 

Then you have to change the functionality according to your requirements. You should change the name of the microservice (in the config_local and config_prod files) (the service_name variable).

The actual functionality is in the main.py, in the "request_handler" method. In general the input is the "msComm" protobuf object, Through it oyu can transfer pandas dataframes (and other things but I would advise to go with pandas DF). 

### msComm.metadata
Specifically the msComm.metadata can have (if you put them there) metadata about the services that were applied so far.  
The following code registers the current microservice in the metadata. This is not strictly necessary but is like a log of the processes that were applied.
```python
logger.debug(f"msComm metadata original: {str(mscomm_metadata)}")
mscomm_metadata = register_service_on_metadata(mscomm_metadata, service_name=service_name)
logger.debug(f"msComm metadata updated: {str(mscomm_metadata)}")
```

Moreover, in the metadata you can transfer the schema of the pandas dataframe. It might be useful so that you don't have to infer the dtypes. 
```python
dataframe_metadata_dict = json.loads(msComm.metadata['dataframe_metadata'])
data_df = protobuf_to_dataframe(msComm.data, dataframe_metadata_dict)
logger.debug(f"df head: {data_df.head()}")
```

The dataframe metadata is not necessary, but in that case you would have to infer the data types by yourself.

### msComm.data

The msComm.data part of the protobuf object holds the actual information that is propagated between the microservices.
To parse the data as pandas dataframe you can do: 
```python
dataframe_metadata_dict = json.loads(msComm.metadata['dataframe_metadata'])
data_df = protobuf_to_dataframe(msComm.data, dataframe_metadata_dict)
logger.debug(f"df head: {data_df.head()}")
```
If you do not want to transfer metadata with the schema you can adjust the protobuf_to_dataframe to infer the types.

Then you can apply the actual functionality to the dataframe and then encode it back to protobuf: 
E.g.:
```python
synthetic_df = generate_synthetic_dataset(data_df)

data, dataframe_metadata = dataframe_to_protobuf(synthetic_df)
mscomm_metadata['dataframe_metadata'] = json.dumps(dataframe_metadata)
```

## Virtual Environment
It is recommended to create a virtual environment (venv/conda) and install DYNAMOS according to the `python/dynamos-python-lib` to ensure that you have the correct versions of the requirements. Especially protobuf needs to be the same version so that it works.

If your module has specific reuirements you can adjust the `ml-anonymize-dataset/requirements.txt` accordingly. Ideally with specific versions. 


## Build images

To build the image you have to add the new module as a target in the `python/Makefile`.

Then in the venv you created for DYNAMOS run: `make module-name` (same as in the Makefile).

You can remove the docker push commands from the Make file if you do not have access to the repo. 
Alternatively you could create your own dockerhub repo and rebuild & push everything. But it is not necessary.

However, if you go with the option to have the images locally, then you need to adjust the configurations so that it know to not pull the images from the repo.
See below.

## imagePullPolicy

By default, all the components pull the new image from the dockerhub repo (so that if you build the image from another machine it is still updated to the latest one).
If you do not want to pull the images from the repo you need to change the image pull policy. 

**From:**
```yaml
imagePullPolicy: Always
```

**To:**
```yaml
imagePullPolicy: IfNotPresent
```

### agents 

Specifically the agent spins up the new microservices from within the code.
```python
return v1.Container{
		Name:            sidecarName,
		Image:           fullImage,
		ImagePullPolicy: "Always",
```

So in this case you would have to modify the code and change the policy from Always to IfNotPresent, in the source code for the agent. Then recreate the agent module (with the go/makefile) (make sure to change to remove the push if you do not have access to the repo). 
And then you would have the update image locally. This image would know to not look at the repo for the images if they exist locally.

## Register microservices in etcd

When you create new microservices you have to register them in the etcd configuration files. Those files are registered in the system on startup (so you would need to restart dynamos and run the dynaos-configuration.sh script again). 
In the directory: `configuration/etcd_launch_files``

`microservices.json` -> add new entry for the new microservice and determine the allowed output (if any).

`requestType.json` -> Specify the required and optional services for the request. 

`optional_microservices.json` -> (Optional) it it not necessary to register them here

`agreements.json` -> (Optional) You can change if you plan to change the name of the requestor. Not necessary.

`archetype.json` -> (Optional) If you want to add a new archetype. If you want to work on an existing archetype it is not necessary to modify this. Note that the weight determines which archetype to select if multiple archetypes are allowed. It selects the lowest weight out of the ones that are allowed. As a first step I would not change this, I would modify an existing archetype.

`datasets.json` -> Not implemented. 

### Data
If you need to put some datasets in the microservice you can add the data in the `module/data` folder. E.g. see `ml-merge-datasets`
This is not the canonical way. The canonical way is via PVC (Kubernetes Persistent Volumes).