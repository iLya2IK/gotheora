/* GoTheora
Wrapper for Theora library

Copyright (c) 2024 by Ilya Medvedkov

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
*/

package gotheora

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -ltheora -ltheoraenc -ltheoradec
#include "theora/theora.h"
#include "theora/theoradec.h"
#include "theora/theoraenc.h"
#include "theora/codec.h"
#include <stdlib.h>
#include <string.h>

int size_of_struct_yuv_buffer() {
    return sizeof(yuv_buffer);
}

int size_of_struct_theora_info() {
    return sizeof(theora_info);
}

int size_of_struct_theora_state() {
    return sizeof(theora_state);
}

int size_of_struct_theora_comment() {
    return sizeof(theora_comment);
}

*/
import "C"
import (
	"fmt"
	"image"
	"image/color"
	"io"
	"math/rand"
	"runtime"
	"time"
	"unsafe"

	OGG "github.com/ilya2ik/googg"
)

type Colorspace int

const (
	Unspec Colorspace = iota
	ITURec470M
	ITURec470BG
	NSpaces
)

type ITheoraYUVbuffer interface {
	Ref() *C.yuv_buffer

	Done()

	GetYWidth() int
	SetYWidth(value int)
	GetYHeight() int
	SetYHeight(value int)
	GetYStride() int
	SetYStride(value int)

	GetUVWidth() int
	SetUVWidth(value int)
	GetUVHeight() int
	SetUVHeight(value int)
	GetUVStride() int
	SetUVStride(value int)

	GetYData() []byte
	SetYData(value []byte)
	GetUData() []byte
	SetUData(value []byte)
	GetVData() []byte
	SetVData(value []byte)

	GetOwnData() bool
	SetOwnData(value bool)

	ConvertFromRasterImage(chroma_format image.YCbCrSubsampleRatio, aData image.Image) bool
}

type ITheoraInfo interface {
	Ref() *C.theora_info

	Init()
	Done()

	GetAspectDenominator() int
	GetAspectNumerator() int
	GetCodecSetup() unsafe.Pointer
	GetColorspace() Colorspace
	GetDropFrames() bool
	GetFPSDenominator() int
	GetFPSNumerator() int
	GetFrameHeight() int
	GetFrameWidth() int
	GetHeight() int
	GetKeyframeAuto() bool
	GetKeyframeAutoThreshold() int
	GetKeyframeDataTargetBitrate() int
	GetKeyframeFrequency() int
	GetKeyframeFrequencyForce() int
	GetKeyframeMindistance() int
	GetNoiseSensitivity() int
	GetOffsetX() int
	GetOffsetY() int
	GetPixelFormat() image.YCbCrSubsampleRatio
	GetQuality() int
	GetQuick() bool
	GetSharpness() int
	GetTargetBitrate() int
	GetWidth() int

	GetVersionMajor() byte
	GetVersionMinor() byte
	GetVersionSubminor() byte
	SetAspectDenominator(AValue int)
	SetAspectNumerator(AValue int)
	SetCodecSetup(AValue unsafe.Pointer)
	SetColorspace(AValue Colorspace)
	SetDropFrames(AValue bool)
	SetFPSDenominator(AValue int)
	SetFPSNumerator(AValue int)
	SetFrameHeight(AValue int)
	SetFrameWidth(AValue int)
	SetHeight(AValue int)
	SetKeyframeAuto(AValue bool)
	SetKeyframeAutoThreshold(AValue int)
	SetKeyframeDataTargetBitrate(AValue int)
	SetKeyframeFrequency(AValue int)
	SetKeyframeFrequencyForce(AValue int)
	SetKeyframeMindistance(AValue int)
	SetNoiseSensitivity(AValue int)
	SetOffsetX(AValue int)
	SetOffsetY(AValue int)
	SetPixelFormat(AValue image.YCbCrSubsampleRatio)
	SetQuality(AValue int)
	SetQuick(AValue bool)
	SetSharpness(AValue int)
	SetTargetBitrate(AValue int)
	SetWidth(AValue int)

	GranuleShift() int
}

type ITheoraComment interface {
	Ref() *C.theora_comment

	Init()
	Done()

	GetVendor() string
	SetVendor(s string)

	Add(comment string)
	AddTag(tag, value string)
	TagsCount() int
	GetTag(index int) string
	Query(tag string, index int) string
	QueryCount(tag string) int
}

type ITheoraState interface {
	Ref() *C.theora_state

	Init(inf ITheoraInfo)
	Done()

	Info() ITheoraInfo
	GetGranulePos() int64
	SetGranulePos(value int64)

	GranuleFrame(granulepos int64) int64
	GranuleTime(granulepos int64) float64
}

