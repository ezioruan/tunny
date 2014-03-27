Skank
=====

Skank is a golang library for creating and managing a thread pool of fixed size. Skank attempts to provide a simple and intuitive interface for defining and using a pool of dedicated, independant worker threads.

Use cases for Skank are any situation where a machine has multiple cores available and needs to distribute a large volume of work across them. Skank isn't intended as a pool for jobs that are IO heavy or involve other long sleep states, since this results in needlessly blocking threads when no real work is being performed.
