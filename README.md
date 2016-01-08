# hfm (High Frequency Monitor)

hfm is an application to run tests in parallel at a high frequency. If the
outcome of the test results in a state change, other commands can be executed.

It is designed to be a general purpose tool, by having both the tests and the
state change commands be interpreted by a shell, such as /bin/sh.

An example application is to poll other network services for health, and to
take actions based on their health status changes.

## Design

hfm is not currently a real-time monitor.

Delays are managed using Go's time.Sleep, which currently only guarantees it
"pauses the current goroutine for at least the duration d".  Therefore tests
will run by "at least" the interval you specify.  This ensures that we are not
creating a flood of tests, should a test execute for longer than the
specified interval.

This may give a false confidence about any statistics that are 
generated by the system, due to Coordinated Omission.

## Building

There's a patch-local-go-libucl make target that will allow you to use the 
locally installed libucl vs. a vendorized version.

