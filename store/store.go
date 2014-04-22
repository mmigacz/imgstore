package imgstore

import (
	"net/http"
	"html/template"
	"regexp"
	"errors"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"time"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	_ "image/jpeg"
	"bytes"
	"log"
	"strings"
	"strconv"
)



type Store struct {
	accessKey string
	secretKey string
	awsRegion aws.Region
	bucketName string

	bucket *s3.Bucket
}


func (s *Store) getBucket() *s3.Bucket {
	if s.bucket == nil {
		auth := aws.Auth {
			AccessKey: s.accessKey,
			SecretKey: s.secretKey,
		}
		connection := s3.New(auth, s.awsRegion)
		s.bucket = connection.Bucket(s.bucketName)
	}
	return s.bucket
}



func (s *Store) storeImg(path string, img image.Image) error {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		return err
	}

	return s.getBucket().Put(path, buf.Bytes(), "image/png", s3.Private, s3.Options{})
}


func (s *Store) getImageUrl(path string, width uint, height uint) (url string, err error) {

	rPath := fmt.Sprintf("%s_%d_%d", path, width, height)

	b := s.getBucket()

	if rExists, _ := b.Exists(rPath); (width > 0 || height > 0) && !rExists {
		body, err := b.GetReader(path)
		if err != nil {
			return "", err
		}

		img, _, err := image.Decode(body)
		if err != nil {
			return "", err
		}
		body.Close()


		rImage := resize.Resize(width, height, img, resize.Lanczos3)
		err = s.storeImg(rPath, rImage)


		if err != nil {
			return "", err
		}

		return b.SignedURL(rPath, time.Now().Add(time.Duration(30) * time.Second)), nil

	} else if exists, er  := b.Exists(path); exists {
		return b.SignedURL(path, time.Now().Add(time.Duration(30) * time.Second)), nil
	} else {
		return "", er
	}

}
