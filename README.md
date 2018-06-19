# mehdb

This is `mehdb`, an educational Kubernetes-native NoSQL data store. It is not meant for production usage but purely to learn and experiment with [StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/).

## Usage

```bash
$ kubectl create ns mehdb
$ kubectl -n=mehdb apply -f app.yaml
```

Access it from within the cluster:

```
$ kubectl run -i -t --rm jumpod --restart=Never --image=quay.io/mhausenblas/jump:0.2 -- sh
$ echo "test data" > /tmp/test
$ curl -L -XPUT -T /tmp/test mehdb:9876/set/test
$ curl mehdb:9876/get/test
$ curl mehdb-1.mehdb:9876/get/test
```

## Local development

Run a leader shard like so:

```bash
$ MEHDB_HOST=mehdb-0 MEHDB_PORT=9999 go run main.go
```

Run a follower shard like so:

```bash
$ MEHDB_LOCAL=yes MEHDB_DATADIR=./follower-data go run main.go
```

Now you can for example write to and/or read from the leader:

```bash
$ http PUT localhost:9999/set/abc < test/somedata
$ http localhost:9999/get/abc
```

Also, you can read from the follower:

```bash
$ http localhost:9876/get/abc
```