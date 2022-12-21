package gif

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	image_gif "image/gif"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type meta struct {
	scale int
	delay int
}

type palette []colour

type colour struct {
	name  string
	value color.Color
}

type img struct {
	data  []string
	w, h  int
	delay int
}

func Create(input string, w io.Writer) error {
	parts := strings.Split(input, "---")
	if len(parts) != 3 {
		return fmt.Errorf("invalid input file: expecting meta, palette, and image sections separated by `---` lines")
	}
	inMeta, err := parseMeta(parts[0])
	if err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	inPalette, err := parsePalette(parts[1])
	if err != nil {
		return fmt.Errorf("invalid palette: %w", err)
	}

	imgs, err := parseImages(parts[2])
	if err != nil {
		return fmt.Errorf("invalid imgs: %w", err)
	}

	paletteImages := make([]*image.Paletted, len(imgs))
	delays := make([]int, len(imgs))
	for idx, i := range imgs {
		if i.delay > 0 {
			delays[idx] = i.delay
		} else {
			delays[idx] = inMeta.delay
		}
		paletteImages[idx] = i.toPaletted(inMeta.scale, inPalette)
	}

	return image_gif.EncodeAll(w,
		&image_gif.GIF{Image: paletteImages, Delay: delays})
}

func parseMeta(s string) (meta, error) {
	m := meta{
		scale: 1,
		delay: 1,
	}
	for _, l := range strings.Split(s, "\n") {
		line := strings.TrimSpace(l)
		if line == "" {
			continue // ignore empty lines
		}
		parts := strings.Split(line, ":")
		switch parts[0] {
		case "scale":
			scale, err := strconv.Atoi(parts[1])
			if err != nil {
				return m, errors.New("invalid scale")
			}
			m.scale = scale
		case "delay":
			delay, err := strconv.Atoi(parts[1])
			if err != nil {
				return m, errors.New("invalid scale")
			}
			m.delay = delay
		default:
			return m, fmt.Errorf("invalid option: `%s`", l)
		}
	}
	if m.scale < 1 {
		m.scale = 1
	}
	if m.delay < 1 {
		m.delay = 1
	}
	return m, nil
}

func parsePalette(s string) (palette, error) {
	p := palette{}
	for _, l := range strings.Split(s, "\n") {
		line := strings.TrimSpace(l)
		if line == "" {
			continue // ignore empty lines
		}
		c, cErr := newColour(line)
		if cErr != nil {
			return nil, cErr
		}
		p = append(p, c)
	}
	return p, nil
}

func parseImages(s string) ([]img, error) {
	var ii []img
	for _, i := range strings.Split(s, "-") {
		imgText := strings.TrimSpace(i)
		if imgText == "" {
			continue // ignore empty strings
		}
		newImg, err := newImage(imgText)
		if err != nil {
			return nil, err
		}
		ii = append(ii, newImg)
	}
	return ii, nil
}

var colorRegex = regexp.MustCompile(`(\w+):([0-9a-fA-F]{2}),([0-9a-fA-F]{2}),([0-9a-fA-F]{2}),([0-9a-fA-F]{2})`)

func newColour(input string) (colour, error) {
	m := colorRegex.FindStringSubmatch(input)
	if len(m) == 0 {
		return colour{}, errors.New(fmt.Sprintf("failed to parse colour: `%s`", input))
	}
	return colour{
		name: m[1],
		value: color.RGBA{
			R: uint8(hexStrToInt(m[2])),
			G: uint8(hexStrToInt(m[3])),
			B: uint8(hexStrToInt(m[4])),
			A: uint8(hexStrToInt(m[5])),
		},
	}, nil
}

func hexStrToInt(s string) int {
	var v int
	rr := []rune(s)
	for i := 0; i < len(s); i++ {
		r := rr[len(s)-1-i]
		rv, err := hexRuneToInt(r)
		if err != nil {
			panic(err)
		}
		v += rv << (i * 4)
	}
	return v
}

var (
	ascii0 = int('0')
	ascii9 = int('9')
	asciiA = int('A')
	asciiF = int('F')
	asciia = int('a')
	asciif = int('f')
)

func hexRuneToInt(r rune) (int, error) {
	v := int(r)
	if v >= ascii0 && v <= ascii9 {
		return v - ascii0, nil
	}
	if v >= asciiA && v <= asciiF {
		return v - asciiA + 10, nil
	}
	if v >= asciia && v <= asciif {
		return v - asciia + 10, nil
	}
	return 0, nil
}

func newImage(input string) (img, error) {
	in := strings.TrimSpace(input)
	if in == "" {
		return img{}, errors.New("image data is empty")
	}

	var data []string
	var w, h int
	rows := strings.Split(in, "\n")

	var delay int
	if strings.Contains(rows[0], ":") {
		for _, opt := range strings.Split(rows[0], ";") {
			optParts := strings.Split(opt, ":")
			switch optParts[0] {
			case "d", "delay":
				d, err := strconv.Atoi(optParts[1])
				if err != nil {
					return img{}, fmt.Errorf("invalid delay image option: `%s`", opt)
				}
				delay = d
			}
		}
		rows = rows[1:]
	}

	h = len(rows)
	for ri, row := range rows {
		cols := strings.Split(row, ",")
		if ri == 0 {
			w = len(cols)
		}
		if w != len(cols) {
			return img{}, errors.New("image data have rows of different length")
		}
		data = append(data, cols...)
	}
	return img{data: data, w: w, h: h, delay: delay}, nil
}

func (i img) value(x, y int) string {
	return i.data[x+y*i.w]
}

func (i img) toPaletted(scale int, palette palette) *image.Paletted {
	rect := image.Rect(0, 0, i.w*scale, i.h*scale)
	var p color.Palette
	for _, c := range palette {
		p = append(p, c.value)
	}
	pmap := make(map[string]color.Color)
	for _, p := range palette {
		pmap[p.name] = p.value
	}
	pi := image.NewPaletted(rect, p)
	for x := 0; x < i.w; x++ {
		for y := 0; y < i.h; y++ {
			c := pmap[i.value(x, y)]
			for sx := 0; sx < scale; sx++ {
				for sy := 0; sy < scale; sy++ {
					pi.Set(x*scale+sx, y*scale+sy, c)
				}
			}
		}
	}
	return pi
}
