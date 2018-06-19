# mehdb

This is `mehdb`, an educational Kubernetes-native NoSQL data store. It is not meant for production usage but purely to learn and experiment with [StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/).

## Usage

Deploy it:

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

Scale to 3 shards (1 leader, 2 followers):

```bash
$ kubectl scale sts web --replicas=3
```

Clean up:

```bash
$ kubectl delete sts/mehdb
$ kubectl delete pvc/data-mehdb-0
$ kubectl delete pvc/data-mehdb-1
$ kubectl delete svc/mehdb
```

Note: I tested it in OpenShift Online with Kubernetes in version 1.9 and the setup assumes that a storage class `ebs` exists.

## API

Once deployed you can use `mehdb` to store and retrieve data. Keep the following in mind:

- The keys are restricted, that is, they must match `[a-z]+`. For example, `abc` is a valid key, `123` or `_mykey42` is not.
- The leader shard accepts both reads and writes, a follower shard will redirect to the leader shard if you attempt a write operation.

The following public endpoints are available:

`/get/$KEY` … a HTTP `GET` at this endpoint retrieves the payload available under the key `$KEY` or a `404` if the key does not exist.

`/set/$KEY` … a HTTP `PUT` at this endpoint stores the payload provided under the key `$KEY`.

`/status` … by default returns a `200` and the role (leader or follower), which can be used for a liveness probe, with `?level=full` it returns a `200` and the number of keys it can serve, which can be used for a readiness probe.

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
