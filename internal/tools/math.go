// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package tools

// return max of two ints
func MathMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// return min of two ints
func MathMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
