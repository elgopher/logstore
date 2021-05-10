# What is it?

Go library for **writing** and **reading** append-only **application logs** storing events (**event store**), transactions (commit, redo, undo log) or any other entries.

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

## MVP - minimal number of features, not optimized yet, final API proposal

* [x] API for writing and reading entries from a log
* [x] Use segments in order to implement efficient compaction
* [x] Add segment limits
* [x] Implement compaction (manual and goroutine)
* [x] Reader should allow reading entries starting from given time
* [x] Add higher level functions for reading and writing using structs (instead of byte slices)

## To be done later

* [ ] Add replication to other filesystems
* [ ] Verify integrity using checksums
* [ ] Improve performance of Write by using batch
* [ ] Improve performance of Read with starting time option by using binary search
* [ ] Decrease number of allocations in Write, Read and codec
* [ ] CLI for listing entries and compaction
* [ ] Metrics
