[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-kafa-handler)
![Go Test](https://github.com/sensu/sensu-kafka-handler/workflows/Go%20Test/badge.svg)(https://github.com/sensu/sensu-kafka-handler/actions?query=workflow%3A%22Go+Test%22)
![goreleaser](https://github.com/sensu/sensu-kafka-handler/workflows/goreleaser/badge.svg)(https://github.com/sensu/sensu-kafka-handler/actions?query=workflow%3Agoreleaser)


***

# sensu-kafka-handler

## Table of Contents
- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
  - [Annotations](#annotations)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview
This handler translate Sensu events into Kafka key/value messages for a configurable Kafka topic. The full Sensu event is passed as the message value, using the event UUID as the message key 

Note handler is in in active development.

## Files

## Usage examples
```
./sensu-kafa-handler --help
Sensu handler to convert Sensu events to Kafka messages

Usage:
  sensu-kafa-handler [flags]
  sensu-kafa-handler [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -n, --dryrun         Dryrun, do not connect to Kafka broker
  -h, --help           help for sensu-kafa-handler
  -H, --host string    The Kafka broker host, defaults to value of KAFKA_HOST env variable (default "localhost:9092")
  -t, --topic string   Kafka topic to post to, defaults to value of KAFKA_TOPIC env variable (default "sensu-event")
  -v, --verbose        Verbose output to stdout, useful for testing

Use "sensu-kafa-handler [command] --help" for more information about a command.

```
## Configuration
`host` argument should point to the Kafka listener `<host:port>`. Currently only plaintext is supported
`topic` argument  should match an existing topic already created on the Kafka server.

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add sensu/sensu-kafa-handler
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/sensu/sensu-kafa-handler].


### Handler definition

```yml
---
type: Handler
api_version: core/v2
metadata:
  name: sensu-kafa-handler
  namespace: default
spec:
  command: sensu-kafa-handler --host localhost:9092 --topic sensu-events
  type: pipe
  runtime_assets:
  - sensu/sensu-kafa-handler
```

#### Proxy Support

This handler supports the use of the environment variables HTTP_PROXY,
HTTPS_PROXY, and NO_PROXY (or the lowercase versions thereof). HTTPS_PROXY takes
precedence over HTTP_PROXY for https requests.  The environment values may be
either a complete URL or a "host[:port]", in which case the "http" scheme is assumed.

### Annotations

The `host` and `topic` arguments for this handler are tunable on a per entity or check basis based on annotations.  The
annotations keyspace for this handler is `sensu.io/plugins/sensu-kafa-handler/config`.

#### Examples

To change the example argument for a particular check, for that checks's metadata add the following:

```yml
type: CheckConfig
api_version: core/v2
metadata:
  annotations:
    sensu.io/plugins/sensu-kafa-handler/config/topic: "custom-topic"
[...]
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-kafa-handler repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu-community/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/handler-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/handler-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/handlers/
[7]: https://github.com/sensu-community/handler-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-community/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
