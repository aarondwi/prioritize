# prioritize

Prioritizing some tasks over the others, to prevent higher latency for more important tasks.
Primarily intended for usages in business logic webserver/batch/pipeline, in which some customer
or their orders should be prioritized over the others.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Installation
-------------------------

```bash
go get -u github.com/aarondwi/prioritize
```

Usages
-------------------------

See the [tests](https://github.com/aarondwi/prioritize/blob/main/engine_test.go) directly for the most up-to-date example.

Notes
-------------------------

1. This library only does local prioritization. So your app will still parse the message before coming to this library. That means that this solution is not for load-shedding, but instead only to give better latency to a proportion of users.
2. This library try to make internal queue as allocation-free as possible, but as it is intended for webserver/batch/pipeline, some allocation should be expected (as the path not that critical). Allocations are used for task mapping (ofc, all references are removed automatically after used).
3. There would be **NO** panic handling, as imo, it is bad practice. `panic` should only be used if the application, for some external reason, can't continue at all (e.g. OOM, disk full, etc). Handling this means going forward in a very unrecoverable, broken state, and it is dangerous.
4. The internal queue (if you choose to implement one yourself, implement `QInterface`) should (for the built-in, is) goroutine-safe. Mostly using locks, so expect around 5-10 million push/pop per second. We probably can make it faster (a la [disruptor](https://lmax-exchange.github.io/disruptor/)), but given for business logic application usage, my target is around 20K/s, which is already far surpassed.

Built-in Supported Queues
-------------------------

1. [Heap](https://github.com/aarondwi/prioritize/tree/main/heap): Bounded, priority straight based on higher priority first
2. [Roundrobin](https://github.com/aarondwi/prioritize/tree/main/roundrobin): Bounded, priority is roundrobin, starting from first item put. The same priority is prioritized last after that

TODO
-------------------------

1. Allow tuning of worker/queue size, dynamically (or preferably, via dynamic concurrency-limit).
2. Add new interface (allow kicking lower priority job when full)
3. Add put timeout, instead of straight error
