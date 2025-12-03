# book-service

## Overview

A simple book creation service. Entry points via GRPC and telegram bot. 
The service consists of 3 applications:
- the main place where the grpc service rises,
- bot, creating books through telegram bot commands,
- publisher, sending events to kafka.

Technologies used. 
- logs are sent to Graylog,
- metrics and traces are implemented through OpenTelemetry, sending to Prometheus and Jaeger respectively,
- events are sent to Kafka.

## Quick start

### Local development

```sh
# Run app main
make run-app
# Run app publisher
make run-publisher
# Run app bot
make run-bot
```