type ITheoraEncoder interface {
	Header(op OGG.IOGGPacket) error
	PacketOut(last_p bool, op OGG.IOGGPacket) error
	DoPacketOut(last_p bool) (OGG.IOGGPacket, error)
	YUVin(yuv ITheoraYUVbuffer) error
	Comment(tc ITheoraComment, op OGG.IOGGPacket) error
	Tables(op OGG.IOGGPacket) error

	Control(req int, buf []byte) int

	SaveDefHeadersToStream() error
	SaveCustomHeadersToStream(tc ITheoraComment) error
	SaveYUVBufferToStream(buf ITheoraYUVbuffer, is_last bool) error
	Flush() error
	Close() error
}

type ITheoraDecoder interface {
	Header(cc ITheoraComment, op OGG.IOGGPacket) error
	PacketIn(op OGG.IOGGPacket) error
	YUVout(yuv ITheoraYUVbuffer) error
}

/* Exceptions */

type errTheoraException struct{ r int }

var ETheoraException = errTheoraException{0}

func (v errTheoraException) Error() string {
	switch v.r {
	case C.OC_FAULT:
		return "Unspecified error"
	case C.OC_EINVAL:
		return "General failure"
	case C.OC_DISABLED:
		return "Requested action is disabled"
	case C.OC_BADHEADER:
		return "Header packet was corrupt/invalid"
	case C.OC_NOTFORMAT:
		return "Packet is not a theora packet"
	case C.OC_VERSION:
		return "Bitstream version is not handled"
	case C.OC_IMPL:
		return "Feature or action not implemented"
	case C.OC_BADPACKET:
		return "Packet is corrupt"
	case C.OC_NEWPACKET:
		return "Packet is an (ignorable) unhandled extension"
	case C.OC_DUPFRAME:
		return "Packet is a dropped frame"
	default:
		return "Unspecified encoder error"
	}
}

type errTheoraDecException struct{}

var ETheoraDecException = errTheoraDecException{}

func (v errTheoraDecException) Error() string {
	return "Unspecified decoder error"
}

type errTheoraDecBadPacketException struct{}

var ETheoraDecBadPacketException = errTheoraDecBadPacketException{}

func (v errTheoraDecBadPacketException) Error() string {
	return "Packet is corrupt. Packet does not contain encoded video data"
}

type errTheoraEncException struct{}

var ETheoraEncException = errTheoraEncException{}

func (v errTheoraEncException) Error() string {
	return "Unspecified encoder error"
}

type errTheoraEncCompletedException struct{}

var ETheoraEncCompletedException = errTheoraEncCompletedException{}

func (v errTheoraEncCompletedException) Error() string {
	return "The encoding process has completed"
}

type errTheoraEncNotReadyException struct{}

var ETheoraEncNotReadyException = errTheoraEncNotReadyException{}

func (v errTheoraEncNotReadyException) Error() string {
	return "Encoder is not ready"
}

type errTheoraEncPackNotReadyException struct{}

var ETheoraEncNotPackReadyException = errTheoraEncPackNotReadyException{}

func (v errTheoraEncPackNotReadyException) Error() string {
	return "No internal storage exists OR no packet is ready"
}

type errTheoraEncDifferException struct{}

var ETheoraEncDifferException = errTheoraEncDifferException{}

func (v errTheoraEncDifferException) Error() string {
	return "The size of the given frame differs from those previously input"
}

type errTheoraOutOfMemory struct{ err error }

func (v errTheoraOutOfMemory) Error() string {
	if v.err != nil {
		return fmt.Sprintf("Fatal. Out of memory. %s", v.err.Error())
	} else {
		return "Fatal. Out of memory"
	}
}

var ETheoraOutOfMemory = errTheoraOutOfMemory{nil}

/* Common methods */

func Version() string {
	return C.GoString(C.theora_version_string())
}

func VersionNumber() uint32 {
	return uint32(C.theora_version_number())
}

/* TheoraComment */

type TheoraComment struct {
	fValue *C.theora_comment
}

func NewTheoraComment() (ITheoraComment, error) {
	value := new(TheoraComment)

	mem, err := C.calloc(1, (C.size_t)(C.size_of_struct_theora_comment()))
	if mem == nil {
		return nil, errTheoraOutOfMemory{err}
	}
	value.fValue = (*C.theora_comment)(mem)
	runtime.SetFinalizer(value, func(a *TheoraComment) {
		if a.fValue != nil {
			a.Done()
		}
	})
	return value, nil
}

func (v *TheoraComment) Ref() *C.theora_comment {
	return v.fValue
}

