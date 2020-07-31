package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/h2non/bimg"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 Configuration
var (
	S3Endpoint  = os.Getenv("S3_ENDPOINT")
	S3AccessKey = os.Getenv("S3_ACCESS_KEY")
	S3SecretKey = os.Getenv("S3_SECRET_KEY")
	S3Bucket    = os.Getenv("S3_BUCKET")
)

func mustOpenS3() *minio.Client {
	minioClient, err := minio.New(S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(S3AccessKey, S3SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Panicln(err)
	}

	return minioClient
}

func loadImage(bucket, filepath string) ([]byte, error) {
	minioClient := mustOpenS3()
	obj, err := minioClient.GetObject(context.Background(), bucket, filepath, minio.GetObjectOptions{})
	if err != nil {
		return []byte{}, err
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

// https://ewr1.vultrobjects.com/zombia/img/Eizouken_Anime_Teaser.jpg
func main() {

	r := gin.Default()
	r.GET("/img-fallback/*path", func(c *gin.Context) {
		path := c.Param("path")[1:]
		originalImage, err := loadImage(S3Bucket, path)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		img, err := bimg.Resize(originalImage, bimg.Options{
			Width: 750,
		})
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Header("Cache-Control", "public, max-age: 63072000")
		c.Writer.Write(img)
	})
	r.GET("/img-thumbnail/*path", func(c *gin.Context) {
		path := c.Param("path")[1:]
		originalImage, err := loadImage(S3Bucket, path)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		img, err := bimg.NewImage(originalImage).Thumbnail(256)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Header("Cache-Control", "public, max-age: 63072000")
		c.Writer.Write(img)
	})

	r.GET("/img/*path", func(c *gin.Context) {
		path := c.Param("path")[1:]
		originalImage, err := loadImage(S3Bucket, path)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		imgWebp, _ := bimg.Resize(originalImage, bimg.Options{
			Width: 750,
			Type:  bimg.WEBP,
		})
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Writer.Write(imgWebp)
	})

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age: 63072000")
		c.File("favicon.ico")
	})

	r.Run(":2222")
}
