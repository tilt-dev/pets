## How to write a Petsfile

When `pets` starts, it reads the `Petsfile` in the current directory.

The `Petsfile` contains the configuration for how to start your servers. A typical `Petsfile` looks like:

```python
def backend_local():
  server = start("go run ./cmd/backend/main.go")
  return service(server, "localhost", 8080)

backend = "backend"
register(backend, "local", backend_local)

def frontend_local(b):
  server = start("go run ./cmd/frontend/main.go -- --backend=%s", b["host"])
  return service(server, "localhost", 8081)

frontend = "frontend"
register(frontend, "local", frontend_local, deps=[backend])
```

### Configuration Language

A `Petsfile` is written in Skylark.

Skylark is a configuration language designed by Google, most notably used in
Bazel. Skylark looks a lot like Python, but with most Python features removed.
Skylark is less powerful than a general-purpose programming langage but more
expressive than a typical JSON or YAML file.

For more detail, read Bazel's 
[introduction to Skylark](https://docs.bazel.build/versions/master/skylark/language.html).

### Built-in functions

#### run(cmd)

Runs a shell script, and waits until the shell script completes.

If the shell script has a non-zero exit code, `pets` will fail.

The current working directory of the shell script is the directory of the current Petsfile.

Arguments:

```
  cmd: string
```

Returns: None

#### start(cmd)

Starts a shell script in the background, returning immediately with info on the running process.

The current working directory of the shell script is the directory of the current Petsfile.

Arguments:

```
  cmd: string
```

Returns: A dictionary with a field "pid" with the process id.


#### service(server, hostname, port)

Tells pets that we expect the server to listen on the given hostname and port.

Pets will wait until the server passes a TCP health check. If the process dies before
it passes the health check, `pets` will fail.

Arguments:

```
  server: A process object returned by `start`
  hostname: string, the hostname of the running process, often "localhost"
  port: int, the port of the running process
```

Returns: A dictionary with fields "pid",  "hostname", "port", and "host".

#### register(name, tier, provider, deps)

Registers a function for starting a server. Once the function is registered, you
run it with `pets up <my-server-name>`

Arguments:

```
  name: `string`, the name of the server
  tier: `string`, a tier for the server. Most CLI commands default to "local" tier.
    A server can have multiple providers under different tiers (e.g., "local", "minikube", etc.)
  provider: `function`, a function to run to start the service
  deps: a list of `string`s. If specified, pets will automatically start those servers
    before running `provider`, and pass them as arguments to the provider function.
```

Returns: `None`

#### print(msg)

Prints a message to standard error.

#### load(path)

Loads the Petsfile at the given path. 

Used when your server depends on servers in other repos or directories.

Arguments:

```
  path: A path to a directory with a Petsfile. Must be a relative path.

        Alternative URL schemes allow you to fetch remote repositories:

        `go-get://path/to/repo`: Fetch a remote repo with `go get'
```

Returns: `None`

