package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrCheck - obsługa błedów
// ========================================================
func ErrCheck(errNr error) bool {
	if errNr != nil {
		fmt.Println(errNr)
		return false
	}
	return true
}

// AudioGet - granie utworu do testów
// ========================================================
func AudioGet(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Connection", "Keep-Alive")
	c.Header("Transfer-Encoding", "chunked")

	var err = os.Remove("music.wav")
	ErrCheck(err)

	cmd := exec.Command("sidplayfp/sidplayfp.exe", "-wmusic", "-t600", "sidplayfp/Incoherent_Nightmare_tune_3.sid")
	err = cmd.Start()
	ErrCheck(err)

	time.Sleep(1 * time.Second)

	const bufferSize = 1024 * 4

	var offset int64
	p := make([]byte, bufferSize)

	for {

		f, _ := os.Open("music.wav")
		readed, _ := f.ReadAt(p, offset)
		f.Close()
		offset += bufferSize
		// if err == io.EOF {
		// 	break
		// }
		if readed < bufferSize {
			time.Sleep(1 * time.Second)
		}
		if readed > bufferSize {
			time.Sleep(1 * time.Second)
		}
		c.Data(http.StatusOK, "audio/wav", p)
		ErrCheck(err)

	}
}

// AudioPost - Granie utworu wysłanego
// ========================================================
func AudioPost(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	// c.JSON(http.StatusOK, gin.H{"Status": "AudioPost OK"})
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

	r := gin.Default()

	r.Use(Options)

	r.GET("/api/v1/audio", AudioGet)
	r.POST("/api/v1/audio", AudioPost)

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8099")
}
