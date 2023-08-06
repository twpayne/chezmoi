Harmonica
=========

<p>
    <a href="https://stuff.charm.sh/harmonica/harmonica-art.png"><img src="https://stuff.charm.sh/harmonica/harmonica-readme.png" alt="Harmonica Image" width="325"></a><br>
    <a href="https://github.com/charmbracelet/harmonica/releases"><img src="https://img.shields.io/github/release/charmbracelet/harmonica.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/charmbracelet/harmonica?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/charmbracelet/harmonica/actions"><img src="https://github.com/charmbracelet/harmonica/workflows/build/badge.svg" alt="Build Status"></a>
</p>

A simple, efficient spring animation library for smooth, natural motion.

<img src="https://stuff.charm.sh/harmonica/harmonica-opengl.gif" width="500" alt="Harmonica OpenGL Demo">

It even works well on the command line.

<img src="https://stuff.charm.sh/harmonica/harmonica-tui.gif" width="900" alt="Harmonica TUI Demo">

[examples]: https://github.com/charmbracelet/harmonica/tree/master/examples
[docs]: https://pkg.go.dev/github.com/charmbracelet/harmonica?tab=doc

## Usage

Harmonica is framework-agnostic and works well in 2D and 3D contexts. Simply
call [`NewSpring`][newspring] with your settings to initialize and
[`Update`][update] on each frame to animate.

```go
import "github.com/charmbracelet/harmonica"

// A thing we want to animate.
sprite := struct{
    x, xVelocity float64
    y, yVelocity float64
}{}

// Where we want to animate it.
const targetX = 50.0
const targetY = 100.0

// Initialize a spring with framerate, angular frequency, and damping values.
spring := harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.5)

// Animate!
for {
    sprite.x, sprite.xVelocity = spring.Update(sprite.x, sprite.xVelocity, targetX)
    sprite.y, sprite.yVelocity = spring.Update(sprite.y, sprite.yVelocity, targetY)
    time.Sleep(time.Second/60)
}
```

For details, see the [examples][examples] and the [docs][docs].

[newspring]: https://pkg.go.dev/github.com/charmbracelet/harmonica#NewSpring
[update]: https://pkg.go.dev/github.com/charmbracelet/harmonica#Update

## Settings

[`NewSpring`][newspring] takes three values:

* **Time Delta:** the time step to operate on. Game engines typically provide
  a way to determine the time delta, however if that's not available you can
  simply set the framerate with the included `FPS(int)` utility function. Make
  the framerate you set here matches your actual framerate.
* **Angular Velocity:** this translates roughly to the speed. Higher values are
  faster.
* **Damping Ratio:** the springiness of the animation, generally between `0`
  and `1`, though it can go higher. Lower values are springier. For details,
  see below.

## Damping Ratios

The damping ratio affects the motion in one of three different ways depending
on how it's set.

### Under-Damping

A spring is under-damped when its damping ratio is less than `1`. An
under-damped spring reaches equilibrium the fastest, but overshoots and will
continue to oscillate as its amplitude decays over time.

### Critical Damping

A spring is critically-damped the damping ratio is exactly `1`. A critically
damped spring will reach equilibrium as fast as possible without oscillating.

### Over-Damping

A spring is over-damped the damping ratio is greater than `1`. An over-damped
spring will never oscillate, but reaches equilibrium at a slower rate than
a critically damped spring.

## Acknowledgements

This library is a fairly straightforward port of [Ryan Juckett][juckett]’s
excellent damped simple harmonic oscillator originally written in C++ in 2008
and published in 2012. [Ryan’s writeup][writeup] on the subject is fantastic.

[juckett]: https://www.ryanjuckett.com/
[writeup]: https://www.ryanjuckett.com/damped-springs/

## License

[MIT](https://github.com/charmbracelet/harmonica/raw/master/LICENSE)

***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge-unrounded.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source
