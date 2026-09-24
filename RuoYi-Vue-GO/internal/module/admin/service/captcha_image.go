// 验证码图片生成（对位 kaptcha ProducerMath.createImage：题目文字渲染为 jpg）。
// 标准库 image/jpeg 实现，无第三方图形库依赖；噪点干扰从简（清晰度以人眼可读为准，见 tasks）。
package service

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math/rand"
)

// constWidth / constHeight 画布尺寸（kaptcha 默认 200x50 邻近值）。
const (
	constWidth  = 200
	constHeight = 60
)

// RenderTextJPEG 把题目文字渲染为 jpg 字节（简单像素字体 + 干扰线）。
func RenderTextJPEG(text string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, constWidth, constHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// 干扰点（随机灰点，对位 kaptcha 噪声的弱化版）
	rnd := rand.New(rand.NewSource(rand.Int63()))
	for i := 0; i < 300; i++ {
		x, y := rnd.Intn(constWidth), rnd.Intn(constHeight)
		img.Set(x, y, color.RGBA{R: 160, G: 160, B: 160, A: 255})
	}
	// 干扰线
	for i := 0; i < 4; i++ {
		x0, y0 := rnd.Intn(constWidth), rnd.Intn(constHeight)
		x1, y1 := rnd.Intn(constWidth), rnd.Intn(constHeight)
		drawLine(img, x0, y0, x1, y1, color.RGBA{R: 120, G: 120, B: 180, A: 255})
	}

	// 5x7 像素字体放大 3 倍逐字符绘制
	scale := 3
	charW := 6 * scale
	totalW := len(text)*charW + 8
	startX := (constWidth - totalW) / 2
	if startX < 2 {
		startX = 2
	}
	y0 := (constHeight - 7*scale) / 2
	x := startX
	for _, ch := range text {
		drawGlyph(img, x, y0, ch, scale, color.RGBA{R: 30, G: 30, B: 30, A: 255})
		x += charW
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// drawLine Bresenham 直线。
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		img.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