func (v *TheoraComment) Init() {
	C.theora_comment_init(v.Ref())
}

func (v *TheoraComment) Done() {
	if v.fValue != nil {
		C.theora_comment_clear(v.Ref())
		C.free(unsafe.Pointer(v.fValue))
		v.fValue = nil
	}
}

func (v *TheoraComment) GetVendor() string {
	return C.GoString(v.Ref().vendor)
}

func (v *TheoraComment) SetVendor(s string) {
	panic("not implemented") // TODO: Implement
}

func (v *TheoraComment) Add(comment string) {
	cs := C.CString(comment)
	defer C.free(unsafe.Pointer(cs))
	C.theora_comment_add(v.Ref(), cs)
}

func (v *TheoraComment) AddTag(tag string, value string) {
	ctag := C.CString(tag)
	defer C.free(unsafe.Pointer(ctag))
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))
	C.theora_comment_add_tag(v.Ref(), ctag, cvalue)
}

func (v *TheoraComment) TagsCount() int {
	panic("not implemented") // TODO: Implement
}

func (v *TheoraComment) GetTag(index int) string {
	panic("not implemented") // TODO: Implement
}

func (v *TheoraComment) Query(tag string, index int) string {
	ctag := C.CString(tag)
	defer C.free(unsafe.Pointer(ctag))
	return C.GoString(C.theora_comment_query(v.Ref(), ctag, C.int(index)))
}

func (v *TheoraComment) QueryCount(tag string) int {
	ctag := C.CString(tag)
	defer C.free(unsafe.Pointer(ctag))
	return int(C.theora_comment_query_count(v.Ref(), ctag))
}

/* TheoraInfo */

type TheoraInfo struct {
	fValue *C.theora_info
}

func NewTheoraInfo() (ITheoraInfo, error) {
	value := new(TheoraInfo)

	mem, err := C.calloc(1, (C.size_t)(C.size_of_struct_theora_info()))
	if mem == nil {
		return nil, errTheoraOutOfMemory{err}
	}
	value.fValue = (*C.theora_info)(mem)
	runtime.SetFinalizer(value, func(a *TheoraInfo) {
		a.Done()
	})
	return value, nil
}

func (v *TheoraInfo) Ref() *C.theora_info {
	return v.fValue
}

func (v *TheoraInfo) Init() {
	C.theora_info_init(v.Ref())
}

func (v *TheoraInfo) Done() {
	if v.fValue != nil {
		C.theora_info_clear(v.fValue)
		C.free(unsafe.Pointer(v.fValue))
		v.fValue = nil
	}
}

func (v *TheoraInfo) GetAspectDenominator() int {
	return int(v.fValue.aspect_denominator)
}

func (v *TheoraInfo) GetAspectNumerator() int {
	return int(v.fValue.aspect_numerator)
}

func (v *TheoraInfo) GetCodecSetup() unsafe.Pointer {
	return v.fValue.codec_setup
}

func (v *TheoraInfo) GetColorspace() Colorspace {
	return Colorspace(v.fValue.colorspace)
}

func (v *TheoraInfo) GetDropFrames() bool {
	return (v.fValue.dropframes_p > 0)
}

func (v *TheoraInfo) GetFPSDenominator() int {
	return int(v.fValue.fps_denominator)
}

func (v *TheoraInfo) GetFPSNumerator() int {
	return int(v.fValue.fps_numerator)
}

func (v *TheoraInfo) GetFrameHeight() int {
	return int(v.fValue.frame_height)
}

func (v *TheoraInfo) GetFrameWidth() int {
	return int(v.fValue.frame_width)
}

func (v *TheoraInfo) GetHeight() int {
	return int(v.fValue.height)
}

func (v *TheoraInfo) GetKeyframeAuto() bool {
	return (v.fValue.keyframe_auto_p > 0)
}

func (v *TheoraInfo) GetKeyframeAutoThreshold() int {
	return int(v.fValue.keyframe_auto_threshold)
}

func (v *TheoraInfo) GetKeyframeDataTargetBitrate() int {
	return int(v.fValue.keyframe_data_target_bitrate)
}

func (v *TheoraInfo) GetKeyframeFrequency() int {
	return int(v.fValue.keyframe_frequency)
}

func (v *TheoraInfo) GetKeyframeFrequencyForce() int {
	return int(v.fValue.keyframe_frequency_force)
}

func (v *TheoraInfo) GetKeyframeMindistance() int {
	return int(v.fValue.keyframe_mindistance)
}

func (v *TheoraInfo) GetNoiseSensitivity() int {
	return int(v.fValue.noise_sensitivity)
}

