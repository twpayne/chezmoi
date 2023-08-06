// Package harmonica is a set of physics-based animation tools for 2D and 3D
// applications. There's a spring animation simulator for for smooth, realistic
// motion and a projectile simulator well suited for projectiles and particles.
//
// Example spring usage:
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
// Example projectile usage:
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
package harmonica
