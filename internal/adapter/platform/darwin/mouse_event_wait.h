#ifndef NERU_MOUSE_EVENT_WAIT_H
#define NERU_MOUSE_EVENT_WAIT_H

#include <errno.h>
#include <time.h>

/// Honor mouse-event pacing without running callbacks between down and up.
static inline void NeruWaitForMouseEvents(double seconds) {
	if (seconds <= 0)
		return;
	struct timespec delay = {
	    .tv_sec = (time_t)seconds,
	    .tv_nsec = (long)((seconds - (time_t)seconds) * 1000000000.0),
	};
	while (nanosleep(&delay, &delay) != 0 && errno == EINTR) {
	}
}

#endif
