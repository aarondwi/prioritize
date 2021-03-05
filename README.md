# prioritize

Prioritizing some tasks over the others, to prevent higher latency for more important tasks.
Primarily intended for usages in webserver/batch/pipeline, in which some customer
or their orders should be prioritized over the others.

[![Build Status](https://travis-ci.com/aarondwi/prioritize.svg?branch=main)](https://travis-ci.com/aarondwi/prioritize)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

usages
-------------------------

See the [tests](https://github.com/aarondwi/prioritize/blob/main/engine_test.go) directly for the most up-to-date example.

notes
-------------------------

1. This library only does local prioritization, making it easier to scale (no cross network consensus needed)
2. This library try to make internal queue as allocation-free as possible, but as it is intended for webserver/batch/pipeline, some allocation should be expected (as the path not that critical). Allocations are used for task mapping (ofc, all references are removed automatically after used)
3. There would be **NO** panic handling, as imo, it is bad practice. `panic` should only be used if the application, for some external reason, can't continue at all (e.g. OOM, disk full, etc). Handling this means going forward in a very unrecoverable, broken state, and it is dangerous.

todo
-------------------------

1. Add `Fair` queue. This type of queue has starvation prevention mechanism.
2. Allow tuning of worker/queue size, dynamically (or preferably, via dynamic concurrency-limit).
