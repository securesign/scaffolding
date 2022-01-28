# Getting Started

This document should allow you to stand up a fully functioning sigstore stack,
including:

 * Fulcio
 * Rekor
 * CTLog
 * Trillian - backing Rekor and CTLog

# Running on GitHub actions

If you want to just incorporate into your tests, you can look at this
[PR](https://github.com/nsmith5/rekor-sidekick/pull/27)] to
see how to incorporate. It includes standing up a KinD cluster as well as how
to install the bits and run a smoke test that exercises all the pieces. Rest of
this document talks about howto run locally on KinD.

# Running locally on KinD

You should be able to install KinD and Knative bits with one of the following
depending on your machine:

```shell
./hack/setup-kind.sh
```

For a Mac,  please use the script below (also please note that the airplay
receiver uses the 5000 port and may need to be disabled details
[here](https://developer.apple.com/forums/thread/682332)). Alternatively, you
can manually modify the script and change the
[REGISTRY_PORT](https://github.com/vaikas/sigstore-scaffolding/blob/main/hack/setup-mac-kind.sh#L19)

```shell
./hack/setup-mac-kind.sh
```

*NOTE* You may have to uninstall the docker registry container between running
the above scripts because it spins up a registry container in a daemon mode.
To clean a previously running registry, you can do one of these:

YOLO:

```shell
docker rm -f `docker ps -a | grep 'registry:2' | awk -F " " '{print $1}'`
```

Or to check things first:

```shell
docker ps -a | grep registry
b1e3f3238f7a   registry:2                        "/entrypoint.sh /etc…"   15 minutes ago   Up 15 minutes               0.0.0.0:5000->5000/tcp, :::5000->5000/tcp   registry.local
```

So that's the running version of the registry, so first kill and then remove it:
```shell
docker rm -rf b1e3f3238f7a
```

# Install sigstore-scaffolding pieces

```shell
curl -L https://github.com/vaikas/sigstore-scaffolding/releases/download/v0.1.13-alpha/release.yaml | kubectl apply -f -
```

Or for Arm64 based (M1 for example):

```shell
curl -L https://github.com/vaikas/sigstore-scaffolding/releases/download/v0.1.13-alpha/release-arm.yaml | kubectl apply -f -
```

The reason for different releases is the mysql binary used in the Intel based
release does not have Arm64 version.


# Then wait for the jobs that setup dependencies to finish

```shell
kubectl wait --timeout=10m -A --for=condition=Complete jobs --all
```

Obviously if you have other jobs running, you might have to tune this, for deets
see [below](#outputs) what gets deployed and where.

 # Outputs

The deployment above creates 4 namespaces:

 * trillian-system
 * ctlog-system
 * fulcio-system
 * rekor-system

## trillian-system namespace

`trillian-system` namespace contains all things related to
[Trillian](https://github.com/google/trillian). This means there will be two
services `log-server`, `log-signer`, and a mysql pod.

To access these services from the cluster, you'd use:

 * `log-server.trillian-system.svc`
 * `log-server.trillian-system.svc`
 * `mysql-trillian.trillian-system.svc`

 ## ctlog-system namespace

 `ctlog-system` namespace contains the
 [ctlog](https://github.com/google/certificate-transparency-go) service and
 can be accessed with:

  * `ctlog.ctlog-system.svc`

## fulcio-system namespace

`fulcio-system` namespace contains [Fulcio](https://github.com/sigstore/fulcio)
and Fulcio can be accessed in the cluster with:

 * `fulcio.fulcio-system.svc`

## rekor-system namespace

`rekor-system` namespace contains [Rekor](https://github.com/sigstore/rekor)
and Rekor can be accessed in the cluster with:

 * `rekor.rekor-system.svc`

## Testing Your new Sigstore Kind Cluster
1) Get ctlog-public-key and add to default namespace
```shell
kubectl -n ctlog-system get secrets ctlog-public-key -oyaml | sed 's/namespace: .*/namespace: default/' | kubectl apply -f -
```

3) Create the two test jobs (checktree and check-oidc)  using this yaml (this may take a bit, since the two jobs are launched simultaneously)
```shell
curl -L https://github.com/vaikas/sigstore-scaffolding/releases/download/v0.1.13-alpha/testrelease.yaml | kubectl apply -f -
```

4) To view if jobs have completed
```shell
kubectl get jobs/checktree jobs/check-oidc
```

## Example e2e test and cosign invocation using all of the above

There's an [E2E](./github/workflows/fulcio-rekor-kind.yaml) test that spins all
these up and before the documentation here catches up is probably the best place
to look to see how things are spun up if you run into trouble or want to use it
in your tests.

As part of the E2E test we use [cosign](https://github.com/sigstore/cosign) to
sign an image (and verify an entry made it Rekor), that should hopefully allow
you to use it in your tests as well. The invocation is
[here](./testdata/config/sign-job/sign-job.yaml) and while it's wrapped in a k8s
Job and it uses a container, it basically executes this against the stack
deployed above:

```shell
COSIGN_EXPERIMENTAL=true SIGSTORE_CT_LOG_PUBLIC_KEY_FILE=/var/run/sigstore-root/rootfile.pem
cosign sign --fulcio-url=http://fulcio.fulcio-system.svc \
--rekor-url=http://rekor.rekor-system.svc \
ko://github.com/vaikas/sigstore-scaffolding/cmd/rekor/checktree
```

Where the `rootfile.pem` gets mounted by the job, but you can get this Public
key of the CTLog, so that you can verify the SCT coming back from Fulcio, by
doing this:

```shell
kubectl -n ctlog-system get secrets ctlog-public-key -o=jsonpath='{.data.public}' | base64 -d
```

For example for my invocation it looks like this:

```shell
kubectl -n ctlog-system get secrets ctlog-public-key -o=jsonpath='{.data.public}' | base64 -d
-----BEGIN RSA PUBLIC KEY-----
MIICCgKCAgEAtPFdYCIKeK9yIZioAqk1JZnkxQVaisxJf17iMgZ6zRxoOh/E/Owh
GIHLUE53P3Zucezq5haJtUz0U8DQd5SDDt7KkT6hyWddLBqVpeHsvIVFb+fva4pn
HgUhXSLiPDuKFP6a4d1b/g9FX0PUFRfUtt1pBL6w+r/oEZtvNgt6xj2a/YK1Vfef
MzJowQ86qAuvAUko9Rt2wkSyjlk3hAq3T+9zTdrR8mfJ6Q0TfFjqVDgZIlVu92TN
BneIzkSp+CnpJo6ghQiVlCMKqaBzYmNWBusNGShGBXuH976nW4AWPMagyUlYDVu3
1L2vHyEuEZpwJgGxWH0SQ2uV94rYqnrdNRfIMvkCQsW/U7zVRf9S3u1lA2sRp7h2
fBe0D27Bu0sCOPH4fwabsKNcrNN/7vTNvujmoS7LqYwI6DvzyjTXSazy9mImB6Ik
0Izm94FL12vQSSsHRXXT0lkvL3cBRHbCd/qk54LKWisO098Nsx24W6dIjFXevnPg
SPqIvN546ELoE5Opa5p8KCEw7IAkkb0OvnfnfciRwmPVBR19fslkm1qAS17ayAKq
OFShVnTiiecQpssOUTCpe9n0y+GRnsx6KazyDHZ6iTCNrXTFDYzocoX3xJV62vyo
uiXoqBju311tisbKmtUX+g8JNlyH3p/eN0TePflERtS4yTcNkDvxVrUCAwEAAQ==
-----END RSA PUBLIC KEY-----
```

So you can pipe that to file and replace the `/var/run/sigstore-root/rootfile.pem`
with that location.

If the services of the cluster are not publically accessible, you can
port-forward to your cluster like so (assuming you installed Knative with
kourier):

```shell
kubectl -n kourier-system port-forward service/kourier-internal 8080:80 &
```

and adding entries to your /etc/hosts

```
127.0.0.1 rekor.rekor-system.svc
127.0.0.1 fulcio.fulcio-system.svc
127.0.0.1 ctlog.ctlog-system.svc
```

If you do this port forwarding, then you also have to modify the --fulcio-url
and --rekor-url above to have the local port number, so for example:

```
--fulcio-url=http://fulcio.fulcio-system.svc:8080
--rekor-url=http://rekor.rekor-system.svc:8080
```

You can also verify that the entries were added to the CTLog (this is assuming)
you successfully ran the jobs to completion above. For example, this is what
I get:

```shell
curl http://ctlog.ctlog-system.svc:8080/sigstorescaffolding/ct/v1/get-sth
{"tree_size":1,"timestamp":1643137195022,"sha256_root_hash":"i3NpxGSUw0/Ol0NmIba9ssMbYsogHHpwD3fHIGS84AI=","tree_head_signature":"BAECAAHqdlMlgLN6lBptnCd4fWaF/m99j877NRu0RerfO8eA/t7b15UKfSRzVmXG60tk1WVCPr0I2yqckIBFMB5tb/d7WCddlWVLbhtwTFG8hJwv1NbAFuSGmkKFHaZCt8rc9Ht9hEYiZTjKjIfXDP+FudSrQt0YsmZCAcVBM0LszowLGhjrx1oUrYNKbfnihI8CEyQMN7qCz83Cebm2acDgaGfv3NgfsC7LMPVFJNN9ryG1F2kK9Mb4UF6us3k8WeAAaOzxO7+NfHNQuaCt7FbFBjrDDtTAuTSfp9rgcGgJ4AiHvrx6X8PhXt7RdTQVDMPlWKPZg/gBYKla+19GSsCLf60msrF57295c2fPoL87mDZGuxLVaSBQKOlKWb+CJAe5+WUqG22Veq8xxlHrZkaXTRToNTA6P8/Z3Vy/QHTDaDd4k3Fm33sERu7F/R46U4UMQd1eMGTIoknKupDbjbW2JigbmAJMdT8dw5EskFLY/+Usn6ykZI6MBp2hHMiW84Son7CiniIzwqYBW0JFPSm/Pf/6LAYYNg/q5DX+AmHWTB8GM58+3IodlAO3qorF7TXUZqpZCiMjlJWP5AC5NRwVfnLOu8nS+fS27dWxK0MqIcAOVTUZ+T6vFAbKCiDIDp6MwggJ2tSUx4lk3wp4i0PF9nEH8S2VeiX7c9cwsP6GXGLV"}%
```

And if you check the rekor state for example with loginfo, you can do so like:

```shell
rekor-cli --store_tree_state=false --rekor_server http://rekor.rekor-system.svc:8080 loginfo
No previous log state stored, unable to prove consistency
Verification Successful!
Tree Size: 1
Root Hash: 68034bc4c888a307cd2f3289aecc4ebf80c5b720a4655bc2b3a073671ca2d54a
Timestamp: 2022-01-25T19:28:56Z
```