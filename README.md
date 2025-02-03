# gphoto2

A Go library and control script for interfacing with DSLR cameras using the gphoto2 library.

## Overview
This project provides a comprehensive Go wrapper around the gphoto2 library, enabling programmatic control of DSLR cameras. It includes both a library for integration into other projects and standalone control scripts.

## Features
- Complete DSLR camera control
- File transfer and management
- Camera settings manipulation
- Live preview support
- Event handling
- Extensive API coverage

## Components
- Camera control interface
- File management
- Settings manipulation
- Widget handling
- Callback system
- Context management

## Usage

### Library Usage
```go
import "github.com/ZeroSpace-Studios/gphoto2"

func main() {
    ctx := gphoto2.NewContext()
    defer ctx.Free()
    
    camera, err := gphoto2.NewCamera()
    if err != nil {
        log.Fatal(err)
    }
    defer camera.Free()
}
```

### Camera Operations
```go
// Capture photo
camera.CaptureImage()

// Download files
camera.DownloadFile("path/to/file")

// Adjust settings
camera.SetConfig("shutterspeed", "1/1000")
```

## Project Structure
```
├── camera.go          # Core camera operations
├── camera_file.go     # File handling
├── camera_photo.go    # Photo operations
├── camera_settings.go # Settings management
├── camera_widget.go   # Widget interface
├── callbacks.go       # Event system
└── types.go          # Data structures
```

## Requirements
- libgphoto2 installed
- Go 1.13 or later
- CGO enabled
- Compatible DSLR camera

## Examples
Check the `examples/` directory for:
- Basic camera control
- File transfer
- Settings adjustment
- Event handling

## History

Much of the code is copied from and/or insipired of https://github.com/szank/gphoto which seems abandonend and had a couple of bugs. I waanted to write a simple stop motion program and this is an artefact from it. So it's claiming to or trying to cover the whole library but you can do a lot, and I'm very open to suggestions and/or merge requests. I only have access to one Nikon camera so it's not tested with any other brand, and it seems like all camera manufacturers have their own quirks.

## Installlation

To build the library you need to have libgphoto2-6 and libgphoto2-port12 or later installed.

## Notes

For Nikon (I don't know about other cameras as I haven't been able to test with them) you want to put the DSLR in manual mode, and turn auto focus off typically. Manual mode is necessary to be able to change many of the settings, and to be able to enter live view mode for instance. Manual focus makes taking shots a lot more reliable and fast as the camera won't have to focus for each shot.

