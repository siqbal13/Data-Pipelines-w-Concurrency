package main

import (
	"fmt"
	"image"
	"strconv"
	"strings"
	"sync"

	imageprocessing "goroutines_pipeline/image_processing"
)

// Job represents a job for image processing
type Job struct {
	InputPath string      // Path of the input image file
	Image     image.Image // Image to be processed
	OutPath   string      // Output path for the processed image file
}

// loadImage loads images from the given paths
func loadImage(paths []string) (<-chan Job, <-chan error) {
	out := make(chan Job)
	errChan := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errChan)
		for _, p := range paths {
			img, err := imageprocessing.ReadImage(p)
			if err != nil {
				errChan <- fmt.Errorf("Error loading image from %s: %w", p, err)
				continue
			}
			out <- Job{InputPath: p, Image: img, OutPath: strings.Replace(p, "images/", "images/output/", 1)}
		}
	}()
	return out, errChan
}

// resize resizes the image
func resize(input <-chan Job) <-chan Job {
	out := make(chan Job)
	go func() {
		defer close(out)
		for job := range input {
			job.Image = imageprocessing.Resize(job.Image)
			out <- job
		}
	}()
	return out
}

// convertToGrayscale converts the image to grayscale
func convertToGrayscale(input <-chan Job) <-chan Job {
	out := make(chan Job)
	go func() {
		defer close(out)
		for job := range input {
			job.Image = imageprocessing.Grayscale(job.Image)
			out <- job
		}
	}()
	return out
}

// saveImage saves the image to the output path
func saveImage(input <-chan Job) (<-chan bool, <-chan error) {
	out := make(chan bool)
	errChan := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errChan)
		for job := range input {
			err := imageprocessing.WriteImage(job.OutPath, job.Image)
			if err != nil {
				errChan <- fmt.Errorf("Error saving image to %s: %w", job.OutPath, err)
				continue
			}
			out <- true
		}
	}()
	return out, errChan
}

// rotateImage rotates the image
func rotateImage(input <-chan Job) <-chan Job {
	out := make(chan Job)
	go func() {
		defer close(out)
		for job := range input {
			rotationAmounts := []float64{90, 180, 270, 360}
			for _, rotation := range rotationAmounts {
				rImg := imageprocessing.RotateImage(job.Image, rotation)
				newPath := strings.Replace(job.OutPath, ".jpg", "_"+strconv.Itoa(int(rotation))+".jpg", 1)
				out <- Job{InputPath: job.InputPath, Image: rImg, OutPath: newPath}
			}
		}
	}()
	return out
}

func main() {
	var wg sync.WaitGroup

	imagePaths := []string{"images/image1.jpg",
		"images/image2.jpg",
		"images/image3.jpg",
		"images/image4.jpg",
	}

	// Load images from the given paths
	channel1, loadErrors := loadImage(imagePaths)

	// Resize images
	channel2 := resize(channel1)

	// Convert images to grayscale
	channel3 := convertToGrayscale(channel2)

	// Rotate images
	channel4 := rotateImage(channel3)

	// Save processed images
	writeResults, saveErrors := saveImage(channel4)

	wg.Add(2)
	go func() {
		// Error handling for image loading
		for err := range loadErrors {
			fmt.Println(err)
		}
		wg.Done()
	}()

	go func() {
		// Error handling for image saving
		for err := range saveErrors {
			fmt.Println(err)
		}
		wg.Done()
	}()

	// Print success or failure for image saving
	for success := range writeResults {
		if success {
			fmt.Println("Success!")
		} else {
			fmt.Println("Failed!")
		}
	}

	wg.Wait()
}
