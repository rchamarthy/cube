# Introduction

A simple and generic messaging framework with some guarantees.

>Implementation of [spec][spec]

[spec]: https://cto-github.cisco.com/atom/rfc/blob/master/proposals/messaging.md

# Design

The two key concepts are **MsgBus** (message bus) and **Msg** (message). The former abstracts the communication bus and 
the latter abstracts how payload is sent and received.
These are de-coupled in the architecture so that developers can specify semantics
and requirements on MsgBus and Msg independently. Furthermore, the implementation lends itself to be changed.

# Candidate Message Bus Implementations

## RabbitMQ
Erlang based [project][rabbitmq].

[rabbitmq]: https://www.rabbitmq.com/

Docker [image and instructions][docker-rabbitmq]

[docker-rabbitmq]: https://hub.docker.com/_/rabbitmq/

## NATS
Golang based [project][nats.io].

[nats.io]: https://nats.io/documentation/streaming/nats-streaming-intro/

Docker [image and instructions][docker-nats].

[docker-nats]: https://hub.docker.com/_/nats-streaming/

## Kafka
Java-based Apache [project][kafka].

[kafka]: https://kafka.apache.org/

# Basic Usage

## Receiver-side 
```
import "cto-github.cisco.com/rchincha/messaging"

mb := RegisterMsgBus()  // opens a connection context to our favorite msg bus

func testMsgHandler(data []byte, respExpected bool) ([]byte, error) {
    if respExpected {
        // this payload arrived as a request-reply, so return a response!
        return []byte("this is a reply"), nil
    } else {
        // this payload arrives a a fire-and-forget, so nothing more to do!
    }
    return nil, nil
}

err := RegisterMsgHandler("target", testMsgHandler)

// stay alive so we can receive and process msgs in testMsgHandler()

// time to tune out now
err := UnregisterMsgHandler("target")

// we are done now!
UnregisterMsgBus(mb)

```

## Sender-side
```
import "cto-github.cisco.com/rchincha/messaging"

mb := RegisterMsgBus()  // opens a connection context to our favorite msg bus

// send and block-wait for response here
r, h, err := SendAndWaitResponse([]byte("this is a request"), "target", timeout)
// r: response payload
// h: handle that was used to help track conversation in logs
// err: error
// timeout: 0 => block forever, else timeout

// fire-and-forget
err := Send([]byte("firing and forgetting"), "target")

// we are done now!
UnregisterMsgBus(mb)

```

Please see _messaging_test.go_ for more example usages.

# Testing

## RabbitMQ

$ docker run -it -p 4369:4369 -p 5671:5671 -p 5672:5672 -p 25672:25672 --hostname my-rabbit --name some-rabbit rabbitmq:latest

$ go test

$ go test -bench .

## NATS
$ go get -u github.com/nats-io/nats-streaming-server

$ $GOPATH/bin/nats-streaming-server -cid "atom-msgbus" -mc 0 -m 8222

$ go test

$ go test -bench .

# Changelog
## 0.1
* Basic API implementation
* Support for RabbitMQ broker
* Support for NATS streaming broker

# TODO
* Better error handling and response codes/behavior
* Stats
