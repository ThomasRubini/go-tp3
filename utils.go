package main

import "math"

type Vec2f struct {
	x, y float32
}

type Vec3f struct {
	x, y, z float32
}

func (v Vec3f) inverte() Vec3f {
	return Vec3f{-v.x, -v.y, -v.z}
}

func Add(v1, v2 Vec3f) Vec3f {
	return Vec3f{v1.x + v2.x, v1.y + v2.y, v1.z + v2.z}
}

func (v Vec3f) mul(f float32) Vec3f {
	return Vec3f{v.x * f, v.y * f, v.z * f}
}

func Mul(v1, v2 Vec3f) Vec3f {
	return Vec3f{v1.x * v2.x, v1.y * v2.y, v1.z * v2.z}
}
func Dot(v1, v2 Vec3f) float32 {
	return v1.x*v2.x + v1.y*v2.y + v1.z*v2.z
}

func cross(v1, v2 Vec3f) Vec3f {
	return Vec3f{v1.y*v2.z - v2.y*v1.z, v1.z*v2.x - v2.z*v1.x, v1.x*v2.y - v2.x*v1.y}
}

// -------------------------------

func (v Vec3f) norme() float32 {
	return float32(math.Sqrt(float64(v.x*v.x + v.y*v.y + v.z*v.z)))
}
func (v *Vec3f) normalize() {
	norme := v.norme()
	v.x /= norme
	v.y /= norme
	v.z /= norme
}
func (v Vec3f) normalized() Vec3f {
	norme := v.norme()
	return Vec3f{v.x / norme, v.y / norme, v.z / norme}
}

// --------------------------------

type rgbRepresentation struct {
	r, g, b uint8
}
