// ================================================================================================
// Sidcloud by DKT/Samar
// ================================================================================================
// TODO:
// Niepoprawnie odczytuje XML
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
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	// "code.google.com/p/go-charset/charset"
	// _ "code.google.com/p/go-charset/data" // Import charset configuration files
	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/charmap"
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

	logFileGin, err := os.Open(filepath)
	ErrCheck(err)
	_, err = logFileGin.Read(p)
	ErrCheck(err)
	logFileGin.Close()

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
	log.Println("CSDBGetLatestReleases()")
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
	log.Println("CSDBGetRelease id=" + id)
	// log.Println(dataString)

	c.JSON(http.StatusOK, dataString)
}

// AudioGet - granie utworu
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

		// dir, _ := os.Getwd()
		// log.Println("Current working dir " + dir)

		// Przygotowanie nazw plików
		filenameWAV := "music" + strconv.Itoa(GlobalFileCnt) + ".wav"
		filenameSID := "music" + strconv.Itoa(GlobalFileCnt) + ".sid"
		filenamePRG := "music" + strconv.Itoa(GlobalFileCnt) + ".prg"

		var filename = ""
		if fileExists(filenameSID) {
			filename = filenameSID
		} else if fileExists(filenamePRG) {
			filename = filenamePRG
		}

		// Odczytujemy parametr - typ playera
		player := c.Param("player")

		// ==============================
		// SIDPLAYFP
		// ==============================
		if player == "sidplayfp" {

			name := "music" + strconv.Itoa(GlobalFileCnt)
			paramName := "-w" + name

			var cmdName string

			czas := "-t600"
			// bits := "-p16"
			// freq := "-f44100"

			// Odpalenie sidplayfp
			if runtime.GOOS == "windows" {
				cmdName = "sidplayfp/sidplayfp.exe"
			} else {
				cmdName = "sidplayfp" // zakładamy że jest zainstalowany
			}

			log.Println("Starting sidplayfp... cmdName(" + cmdName + " " + czas + " " + paramName + " " + filename + ")")
			cmd := exec.Command(cmdName, czas, paramName, filename)
			err := cmd.Start()
			ErrCheck(err)

			defer cmd.Process.Kill()
		}

		// ==============================
		// JSIDPLAY2
		// ==============================
		if player == "jsidplay2" {

			par1 := "-jar"
			par2 := "jsidplay2_console-4.1.jar"
			par3 := "-q"
			par4 := "-a"
			par5 := "WAV"
			par6 := "-r"

			var cmdName string
			cmdName = "java"

			// var out bytes.Buffer
			// var stderr bytes.Buffer

			cmd := exec.Command(cmdName, par1, par2, par3, par4, par5, par6, filenameWAV, filename)

			// cmd.Stdout = &out
			// cmd.Stderr = &stderr

			// log.Println("Path   = " + cmd.Path)

			// for i, arg := range cmd.Args {
			// 	log.Println("arg[" + strconv.Itoa(i) + "]=" + arg)
			// }

			// log.Println("start cmd")
			err := cmd.Start()
			ErrCheck(err)

			// log.Println("Result: " + out.String())
			// log.Println("Errors: " + stderr.String())

			defer cmd.Process.Kill()
		}

		// Gdyby cos poszło nie tak to zamykamy sidplayfp i kasujemy pliki
		defer os.Remove(filenameSID)
		defer os.Remove(filenamePRG)
		defer os.Remove(filenameWAV)

		// czekamy aż plik wav powstanie - dodać TIMEOUT
		log.Println(filenameWAV + " is creating...")
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
			logFileGin, _ := os.Open(filenameWAV)
			// ErrCheck(errOpen)

			// Gdyby cos poszło nie tak zamykamy plik, zamykamy sidplayfp i kasujemy pliki
			defer logFileGin.Close()
			defer os.Remove(filenameSID)
			defer os.Remove(filenamePRG)
			defer os.Remove(filenameWAV)

			// Czytamy z pliku kolejne dane do bufora
			readed, _ := logFileGin.ReadAt(p, offset)
			// ErrCheck(err)
			logFileGin.Close()

			// Jeżeli coś odczytaliśmy to wysyłamy
			if readed > 0 {

				// if offset > 44 {
				// 	// log.Print("readed " + strconv.Itoa(readed))
				// 	var ix int
				// 	for ix = 0; ix < readed; ix = ix + 2 {
				// 		var valInt1 int16
				// 		valInt1 = int16(p[ix]) + 256*int16(p[ix+1])

				// 		var valFloat float64
				// 		valFloat = float64(valInt1) * 1.25
				// 		if valFloat > 32766 {
				// 			valFloat = 32766
				// 		}
				// 		if valFloat < -32766 {
				// 			valFloat = -32766
				// 		}
				// 		var valInt2 int16
				// 		valInt2 = int16(math.Round(valFloat))
				// 		var valInt3 uint16
				// 		valInt3 = uint16(valInt2)

				// 		p[ix] = byte(valInt3 & 0xff)
				// 		p[ix+1] = byte((valInt3 & 0xff00) >> 8)
				// 	}
				// }
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

// RssItem - pojednyczy wpis w XML
// ================================================================================================
type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

// RssFeed - tabela XML
// ================================================================================================
type RssFeed struct {
	Items []RssItem `xml:"channel>item"`
}

// Handle - kto jest autorem wydał
// ================================================================================================
type Handle struct {
	ID     string `xml:"ID"`
	Handle string `xml:"Handle"`
}

// Group - kto jest autorem wydał
// ================================================================================================
type Group struct {
	ID   string `xml:"ID"`
	Name string `xml:"Name"`
}

// ReleasedBy - kto wydał
// ================================================================================================
type ReleasedBy struct {
	Handle []Handle `xml:"Handle"`
	Group  []Group  `xml:"Group"`
}

// Credit - Credit za produkcję
// ================================================================================================
type Credit struct {
	CreditType string `xml:"CreditType"`
	Handle     Handle `xml:"Handle"`
}

// Release - wydanie produkcji na csdb
// ================================================================================================
type Release struct {
	ReleaseID         string     `xml:"Release>ID"`
	ReleaseName       string     `xml:"Release>Name"`
	ReleaseType       string     `xml:"Release>Type"`
	ReleaseScreenShot string     `xml:"Release>ScreenShot"`
	ReleasedBy        ReleasedBy `xml:"Release>ReleasedBy"`
	Credits           []Credit   `xml:"Release>Credits>Credit"`
}

// makeCharsetReader - decode reader
// ================================================================================================
func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" {
		// Windows-1252 is a superset of ISO-8859-1, so should do here
		return charmap.Windows1252.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("Unknown charset: %s", charset)
}

