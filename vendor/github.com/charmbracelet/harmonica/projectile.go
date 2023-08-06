package harmonica

// This file defines simple physics projectile motion.
//
// Example usage:
//
//    // Run once to initialize.
//    projectile := NewProjectile(
//        FPS(60),
//        Point{6.0, 100.0, 0.0},
//        Vector{2.0, 0.0, 0.0},
//        Vector{2.0, -9.81, 0.0},
//    )
//
//    // Update on every frame.
//    someUpdateLoop(func() {
//        pos := projectile.Update()
//    })
//
// For background on projectile motion see:
// https://en.wikipedia.org/wiki/Projectile_motion

// Projectile is the representation of a projectile that has a position on
// a plane, an acceleration, and velocity.
type Projectile struct {
	pos       Point
	vel       Vector
	acc       Vector
	deltaTime float64
}

// Point represents a point containing the X, Y, Z coordinates of the point on
// a plane.
type Point struct {
	X, Y, Z float64
}

// Vector represents a vector carrying a magnitude and a direction. We
// represent the vector as a point from the origin (0, 0) where the magnitude
// is the euclidean distance from the origin and the direction is the direction
// to the point from the origin.
type Vector struct {
	X, Y, Z float64
}

// Gravity is a utility vector that represents gravity in 2D and 3D contexts,
// assuming that your coordinate plane looks like in 2D or 3D:
//
//   y             y ±z
//   │             │ /
//   │             │/
//   └───── ±x     └───── ±x
//
// (i.e. origin is located in the bottom-left corner)
var Gravity = Vector{0, -9.81, 0}

// TerminalGravity is a utility vector that represents gravity where the
// coordinate plane's origin is on the top-right corner
var TerminalGravity = Vector{0, 9.81, 0}

// NewProjectile creates a new projectile. It accepts a frame rate and initial
// values for position, velocity, and acceleration. It returns a new
// projectile.
func NewProjectile(deltaTime float64, initialPosition Point, initialVelocity, initalAcceleration Vector) *Projectile {
	return &Projectile{
		pos:       initialPosition,
		vel:       initialVelocity,
		acc:       initalAcceleration,
		deltaTime: deltaTime,
	}
}

// Update updates the position and velocity values for the given projectile.
// Call this after calling NewProjectile to update values.
func (p *Projectile) Update() Point {
	p.pos.X += (p.vel.X * p.deltaTime)
	p.pos.Y += (p.vel.Y * p.deltaTime)
	p.pos.Z += (p.vel.Z * p.deltaTime)

	p.vel.X += (p.acc.X * p.deltaTime)
	p.vel.Y += (p.acc.Y * p.deltaTime)
	p.vel.Z += (p.acc.Z * p.deltaTime)

	return p.pos
}

// Position returns the position of the projectile.
func (p *Projectile) Position() Point {
	return p.pos
}

// Velocity returns the velocity of the projectile.
func (p *Projectile) Velocity() Vector {
	return p.vel
}

// Acceleration returns the acceleration of the projectile.
func (p *Projectile) Acceleration() Vector {
	return p.acc
}
