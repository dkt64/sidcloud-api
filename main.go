// ================================================================================================
// Sidcloud by DKT/Samar
// ================================================================================================
// TODO:
// Android/Chrome wielokrotne GET i przerwanie pipe
// zwracać info z SID
// używać czasu trwania z pliku i dać możliwość ustawienia
// wyświetlać w kontrolce poprawny czas
// używanie ID poprzez Cookies
// ================================================================================================
// DONE:
// dodać obsługę PRG
// sprawdzać rodzaj pliku i inne błędy
// ================================================================================================

package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GlobalFileCnt - numer pliku
// ================================================================================================
var GlobalFileCnt int
var posted bool

// fileExists - sprawdzenie czy plik istnieje
// ================================================================================================
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// ErrCheck - obsługa błedów
// ================================================================================================
func ErrCheck(errNr error) bool {
	if errNr != nil {
		fmt.Println(errNr)
		return false
	}
	return true
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
// ================================================================================================
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
	ErrCheck(err)
	out.Close()

	// Sprawdzzamy rozmiar pliku
	fi, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	// get the size
	size := fi.Size()
	log.Println("Rozmiar pliku " + strconv.Itoa(int(size)))

	if size < 8 || size > 65535 {
		err := errors.New("Rozmiar pliku niewłaściwy")
		return err
	}
	// Odczytujemy 4 pierwsze bajty żeby sprawdzić czy to SID
	p := make([]byte, 4)

	f, err := os.Open(filepath)
	ErrCheck(err)
	_, err = f.Read(p)
	ErrCheck(err)
	f.Close()

	// log.Println("Sprawdzanie pliku " + strconv.Itoa(n))

	var newName string

	if p[1] == 0x53 && p[2] == 0x49 && p[3] == 0x44 {
		newName = filepath + ".sid"
		err := os.Rename(filepath, newName)
		ErrCheck(err)
	} else {
		newName = filepath + ".prg"
		err := os.Rename(filepath, newName)
		ErrCheck(err)
	}

	return err

}

// CSDBGetLatestReleases - ostatnie release'y
// ================================================================================================
func CSDBGetLatestReleases(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	resp, errGet := http.Get("https://csdb.dk/rss/latestreleases.php")
	ErrCheck(errGet)

	data, errRead := ioutil.ReadAll(resp.Body)
	ErrCheck(errRead)

	dataString := string(data)

	// Info o wejściu do GET
	// log.Println("CSDBGetLatestReleases")
	// log.Println(dataString)

	c.JSON(http.StatusOK, dataString)
}

// CSDBGetRelease - ostatnie release'y
// ================================================================================================
func CSDBGetRelease(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	id := c.Query("id")

	resp, errGet := http.Get("https://csdb.dk/webservice/?type=release&id=" + id)
	ErrCheck(errGet)

	data, errRead := ioutil.ReadAll(resp.Body)
	ErrCheck(errRead)

	dataString := string(data)

	// Info o wejściu do GET
	// log.Println("CSDBGetLatestReleases")
	// log.Println(dataString)

	c.JSON(http.StatusOK, dataString)
}

// AudioGet - granie utworu do testów
// ================================================================================================
func AudioGet(c *gin.Context) {

	if posted {
		posted = false

		// Typ połączania
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Connection", "Keep-Alive")
		c.Header("Transfer-Encoding", "chunked")

		// Info o wejściu do GET
		log.Println("AudioGet start with GlobalFileCnt = " + strconv.Itoa(GlobalFileCnt))

		// Przygotowanie nazw plików
		name := "music" + strconv.Itoa(GlobalFileCnt)
		paramName := "-w" + name
		filenameWAV := "music" + strconv.Itoa(GlobalFileCnt) + ".wav"

		filenameSID := "music" + strconv.Itoa(GlobalFileCnt) + ".sid"
		filenamePRG := "music" + strconv.Itoa(GlobalFileCnt) + ".prg"

		var filename = ""
		if fileExists(filenameSID) {
			filename = filenameSID
		} else if fileExists(filenamePRG) {
			filename = filenamePRG
		}

		czas := "-t600"
		// bits := "-p16"
		// freq := "-f44100"

		// Odpalenie sidplayfp

		var cmdName string

		if runtime.GOOS == "windows" {
			cmdName = "sidplayfp/sidplayfp.exe"
		} else {
			cmdName = "./sidplayfp/sidplayfp"
		}

		log.Println("Starting sidplayfp... cmdName(" + cmdName + " " + czas + " " + paramName + " " + filename + ")")
		cmd := exec.Command(cmdName, czas, paramName, filename)
		err := cmd.Start()
		ErrCheck(err)

		// Gdyby cos poszło nie tak to zamykamy sidplayfp i kasujemy pliki
		defer cmd.Process.Kill()
		defer os.Remove(filenameSID)
		defer os.Remove(filenamePRG)
		defer os.Remove(filenameWAV)

		// czekamy aż plik wav powstanie - dodać TIMEOUT
		for !fileExists(filenameWAV) {
			time.Sleep(200 * time.Millisecond)
		}
		log.Println(filenameWAV + " created.")

		// Przygotowanie bufora do streamingu
		const bufferSize = 1024 * 64
		var offset int64
		p := make([]byte, bufferSize)

		log.Println("Sending...")

		// Streaming LOOP...
		// ----------------------------------------------------------------------------------------------

		for {

			// Wysyłamy pakiet co 500 ms
			time.Sleep(500 * time.Millisecond)

			// Jeżeli doszliśmy w pliku do 50MB to koniec
			if offset > 50000000 {
				log.Println("ERR! EOF (50MB).")
				break
			}

			// Jeżeli straciimy kontekst to wychodzimy
			if c.Request.Context() == nil {
				log.Println("ERR! c.Request.Context() == nil")
				break
			}

			// Otwieraamy plik - bez sprawdzania błędów
			f, _ := os.Open(filenameWAV)
			// ErrCheck(errOpen)

			// Gdyby cos poszło nie tak zamykamy plik, zamykamy sidplayfp i kasujemy pliki
			defer f.Close()
			defer cmd.Process.Kill()
			defer os.Remove(filenameSID)
			defer os.Remove(filenamePRG)
			defer os.Remove(filenameWAV)

			// Czytamy z pliku kolejne dane do bufora
			readed, _ := f.ReadAt(p, offset)
			// ErrCheck(err)
			f.Close()

			// Jeżeli coś odczytaliśmy to wysyłamy
			if readed > 0 {

				if offset > 44 {
					// log.Print("readed " + strconv.Itoa(readed))
					var ix int
					for ix = 0; ix < readed; ix = ix + 2 {
						var valInt1 int16
						valInt1 = int16(p[ix]) + 256*int16(p[ix+1])

						var valFloat float64
						valFloat = float64(valInt1) * 1.25
						if valFloat > 32766 {
							valFloat = 32766
						}
						if valFloat < -32766 {
							valFloat = -32766
						}
						var valInt2 int16
						valInt2 = int16(math.Round(valFloat))
						var valInt3 uint16
						valInt3 = uint16(valInt2)

						p[ix] = byte(valInt3 & 0xff)
						p[ix+1] = byte((valInt3 & 0xff00) >> 8)
					}
				}
				c.Data(http.StatusOK, "audio/wav", p)
				offset += int64(readed)
				// log.Print(".")
			}

		}

		// Feedback gdybyśmy wyszli z LOOP
		c.JSON(http.StatusOK, "Loop ended.")
		log.Println("Loop ended.")

	} else {

		// Przy powturzonym Get
		c.JSON(http.StatusOK, "ERR! Repeated GET.")
		log.Println("ERR! Repeated GET.")
	}
}

