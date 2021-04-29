# What is it?

Go library for writing and reading append-only application logs, which can be used as an:

* event store
* commit log
* transaction log
* redo log

# Why it is needed?

* we need to store changes made to a big database after every (even tiny) change
* the most efficient way of storing the change is to store only the data that has been modified:
  this can be an event, command or a transaction.

# Install

`go get -u github.com/jacekolszak/logstore`

# Quick Start

See [example/write/main.go](example/write/main.go). More examples in [example directory](example).

# Project Status

The library is under heavy development, not ready for production use.

# Project Plan

* [ ] Use segments in order to implement efficient compaction
* [ ] Implement compaction 
* [ ] Reader should allow to read entries starting from given time
* [ ] Verify integrity using checksums
* [ ] Add replication to other filesystems
* [ ] Add higher level functions for reading and writing using structures (instead of byte slices)  
* [ ] Improve performance of Write by using batch
* [ ] CLI for listing entries and compaction
