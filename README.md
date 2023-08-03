# udpp
A Point-to-Point UDP Tunnel.

## Usage

Generate an example config file:

```shell
udpp new
```

A new file `config.yml` was created in current directory.

```yaml
id: 1f0e4346-b633-4525-af3c-33f8f4aa841b
server: redis://127.0.0.1:6379
local: 127.0.0.1:9999
```

The server only supports redis for now.

### Server

The following configuration exposes a service listening on udp port 9999.

```yaml
id: 1f0e4346-b633-4525-af3c-33f8f4aa841b
server: redis://127.0.0.1:6379
local: 127.0.0.1:9999
```

### Client

Access your server using `peer` config:

```yaml
id: e5958a55-1c94-4368-82fa-86de1e1af24a
server: redis://127.0.0.1:6379
local: 127.0.0.1:5321
peer:
  id: 1f0e4346-b633-4525-af3c-33f8f4aa841b
  bind: 127.0.0.1:6565
```

The client listens on `127.0.0.1:6565` for incoming traffic and forwards to the target peer.

Note that `127.0.0.1:5321` is the source address of your local visitor client.
