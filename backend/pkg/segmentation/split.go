package segmentation

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
)

// splitImage chia ảnh thành nhiều đoạn văn bản nhỏ
func SplitImage(filePath string, prefix string) []string {
	// Đọc file ảnh
	imgFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Không thể mở ảnh:", err)
	}
	defer imgFile.Close()

	// Decode ảnh
	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Fatal("Không thể decode ảnh:", err)
	}

	// Chuyển ảnh sang grayscale
	grayImg := convertToGray(img)

	// Tìm và cắt đoạn
	segments := findParagraphs(grayImg)

	return SaveSegments(segments, prefix)
}

// convertToGray chuyển ảnh thành grayscale
func convertToGray(img image.Image) *image.Gray {
	gray := image.NewGray(img.Bounds())
	draw.Draw(gray, gray.Bounds(), img, img.Bounds().Min, draw.Src)
	return gray
}

// findParagraphs tìm các đoạn văn bản bằng cách xác định các vùng ít thay đổi cường độ màu
func findParagraphs(grayImg *image.Gray) []image.Image {
	var segments []image.Image
	bounds := grayImg.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	sequenceLimit := 22
	var Density = make([]int, height)
	for y := 0; y < height; y++ {
		var rowIntensity int
		for x := 0; x < width; x++ {
			rowIntensity += int(grayImg.GrayAt(x, y).Y)
		}
		Density[y] = rowIntensity / width
	}

	// fmt.Println(Density)
	count := 0
	splitPosition := make([]int, 0, height) // Khởi tạo slice với make() thay vì gọi biến chưa khai báo
	for line := 1; line < height; line++ {
		if (Density[line] == Density[line-1]) {
			count++
		} else {
			count = 0
		}
		if (count == sequenceLimit) {
			splitPosition = append(splitPosition, line)
			count = 0 
		}
	}

	segmentStart := 0
	for i := 0; i < len(splitPosition); i++ {
		segment := grayImg.SubImage(image.Rect(0, segmentStart, width, splitPosition[i]))
		segments = append(segments, segment)
		segmentStart = splitPosition[i]
	}

	if segmentStart < height {
		segment := grayImg.SubImage(image.Rect(0, segmentStart, width, height))
		segments = append(segments, segment)
	}

	return segments
}

// SaveSegments lưu các đoạn ảnh vào các file ảnh riêng biệt
func SaveSegments(segments []image.Image, prefix string) []string {
	var paths []string
	dirPath := "segments/"
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		log.Fatalf("Không thể tạo thư mục: %v", err)
	}
	for i, segment := range segments {
		filename := fmt.Sprintf("%s%s_segment_%d.jpg", dirPath, prefix, i)
		file, err := os.Create(filename)
		if err != nil {
			log.Printf("Không thể tạo file ảnh: %v", err)
			continue
		}
		defer file.Close()

		err = jpeg.Encode(file, segment, nil)
		if err != nil {
			log.Printf("Không thể lưu file ảnh: %v", err)
			continue
		}
		paths = append(paths, filename)
	}
	return paths
}