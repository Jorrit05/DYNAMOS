# ---- Base python ----
# FROM python:3.9-slim-buster AS base
FROM  python:3.10-slim

ARG NAME
ENV NAME=$NAME

WORKDIR /app
RUN mkdir $NAME
RUN mkdir grpc_lib
RUN mkdir datasets

COPY requirements.txt .
COPY ./*.py /app/$NAME
COPY ./grpc_lib/*.py /app/grpc_lib
COPY ./*.csv /app/datasets

RUN pip install --no-cache-dir -r requirements.txt

ENV  ENV="PROD"
ENV PYTHONPATH="/app/grpc_lib"

CMD ["sh", "-c", "python /app/${NAME}/main.py"]
