// Package clock provides a System implementation of the wall clock that
// satisfies any single-method `Now() time.Time` interface. Each consumer
// declares its own Clock interface (Go convention for tiny interfaces);
// System is the production implementation injected by the composition root.
package clock

import "time"

type System struct{}

func (System) Now() time.Time { return time.Now() }
