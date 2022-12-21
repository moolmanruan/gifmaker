package gif_test

import (
	"bytes"
	"image"
	"image/color"
	img_gif "image/gif"
	"testing"

	"github.com/moolmanruan/gifmaker/gif"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	input := `
---
b:00,00,00,FF
w:FF,FF,FF,FF
---
b,w
w,b
`
	got := new(bytes.Buffer)
	err := gif.Create(input, got)
	require.Nil(t, err)

	want := new(bytes.Buffer)
	err = img_gif.EncodeAll(want, &img_gif.GIF{
		Delay: []int{1},
		Image: []*image.Paletted{
			{
				Pix: []uint8{
					0, 1,
					1, 0,
				},
				Stride: 2,
				Rect: image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: 2, Y: 2},
				},
				Palette: color.Palette{
					color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
					color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				},
			},
		},
	})
	require.Nil(t, err)

	require.Equal(t, got, want)
}
