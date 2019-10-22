package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GlobalFileCnt - numer pliku
// ========================================================
var GlobalFileCnt int

// ErrCheck - obsługa błedów
// ========================================================
func ErrCheck(errNr error) bool {
	if errNr != nil {
		fmt.Println(errNr)
		return false
	}
	return true
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
// ========================================================
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// AudioGet - granie utworu do testów
// ========================================================
func AudioGet(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Connection", "Keep-Alive")
	// c.Header("Transfer-Encoding", "chunked")

	GlobalFileCnt++
	name := "music" + strconv.Itoa(GlobalFileCnt)
	paramName := "-w" + name
	filename := "music" + strconv.Itoa(GlobalFileCnt) + ".wav"

	cmd := exec.Command("sidplayfp/sidplayfp.exe", paramName, "-t600", "music.sid")
	err := cmd.Start()
	ErrCheck(err)

	defer cmd.Process.Kill()

	time.Sleep(1 * time.Second)

	const bufferSize = 1024
	var offset int64
	p := make([]byte, bufferSize)

	for {

		if c.Request.Context() == nil {
			break
		}

		f, _ := os.Open(filename)
		defer f.Close()

		readed, _ := f.ReadAt(p, offset)
		f.Close()

		offset += bufferSize

		if readed < bufferSize {
			time.Sleep(1 * time.Second)
		}

		c.Data(http.StatusOK, "audio/wav", p)
	}

	// c.JSON(http.StatusOK, "Connection lost.")
}

// AudioPost - Granie utworu wysłanego
// ========================================================
func AudioPost(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Connection", "Keep-Alive")
	// c.Header("Transfer-Encoding", "chunked")

	// Ściągnięcie pliku SID

	sidURL := c.Query("sid_url")

	err := DownloadFile("music.sid", sidURL)
	ErrCheck(err)

	c.JSON(http.StatusOK, "Odebrałem: "+sidURL)

}

// Options - Obsługa request'u OPTIONS (CORS)
// ========================================================
func Options(c *gin.Context) {
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
		c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
		// c.Header("Content-Type", "application/json")
		c.AbortWithStatus(http.StatusOK)
	}
}

// MAIN()
// ========================================================
func main() {

	// var err = os.Remove("music.wav")
	// ErrCheck(err)

	r := gin.Default()

	r.Use(Options)

	r.GET("/api/v1/audio", AudioGet)
	r.POST("/api/v1/audio", AudioPost)

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8099")
}
