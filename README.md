# What is it?

Go library for writing and reading application logs, which can be used as an:

* event store
* commit log
* transaction log
* redo log

# Why it is needed?

* we need to store changes made to a big database after every (even tiny) change
* the most efficient way of storing the change is to store only the data that has been modified:
  this can be an event, command or a transaction.
