package harmonica

// This file defines a simplified damped harmonic oscillator, colloquially
// known as a spring. This is ported from Ryan Juckett’s simple damped harmonic
// motion, originally written in C++.
//
// Example usage:
//
//     // Run once to initialize.
//     spring := NewSpring(FPS(60), 6.0, 0.2)
//
//     // Update on every frame.
//     pos := 0.0
//     velocity := 0.0
//     targetPos := 100.0
//     someUpdateLoop(func() {
//         pos, velocity = spring.Update(pos, velocity, targetPos)
//     })
//
// For background on the algorithm see:
// https://www.ryanjuckett.com/damped-springs/

/******************************************************************************

  Copyright (c) 2008-2012 Ryan Juckett
  http://www.ryanjuckett.com/

  This software is provided 'as-is', without any express or implied
  warranty. In no event will the authors be held liable for any damages
  arising from the use of this software.

  Permission is granted to anyone to use this software for any purpose,
  including commercial applications, and to alter it and redistribute it
  freely, subject to the following restrictions:

  1. The origin of this software must not be misrepresented; you must not
     claim that you wrote the original software. If you use this software
     in a product, an acknowledgment in the product documentation would be
     appreciated but is not required.

  2. Altered source versions must be plainly marked as such, and must not be
     misrepresented as being the original software.

  3. This notice may not be removed or altered from any source
     distribution.

*******************************************************************************

  Ported to Go by Charmbracelet, Inc. in 2021.

******************************************************************************/

import (
	"math"
	"time"
)

// FPS returns a time delta for a given number of frames per second. This
// value can be used as the time delta when initializing a Spring. Note that
// game engines often provide the time delta as well, which you should use
// instead of this function, if possible.
//
// Example:
//
//     spring := NewSpring(FPS(60), 5.0, 0.2)
//
func FPS(n int) float64 {
	return (time.Second / time.Duration(n)).Seconds()
}

// In calculus ε is, in vague terms, an arbitrarily small positive number. In
// the original C++ source ε is represented as such:
//
//     const float epsilon = 0.0001
//
//  Some Go programmers use:
//
//     const epsilon float64 = 0.00000001
//
// We can, however, calculate the machine’s epsilon value, with the drawback
// that it must be a variable versus a constant.
var epsilon = math.Nextafter(1, 2) - 1

// Spring contains a cached set of motion parameters that can be used to
// efficiently update multiple springs using the same time step, angular
// frequency and damping ratio.
//
// To use a Spring call New with the time delta (that's animation frame
// length), frequency, and damping parameters, cache the result, then call
// Update to update position and velocity values for each spring that neeeds
// updating.
//
// Example:
//
//     // First precompute spring coefficients based on your settings:
//     var x, xVel, y, yVel float64
//     deltaTime := FPS(60)
//     s := NewSpring(deltaTime, 5.0, 0.2)
//
//     // Then, in your update loop:
//     x, xVel = s.Update(x, xVel, 10) // update the X position
//     y, yVel = s.Update(y, yVel, 20) // update the Y position
//
type Spring struct {
	posPosCoef, posVelCoef float64
	velPosCoef, velVelCoef float64
}

// NewSpring initializes a new Spring, computing the parameters needed to
// simulate a damped spring over a given period of time.
//
// The delta time is the time step to advance; essentially the framerate.
//
// The angular frequency is the angular frequency of motion, which affects the
// speed.
//
// The damping ratio is the damping ratio of motion, which determines the
// oscillation, or lack thereof. There are three categories of damping ratios:
//
// Damping ratio > 1: over-damped.
// Damping ratio = 1: critlcally-damped.
// Damping ratio < 1: under-damped.
//
// An over-damped spring will never oscillate, but reaches equilibrium at
// a slower rate than a critically damped spring.
//
// A critically damped spring will reach equilibrium as fast as possible
// without oscillating.
//
// An under-damped spring will reach equilibrium the fastest, but also
// overshoots it and continues to oscillate as its amplitude decays over time.
func NewSpring(deltaTime, angularFrequency, dampingRatio float64) (s Spring) {
	// Keep values in a legal range.
	angularFrequency = math.Max(0.0, angularFrequency)
	dampingRatio = math.Max(0.0, dampingRatio)

	// If there is no angular frequency, the spring will not move and we can
	// return identity.
	if angularFrequency < epsilon {
		s.posPosCoef = 1.0
		s.posVelCoef = 0.0
		s.velPosCoef = 0.0
		s.velVelCoef = 1.0
		return s
	}

	if dampingRatio > 1.0+epsilon {
		// Over-damped.
		var (
			za = -angularFrequency * dampingRatio
			zb = angularFrequency * math.Sqrt(dampingRatio*dampingRatio-1.0)
			z1 = za - zb
			z2 = za + zb

			e1 = math.Exp(z1 * deltaTime)
			e2 = math.Exp(z2 * deltaTime)

			invTwoZb = 1.0 / (2.0 * zb) // = 1 / (z2 - z1)

			e1_Over_TwoZb = e1 * invTwoZb
			e2_Over_TwoZb = e2 * invTwoZb

			z1e1_Over_TwoZb = z1 * e1_Over_TwoZb
			z2e2_Over_TwoZb = z2 * e2_Over_TwoZb
		)

		s.posPosCoef = e1_Over_TwoZb*z2 - z2e2_Over_TwoZb + e2
		s.posVelCoef = -e1_Over_TwoZb + e2_Over_TwoZb

		s.velPosCoef = (z1e1_Over_TwoZb - z2e2_Over_TwoZb + e2) * z2
		s.velVelCoef = -z1e1_Over_TwoZb + z2e2_Over_TwoZb

	} else if dampingRatio < 1.0-epsilon {
		// Under-damped.
		var (
			omegaZeta = angularFrequency * dampingRatio
			alpha     = angularFrequency * math.Sqrt(1.0-dampingRatio*dampingRatio)

			expTerm = math.Exp(-omegaZeta * deltaTime)
			cosTerm = math.Cos(alpha * deltaTime)
			sinTerm = math.Sin(alpha * deltaTime)

			invAlpha = 1.0 / alpha

			expSin                     = expTerm * sinTerm
			expCos                     = expTerm * cosTerm
			expOmegaZetaSin_Over_Alpha = expTerm * omegaZeta * sinTerm * invAlpha
		)

		s.posPosCoef = expCos + expOmegaZetaSin_Over_Alpha
		s.posVelCoef = expSin * invAlpha

		s.velPosCoef = -expSin*alpha - omegaZeta*expOmegaZetaSin_Over_Alpha
		s.velVelCoef = expCos - expOmegaZetaSin_Over_Alpha

	} else {
		// Critically damped.
		var (
			expTerm     = math.Exp(-angularFrequency * deltaTime)
			timeExp     = deltaTime * expTerm
			timeExpFreq = timeExp * angularFrequency
		)

		s.posPosCoef = timeExpFreq + expTerm
		s.posVelCoef = timeExp

		s.velPosCoef = -angularFrequency * timeExpFreq
		s.velVelCoef = -timeExpFreq + expTerm
	}

	return s
}

// Update updates position and velocity values against a given target value.
// Call this after calling NewSpring to update values.
func (s Spring) Update(pos, vel float64, equilibriumPos float64) (newPos, newVel float64) {
	oldPos := pos - equilibriumPos // update in equilibrium relative space
	oldVel := vel

	newPos = oldPos*s.posPosCoef + oldVel*s.posVelCoef + equilibriumPos
	newVel = oldPos*s.velPosCoef + oldVel*s.velVelCoef

	return newPos, newVel
}
