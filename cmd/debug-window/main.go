package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/g-wilson/led/clock"
	"github.com/g-wilson/led/internal/framestreamer"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/joho/godotenv"
)

const (
	vertexShaderSource = `#version 330 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec2 aTexCoord;
out vec2 TexCoord;
uniform mat4 projection;
void main() {
    gl_Position = projection * vec4(aPos, 0.0, 1.0);
    TexCoord = aTexCoord;
}
`

	fragmentShaderSource = `#version 330 core
in vec2 TexCoord;
out vec4 FragColor;
uniform sampler2D texture1;
void main() {
    FragColor = texture(texture1, TexCoord);
}
`
)

// WindowRenderer manages the native window and OpenGL rendering state
type WindowRenderer struct {
	window       *glfw.Window
	windowWidth  int
	windowHeight int
	ledRows      int
	ledCols      int
	texture      uint32
	shaderProg   uint32
	vao          uint32
	vbo          uint32
	ebo          uint32
	frameChan    chan image.Image
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// Lock the main goroutine to the OS thread (required for GLFW on macOS)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rows, _ := strconv.ParseInt(os.Getenv("LED_ROWS"), 10, 32)
	cols, _ := strconv.ParseInt(os.Getenv("LED_COLS"), 10, 32)
	ledRows := int(rows)
	ledCols := int(cols)

	// Initialize GLFW (must be on main thread on macOS)
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	// Create window renderer
	renderer, err := NewWindowRenderer(ledRows, ledCols)
	if err != nil {
		log.Fatalln("failed to create window renderer:", err)
	}
	defer renderer.Cleanup()

	// Create clock renderer
	clockApp, err := clock.New()
	if err != nil {
		log.Fatalln(err)
	}

	// Create framestreamer
	fs := framestreamer.New(framestreamer.Params{
		Bounds: image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: ledCols, Y: ledRows},
		},
		FrametimeMs: framestreamer.OneFPS,
		Renderer:    clockApp,
	})

	// Start framestreamer
	go fs.Start()

	// Frame receiver goroutine - sends frames to channel for main thread processing
	go func() {
		for {
			select {
			case err := <-fs.E:
				if err != nil {
					log.Fatalln("framestreamer error:", err)
				}
			case frame := <-fs.C:
				if frame != nil {
					// Send frame to channel (non-blocking if channel is full)
					select {
					case renderer.frameChan <- frame:
					default:
						// Skip frame if channel is full (we're rendering faster than frames arrive)
					}
				}
			}
		}
	}()

	// Main render loop - processes frames on main thread
	renderer.Run()

	// Stop framestreamer
	fs.Stop()
}

// NewWindowRenderer creates and initializes a new window renderer
func NewWindowRenderer(ledRows, ledCols int) (*WindowRenderer, error) {
	wr := &WindowRenderer{
		windowWidth:  800,
		windowHeight: 600,
		ledRows:      ledRows,
		ledCols:      ledCols,
		frameChan:    make(chan image.Image, 1), // Buffered channel to avoid blocking
	}

	// Configure GLFW window hints
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.True)

	// Create window
	win, err := glfw.CreateWindow(wr.windowWidth, wr.windowHeight, "LED Matrix Debug", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	wr.window = win

	wr.window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		wr.window.Destroy()
		return nil, fmt.Errorf("failed to initialize OpenGL: %w", err)
	}

	// Enable VSync
	glfw.SwapInterval(1)

	// Set up OpenGL resources
	if err := wr.setupOpenGL(); err != nil {
		wr.window.Destroy()
		return nil, fmt.Errorf("failed to setup OpenGL: %w", err)
	}

	// Set up window resize callback
	wr.window.SetFramebufferSizeCallback(wr.framebufferSizeCallback)

	return wr, nil
}

// Cleanup releases all OpenGL resources and destroys the window
func (wr *WindowRenderer) Cleanup() {
	wr.cleanupOpenGL()
	if wr.window != nil {
		wr.window.Destroy()
	}
}