// AudioPost - Odernanie linka do SID lub PRG
// ================================================================================================
func AudioPost(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	log.Println("AudioPost start with GlobalFileCnt = " + strconv.Itoa(GlobalFileCnt))

	GlobalFileCnt++
	posted = true

	var filename = ""

	sidURL := c.Query("sid_url")

	filename = "music" + strconv.Itoa(GlobalFileCnt)

	// Ściągnięcie pliku SID

	// Gdy to nie link do SID albo PRG można podejrzewać skrypt
	// Trzeba sprawdzić zawartość pliku
	err := DownloadFile(filename, sidURL)

	ErrCheck(err)
	if err != nil {
		log.Println("ERR! Error downloading file: " + sidURL)
		c.JSON(http.StatusOK, "ERR! Error downloading file: "+sidURL)
	} else {
		log.Println("Downloaded file: " + sidURL)
		c.JSON(http.StatusOK, "Downloaded file: "+sidURL)
	}

	log.Println("AudioPost end with GlobalFileCnt = " + strconv.Itoa(GlobalFileCnt))
}

// AudioPut - Odebranie pliku SID lub PRG
// ================================================================================================
func AudioPut(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	log.Println("AudioPut start with GlobalFileCnt = " + strconv.Itoa(GlobalFileCnt))

	file, fileErr := c.FormFile("file")
	ErrCheck(fileErr)
	log.Println("Odebrałem plik " + file.Filename)

	GlobalFileCnt++
	posted = true

	var properExtension = false
	var filenameSID = ""

	if strings.HasSuffix(file.Filename, ".sid") {
		filenameSID = "music" + strconv.Itoa(GlobalFileCnt) + ".sid"
		properExtension = true
	}
	if strings.HasSuffix(file.Filename, ".prg") {
		filenameSID = "music" + strconv.Itoa(GlobalFileCnt) + ".prg"
		properExtension = true
	}

	if properExtension {
		// Zapis SID'a
		saveErr := c.SaveUploadedFile(file, filenameSID)
		ErrCheck(saveErr)
		c.JSON(http.StatusOK, "Got the file: "+file.Filename)
	} else {
		c.JSON(http.StatusOK, "Wrong file extension: "+file.Filename)
	}

	log.Println("AudioPut end with GlobalFileCnt = " + strconv.Itoa(GlobalFileCnt))
}

// Options - Obsługa request'u OPTIONS (CORS)
// ================================================================================================
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

// ================================================================================================
// MAIN()
// ================================================================================================
func main() {

	r := gin.Default()

	r.Use(Options)

	r.LoadHTMLGlob("./dist/*.html")

	r.StaticFS("/css", http.Dir("./dist/css"))
	r.StaticFS("/js", http.Dir("./dist/js"))

	r.StaticFile("/", "./dist/index.html")
	r.StaticFile("favicon.ico", "./dist/favicon.ico")

	r.GET("/api/v1/audio", AudioGet)
	r.POST("/api/v1/audio", AudioPost)
	r.PUT("/api/v1/audio", AudioPut)
	r.GET("/api/v1/csdb_releases", CSDBGetLatestReleases)
	r.POST("/api/v1/csdb_release", CSDBGetRelease)

	r.Run(":80")
}
