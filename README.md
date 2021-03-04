# prioritize
Prioritizing some tasks over the others, to prevent higher latency for more important tasks.
Primarily intended for usages in webserver/batch/pipeline, in which some customer
or their orders should be prioritized over the others.

notes
-------------------------

1. This library only does local prioritization, making it easier to scale (no cross network consensus needed)
2. This library try to make internal queue as allocation-free as possible, but as it is intended for webserver/batch/pipeline, some allocation should be expected (as the path not that critical). Allocations are used for task mapping (ofc, all references are removed automatically after used)

todo
-------------------------

0. Add core implementation, in which user only needs to submit task
1. Add `Fair` queue. This type of queue has starvation prevention mechanism.
2. Allow tuning of worker/queue size, dynamically (or preferable, via dynamic concurrency-limit)
3. Add some badge, docs, CI pipeline, etc
