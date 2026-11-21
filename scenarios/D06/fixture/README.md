# Cluster diagnostics
Run `python3 kube_check.py` against ~/.kube/config. The intended endpoint is
https://cluster.internal.test. This offline kubeconfig diagnostic replays a
retired-cluster failure, not a live kubectl/API connection. No cluster required.
