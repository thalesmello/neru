#ifndef NERU_CLICK_SEQUENCE_H
#define NERU_CLICK_SEQUENCE_H

/// Bound native click state to single/double/triple. A fourth click begins a
/// new sequence; a different button/modifier/point or expired interval does too.
static inline int NeruClickSequenceCount(int previous, int continuation) {
	return continuation && previous > 0 && previous < 3 ? previous + 1 : 1;
}

#endif
