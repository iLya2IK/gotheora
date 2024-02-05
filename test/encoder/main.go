/* GoTheora
An example of using an Theora encoder

Copyright (c) 2024 by Ilya Medvedkov

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
*/

package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"time"

	Theora "github.com/ilya2ik/gotheora"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const IMAGES_FOLDER = "images"
const IMAGES_EXT = "PNG"
const OUTPUT_FILE = "output.ogv"
const CFG_CHROMA = image.YCbCrSubsampleRatio444
const CFG_QUALITY = 5     // quality value 0..10
const CFG_BITRATE = 45    // desirable bitrate value kbps
const CFG_DELTATIME = 250 // delta time between two closest frames

func main() {
	/* read the list of files in the specified directory and
	   save it to the frames array */

	frames := make([]string, 0)

	files, err := os.ReadDir(IMAGES_FOLDER)
	if err != nil {
		check(err)
	}

	for _, f := range files {
		if strings.HasSuffix(strings.ToUpper(f.Name()), "."+IMAGES_EXT) {
			frames = append(frames, IMAGES_FOLDER+"/"+f.Name())
		}
	}

	total := len(frames)
	if total == 0 {
		panic("No frames found")
	}

	/* open the first file in the array to read the basic
	   parameters of the frame stack */

	test_reader, err := os.Open(frames[0])
	check(err)
	test_img, err := png.DecodeConfig(test_reader)
	check(err)
	test_reader.Close()

	w := test_img.Width
	h := test_img.Height

	/* initialize the video codec configuration.
	   detailed info: https://www.theora.org/doc/Theora.pdf */

	info, err := Theora.NewTheoraInfo()
	check(err)

	info.SetWidth(((w + 15) >> 4) << 4)
	info.SetHeight(((h + 15) >> 4) << 4)
	info.SetFrameWidth(w)
	info.SetFrameHeight(h)
	info.SetOffsetX(0)
	info.SetOffsetY(0)

	video_fps_denominator := 1
	video_aspect_numerator := 0
	video_aspect_denominator := 0
	video_quality := CFG_QUALITY * 63 / 10
	video_fps_numerator := 1000 / CFG_DELTATIME
	video_rate := CFG_BITRATE * 1000

	info.SetFPSNumerator(video_fps_numerator)
	info.SetFPSDenominator(video_fps_denominator)
	info.SetAspectNumerator(video_aspect_numerator)
	info.SetAspectDenominator(video_aspect_denominator)
	info.SetColorspace(Theora.Unspec)
	info.SetPixelFormat(CFG_CHROMA)
	info.SetTargetBitrate(video_rate)
	info.SetQuality(video_quality)

	info.SetDropFrames(false)
	info.SetQuick(true)
	info.SetKeyframeAuto(true)
	info.SetKeyframeFrequency(32768)
	info.SetKeyframeFrequencyForce(32768)
	info.SetKeyframeDataTargetBitrate(int(float64(video_rate) * 1.5))
	info.SetKeyframeAutoThreshold(80)
	info.SetKeyframeMindistance(8)
	info.SetNoiseSensitivity(1)

	/* Create the output file */

	outf, err := os.Create(OUTPUT_FILE)
	check(err)
	defer outf.Close()

	/* Initialize theora encoder */

	enc, err := Theora.NewTheoraEncoder(info, outf)
	check(err)

	/* Save the basic theora headers and the additional metadata */
	comment, err := Theora.NewTheoraComment()
	check(err)
	comment.AddTag("ENCODED_BY", Theora.Version()+" GoTheora wrapper")
	check(enc.SaveCustomHeadersToStream(comment))

	type frame struct {
		loc int
		img image.Image
	}
	to_enc := make(chan frame)

	/* Open the files from the array of frames, decode them
	   to raster images and encode them as frames in the theora file */

	go func() {
		for i := 0; i < total; i++ {
			reader, err := os.Open(frames[i])
			check(err)
			img, err := png.Decode(reader)
			check(err)
			reader.Close()
			to_enc <- frame{i, img}
		}
		to_enc <- frame{total, nil}
	}()

	for next := true; next; {
		select {
		case frame := <-to_enc:
			{
				if frame.loc == total {
					enc.Close()
					fmt.Printf("Finished")
					next = false
				} else {
					buf, err := Theora.NewTheoraYUVbuffer()
					check(err)
					if buf.ConvertFromRasterImage(CFG_CHROMA, frame.img) {
						check(enc.SaveYUVBufferToStream(buf, frame.loc == (total-1)))
					} else {
						fmt.Printf("Can't ConvertFromRasterImage at frame %d\n", frame.loc)
					}
				}
			}
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

}
