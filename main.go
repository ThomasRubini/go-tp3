package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"runtime/pprof"
)

type Image struct {
	frameBuffer   []rgbRepresentation
	width, height int
}

func (i Image) save(path string) error {
	// Création de l'image
	img := image.NewRGBA(image.Rect(0, 0, i.width, i.height))
	for y := 0; y < i.height; y++ {
		for x := 0; x < i.width; x++ {
			idx := (y*i.width + x)
			r, g, b := i.frameBuffer[idx].r, i.frameBuffer[idx].b, i.frameBuffer[idx].g
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	pngFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer pngFile.Close()
	if err := png.Encode(pngFile, img); err != nil {
		return err
	}
	return nil
}

// ------------------
type Light struct {
	color    Vec3f
	position Vec3f
}

// --------------------------------
type Scene struct {
	objects []GeometricObject
	lights  []Light
}

func (s *Scene) addLight(l Light) {
	s.lights = append(s.lights, l)
}
func (s *Scene) addElement(g GeometricObject) {
	s.objects = append(s.objects, g)
}

// ----------------------------------
type Materials interface {
	render(rio, rdi, n Vec3f, t float32, scene Scene) rgbRepresentation
}

// Lambert represents a Lambertian reflectance model which is used in computer graphics
// to simulate the way light reflects off a diffuse surface. The kd field represents
// the diffuse reflection coefficient, which is a vector of three floating-point values
// corresponding to the RGB color components.
type Lambert struct {
	kd Vec3f
}

// render calculates the Lambertian reflectance for a given point in the scene.
// It takes the incoming ray direction (rio), the reflected ray direction (rdi),
// the normal at the intersection point (n), the intersection distance (t), and
// the scene information (scene). It returns the RGB representation of the
// reflected light.
//
// Parameters:
// - rio: Vec3f representing the incoming ray direction.
// - rdi: Vec3f representing the reflected ray direction.
// - n: Vec3f representing the normal at the intersection point.
// - t: float32 representing the intersection distance.
// - scene: Scene containing the scene information including lights.
//
// Returns:
// - rgbRepresentation: The RGB representation of the reflected light.
func (l Lambert) render(rio, rdi, n Vec3f, t float32, scene Scene) rgbRepresentation {
	// res := Mul(l.kd, scene.lights[0].color) // res := l.kd
	// return rgbRepresentation{uint8(res.x), uint8(res.y), uint8(res.z)}
	omega := Add(rio, rdi.mul(t))
	Li := Mul(l.kd, scene.lights[0].color.mul(Dot(n, omega))).mul(1 / 3.14)
	return rgbRepresentation{uint8(Li.x * 255), uint8(Li.y * 255), uint8(Li.z * 255)}
}

type GeometricObject interface {
	isIntersectedByRay(ro, rd Vec3f) (bool, float32)
	render(rio, rdi Vec3f, t float32, scene Scene) rgbRepresentation
}

// -------------------------------
// Sphere represents a 3D sphere with a specific radius, position, and material properties.
type Sphere struct {
	radius   float32
	position Vec3f
	Material Materials
}

// render calculates the color representation of a sphere when rendered in a scene.
// It takes the incident ray origin (rio), the incident ray direction (rdi),
// the intersection distance (t), and the scene as parameters.
// The normal on a sphere is the inverse of the incident ray direction.
// This function returns the RGB representation of the rendered sphere.
func (s Sphere) render(rio, rdi Vec3f, t float32, scene Scene) rgbRepresentation {
	/*
	* Le calcul de la normal sur une sphère est l'inverse du rayon incident.
	* C'est pourquoi n = rd1.inverte()
	 */
	return s.Material.render(rio, rdi, rdi.inverte(), t, scene)
}

// isIntersectedByRay determines if a ray intersects with the sphere.
// It takes the ray origin (ro) and ray direction (rd) as Vec3f parameters.
// It returns a boolean indicating if there is an intersection, and a float32
// representing the distance from the ray origin to the intersection point.
//
// Parameters:
//   - ro: Vec3f representing the origin of the ray.
//   - rd: Vec3f representing the direction of the ray.
//
// Returns:
//   - bool: true if the ray intersects the sphere, false otherwise.
//   - float32: the distance from the ray origin to the intersection point if there is an intersection, 0.0 otherwise.
func (s Sphere) isIntersectedByRay(ro, rd Vec3f) (bool, float32) {
	L := Add(ro, Vec3f{-s.position.x, -s.position.y, -s.position.z})

	a := Dot(rd, rd)
	b := 2.0 * Dot(rd, L)
	c := Dot(L, L) - s.radius*s.radius
	delta := b*b - 4.0*a*c

	t0 := (-b - float32(math.Sqrt(float64(delta)))) / 2 * a
	t1 := (-b + float32(math.Sqrt(float64(delta)))) / 2 * a
	t := t0
	t = min(t, t1)

	if delta > 0 {
		return true, t
	}
	return false, 0.0
}

// ------------------------------
type Camera struct {
	position, up, at Vec3f
}

// direction calculates the direction vector of the camera by subtracting
// the camera's position from its target point (at), normalizing the resulting
// vector, and returning it. The returned vector is a unit vector pointing
// from the camera's position to its target point.
func (c Camera) direction() Vec3f {
	dir := Add(c.at, c.position.inverte())
	return dir.mul(float32(1) / dir.norme())
}

// ------------------------------

// renderPixel computes the color of a pixel by tracing a ray through the scene.
// It iterates over all objects in the scene to find the closest intersection point
// and then calculates the color at that point.
//
// Parameters:
// - scene: The Scene containing all objects to be rendered.
// - ro: The origin of the ray (Vec3f).
// - rd: The direction of the ray (Vec3f).
//
// Returns:
// - rgbRepresentation: The color of the pixel as an rgbRepresentation struct.
func renderPixel(scene Scene, ro, rd Vec3f) rgbRepresentation {
	var tmin float32
	tmin = 9999999999.0
	res := rgbRepresentation{}
	for _, object := range scene.objects {
		isIntersected, t := object.isIntersectedByRay(ro, rd)
		if isIntersected && t < tmin {
			tmin = t
			res = object.render(ro, rd, t, scene)
		}
	}
	return res
}

// renderFrame renders a frame of the scene from the perspective of the camera onto the image.
//
// Parameters:
//   - image: The Image object that contains the frame buffer where the rendered frame will be stored.
//   - camera: The Camera object that defines the position and orientation of the camera.
//   - scene: The Scene object that contains all the objects and lights to be rendered.
//
// The function calculates the ray direction for each pixel in the image based on the camera's position and orientation.
// It then traces the ray through the scene to determine the color of the pixel and stores the result in the image's frame buffer.
func renderFrame(image Image, camera Camera, scene Scene) {
	ro := camera.position
	cosFovy := float32(0.66)

	aspect := float32(image.width) / float32(image.height)
	horizontal := (cross(camera.direction(), camera.up)).normalized().mul(cosFovy * aspect)
	vertical := (cross(horizontal, camera.direction())).normalized().mul(cosFovy)

	for x := 0; x < image.width; x++ {
		for y := 0; y < image.height; y++ {

			uvx := (float32(x) + float32(0.5)) / float32(image.width)
			uvy := (float32(y) + float32(0.5)) / float32(image.height)

			rd := Add(Add(camera.direction(), horizontal.mul(uvx-float32(0.5))), vertical.mul(uvy-float32(0.5))).normalized()

			image.frameBuffer[y*image.width+x] = renderPixel(scene, ro, rd)
		}
	}

}

func populateScene(scene *Scene) {

	//Intégrer dans l'objet Scène
	scene.addElement(Sphere{1, Vec3f{0, 0, 8}, Lambert{Vec3f{1.0, 0, 0}}})
	scene.addElement(Sphere{0.3, Vec3f{2, 1.5, 4}, Lambert{Vec3f{0.0, 1.0, 0}}})
	scene.addElement(Sphere{0.9, Vec3f{0, -1, 5}, Lambert{Vec3f{0.0, 0, 1.0}}})
	scene.addElement(Sphere{0.5, Vec3f{-2, -2, 5}, Lambert{Vec3f{1.0, 1.0, 1.0}}})

	scene.addLight(Light{Vec3f{1.0, 1.0, 1.0}, Vec3f{0, 10, 0}})
}

func main() {
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	width := 4096
	height := 4096
	//Créer un objet Scène
	scene := Scene{}

	//Initialiser la scène
	populateScene(&scene)
	//Créer une caméra
	camera := Camera{Vec3f{0, 0, -5}, Vec3f{0, 1, 0}, Vec3f{0, 0, 5}}

	image := Image{make([]rgbRepresentation, width*height), width, height}
	//fonction de rendu
	renderFrame(image, camera, scene)
	//Sauvegarde de l'image
	if err := image.save("./result.png"); err != nil {
		panic(err)
	}
}
