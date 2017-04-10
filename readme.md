# GCE Cache Cluster

Easy groupcache clustering on GCE

## Overview

[Groupcache](https://github.com/golang/groupcache) is amazing, it's like
memcache on steroids but offering some new features:

* Can run on your existing servers rather than as a separate service.
* De-dupes cache requests to avoid "stempeding herd" cache filling.
* Improves response time as hot items can kept in-memory (no network request).

Fun fact - Groupcache was written by the same author as memcache (so yeah, he
_probably_ knows a thing or two about caching). It's also used within Google
so you can be confident it works at scale.

There is a lot to love about groupcache so, if you're using Go, it's
really a no-brainer. There's just one issue ...

## Peer Discovery

In order for groupcache to work, the servers need to know about each other.
That is, each server needs to keep an updated list of it's peers so it knows
which node is responsible for loading and caching any item.

This package is designed to make that easier if you're using Google Compute
Engine (GCE) especially if you have any kind of auto-scaling. It will handle
the job of maintaining the list of peers and can be configured to filter the
nodes that should be considered part of any cluster (you may have more than
one "service" within a project which should each have their own independent
groupcache).

## Usage

Import the package:

```go
import (
    "github.com/captaincodeman/gce-cache-cluster"
)
```

Create a configuration or use the `LoadConfig` function to load from an `.ini`
file (defaults to `cachecluster.ini`):

```go
config, _ := cachecluster.LoadConfig("")
```

Example configuration file:

```ini
[cache]
port = 9080               # port that groupcache will run on

[cluster]
port = 9999               # port for cluster traffic
heartbeat = 250           # heartbeat in milliseconds

[match]
project = my-project      # project name (see note)
zone = us-central1-b      # zone name (see note)
network_interface = nic0  # network interface to use
tags = http-server        # tags used to match instances

[meta]
service = web             # metadata used to match instances
```

NOTE: The project and zone are populated automatically from the GCE metadata.

Create a new cache instance with the config:

```go
cache, _ := cachecluster.New(config)
```

Start the groupcache webserver:

```go
go http.ListenAndServe(cache.ListenOn(), cache)
```

Run your own web server as normal using groupcache to fill data from the
cache when needed.

The groupcache peers will be kept updated automatically whenever any GCE
instances are added or removed. The project, zone, tags and meta settings
are used to limit which instances should be considered part of the cluster.

## How it works

When a machine starts up the GCE metadata service is used to retrieve a list
of running instances within the project & zone and the configuration settings
used to filter these to those that should be considered part of the cluster.

A [clustering package](https://github.com/clockworksoul/smudge) based on
[SWIM](https://www.cs.cornell.edu/~asdas/research/dsn02-swim.pdf)
(Scalable Weakly-consistent Infection-style Membership) is then used to keep
track of which nodes are alive. As new instances start the existing nodes are
notified and dead nodes are removed.

The end result is that each node in the cluster maintains a complete list of
peers that is kept uptodate very quickly.

All this is done in a separate goroutine so startup of groupcache and the web
service is not delayed. The groupcache will initially be operating in 'local'
mode so the app will work fine while any cluster discovery is being performed.

## Example

See the example project for a working example.
