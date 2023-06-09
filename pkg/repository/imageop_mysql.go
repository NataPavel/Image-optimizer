package repository

import (
	"database/sql"
	"fmt"
	"image/jpeg"
	"imageOptimisation/entities"

	"log"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

type ImageOp struct {
	db *sql.DB
}

const localStorageDir = "./localStorage/"

func NewImageOp(db *sql.DB) *ImageOp {
	return &ImageOp{db: db}
}
func (r *ImageOp) CreateImage(image entities.Image, filename string) (int64, error) {
	sizes := []int{100, 75, 50, 25}

	var wg sync.WaitGroup
	wg.Add(len(sizes))

	for i, size := range sizes {
		go resizeImage(size, filename, &wg)

		switch i {
		case 0:
			image.Image100 = fmt.Sprintf("%d_%s", size, filename)
		case 1:
			image.Image75 = fmt.Sprintf("%d_%s", size, filename)
		case 2:
			image.Image50 = fmt.Sprintf("%d_%s", size, filename)
		case 3:
			image.Image25 = fmt.Sprintf("%d_%s", size, filename)
		}
	}
	wg.Wait()

	var id int64

	query := "INSERT INTO images (image100, image75, image50, image25) VALUES (?, ?, ?, ?)"
	result, err := r.db.Exec(query, image.Image100, image.Image75, image.Image50, image.Image25)
	if err != nil {
		return 0, err
	}
	
	id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ImageOp) GetImageById(image entities.Image, id int, c *gin.Context) (string, error) {
	err := r.db.QueryRow("SELECT image100, image75, image50, image25 FROM images WHERE id = ?", id).Scan(&image.Image100, &image.Image75, &image.Image50, &image.Image25)
	if err != nil {
		return "", err
	}

	quality := c.Query("quality")
	var filename string

	switch quality {
	case "100":
		filename = image.Image100
	case "75":
		filename = image.Image75
	case "50":
		filename = image.Image50
	case "25":
		filename = image.Image25
	default:
		filename = image.Image100
	}

	return filename, nil
}

func resizeImage(sizePercentage int, filename string, wg *sync.WaitGroup) {
	defer wg.Done()

	img, err := os.Open(fmt.Sprintf("%s%s", localStorageDir, filename))
	if err != nil {
		log.Println(err)
		return
	}
	defer img.Close()

	decodedImg, err := jpeg.Decode(img)
	if err != nil {
		log.Println(err)
		return
	}

	newHeight := uint(float64(decodedImg.Bounds().Dy()*sizePercentage) / 100)
	newImg := resize.Resize(0, newHeight, decodedImg, resize.Lanczos3)
	if err != nil {
		log.Println(err)
		return
	}

	newFilename := fmt.Sprintf("%d_%s", sizePercentage, filename)
	path := fmt.Sprintf("%s%s", localStorageDir, newFilename)
	// create a new file to save the modified image
	file, err := os.Create(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	// save the modified image as a JPEG to the file
	err = jpeg.Encode(file, newImg, &jpeg.Options{Quality: 90})
	if err != nil {
		log.Println(err)
		return
	}
}
