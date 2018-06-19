# mehdb

This is `mehdb`, an educational Kubernetes-native NoSQL datastore. It is not meant for production usage but purely to learn and experiment with [StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/).

## Deployment

## Local development

```bash
$ MEHDB_HOST=meh-shard-0 go run main.go
```

```bash
$ http PUT localhost:9876/set/abc < test/somedata
$ http localhost:9876/get/abc
```