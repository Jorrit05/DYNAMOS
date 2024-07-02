from setuptools import setup, find_packages

setup(
    name='dynamos',
    version='0.1',
    packages=find_packages(),
    install_requires=[
        'grpcio>=1.64.1',
        'google>=3.0.0',
        'grpcio-tools>=1.64.1',
        'retrying>=1.3.4',
        'grpclib>=0.4.5',
        'protobuf==5.26.1'
    ],
    author='Jorrit S.',
    author_email='',
    description='Python lib to interface microservice to DYNAMOS',
    url='https://github.com/jorrit05/DYNAMOS'
)
