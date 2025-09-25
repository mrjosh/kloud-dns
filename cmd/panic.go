package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
)

var panicMutex sync.Mutex

// This output is shown if a panic happens.
const panicOutput = `
 _    _                 _ 
| | _| | ___  _   _  __| |
| |/ / |/ _ \| | | |/ _' |
|   <| | (_) | |_| | (_| |
|_|\_\_|\___/ \__,_|\__,_|
                          

kloud crashed! This is always indicative of a bug within kloud.
`

// PanicHandler is called to recover from an internal panic in kloud, and
// augments the standard stack trace with a more user friendly error message.
// PanicHandler must be called as a defered function, and must be the first
// defer called at the start of a new goroutine.
func PanicHandler() {
	// Have all managed goroutines checkin here, and prevent them from exiting
	// if there's a panic in progress. While this can't lock the entire runtime
	// to block progress, we can prevent some cases where kloud may return
	// early before the panic has been printed out.
	panicMutex.Lock()
	defer panicMutex.Unlock()

	recovered := recover()
	if recovered == nil {
		return
	}

	fmt.Fprint(os.Stderr, panicOutput)
	fmt.Fprint(os.Stderr, recovered, "\n")

	// When called from a deferred function, debug.PrintStack will include the
	// full stack from the point of the pending panic.
	debug.PrintStack()

	// An exit code of 11 keeps us out of the way of the detailed exitcodes
	// from plan, and also happens to be the same code as SIGSEGV which is
	// roughly the same type of condition that causes most panics.
	os.Exit(11)
}
