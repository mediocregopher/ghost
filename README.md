# ghost

Durable-but-not-reliable remote message passing for go.

Ghost is a small library that lets you create long-lived client/servers with the
following properties:

* Automatic, transparrent connection resurrection

* Messages can be any go structure, not just strings or bytes. Encoding and
  decoding is done transparently

* Messages only travel from client to server. To go the other way a listen
  must be created on the client and a response message is sent to that. How you
  route messages is up to you.

Ghost is useful where you have a cluster of processes that you want to have
long-running communication connections to each other on.

# Usage

`go get github.com/mediocregopher/ghost`

or

[goat][0]:
```json
{
    "loc":"https://github.com/mediocregopher/ghost.git",
    "type":"git",
    "reference":"v0.2.1",
    "path":"github.com/mediocregopher/ghost"
}
```

Then when you want to use it import `github.com/mediocregopher/ghost/src/ghost`

# Docs

Check out [docs][1] for externally available methods.  Also check out the
[example](/example) code to see actual usage.

[0]: http://github.com/mediocregopher/goat
[1]: http://godoc.org/github.com/mediocregopher/ghost
