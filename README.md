# httptap

## serve command

```
--address
    The address the proxy listens on, defaults to ":8080"

--upstream
    The service that we tap. Must be a valid url.

--loglevel
    The log level. Valid values are ERROR, INFO, WARN and DEBUG.
    Defaults to INFO

--logfile
    File to log to, defaults to /dev/stdout

--config-file
    Name of the file that contains the yaml configuration.
```

The audit records are interleaved with the system log records.

To easily filter them, the audit info is put in a group for each tap.

```json
{
  "tap_0": {
    "paths":["path1"],
    "data
  }
}
```

## Specification schema

```yaml
listenAddress: 0.0.0.0:18080
logging:
  logLevel: info
  logFile: /dev/stdout
upstream: http://localhost:8080
header:                      # Include and exclude can be both specified
  exclude: ["Authorization"] # Exclude always
  include: ["X-Api-Key"]     # Include
taps:
  - name: log tap
    patterns:
      - "PUT /"
      - "GET /"
    logTap: 
      logFile: /dev/stdout
    body: true
    bodyPatch: |-       # Json path written in Yaml, like Kustomize does
      - op: add
        path: /add/here/key
        value: value
    header:
      exclude: ["Authorization"]
      include: ["X-Api-Key"]
  - name: template tap
    patterns:
      - "PUT /"
      - "GET /"
    templateTap: 
      logFile: /dev/stdout
      template: |-
        multiline
        template is here
```

