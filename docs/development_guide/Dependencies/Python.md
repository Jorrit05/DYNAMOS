
# Python
Some DYNAMOS services are based on python. For the microservices based in Python, we have created a `dynamos`python pip library to ease integration. To use this library we recommend taking the following approach:

Firstly, you need to have python installed. Most modern Linux distros have Python3 pre-installed, you can check this by running the following: 

```sh 
python3 --version # Python 3.10.12
```
In the rare case that you do not have Python installed, use the following:
```sh 
sudo apt update
sudo apt install python3
```

## venv
To handle dependencies, we recommend to use venv (you are free to use other dependency managers such as anaconda)

If you're using Python3.4+, `venv` is available directly in python. Else, you can install it with pip, with the following command:
```sh 
# (Optional, since most distros have venv)
pip install virtualenv
```
Create a venv, in this case we will call the directory venv (second argument), but it can be anything you like:
```sh 
python -m venv venv
```
When developing, you can activate your venv with the following command:
```sh 
source venv/bin/activate
```
Keep in mind that the `venv` directory is wherever you created in the previous command.

## Prerequisite pip package 
To be able to use create and build a PIP package, the `wheel` dependency is required, all other dependencies are handled as requirements in the services themselves.

Activate your venv and run the following:
```sh 
pip install wheel
```

## DYNAMOS Pip package
As previously mentioned, we developed a python library that handles the initialization and configuration of microservices, to build this locally you must:

1. Activate venv
```sh 
source venv/bin/activate
```

2. Change directory to `dynamos-python-lib`
```sh 
cd python/dynamos-python-lib
```

3. Run pip install:
```sh 
pip install .
```

4. `editable` mode
When testing locally, changes in the library will be directly reflected without needing to re-install the package. 

```sh 
pip install -e .
```


The output of above command should look something like this:
```
...
Successfully built dynamos
Successfully installed dynamos-0.1
```
