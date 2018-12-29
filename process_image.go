// This file is designed to process subtitles

package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

type File struct {
	name           *image.Image
	horizontalrate int
	verticalrate   int
}

type Config struct {
	ImageDirectory string
	HorizontalRate int
	VerticalRate   int
	OutputFilename string
}

var config Config

func init() {
	readConfig()
}

func readConfig() {

	configfile := "config.toml"

	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}
}

func generatenewImage(files []File) {
	fmt.Printf("共有 %d 个图片需要处理\n", len(files))

	// 首先建立一个空白图形保证可以存储所有图片
	// draw.Draw(dst, r, src, sp, op)
	// 先计算长度，再计算宽度
	AllWidth := (*(files[0]).name).Bounds().Dx()
	height := (*(files[0]).name).Bounds().Dy()
	AllHeight := height * len(files)
	fmt.Println(AllWidth)

	target := image.Rectangle{image.Point{0, 0}, image.Point{AllWidth, AllHeight}}
	rgba := image.NewRGBA(target)

	var tempRectangle image.Rectangle
	for index := 0; index < len(files); index++ {
		sp := image.Point{0, height * index}
		/** 第一个加入图片的初始坐标为 (0,0)
			第二个加入图片的坐标为 (0,第一个图片的 weight)
			依此类推
			坐标系参考：https://blog.golang.org/go-imagedraw-package
		**/
		tempRectangle = image.Rectangle{sp, sp.Add((*files[index].name).Bounds().Size())}
		draw.Draw(rgba, tempRectangle, *files[index].name,
			image.Point{0, 0}, draw.Src)
	}

	out, err := os.Create(config.OutputFilename)
	defer out.Close()
	if err != nil {
		fmt.Println(err)
	}

	jpeg.Encode(out, rgba, nil)

}

func subImage(fileName string, verticalrate int, horizontalrate int) (output image.Image) {
	fmt.Println(fileName)
	reader, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	width, height := m.Bounds().Dx(), m.Bounds().Dy()

	// TODO variabledivision is confusing
	absoluteHorizontalLength := (height * horizontalrate) / 10.0

	x := m.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(0, height-absoluteHorizontalLength, width, height))
	fmt.Println(x.Bounds().Size())

	return x
}

func getFileNumber(dir string) int {
	files, _ := ioutil.ReadDir(dir)
	return len(files)
}

func saveImageToFile(input image.Image, name string) {
	outf, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
	}

	jpeg.Encode(outf, input, nil)
}

func openImage(fileName string) (target image.Image) {
	reader, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return m
}

func main() {
	config.HorizontalRate = 2

	s := make([]File, 0)

	fileNumber := getFileNumber(config.ImageDirectory)

	for index := 0; index < fileNumber; index++ {
		// string() tricky...
		filename := config.ImageDirectory + string(os.PathSeparator) + strconv.Itoa(index+1) + ".JPG"
		// crop image
		sideProduct := subImage(filename, config.VerticalRate, config.HorizontalRate)
		// save to file
		saveImageToFile(sideProduct, fmt.Sprintf("%d_middle.jpg", index))
		// reopen to get Image Object
		target := openImage(fmt.Sprintf("%d_middle.jpg", index))
		// add to struct
		file := File{
			name:           &target,
			horizontalrate: 2,
			verticalrate:   0,
		}
		// add struct to slice
		s = append(s, file)
	}

	generatenewImage(s)

	fmt.Println("截图已经输出完成")
}
