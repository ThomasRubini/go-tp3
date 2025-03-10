package main

type Phong struct {
	ka, kd, ks Vec3f
	n          float32
}

func (l Phong) render(rio, rdi, n Vec3f, t float32, scene Scene) rgbRepresentation {
	// --- Etape 1
	Ia := Mul(l.ka, scene.ambiantLight)

	// --- Etape 2
	// Point d'intersection
	omega := Add(rio, rdi.mul(t))
	// Vecteur point d'intersection -> lumière
	vec_intersect_light := Sub(scene.lights[0].position, omega).normalized()
	L := vec_intersect_light

	n.normalize()

	// Intensité lumineuse
	I := scene.lights[0].color
	Id := Mul(l.kd, I.mul(Dot(L, n)))

	// --- Etape 3
	R := vec_intersect_light.normalized()
	V := rdi.inverte().normalized()
	Is := Mul(l.ks, I).mul(Pow(Dot(R, V), l.n))
	// fmt.Println("-----")
	// fmt.Println(R)
	// fmt.Println(V)

	// --- Finish
	res := Add(Add(Ia, Id), Is)
	// res := Add(Ia, Id)
	_ = Ia
	_ = Id
	_ = Is
	return rgbRepresentation{uint8(res.x * 255), uint8(res.y * 255), uint8(res.z * 255)}
}
