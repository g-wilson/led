// Package windowrenderer provides a native window renderer for displaying LED matrix frames
// using GLFW and OpenGL.
package windowrenderer

import (
	"fmt"
	"image"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
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
` + "\x00"

	fragmentShaderSource = `#version 330 core
in vec2 TexCoord;
out vec4 FragColor;
uniform sampler2D texture1;
void main() {
    FragColor = texture(texture1, TexCoord);
}
` + "\x00"
)

// Renderer manages the native window and OpenGL rendering state
type Renderer struct {
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
	frameChan    chan *image.RGBA
}

// New creates and initializes a new window renderer.
// IMPORTANT: Must be called from the main goroutine with runtime.LockOSThread() already called.
// This is required for GLFW/OpenGL to work correctly on macOS.
func New(title string, ledRows, ledCols int) (*Renderer, error) {
	// Ensure the current goroutine is locked to an OS thread
	// This is a best-effort check - the caller should have called runtime.LockOSThread() in main()
	runtime.LockOSThread()

	r := &Renderer{
		windowWidth:  800,
		windowHeight: 600,
		ledRows:      ledRows,
		ledCols:      ledCols,
		frameChan:    make(chan *image.RGBA, 1), // Buffered channel to avoid blocking
	}

	// Configure GLFW window hints
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.True)

	// Create window
	win, err := glfw.CreateWindow(r.windowWidth, r.windowHeight, title, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	r.window = win

	r.window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		r.window.Destroy()
		return nil, fmt.Errorf("failed to initialize OpenGL: %w", err)
	}

	// Enable VSync
	glfw.SwapInterval(1)

	// Set up OpenGL resources
	if err := r.setupOpenGL(); err != nil {
		r.window.Destroy()
		return nil, fmt.Errorf("failed to setup OpenGL: %w", err)
	}

	// Set up window resize callback
	r.window.SetFramebufferSizeCallback(r.framebufferSizeCallback)

	return r, nil
}

// Cleanup releases all OpenGL resources and destroys the window
func (r *Renderer) Cleanup() {
	r.cleanupOpenGL()
	if r.window != nil {
		r.window.Destroy()
	}
}

// SendFrame sends a frame to the renderer for display.
// This is safe to call from any goroutine.
// If a frame is already pending, the new frame will be dropped (non-blocking).
func (r *Renderer) SendFrame(frame *image.RGBA) {
	select {
	case r.frameChan <- frame:
	default:
		// Skip frame if channel is full (we're rendering faster than frames arrive)
	}
}

// Run executes the main render loop until the window is closed.
// All OpenGL/GLFW operations happen on the main thread (required on macOS).
func (r *Renderer) Run() {
	for !r.window.ShouldClose() {
		// Poll events (must be on main thread on macOS)
		glfw.PollEvents()

		// Process any pending frames from the channel (non-blocking)
		select {
		case frame := <-r.frameChan:
			r.updateTexture(frame)
		default:
			// No new frame, continue with rendering
		}

		// Render (must be on main thread on macOS)
		r.render()

		// Swap buffers (must be on main thread on macOS)
		r.window.SwapBuffers()
	}
}

// updateTexture updates the texture with a new frame (must be called from main thread)
func (r *Renderer) updateTexture(frame *image.RGBA) {
	gl.BindTexture(gl.TEXTURE_2D, r.texture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(r.ledCols), int32(r.ledRows), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(frame.Pix))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (r *Renderer) setupOpenGL() error {
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
	r.shaderProg = gl.CreateProgram()
	gl.AttachShader(r.shaderProg, vertexShader)
	gl.AttachShader(r.shaderProg, fragmentShader)
	gl.LinkProgram(r.shaderProg)

	// Check for linking errors
	var success int32
	gl.GetProgramiv(r.shaderProg, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(r.shaderProg, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(r.shaderProg, logLength, nil, &logMsg[0])
		return fmt.Errorf("shader program linking failed: %s", string(logMsg))
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
	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	// Generate and bind VBO
	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Generate and bind EBO
	gl.GenBuffers(1, &r.ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.ebo)
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
	gl.GenTextures(1, &r.texture)
	gl.BindTexture(gl.TEXTURE_2D, r.texture)

	// Set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Initialize texture with black
	blackData := make([]uint8, r.ledCols*r.ledRows*4)
	for i := 0; i < len(blackData); i += 4 {
		blackData[i] = 0     // R
		blackData[i+1] = 0   // G
		blackData[i+2] = 0   // B
		blackData[i+3] = 255 // A
	}
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(r.ledCols), int32(r.ledRows), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(blackData))

	gl.BindTexture(gl.TEXTURE_2D, 0)

	// Set clear color to dark grey to distinguish from LED matrix black background
	gl.ClearColor(0.15, 0.15, 0.15, 1.0)

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
		logMsg := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &logMsg[0])
		return 0, fmt.Errorf("shader compilation failed: %s", string(logMsg))
	}

	return shader, nil
}

func (r *Renderer) render() {
	// Clear screen
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// Use shader program
	gl.UseProgram(r.shaderProg)

	// Calculate aspect ratio and projection matrix
	r.updateProjectionMatrix()

	// Bind texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, r.texture)
	gl.Uniform1i(gl.GetUniformLocation(r.shaderProg, gl.Str("texture1\x00")), 0)

	// Bind VAO and draw
	gl.BindVertexArray(r.vao)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)
}

func (r *Renderer) updateProjectionMatrix() {
	// Get window size
	w, h := r.window.GetFramebufferSize()
	r.windowWidth = w
	r.windowHeight = h

	// Minimum border in pixels
	const minBorder = 20.0

	// Calculate effective drawable area (excluding minimum borders)
	effectiveWidth := float32(r.windowWidth) - 2*minBorder
	effectiveHeight := float32(r.windowHeight) - 2*minBorder

	// Ensure we have positive dimensions
	if effectiveWidth <= 0 {
		effectiveWidth = float32(r.windowWidth)
	}
	if effectiveHeight <= 0 {
		effectiveHeight = float32(r.windowHeight)
	}

	// Calculate aspect ratios
	windowAspect := effectiveWidth / effectiveHeight
	ledAspect := float32(r.ledCols) / float32(r.ledRows)

	// Calculate scale to fit within effective area while maintaining aspect ratio
	var scaleX, scaleY float32
	if windowAspect > ledAspect {
		// Window is wider - scale based on height
		// The LED frame height should fit within effectiveHeight
		scaleY = effectiveHeight / float32(r.windowHeight)
		scaleX = scaleY * ledAspect * float32(r.windowHeight) / float32(r.windowWidth)
	} else {
		// Window is taller - scale based on width
		// The LED frame width should fit within effectiveWidth
		scaleX = effectiveWidth / float32(r.windowWidth)
		scaleY = scaleX * (1.0 / ledAspect) * float32(r.windowWidth) / float32(r.windowHeight)
	}

	// Create orthographic projection matrix
	projection := [16]float32{
		scaleX, 0, 0, 0,
		0, scaleY, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	// Set projection matrix uniform
	projLoc := gl.GetUniformLocation(r.shaderProg, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	// Set viewport
	gl.Viewport(0, 0, int32(r.windowWidth), int32(r.windowHeight))
}

func (r *Renderer) framebufferSizeCallback(_ *glfw.Window, width, height int) {
	r.windowWidth = width
	r.windowHeight = height
	r.updateProjectionMatrix()
}

func (r *Renderer) cleanupOpenGL() {
	gl.DeleteTextures(1, &r.texture)
	gl.DeleteBuffers(1, &r.vbo)
	gl.DeleteBuffers(1, &r.ebo)
	gl.DeleteVertexArrays(1, &r.vao)
	gl.DeleteProgram(r.shaderProg)
}
