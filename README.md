# What is it?

Go library for writing and reading application logs, which can be used as a:

* event store
* commit log
* transaction log
* redo log

# Why it is needed?

* we need to store changes made to a big database after every (even tiny) change
* the most efficient way of storing the change is to store only the data that has been modified:
  this can be an event, command or transaction.

# Replication

Log can be replicated by copying recent entries into another log located on a different FS.

## Async replication

* quickly replicate changes to another log in order to support instant fail-over
* high availability - problem with replica does not interrupt the system, but potential data loss

## Sync replication

* not possible, because it is complicated and requires consensus/quorum/you name it

# File format

* Format
  * Name - 6 bytes
* Entries (for format named "log", version 1.0.0)
  * Start delimiter - "==>" - 3 bytes
  * Time - 16 bytes
  * CRC for header - 4 bytes 
  * Length - 2 bytes
  * Data
  * Checksum - 4 bytes

## Advantages of this file format

* Write data directly to disk (even when bigger than block size)
* Does not require index file
* low use of resources (both CPU and memory) 

## Disadvantages

* Overhead - 3 bytes for delimiter, 16 bytes for time, two CRCs
* Still possible that user data is: preamble + delimiter + time + checksum
  * someone can hack the database (header injection)
   * can be mitigated when CRC is calculated from two fields -> seed (from the beginning of the file) + time
  * the propability that data has correct header accidentally is very low 

## Find specific time in segment file

* Binary Search
 * read file in the middle (size/2)
 * find preamble, find delimiter
 * read time
 * calculate checksum for time
 * found time higher/lower than asked for? (go forward/backward) 
* What to do when checksum for time is not valid?
  * this means that either
    * this is not a beginning of a log entry
    * the log entry is corrupted
  * binary search should skip such entry and go to the next one
  * eventually the T found must be <= than the t which is looking for
  * when reading the first record should drop the T < t
  * when next record is corrupted should return error to the user 

# Why time  was used as a key?

* because it is human friendly
* because it is used for compaction (no need for index)
* when time is not forced during write, then current time is used or time > than last saved entry

# Why use segments at all?

* fast old log eviction by simply removing whole file

# Potential improvements

* is Log.Close() really needed? It's a pita to close log every single time
  * this can be used to flush current batch
* Implement batch write
  * why??? in order to limit number of syscalls (benchmarks showed that 2us takes one Write syscall)
  * Batch should be enabled by default
  * Append is done asynchronously 
  * special goroutine is responsible for writing 
  * when Append is called then entry is immediately written in the background
    * but the program go-routine does not wait for operation to finish
    * next time Append is called and previous write is still taking place then entry is batched
  * the main program will block on Append when batch is too big
    * we need a batch size parameter to control this behaviour
* cmd for interacting with logs
```shell
$ logstore ls --start-time yesterday --colorful
Time string....
Time string....
Time string....

$ logstore ls -f
Time string....
Time string....
Time string....

$ logstore ls --payload-only
string....
string....

$ logstore append "string..."

$ echo "string" | logstore append --force-time "2021-05-01T12:30:00Z" --stdin

$ logstore compact --older-than 10h
```

# Questions

* What to do when replicated log has a different event payload with same Time????
* should sync be done when closing the segment?
* should be possible to force sync?
  * when it can be useful? when some important event is saved?