package gphoto2

/** \file
 *
 * \author Copyright 2020 Jon Molin
 *
 * \note
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2 of the License, or (at your option) any later version.
 *
 * \note
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * \note
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the
 * Free Software Foundation, Inc., 51 Franklin Street, Fifth Floor,
 * Boston, MA  02110-1301  USA
 */

// #cgo LDFLAGS:  -lgphoto2 -lgphoto2_port
// #cgo CFLAGS: -I/usr/include
// #include <gphoto2/gphoto2.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

//Camera struct represents a camera connected to the computer
type Camera struct {
	gpCamera *C.Camera
	Ctx      *Context
	Settings *CameraWidget
}

type CameraList struct {
	Names []string
	Ports []string
}

//Exit Closes a connection to the camera and therefore gives other application
//the possibility to access the camera, too. It is recommended that you call
//this function when you currently don't need the camera. The camera will get
//reinitialized by gp_camera_init() automatically if you try to access the camera again.
func (c Camera) Exit() error {
	if c.gpCamera != nil {
		if res := C.gp_camera_exit(c.gpCamera, c.Ctx.gpContext); res != GPOK {
			return newError("", int(res))
		}
	}
	return nil
}

func (c Camera) Free() error {
	if err := c.Exit(); err != nil {
		return err
	}
	if res := C.gp_camera_unref(c.gpCamera); res != GPOK {
		return newError("", int(res))
	}
	c.Ctx.free()
	return nil
}

// ResetCamera resets the camera port, can be needed at times
// https://github.com/gphoto/gphoto2/blob/7a48ea37832bcd19e17b80afef2f7f2d426419f3/gphoto2/main.c#L1675
func (c *Camera) Reset() error {
	if err := c.Exit(); err != nil {
		return err
	}

	var port *C.GPPort
	var info C.GPPortInfo
	if res := C.gp_port_new(&port); res != GPOK {
		return newError("", int(res))
	}
	if res := C.gp_camera_get_port_info(c.gpCamera, &info); res != GPOK {
		return newError("", int(res))
	}
	if res := C.gp_port_set_info(port, info); res != GPOK {
		return newError("", int(res))
	}
	if res := C.gp_port_open(port); res != GPOK {
		return newError("", int(res))
	}
	if res := C.gp_port_reset(port); res != GPOK {
		return newError("", int(res))
	}
	if res := C.gp_port_close(port); res != GPOK {
		return newError("", int(res))
	}
	if res := C.gp_port_free(port); res != GPOK {
		return newError("", int(res))
	}
	return nil
}

