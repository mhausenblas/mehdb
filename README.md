# mehdb

This is `mehdb`, an educational Kubernetes-native NoSQL datastore. It is not meant for production usage but purely to learn and experiment with [StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/).

## Deployment

## Local development

Run a leader shard like so:

```bash
$ MEHDB_HOST=meh-shard-0 MEHDB_PORT=9999 go run main.go
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