func (v *TheoraInfo) GetOffsetX() int {
	return int(v.fValue.offset_x)
}

func (v *TheoraInfo) GetOffsetY() int {
	return int(v.fValue.offset_y)
}

func (v *TheoraInfo) GetPixelFormat() image.YCbCrSubsampleRatio {
	switch v.fValue.pixelformat {
	case 0:
		{
			return image.YCbCrSubsampleRatio420
		}
	case 2:
		{
			return image.YCbCrSubsampleRatio422
		}
	case 3:
		{
			return image.YCbCrSubsampleRatio444
		}
	}
	return image.YCbCrSubsampleRatio410
}

func (v *TheoraInfo) GetQuality() int {
	return int(v.fValue.quality)
}

func (v *TheoraInfo) GetQuick() bool {
	return (v.fValue.quick_p > 0)
}

func (v *TheoraInfo) GetSharpness() int {
	return int(v.fValue.sharpness)
}

func (v *TheoraInfo) GetTargetBitrate() int {
	return int(v.fValue.target_bitrate)
}

func (v *TheoraInfo) GetWidth() int {
	return int(v.fValue.width)
}

func (v *TheoraInfo) GetVersionMajor() byte {
	return byte(v.fValue.version_major)
}

func (v *TheoraInfo) GetVersionMinor() byte {
	return byte(v.fValue.version_minor)
}

func (v *TheoraInfo) GetVersionSubminor() byte {
	return byte(v.fValue.version_subminor)
}

