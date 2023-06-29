# ---- Base python ----
# FROM python:3.9-slim-buster AS base
FROM  python:3.10-slim

ARG NAME
ENV NAME=$NAME

WORKDIR /app
COPY requirements.txt .
COPY . /app

RUN pip install --no-cache-dir -r requirements.txt

ENV  ENV="PROD"

CMD ["python", "/app/main.py"]
