//+build DEBUG

package maps

// DEBUG is used for debugging builds
const DEBUG = true

// Displaceable debug allow users to walk through walls
func (w *Wall) Displace() bool { return true }
