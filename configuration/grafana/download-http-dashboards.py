import requests
import os

# Define your dashboard IDs, revisions and names
dashboards = [
    {"id": 15474, "revision": 4, "name": "top-line"},
    {"id": 15486, "revision": 3, "name": "health"},
    {"id": 15479, "revision": 2, "name": "kubernetes"},
    {"id": 15478, "revision": 3, "name": "namespace"},
    {"id": 15475, "revision": 6, "name": "deployment"},
    {"id": 15477, "revision": 3, "name": "pod"},
    {"id": 15480, "revision": 3, "name": "service"},
    {"id": 15481, "revision": 3, "name": "route"},
    {"id": 15482, "revision": 3, "name": "authority"},
    {"id": 15483, "revision": 3, "name": "cronjob"},
    {"id": 15487, "revision": 3, "name": "job"},
    {"id": 15484, "revision": 3, "name": "daemonset"},
    {"id": 15491, "revision": 3, "name": "replicaset"},
    {"id": 15493, "revision": 3, "name": "statefulset"},
    {"id": 15492, "revision": 4, "name": "replicationcontroller"},
    {"id": 15489, "revision": 2, "name": "prometheus"},
    {"id": 15490, "revision": 2, "name": "prometheus-benchmark"},
    {"id": 15488, "revision": 3, "name": "multicluster"}
]

# Iterate over all the dashboards and download them
for dashboard in dashboards:
    url = f"https://grafana.com/api/dashboards/{dashboard['id']}/revisions/{dashboard['revision']}/download"
    response = requests.get(url)
    with open(f"{dashboard['name']}.json", 'w') as f:
        f.write(response.text)