// updateTexture updates the texture with a new frame (must be called from main thread)
func (wr *WindowRenderer) updateTexture(frame image.Image) {
	// Convert image to RGBA if needed
	rgba, ok := frame.(*image.RGBA)
	if !ok {
		// Convert to RGBA
		bounds := frame.Bounds()
		rgba = image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rgba.Set(x, y, frame.At(x, y))
			}
		}
	}

	// Upload texture data
	gl.BindTexture(gl.TEXTURE_2D, wr.texture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(wr.ledCols), int32(wr.ledRows), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// Run executes the main render loop until the window is closed
// All OpenGL/GLFW operations happen on the main thread (required on macOS)
func (wr *WindowRenderer) Run() {
	for !wr.window.ShouldClose() {
		// Poll events (must be on main thread on macOS)
		glfw.PollEvents()

		// Process any pending frames from the channel (non-blocking)
		select {
		case frame := <-wr.frameChan:
			wr.updateTexture(frame)
		default:
			// No new frame, continue with rendering
		}

		// Render (must be on main thread on macOS)
		wr.render()

		// Swap buffers (must be on main thread on macOS)
		wr.window.SwapBuffers()
	}
}

func (wr *WindowRenderer) setupOpenGL() error {
	// Create and compile vertex shader
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	// Create and compile fragment shader
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	// Create shader program
	wr.shaderProg = gl.CreateProgram()
	gl.AttachShader(wr.shaderProg, vertexShader)
	gl.AttachShader(wr.shaderProg, fragmentShader)
	gl.LinkProgram(wr.shaderProg)

	// Check for linking errors
	var success int32
	gl.GetProgramiv(wr.shaderProg, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(wr.shaderProg, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		gl.GetProgramInfoLog(wr.shaderProg, logLength, nil, &log[0])
		return fmt.Errorf("shader program linking failed: %s", string(log))
	}

	// Delete shaders (they're linked into the program now)
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// Set up vertex data for fullscreen quad
	vertices := []float32{
		// positions   // texCoords (Y flipped: OpenGL has 0,0 at bottom-left, Go images have 0,0 at top-left)
		-1.0, -1.0, 0.0, 1.0, // bottom-left position -> top-left texcoord
		1.0, -1.0, 1.0, 1.0, // bottom-right position -> top-right texcoord
		1.0, 1.0, 1.0, 0.0, // top-right position -> bottom-right texcoord
		-1.0, 1.0, 0.0, 0.0, // top-left position -> bottom-left texcoord
	}

	indices := []uint32{
		0, 1, 2, // first triangle
		2, 3, 0, // second triangle
	}

	// Generate and bind VAO
	gl.GenVertexArrays(1, &wr.vao)
	gl.BindVertexArray(wr.vao)

	// Generate and bind VBO
	gl.GenBuffers(1, &wr.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, wr.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Generate and bind EBO
	gl.GenBuffers(1, &wr.ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, wr.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Set vertex attributes
	// Position attribute
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Texture coordinate attribute
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	// Unbind VAO
	gl.BindVertexArray(0)

	// Create texture
	gl.GenTextures(1, &wr.texture)
	gl.BindTexture(gl.TEXTURE_2D, wr.texture)

	// Set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Initialize texture with black
	blackData := make([]uint8, wr.ledCols*wr.ledRows*4)
	for i := 0; i < len(blackData); i += 4 {
		blackData[i] = 0     // R
		blackData[i+1] = 0   // G
		blackData[i+2] = 0   // B
		blackData[i+3] = 255 // A
	}
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(wr.ledCols), int32(wr.ledRows), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(blackData))

	gl.BindTexture(gl.TEXTURE_2D, 0)

	// Set clear color
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	return nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])
		return 0, fmt.Errorf("shader compilation failed: %s", string(log))
	}

	return shader, nil
}

func (wr *WindowRenderer) render() {
	// Clear screen
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// Use shader program
	gl.UseProgram(wr.shaderProg)

	// Calculate aspect ratio and projection matrix
	wr.updateProjectionMatrix()

	// Bind texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, wr.texture)
	gl.Uniform1i(gl.GetUniformLocation(wr.shaderProg, gl.Str("texture1\x00")), 0)

	// Bind VAO and draw
	gl.BindVertexArray(wr.vao)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)
}

func (wr *WindowRenderer) updateProjectionMatrix() {
	// Get window size
	w, h := wr.window.GetFramebufferSize()
	wr.windowWidth = w
	wr.windowHeight = h

	// Calculate aspect ratios
	windowAspect := float32(wr.windowWidth) / float32(wr.windowHeight)
	ledAspect := float32(wr.ledCols) / float32(wr.ledRows)

	// Calculate scale to fill window while maintaining aspect ratio
	var scaleX, scaleY float32
	if windowAspect > ledAspect {
		// Window is wider - scale based on height
		scaleY = 1.0
		scaleX = ledAspect / windowAspect
	} else {
		// Window is taller - scale based on width
		scaleX = 1.0
		scaleY = windowAspect / ledAspect
	}

	// Create orthographic projection matrix
	projection := [16]float32{
		scaleX, 0, 0, 0,
		0, scaleY, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	// Set projection matrix uniform
	projLoc := gl.GetUniformLocation(wr.shaderProg, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	// Set viewport
	gl.Viewport(0, 0, int32(wr.windowWidth), int32(wr.windowHeight))
}

func (wr *WindowRenderer) framebufferSizeCallback(w *glfw.Window, width int, height int) {
	wr.windowWidth = width
	wr.windowHeight = height
	wr.updateProjectionMatrix()
}

func (wr *WindowRenderer) cleanupOpenGL() {
	gl.DeleteTextures(1, &wr.texture)
	gl.DeleteBuffers(1, &wr.vbo)
	gl.DeleteBuffers(1, &wr.ebo)
	gl.DeleteVertexArrays(1, &wr.vao)
	gl.DeleteProgram(wr.shaderProg)
}