// NewCamera tries to connect to a camera with name name, if name is empty it tries with the first connected camera. It returns a new Camera struct.
func NewCamera(name string) (*Camera, error) {
	ctx, err := NewContext()
	if err != nil {
		return nil, err
	}
	var gpCamera *C.Camera

	if name != "" {
		var cameraList *C.CameraList
		var abilitiesList *C.CameraAbilitiesList
		var portInfoList *C.GPPortInfoList
		var abilities C.CameraAbilities
		var portInfo C.GPPortInfo

		if res := C.gp_list_new(&cameraList); res != GPOK {
			return nil, newError("Cannot initialize camera list", int(res))
		}

		if res := C.gp_abilities_list_new(&abilitiesList); res != GPOK {
			return nil, newError("Cannot initialize camera abilities list", int(res))
		}

		if res := C.gp_abilities_list_load(abilitiesList, ctx.gpContext); res != GPOK {
			return nil, newError("Cannot load camera abilities list", int(res))
		}

		defer C.free(unsafe.Pointer(cameraList))
		defer C.free(unsafe.Pointer(abilitiesList))

		//Autodetect cameras
		C.gp_camera_autodetect(cameraList, ctx.gpContext)

		size := int(C.gp_list_count(cameraList))

		if size < 0 {
			return nil, newError("Cannot get camera list", size)
		}

		if size == 0 {
			return nil, newError("Unable to detect cameras, Big Fail, Me Sad", size)
		}
		for i := 0; i < size; i++ {
			var cKey *C.char
			var cVal *C.char

			C.gp_list_get_name(cameraList, C.int(i), &cKey)
			C.gp_list_get_value(cameraList, 0, &cVal)
			defer C.free(unsafe.Pointer(cKey))
			defer C.free(unsafe.Pointer(cVal))

			if name == C.GoString(cKey) {
				println("Found camera: " + name)
				if res := C.gp_camera_new((**C.Camera)(unsafe.Pointer(&gpCamera))); res != GPOK {
					return nil, newError("Cannot initialize camera pointer", int(res))
				} else if gpCamera == nil {
					return nil, newError("Cannot initialize camera pointer", Error)
				}

				m := C.gp_abilities_list_lookup_model(abilitiesList, &cKey)
				if m != GPOK {
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Cannot lookup camera model", int(m))
				}

				if res := C.gp_abilities_list_get_abilities(abilitiesList, m, &abilities); res != GPOK {
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Cannot get camera abilities", int(res))
				}

				if res := C.gp_camera_set_abilities(gpCamera, abilities); res != GPOK {
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Cannot set camera abilities", int(res))
				}

				if res := C.gp_port_info_list_new(&portInfoList); res != GPOK {
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Cannot initialize port info list", int(res))
				}

				if res := C.gp_port_info_list_load(portInfoList); res != GPOK {
					C.gp_port_info_list_free(portInfoList)
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Cannot load port info list", int(res))
				}

				p := C.gp_port_info_list_count(portInfoList)
				switch p {
				case C.GP_ERROR_UNKNOWN_PORT:
					C.gp_port_info_list_free(portInfoList)
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Unknown port", int(p))
					break
				default:
					break
				}

				if (p != GPOK) && (p != 0) {
					C.gp_port_info_list_free(portInfoList)
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Cannot get port info list", int(p))
				}

				if res := C.gp_port_info_list_get_info(portInfoList, p, &portInfo); res != GPOK {
					C.gp_port_info_list_free(portInfoList)
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Cannot get port info", int(res))
				}

				if res := C.gp_camera_set_port_info(gpCamera, portInfo); res != GPOK {
					C.gp_port_info_list_free(portInfoList)
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("Cannot set port info", int(res))
				}

				if res := C.gp_camera_init(gpCamera, ctx.gpContext); res != GPOK {
					C.gp_camera_exit(gpCamera, ctx.gpContext)
					C.gp_camera_unref(gpCamera)
					ctx.free()
					return nil, newError("", int(res))
				}

				return &Camera{gpCamera: gpCamera, Ctx: ctx}, nil
			}

		}

	}
	if res := C.gp_camera_new((**C.Camera)(unsafe.Pointer(&gpCamera))); res != GPOK {
		return nil, newError("Cannot initialize camera pointer", int(res))
	} else if gpCamera == nil {
		return nil, newError("Cannot initialize camera pointer", Error)
	}

	if res := C.gp_camera_init(gpCamera, ctx.gpContext); res != GPOK {
		C.gp_camera_exit(gpCamera, ctx.gpContext)
		C.gp_camera_unref(gpCamera)
		ctx.free()
		return nil, newError("", int(res))
	}

	return &Camera{gpCamera: gpCamera, Ctx: ctx}, nil
}

func ListCameras() (CameraList, error) {
	ctx, err := NewContext()
	if err != nil {
		return CameraList{}, err
	}
	st := CameraList{}
	var cameraList *C.CameraList
	C.gp_list_new(&cameraList)
	C.gp_camera_autodetect(cameraList, ctx.gpContext)
	defer C.free(unsafe.Pointer(cameraList))

	size := int(C.gp_list_count(cameraList))

	if size < 0 {
		return CameraList{}, newError("Cannot get camera list", size)
	}

	for i := 0; i < size; i++ {
		var cKey *C.char
		var cVal *C.char

		C.gp_list_get_name(cameraList, C.int(i), &cKey)
		C.gp_list_get_value(cameraList, 0, &cVal)
		defer C.free(unsafe.Pointer(cKey))
		defer C.free(unsafe.Pointer(cVal))

		st.Names = append(st.Names, C.GoString(cKey))
		st.Ports = append(st.Ports, C.GoString(cVal))

	}
	return st, nil
}
