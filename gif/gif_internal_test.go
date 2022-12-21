package gif

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMeta(t *testing.T) {
	cc := []struct {
		in   string
		want meta
	}{
		{"", meta{scale: 1, delay: 1}},
		{"\n", meta{scale: 1, delay: 1}},
		{"scale:10", meta{scale: 10, delay: 1}},
		{"delay:10", meta{scale: 1, delay: 10}},
		{"scale:123\ndelay:321", meta{scale: 123, delay: 321}},
	}
	for _, c := range cc {
		t.Run(c.in, func(t *testing.T) {
			got, err := parseMeta(c.in)
			require.Nil(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

func TestParseMetaFail(t *testing.T) {
	cc := []struct {
		in  string
		err string
	}{
		{"foo:123", "invalid option: `foo:123`"},
	}
	for _, c := range cc {
		t.Run(c.in, func(t *testing.T) {
			_, err := parseMeta(c.in)
			require.ErrorContains(t, err, c.err)
		})
	}
}

func TestParsePalette(t *testing.T) {
	cc := []struct {
		in   string
		want palette
	}{
		{"", palette{}},
		{"\n", palette{}},
		{
			"b:00,00,00,FF",
			palette{
				{"b", color.RGBA{A: 0xff}},
			},
		},
		{
			"b:00,00,00,FF\nw:FF,FF,FF,FF",
			palette{
				{"b", color.RGBA{A: 0xff}},
				{"w", color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}},
			},
		},
		{
			"r:cc,00,00,ff\ng:00,aa,00,ff\nb:12,34,9a,ff",
			palette{
				{"r", color.RGBA{R: 0xcc, A: 0xff}},
				{"g", color.RGBA{G: 0xaa, A: 0xff}},
				{"b", color.RGBA{R: 0x12, G: 0x34, B: 0x9a, A: 0xff}},
			},
		},
	}
	for _, c := range cc {
		t.Run(c.in, func(t *testing.T) {
			got, err := parsePalette(c.in)
			require.Nil(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

func TestParsePaletteFail(t *testing.T) {
	cc := []struct {
		in  string
		err string
	}{
		{"g:00", "invalid option: `g:00`"},
		{"b:43,34", "invalid option: `b:43,34`"},
		{"-:00,00,00,00", "invalid option: `-:00,00,00,00`"},
	}
	for _, c := range cc {
		t.Run(c.in, func(t *testing.T) {
			_, err := parseMeta(c.in)
			require.ErrorContains(t, err, c.err)
		})
	}
}

func TestHexStrToUint8(t *testing.T) {
	cc := []struct {
		in   string
		want int
	}{
		{"00", 0x00},
		{"FF", 0xFF},
		{"ff", 0xFF},

		{"99", 0x99},
		{"AA", 0xAA},
		{"aa", 0xAA},

		{"abc", 0xABC},

		{"ffff", math.MaxUint16},
		{"ffffffff", math.MaxUint32},
	}
	for _, c := range cc {
		t.Run(c.in, func(t *testing.T) {
			require.Equal(t, c.want, hexStrToInt(c.in))
		})
	}
}

func TestNewImage(t *testing.T) {
	cc := []struct {
		in   string
		want img
	}{
		{
			"b",
			img{
				data: []string{"b"},
				w:    1,
				h:    1,
			},
		},
		{
			"a,b,c\nd,e,f",
			img{
				data: []string{"a", "b", "c", "d", "e", "f"},
				w:    3,
				h:    2,
			},
		},
	}
	for _, c := range cc {
		t.Run(c.in, func(t *testing.T) {
			got, err := newImage(c.in)
			require.Nil(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

func TestNewImageFail(t *testing.T) {
	cc := []struct {
		in  string
		err string
	}{
		{
			"",
			"image data is empty",
		},
		{
			"  ",
			"image data is empty",
		},
		{
			"    \n   ",
			"image data is empty",
		},
		{
			"a,b,c\nd,e",
			"image data have rows of different length",
		},
	}
	for _, c := range cc {
		t.Run(c.in, func(t *testing.T) {
			_, err := newImage(c.in)
			require.ErrorContains(t, err, c.err)
		})
	}
}

func TestImgToPaletted(t *testing.T) {
	pal := palette{
		colour{"_", color.RGBA{A: 0xff}},
		colour{"r", color.RGBA{R: 0xff, A: 0xff}},
		colour{"g", color.RGBA{G: 0xff, A: 0xff}},
		colour{"b", color.RGBA{B: 0xff, A: 0xff}},
		colour{"w", color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}},
	}
	cc := []struct {
		name  string
		in    string
		scale int
		want  *image.Paletted
	}{
		{
			"simple 2x1",
			"b,r",
			1,
			&image.Paletted{
				Pix:    []uint8{0x3, 0x1},
				Stride: 2,
				Rect: image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: 2, Y: 1},
				},
				Palette: color.Palette{
					color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
					color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
					color.RGBA{R: 0x0, G: 0xff, B: 0x0, A: 0xff},
					color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff},
					color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				},
			},
		},
		{
			"medium 2x3",
			"_,r\nb,_\ng,w",
			1,
			&image.Paletted{
				Pix:    []uint8{0, 1, 3, 0, 2, 4},
				Stride: 2,
				Rect: image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: 2, Y: 3},
				},
				Palette: color.Palette{
					color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
					color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
					color.RGBA{R: 0x0, G: 0xff, B: 0x0, A: 0xff},
					color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff},
					color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				},
			},
		},
		{
			"scaled 3x2",
			"r,g,b\n_,w,_",
			2,
			&image.Paletted{
				Pix: []uint8{
					1, 1, 2, 2, 3, 3,
					1, 1, 2, 2, 3, 3,
					0, 0, 4, 4, 0, 0,
					0, 0, 4, 4, 0, 0,
				},
				Stride: 6,
				Rect: image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: 6, Y: 4},
				},
				Palette: color.Palette{
					color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
					color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
					color.RGBA{R: 0x0, G: 0xff, B: 0x0, A: 0xff},
					color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff},
					color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				},
			},
		},
	}
	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			i, err := newImage(c.in)
			require.Nil(t, err)
			require.Equal(t, c.want, i.toPaletted(c.scale, pal))
		})
	}
}