func (v *TheoraInfo) SetAspectDenominator(AValue int) {
	v.fValue.aspect_denominator = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetAspectNumerator(AValue int) {
	v.fValue.aspect_numerator = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetCodecSetup(AValue unsafe.Pointer) {
	v.fValue.codec_setup = AValue
}

func (v *TheoraInfo) SetColorspace(AValue Colorspace) {
	v.fValue.colorspace = C.theora_colorspace(AValue)
}

func (v *TheoraInfo) SetDropFrames(AValue bool) {
	if AValue {
		v.fValue.dropframes_p = C.int(0)
	} else {
		v.fValue.dropframes_p = C.int(1)
	}
}

func (v *TheoraInfo) SetFPSDenominator(AValue int) {
	v.fValue.fps_denominator = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetFPSNumerator(AValue int) {
	v.fValue.fps_numerator = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetFrameHeight(AValue int) {
	v.fValue.frame_height = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetFrameWidth(AValue int) {
	v.fValue.frame_width = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetHeight(AValue int) {
	v.fValue.height = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetKeyframeAuto(AValue bool) {
	if AValue {
		v.fValue.keyframe_auto_p = C.int(0)
	} else {
		v.fValue.keyframe_auto_p = C.int(1)
	}
}

func (v *TheoraInfo) SetKeyframeAutoThreshold(AValue int) {
	v.fValue.keyframe_auto_threshold = C.int32_t(AValue)
}

func (v *TheoraInfo) SetKeyframeDataTargetBitrate(AValue int) {
	v.fValue.keyframe_data_target_bitrate = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetKeyframeFrequency(AValue int) {
	v.fValue.keyframe_frequency = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetKeyframeFrequencyForce(AValue int) {
	v.fValue.keyframe_frequency_force = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetKeyframeMindistance(AValue int) {
	v.fValue.keyframe_mindistance = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetNoiseSensitivity(AValue int) {
	v.fValue.noise_sensitivity = C.int32_t(AValue)
}

func (v *TheoraInfo) SetOffsetX(AValue int) {
	v.fValue.offset_x = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetOffsetY(AValue int) {
	v.fValue.offset_y = C.uint32_t(AValue)
}

func (v *TheoraInfo) SetPixelFormat(AValue image.YCbCrSubsampleRatio) {
	switch AValue {
	case image.YCbCrSubsampleRatio420:
		v.fValue.pixelformat = C.OC_PF_420
	case image.YCbCrSubsampleRatio422:
		v.fValue.pixelformat = C.OC_PF_422
	case image.YCbCrSubsampleRatio444:
		v.fValue.pixelformat = C.OC_PF_444
	}
}

func (v *TheoraInfo) SetQuality(AValue int) {
	v.fValue.quality = C.int(AValue)
}

func (v *TheoraInfo) SetQuick(AValue bool) {
	if AValue {
		v.fValue.quick_p = C.int(1)
	} else {
		v.fValue.quick_p = C.int(0)
	}
}

func (v *TheoraInfo) SetSharpness(AValue int) {
	v.fValue.sharpness = C.int32_t(AValue)
}

func (v *TheoraInfo) SetTargetBitrate(AValue int) {
	v.fValue.target_bitrate = C.int(AValue)
}

func (v *TheoraInfo) SetWidth(AValue int) {
	v.fValue.width = C.uint32_t(AValue)
}

func (v *TheoraInfo) GranuleShift() int {
	return int(C.theora_granule_shift(v.Ref()))
}

/* TheoraState */

type TheoraState struct {
	fValue *C.theora_state
	info   ITheoraInfo
}

func NewTheoraState() (ITheoraState, error) {
	value := new(TheoraState)

	mem, err := C.calloc(1, (C.size_t)(C.size_of_struct_theora_state()))
	if mem == nil {
		return nil, errTheoraOutOfMemory{err}
	}
	value.fValue = (*C.theora_state)(mem)
	runtime.SetFinalizer(value, func(a *TheoraState) {
		a.Done()
	})
	return value, nil
}

func (v *TheoraState) Ref() *C.theora_state {
	return v.fValue
}

func (v *TheoraState) Init(inf ITheoraInfo) {
	v.Ref().i = inf.Ref()
	v.info = inf
}

func (v *TheoraState) Done() {
	if v.fValue != nil {
		v.fValue.i = nil
		C.theora_clear(v.fValue)
		C.free(unsafe.Pointer(v.fValue))
		v.fValue = nil
	}
}

func (v *TheoraState) Info() ITheoraInfo {
	return v.info
}

func (v *TheoraState) GetGranulePos() int64 {
	return int64(v.Ref().granulepos)
}

func (v *TheoraState) SetGranulePos(value int64) {
	v.Ref().granulepos = C.int64_t(value)
}

func (v *TheoraState) GranuleFrame(granulepos int64) int64 {
	return int64(C.theora_granule_frame(v.Ref(), C.int64_t(granulepos)))
}

func (v *TheoraState) GranuleTime(granulepos int64) float64 {
	return float64(C.theora_granule_time(v.Ref(), C.int64_t(granulepos)))
}

/* TheoraYUVbuffer */

type TheoraYUVbuffer struct {
	fValue   *C.yuv_buffer
	fyData   []byte
	fuData   []byte
	fvData   []byte
	fOwnData bool
}

func NewTheoraYUVbuffer() (ITheoraYUVbuffer, error) {
	value := new(TheoraYUVbuffer)

	mem, err := C.calloc(1, (C.size_t)(C.size_of_struct_yuv_buffer()))
	if mem == nil {
		return nil, errTheoraOutOfMemory{err}
	}
	value.fValue = (*C.yuv_buffer)(mem)
	runtime.SetFinalizer(value, func(a *TheoraYUVbuffer) {
		a.Done()
	})
	return value, nil
}

func (v *TheoraYUVbuffer) Ref() *C.yuv_buffer {
	return v.fValue
}

func (v *TheoraYUVbuffer) Done() {
	if v.fValue != nil {
		if v.GetOwnData() {
			if v.fValue.y != nil {
				C.free(unsafe.Pointer(v.fValue.y))
				v.fValue.y = nil
			}
			if v.fValue.u != nil {
				C.free(unsafe.Pointer(v.fValue.u))
				v.fValue.u = nil
			}
			if v.fValue.v != nil {
				C.free(unsafe.Pointer(v.fValue.v))
				v.fValue.v = nil
			}
		}
		C.free(unsafe.Pointer(v.fValue))
		v.fValue = nil
	}
}

func (v *TheoraYUVbuffer) GetYWidth() int {
	return int(v.fValue.y_width)
}

func (v *TheoraYUVbuffer) SetYWidth(value int) {
	v.fValue.y_width = C.int(value)
}

func (v *TheoraYUVbuffer) GetYHeight() int {
	return int(v.fValue.y_height)
}

func (v *TheoraYUVbuffer) SetYHeight(value int) {
	v.fValue.y_height = C.int(value)
}

func (v *TheoraYUVbuffer) GetYStride() int {
	return int(v.fValue.y_stride)
}

func (v *TheoraYUVbuffer) SetYStride(value int) {
	v.fValue.y_stride = C.int(value)
}

func (v *TheoraYUVbuffer) GetUVWidth() int {
	return int(v.fValue.uv_width)
}

func (v *TheoraYUVbuffer) SetUVWidth(value int) {
	v.fValue.uv_width = C.int(value)
}

func (v *TheoraYUVbuffer) GetUVHeight() int {
	return int(v.fValue.uv_height)
}

func (v *TheoraYUVbuffer) SetUVHeight(value int) {
	v.fValue.uv_height = C.int(value)
}

func (v *TheoraYUVbuffer) GetUVStride() int {
	return int(v.fValue.uv_stride)
}

func (v *TheoraYUVbuffer) SetUVStride(value int) {
	v.fValue.uv_stride = C.int(value)
}

func (v *TheoraYUVbuffer) GetYData() []byte {
	return v.fyData
}

func (v *TheoraYUVbuffer) SetYData(value []byte) {
	v.fyData = value
	v.fValue.y = (*C.uchar)(OGG.Ptr(v.fyData[0:]))
}

func (v *TheoraYUVbuffer) GetUData() []byte {
	return v.fuData
}

func (v *TheoraYUVbuffer) SetUData(value []byte) {
	v.fuData = value
	v.fValue.u = (*C.uchar)(OGG.Ptr(v.fuData[0:]))
}

func (v *TheoraYUVbuffer) GetVData() []byte {
	return v.fvData
}

func (v *TheoraYUVbuffer) SetVData(value []byte) {
	v.fvData = value
	v.fValue.v = (*C.uchar)(OGG.Ptr(v.fvData[0:]))
}

func (v *TheoraYUVbuffer) GetOwnData() bool {
	return v.fOwnData
}

func (v *TheoraYUVbuffer) SetOwnData(value bool) {
	v.fOwnData = value
}

func (v *TheoraYUVbuffer) ConvertFromRasterImage(chroma_format image.YCbCrSubsampleRatio, aData image.Image) bool {

	/* increadable awfully */
	nrgb := func(v color.Color) (uint32, uint32, uint32) {
		c := color.NRGBAModel.Convert(v).(color.NRGBA)
		return uint32(c.R), uint32(c.G), uint32(c.B)
	}

	clamp := func(v uint32) byte {
		if v > 255 {
			return 255
		}
		return byte(v)
	}

	booltoint := func(v bool) int {
		if v {
			return 1
		}
		return 0
	}

	if !(chroma_format == image.YCbCrSubsampleRatio444 ||
		chroma_format == image.YCbCrSubsampleRatio422 ||
		chroma_format == image.YCbCrSubsampleRatio420) {
		return false
	}

	h := aData.Bounds().Dy()
	w := aData.Bounds().Dx()

	// Must hold: yuv_w >= w
	var yuv_w int = int(uint32(w+15) & ^uint32(0xf))
	// Must hold: yuv_h >= h
	var yuv_h int = int(uint32(h+15) & ^uint32(0xf))

	v.SetYWidth(yuv_w)
	v.SetYHeight(yuv_h)
	v.SetYStride(yuv_w)

	if chroma_format == image.YCbCrSubsampleRatio444 {
		v.SetUVWidth(yuv_w)
	} else {
		v.SetUVWidth(yuv_w >> 1)
	}
	v.SetUVStride(v.GetUVWidth())

	if chroma_format == image.YCbCrSubsampleRatio420 {
		v.SetUVHeight(yuv_h >> 1)
	} else {
		v.SetUVHeight(yuv_h)
	}

	yuv_y := make([]byte, v.GetYStride()*v.GetYHeight())
	yuv_u := make([]byte, v.GetUVStride()*v.GetUVHeight())
	yuv_v := make([]byte, v.GetUVStride()*v.GetUVHeight())

	v.SetYData(yuv_y)
	v.SetUData(yuv_u)
	v.SetVData(yuv_v)

	if chroma_format == image.YCbCrSubsampleRatio420 {
		y := 0
		for y < h {
			y1 := y + booltoint((y+1) < h)
			x := 0
			for x < w {
				x1 := x + booltoint((x+1) < w)
				r0, g0, b0 := nrgb(aData.At(x, y))
				r1, g1, b1 := nrgb(aData.At(x1, y))
				r2, g2, b2 := nrgb(aData.At(x, y1))
				r3, g3, b3 := nrgb(aData.At(x1, y1))

				yuv_y[x+y*yuv_w] = clamp((65481*r0 + 128553*g0 + 24966*b0 + 4207500) / 255000)
				yuv_y[x1+y*yuv_w] = clamp((65481*r1 + 128553*g1 + 24966*b1 + 4207500) / 255000)
				yuv_y[x+y1*yuv_w] = clamp((65481*r2 + 128553*g2 + 24966*b2 + 4207500) / 255000)
				yuv_y[x1+y1*yuv_w] = clamp((65481*r3 + 128553*g3 + 24966*b3 + 4207500) / 255000)

				yuv_u[(x>>1)+(y>>1)*v.GetUVStride()] =
					clamp(((29032005-33488*r0-65744*g0+99232*b0)/4 +
						(29032005-33488*r1-65744*g1+99232*b1)/4 +
						(29032005-33488*r2-65744*g2+99232*b2)/4 +
						(29032005-33488*r3-65744*g3+99232*b3)/4) / 225930)
				yuv_v[(x>>1)+(y>>1)*v.GetUVStride()] =
					clamp(((157024*r0-131488*g0-25536*b0+45940035)/4 +
						(157024*r1-131488*g1-25536*b1+45940035)/4 +
						(157024*r2-131488*g2-25536*b2+45940035)/4 +
						(157024*r3-131488*g3-25536*b3+45940035)/4) / 357510)
				x += 2
			}
			y += 2
		}
	} else if chroma_format == image.YCbCrSubsampleRatio444 {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r0, g0, b0 := nrgb(aData.At(x, y))

				yuv_y[x+y*yuv_w] = clamp((65481*r0 + 128553*g0 + 24966*b0 + 4207500) / 255000)
				yuv_u[x+y*yuv_w] = clamp((29032005 - 33488*r0 - 65744*g0 + 99232*b0) / 225930)
				yuv_v[x+y*yuv_w] = clamp((157024*r0 - 131488*g0 - 25536*b0 + 45940035) / 357510)
			}
		}
	} else { /* TH_PF_422 */
		y := 0
		for y < h {
			x := 0
			for x < w {
				x1 := x + booltoint((x+1) < w)
				r0, g0, b0 := nrgb(aData.At(x, y))
				r1, g1, b1 := nrgb(aData.At(x1, y))

				yuv_y[x+y*yuv_w] = clamp((65481*r0 + 128553*g0 + 24966*b0 + 4207500) / 255000)
				yuv_y[x1+y*yuv_w] = clamp((65481*r1 + 128553*g1 + 24966*b1 + 4207500) / 255000)

				yuv_u[(x>>1)+y*v.GetUVStride()] =
					clamp(((29032005-33488*r0-65744*g0+99232*b0)/2 +
						(29032005-33488*r1-65744*g1+99232*b1)/2) / 225930)
				yuv_v[(x>>1)+y*v.GetUVStride()] =
					clamp(((157024*r0-131488*g0-25536*b0+45940035)/2 +
						(157024*r1-131488*g1-25536*b1+45940035)/2) / 357510)
				x += 2
			}
			y++
		}
	}
	return true
}

/* TheoraEncoder */

type TheoraEncoder struct {
	fState  ITheoraState
	foggs   OGG.IOGGStreamState
	fwriter io.Writer
}

func NewTheoraEncoder(inf ITheoraInfo, str io.Writer) (ITheoraEncoder, error) {
	value := new(TheoraEncoder)
	var err error
	value.fState, err = NewTheoraState()
	if err != nil {
		return nil, err
	}
	value.fState.Init(inf)
	R := int(C.theora_encode_init(value.fState.Ref(), inf.Ref()))
	if R != 0 {
		return nil, errTheoraException{R}
	}
	value.foggs, err = OGG.NewStream(int32(rand.Int63n(time.Now().UnixMilli())))
	if err != nil {
		return nil, err
	}
	value.fwriter = str

	runtime.SetFinalizer(value, func(a *TheoraEncoder) {
		if value.fState != nil {
			value.fState.Done()
			value.fState = nil
		}
	})
	return value, nil
}

func (v *TheoraEncoder) State() ITheoraState {
	return v.fState
}

func (v *TheoraEncoder) Stream() OGG.IOGGStreamState {
	return v.foggs
}

func (v *TheoraEncoder) Header(op OGG.IOGGPacket) error {
	R := int(C.theora_encode_header(v.fState.Ref(), (*C.ogg_packet)(unsafe.Pointer(op.Ref()))))
	if R != 0 {
		return errTheoraException{R}
	}
	return nil
}

func (v *TheoraEncoder) PacketOut(last_p bool, op OGG.IOGGPacket) error {
	var lp C.int
	if last_p {
		lp = 1
	} else {
		lp = 0
	}

	R := int(C.theora_encode_packetout(v.fState.Ref(), lp, (*C.ogg_packet)(unsafe.Pointer(op.Ref()))))
	if R == 1 {
		return nil
	} else if R == -1 {
		return ETheoraEncCompletedException
	} else if R == 0 {
		return ETheoraEncNotReadyException
	} else {
		return errTheoraException{R}
	}
}

func (v *TheoraEncoder) DoPacketOut(last_p bool) (OGG.IOGGPacket, error) {
	p, err := OGG.NewPacket()
	if err != nil {
		return nil, err
	}
	return p, v.PacketOut(last_p, p)
}

func (v *TheoraEncoder) YUVin(yuv ITheoraYUVbuffer) error {
	R := int(C.theora_encode_YUVin(v.fState.Ref(), yuv.Ref()))
	if R == 0 {
		return nil
	} else if R == -1 {
		return ETheoraEncDifferException
	} else if R == C.OC_EINVAL {
		return ETheoraEncNotReadyException
	} else {
		return errTheoraException{R}
	}
}

func (v *TheoraEncoder) Comment(tc ITheoraComment, op OGG.IOGGPacket) error {
	R := int(C.theora_encode_comment(tc.Ref(), (*C.ogg_packet)(unsafe.Pointer(op.Ref()))))
	if R != 0 {
		return errTheoraException{R}
	}
	return nil
}

func (v *TheoraEncoder) Tables(op OGG.IOGGPacket) error {
	R := int(C.theora_encode_tables(v.fState.Ref(), (*C.ogg_packet)(unsafe.Pointer(op.Ref()))))
	if R != 0 {
		return errTheoraException{R}
	}
	return nil
}

func (v *TheoraEncoder) Control(req int, buf []byte) int {
	panic("not implemented") // TODO: Implement
}

func (v *TheoraEncoder) SaveDefHeadersToStream() error {
	tc, err := NewTheoraComment()
	if err != nil {
		return err
	}
	return v.SaveCustomHeadersToStream(tc)
}

func (v *TheoraEncoder) SaveCustomHeadersToStream(tc ITheoraComment) error {
	op, err := OGG.NewPacket()
	if err != nil {
		return err
	}
	err = v.Header(op)
	if err != nil {
		return err
	}
	err = v.foggs.SavePacketToStream(v.fwriter, op)
	if err != nil {
		return err
	}
	err = v.Comment(tc, op)
	if err != nil {
		return err
	}
	err = v.foggs.PacketIn(op)
	if err != nil {
		return err
	}
	err = v.Tables(op)
	if err != nil {
		return err
	}
	err = v.foggs.PacketIn(op)
	if err != nil {
		return err
	}
	err = v.foggs.SavePacketToStream(v.fwriter, op)
	if err != nil {
		return err
	}
	return nil
}

func (v *TheoraEncoder) SaveYUVBufferToStream(buf ITheoraYUVbuffer, is_last bool) error {
	err := v.YUVin(buf)
	if err != nil {
		return err
	}
	op, err := v.DoPacketOut(is_last)
	if err != nil {
		return err
	}
	err = v.foggs.SavePacketToStream(v.fwriter, op)
	if err != nil {
		return err
	}
	return nil
}

func (v *TheoraEncoder) Flush() error {
	return v.foggs.PagesFlushToStream(v.fwriter)
}

func (v *TheoraEncoder) Close() error {
	err := v.Flush()
	if err != nil {
		return err
	}
	if v.foggs != nil {
		v.foggs.Done()
		v.foggs = nil
	}
	return nil
}

/* TheoraDecoder */

type TheoraDecoder struct {
	fState ITheoraState
}

func NewTheoraDecoder(inf ITheoraInfo) (ITheoraDecoder, error) {
	value := new(TheoraDecoder)
	var err error
	value.fState, err = NewTheoraState()
	if err != nil {
		return nil, err
	}
	value.fState.Init(inf)
	R := int(C.theora_decode_init(value.fState.Ref(), inf.Ref()))
	if R != 0 {
		return nil, errTheoraException{R}
	}

	runtime.SetFinalizer(value, func(a *TheoraDecoder) {
		if value.fState != nil {
			value.fState.Done()
			value.fState = nil
		}
	})
	return value, nil
}

func (v *TheoraDecoder) State() ITheoraState {
	return v.fState
}

func (v *TheoraDecoder) Header(cc ITheoraComment, op OGG.IOGGPacket) error {
	R := int(C.theora_decode_header(v.fState.Info().Ref(), cc.Ref(), (*C.ogg_packet)(unsafe.Pointer(op.Ref()))))
	if R != 0 {
		return errTheoraException{R}
	}
	return nil
}

func (v *TheoraDecoder) PacketIn(op OGG.IOGGPacket) error {
	R := int(C.theora_decode_packetin(v.fState.Ref(), (*C.ogg_packet)(unsafe.Pointer(op.Ref()))))
	if R == 0 {
		return nil
	} else if R == C.OC_BADPACKET {
		return ETheoraDecBadPacketException
	} else {
		return errTheoraException{R}
	}
}

func (v *TheoraDecoder) YUVout(yuv ITheoraYUVbuffer) error {
	R := int(C.theora_decode_YUVout(v.fState.Ref(), yuv.Ref()))
	if R != 0 {
		return errTheoraException{R}
	}
	return nil
}
