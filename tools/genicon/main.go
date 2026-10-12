// genicon renders the authoritative SVG using MyGo's vector renderer and
// writes portable PNG, macOS ICNS/iconset and Windows ICO files.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo/ui"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func render(data []byte, size int) []byte {
	svg, err := ui.ParseSVG(data)
	must(err)
	tt := ui.NewTester(func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		ui.Image(c, svg).Size(float32(size), float32(size))
	}, size, size)
	tt.Frame()
	var b bytes.Buffer
	must(png.Encode(&b, tt.Image()))
	return b.Bytes()
}
func write(path string, data []byte) {
	must(os.MkdirAll(filepath.Dir(path), 0o755))
	must(os.WriteFile(path, data, 0o644))
}
func main() {
	svg, err := os.ReadFile("resources/icon.svg")
	must(err)
	sizes := []int{16, 32, 48, 64, 128, 256, 512, 1024}
	images := map[int][]byte{}
	for _, size := range sizes {
		images[size] = render(svg, size)
		write(fmt.Sprintf("resources/icons/rivu-%d.png", size), images[size])
	}
	write("resources/icon.png", images[1024])
	for _, size := range []int{16, 32, 128, 256, 512} {
		write(fmt.Sprintf("resources/rivu.iconset/icon_%dx%d.png", size, size), images[size])
		write(fmt.Sprintf("resources/rivu.iconset/icon_%dx%d@2x.png", size, size), images[size*2])
	}
	var chunks bytes.Buffer
	for _, v := range []struct {
		size int
		kind string
	}{{16, "icp4"}, {32, "icp5"}, {64, "icp6"}, {128, "ic07"}, {256, "ic08"}, {512, "ic09"}, {1024, "ic10"}} {
		chunks.WriteString(v.kind)
		must(binary.Write(&chunks, binary.BigEndian, uint32(len(images[v.size])+8)))
		chunks.Write(images[v.size])
	}
	var icns bytes.Buffer
	icns.WriteString("icns")
	must(binary.Write(&icns, binary.BigEndian, uint32(chunks.Len()+8)))
	icns.Write(chunks.Bytes())
	write("resources/rivu.icns", icns.Bytes())
	icoSizes := []int{16, 32, 48, 64, 128, 256}
	var ico bytes.Buffer
	for _, v := range []uint16{0, 1, uint16(len(icoSizes))} {
		must(binary.Write(&ico, binary.LittleEndian, v))
	}
	offset := 6 + 16*len(icoSizes)
	for _, size := range icoSizes {
		ico.Write([]byte{byte(size % 256), byte(size % 256), 0, 0})
		for _, v := range []uint16{1, 32} {
			must(binary.Write(&ico, binary.LittleEndian, v))
		}
		must(binary.Write(&ico, binary.LittleEndian, uint32(len(images[size]))))
		must(binary.Write(&ico, binary.LittleEndian, uint32(offset)))
		offset += len(images[size])
	}
	for _, size := range icoSizes {
		ico.Write(images[size])
	}
	write("resources/rivu.ico", ico.Bytes())
	tray, err := os.ReadFile("resources/tray.svg")
	must(err)
	write("resources/tray.png", render(tray, 36))
	write("resources/tray@2x.png", render(tray, 72))
	// Validate the generated master before a packaging step consumes it.
	img, _, err := image.Decode(bytes.NewReader(images[1024]))
	must(err)
	if img.Bounds().Dx() != 1024 {
		log.Fatal("invalid master dimensions")
	}
	fmt.Println("Generated Rivu SVG-derived PNG, ICNS, ICO, iconset and tray assets.")
}
