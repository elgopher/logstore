# What is it?

Go library for writing and reading append-only application logs, which can be used as an:

* event store
* commit log
* transaction log
* redo log

# Why it is needed?

Let's say you have a large data structure which is modified by some command. After such modificaton you can either:

* save a whole snapshot of data to disk (simple to implement, but not efficient if data structure is big)
* or store the actual change in a form of event, command or transaction (much more efficient, harder to implement)

Logstore is an API for storing and retrieving such entries.

# Install

`go get -u github.com/jacekolszak/logstore`

# Quick Start

See [example/write/main.go](example/write/main.go). More examples in [example directory](example).

# Project Status

The library is under heavy development, not ready for production use yet.

# Project Plan

* [x] API for writing and reading entries from a log
* [x] Use segments in order to implement efficient compaction
* [x] Add segment max duration 
* [x] Implement compaction (manual and goroutine)
* [ ] Reader should allow reading entries starting from given time
* [ ] Verify integrity using checksums
* [ ] Add replication to other filesystems
* [ ] Add higher level functions for reading and writing using structures (instead of byte slices)  
* [ ] Improve performance of Write by using batch
* [ ] CLI for listing entries and compaction
* [ ] Metrics
