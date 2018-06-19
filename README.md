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

Scale:

```bash
$ kubectl scale sts web --replicas=3
```

Clean up:

```bash
$ kubectl delete sts/mehdb
$ kubectl delete pvc/data-mehdb-0
$ kubectl delete pvc/data-mehdb-1
```

Note: I tested it in OpenShift Online with Kubernetes in version 1.9.

## Endpoints


`/get/$KEY` … A HTTP `GET` at this endpoint retrieves the payload available under the key `$KEY` or a `404` if it doesn't exist.


`/set/$KEY` … A HTTP `PUT` at this endpoint stores the payload provided under the key `$KEY`.


`/status` … by default return `200` and the role (leader or follower), which can be used for a liveness probe, with `?level=full` it returns the number of keys it can serve, which can be used for a readiness probe.


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

If you try to write to the follower, you'll be redirected:

```bash
$ http PUT localhost:9876/set/abc < test/somedata
HTTP/1.1 307 Temporary Redirect
Content-Length: 0
Date: Tue, 19 Jun 2018 12:28:57 GMT
Location: http://localhost:9999/set/abc
```