// toUtf8 - konwersja kodowania
// ================================================================================================
func toUtf8(inputbuf []byte) string {
	buf := make([]rune, len(inputbuf))
	for i, b := range inputbuf {
		buf[i] = rune(b)
	}
	return string(buf)
}

// ReadLatestReleasesThread - Wątek odczygtujący dane z csdb
// ================================================================================================
func ReadLatestReleasesThread() {

	defer func() {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! Koniec watku ScannerThread !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}()

	// for {
	netClient := &http.Client{Timeout: time.Second * 5}
	resp, err := netClient.Get("https://csdb.dk/rss/latestreleases.php")

	if ErrCheck(err) {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		ErrCheck(err)
		// fmt.Println(string(body))
		resp.Body.Close()

		// Przerobienie na strukturę

		var latestReleases RssFeed
		reader := bytes.NewReader(body)
		decoder := xml.NewDecoder(reader)
		decoder.CharsetReader = makeCharsetReader
		// err = xml.Unmarshal([]byte(body), &latestReleases)
		err = decoder.Decode(&latestReleases)
		ErrCheck(err)

		// fmt.Println("Odebrano: ", latestReleases)
		fmt.Println("===================================")
		fmt.Println("Odebrano listę ostatnich realeases")
		fmt.Println("===================================")

		for index := 0; index < 20; index++ {
			rssItem := latestReleases.Items[index]
			// fmt.Println(rssItem.Title)
			url, err := url.Parse(rssItem.GUID)
			ErrCheck(err)
			q := url.Query()
			// fmt.Println(q.Get("id"))

			resp, err := netClient.Get("https://csdb.dk/webservice/?type=release&id=" + q.Get("id"))

			if ErrCheck(err) {
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				ErrCheck(err)
				// fmt.Println(string(body))
				resp.Body.Close()

				// Przerobienie na strukturę

				var entry Release
				reader := bytes.NewReader(body)
				decoder := xml.NewDecoder(reader)
				decoder.CharsetReader = makeCharsetReader
				err = decoder.Decode(&entry)
				ErrCheck(err)

				// var entry Release
				// err = xml.Unmarshal([]byte(body), &entry)
				// ErrCheck(err)

				fmt.Println("Nazwa:  ", entry.ReleaseName)
				fmt.Println("ID:     ", entry.ReleaseID)
				fmt.Println("Typ:    ", entry.ReleaseType)
				for _, group := range entry.ReleasedBy.Group {
					fmt.Println("Group:  ", group.Name)
				}
				for _, handle := range entry.ReleasedBy.Handle {
					fmt.Println("Handle: ", handle.Handle)
				}
				fmt.Println("-----------------------------------")
				for _, credit := range entry.Credits {

					if credit.Handle.Handle == "" {
						for _, releaseHandle := range entry.ReleasedBy.Handle {
							if releaseHandle.ID == credit.Handle.ID {
								fmt.Println(credit.CreditType + ": " + releaseHandle.Handle + " [" + releaseHandle.ID + "]")
							}
						}
					} else {
						fmt.Println(credit.CreditType + ": " + credit.Handle.Handle + " [" + credit.Handle.ID + "]")
					}
				}
				fmt.Println("===================================")

				// ReleaseType       string `xml:"CSDbData>Release>Type"`
				// ReleaseScreenShot string `xml:"CSDbData>Release>ScreenShot"`
				// ReleasedByGroup   string `xml:"CSDbData>Release>ReleasedBy>Group"`
				// ReleasedByHandle  string `xml:"CSDbData>Release>ReleasedBy>Handle"`

			} else {
				fmt.Println("Błąd komunikacji z csdb.dk")
			}
		}

	} else {
		fmt.Println("Błąd komunikacji z csdb.dk")
	}

	//
	// SLEEP
	// ----------------------------------------------------------------------------------------
	//
	// time.Sleep(60 * time.Second)
	// }

}

// ================================================================================================
// MAIN()
// ================================================================================================
func main() {

	// Logowanie do pliku
	//
	logFileApp, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	ErrCheck(err)
	log.SetOutput(io.MultiWriter(os.Stdout, logFileApp))

	gin.DisableConsoleColor()
	logFileGin, err := os.OpenFile("gin.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	ErrCheck(err)
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFileGin)

	log.Println("=========================================")
	log.Println("======          APP START        ========")
	log.Println("=========================================")

	r := gin.Default()

	r.Use(Options)

	r.LoadHTMLGlob("./dist/*.html")

	r.StaticFS("/css", http.Dir("./dist/css"))
	r.StaticFS("/js", http.Dir("./dist/js"))

	r.StaticFile("/", "./dist/index.html")
	r.StaticFile("favicon.ico", "./dist/favicon.ico")

	r.GET("/api/v1/audio/:player", AudioGet)
	r.POST("/api/v1/audio", AudioPost)
	r.PUT("/api/v1/audio", AudioPut)
	r.GET("/api/v1/csdb_releases", CSDBGetLatestReleases)
	r.POST("/api/v1/csdb_release", CSDBGetRelease)

	ReadLatestReleasesThread()

	// r.Run(":8080")
}
