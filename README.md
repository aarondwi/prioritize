# prioritize
Prioritizing some tasks over the others, to prevent higher latency for more important tasks.
Primarily intended for usages in webserver or batch, in which some customer
should be prioritized over the others.

todo
-------------------------

0. Add core implementation, in which user only needs to submit task
1. Add `RoundRobin` and `Fair` queue. This type of queue has starvation prevention mechanism.
2. Allow tuning of worker/queue size, dynamically (or preferable, via dynamic concurrency-limit)
