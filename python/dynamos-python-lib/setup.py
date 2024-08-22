from setuptools import setup, find_packages

setup(
    name='dynamos',
    version='0.1',
    packages=find_packages(),
    include_package_data=True,
    install_requires=[
        'grpcio==1.59.3',
        'google>=3.0.0',
        'grpcio-tools>=1.59.3',
        'retrying>=1.3.4',
        'grpclib>=0.4.5',
        'protobuf==4.25.3',
        'opentelemetry-api==1.19.0',
        'opentelemetry-instrumentation>=0.40b0',
        'opentelemetry-instrumentation-grpc>=0.40b0',
        'opentelemetry-semantic-conventions==0.40b0',
        'opentelemetry-exporter-otlp==1.19.0',
        'opentelemetry-exporter-otlp-proto-common==1.19.0',
        'opentelemetry-exporter-otlp-proto-grpc==1.19.0',
        # 'opentelemetry-proto==1.19.0',
        'opentelemetry-sdk==1.19.0',
    ],
    author='Jorrit S.',
    author_email='',
    description='Python lib to interface microservice to DYNAMOS',
    url='https://github.com/jorrit05/DYNAMOS'
)

