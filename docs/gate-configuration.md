# FrostFS REST Gateway configuration file

This section contains detailed FrostFS REST Gateway configuration file description
including default config values and some tips to set up configurable values.

There are some custom types used for brevity:

* `duration` -- string consisting of a number and a suffix. Suffix examples include `s` (seconds), `m` (minutes), `ms` (
  milliseconds).

# Structure

| Section      | Description                                     |
|--------------|-------------------------------------------------|
| no section   | [General parameters](#general-section)          |
| `wallet`     | [Wallet configuration](#wallet-section)         |
| `pool`       | [Pool configuration](#pool-section)             |
| `logger`     | [Logger configuration](#logger-section)         |
| `pprof`      | [Pprof configuration](#pprof-section)           |
| `prometheus` | [Prometheus configuration](#prometheus-section) |

# General section

```yaml
scheme: [ http ]
cleanup-timeout: 10s
graceful-timeout: 15s
max-header-size: 1000000

listen-address: localhost:8080
listen-limit: 0
keep-alive: 3m
read-timeout: 30s
write-timeout: 30s

tls-listen-address: localhost:8081
tls-certificate: /path/to/tls/cert
tls-key: /path/to/tls/key
tls-ca: /path/to/tls/ca
tls-listen-limit: 0
tls-keep-alive: 3m
tls-read-timeout: 30s
tls-write-timeout: 30s
```

| Parameter              | Type       | Default value    | Description                                                                                                                                                                         |
|------------------------|------------|------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `scheme`               | `[]string` | `[http]`         | The listeners to enable, this can be repeated and defaults to the schemes in the swagger spec.                                                                                      |
| `cleanup-timeout`      | `duration` | `10s`            | Grace period for which to wait before killing idle connections.                                                                                                                     |
| `graceful-timeout`     | `duration` | `15s`            | Grace period for which to wait before shutting down the server.                                                                                                                     |
| `max-header-size`      | `int`      | `1000000`        | Controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body. |
| `listen-address`       | `string`   | `localhost:8080` | The IP and port to listen on.                                                                                                                                                       |
| `listen-limit`         | `int`      | `0`              | Limit the number of outstanding requests. `0` means no limit                                                                                                                        |                                                                                                                         |
| `keep-alive`           | `duration` | `3m`             | Sets the TCP keep-alive timeouts on accepted connections.                                                                                                                           |
| `read-timeout`         | `duration` | `30s`            | Maximum duration before timing out read of the request. It prunes dead TCP connections (e.g. closing laptop mid-download).                                                          |
| `write-timeout`        | `duration` | `30s`            | Maximum duration before timing out write of the response.                                                                                                                           |
| `tls-listen-address`   | `string`   | `localhost:8081` | The IP and port to listen on for TLS.                                                                                                                                               |
| `tls-certificate`      | `string`   |                  | The certificate file to use for secure connections.                                                                                                                                 |
| `tls-key`              | `string`   |                  | The private key file to use for secure connections (without passphrase).                                                                                                            |
| `tls-ca`               | `string`   |                  | The certificate authority certificate file to be used with mutual tls auth.                                                                                                         |
| `tls-listen-limit`     | `int`      | `0`              | Limit the number of outstanding requests for TLS. `0` means no limit                                                                                                                |                                                                                                                         |
| `tls-keep-alive`       | `duration` | `3m`             | Sets the TCP keep-alive timeouts on accepted connections for TLS.                                                                                                                   |
| `tls-read-timeout`     | `duration` | `30s`            | Maximum duration before timing out read of the request for TLS. It prunes dead TCP connections (e.g. closing laptop mid-download).                                                  |
| `tls-write-timeout`    | `duration` | `30s`            | Maximum duration before timing out write of the response for TLS.                                                                                                                   |

# `wallet` section

```yaml
wallet:
  path: /path/to/wallet.json
  address: NfgHwwTi3wHAS8aFAN243C5vGbkYDpqLHP
  passphrase: pwd
```

| Parameter    | Type     | Default value | Description                                                              |
|--------------|----------|---------------|--------------------------------------------------------------------------|
| `path`       | `string` |               | Path to the wallet.                                                      |
| `address`    | `string` |               | Account address to get from wallet. If omitted default one will be used. |
| `passphrase` | `string` |               | Passphrase to decrypt wallet.                                            |

# `pool` section

```yaml
pool:
  node-dial-timeout: 10s
  healthcheck-timeout: 15s
  rebalance-timer: 60s
  error-threshold: 100

  # Nodes configuration
  # This configuration makes the gateway use the first node (node1.frostfs:8080)
  # while it's healthy. Otherwise, gateway uses the second node (node2.frostfs:8080)
  # for 10% of requests and the third node (node3.frostfs:8080) for 90% of requests.
  # Until nodes with the same priority level are healthy
  # nodes with other priority are not used.
  # The lower the value, the higher the priority.
  peers:
    0:
      address: node1.frostfs:8080
      priority: 1
      weight: 1
    1:
      address: node2.frostfs:8080
      priority: 2
      weight: 0.1
    2:
      address: node3.frostfs:8080
      priority: 2
      weight: 0.9
```

| Parameter             | Type       | Default value | Description                                                                     |
|-----------------------|------------|---------------|---------------------------------------------------------------------------------|
| `node-dial-timeout`   | `duration` | `10s`         | Timeout to connect to a node.                                                   |
| `healthcheck-timeout` | `duration` | `15s`         | Timeout to check node health during rebalance.                                  |
| `rebalance-timer`     | `duration` | `60s`         | Interval to check node health.                                                  |
| `error-threshold`     | `uint32`   | `100`         | The number of errors on connection after which node is considered as unhealthy. |

## `peers` section

| Parameter              | Type       | Default value | Description                                                                                                                                             |
|------------------------|------------|---------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| `address`              | `string`   |               | Address of storage node.                                                                                                                                |
| `priority`             | `int`      | `1`           | It allows to group nodes and don't switch group until all nodes with the same priority will be unhealthy. The lower the value, the higher the priority. |
| `weight`               | `float`    | `1`           | Weight of node in the group with the same priority. Distribute requests to nodes proportionally to these values.                                        |

# `logger` section

```yaml
logger:
  level: debug
```

| Parameter | Type     | Default value | Description                                                                                        |
|-----------|----------|---------------|----------------------------------------------------------------------------------------------------|
| `level`   | `string` | `debug`       | Logging level.<br/>Possible values:  `debug`, `info`, `warn`, `error`, `dpanic`, `panic`, `fatal`. |

# `pprof` section

Contains configuration for the `pprof` profiler.

```yaml
pprof:
  enabled: true
  address: localhost:8091
```

| Parameter | Type     | Default value    | Description                             |
|-----------|----------|------------------|-----------------------------------------|
| `enabled` | `bool`   | `false`          | Flag to enable the service.             |
| `address` | `string` | `localhost:8091` | Address that service listener binds to. |

# `prometheus` section

Contains configuration for the `prometheus` metrics service.

```yaml
prometheus:
  enabled: true
  address: localhost:8092
```

| Parameter | Type     | Default value    | Description                             |
|-----------|----------|------------------|-----------------------------------------|
| `enabled` | `bool`   | `false`          | Flag to enable the service.             |
| `address` | `string` | `localhost:8092` | Address that service listener binds to. |
