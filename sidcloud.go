// ================================================================================================
// Sidcloud by DKT/Samar
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
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
)

var mutex = &sync.Mutex{}

// GlobalFileCnt - numer pliku
// ================================================================================================
var GlobalFileCnt int

// var posted bool

// RssItem - pojednyczy wpis w XML
// ------------------------------------------------------------------------------------------------
type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

// XMLRssFeed - tabela XML
// ------------------------------------------------------------------------------------------------
type XMLRssFeed struct {
	Items []RssItem `xml:"channel>item"`
}

// XMLHandle - kto jest autorem wydał
// ------------------------------------------------------------------------------------------------
type XMLHandle struct {
	ID        string `xml:"ID"`
	XMLHandle string `xml:"Handle"`
}

// XMLGroup - kto jest autorem wydał
// ------------------------------------------------------------------------------------------------
type XMLGroup struct {
	ID   string `xml:"ID"`
	Name string `xml:"Name"`
}

// XMLReleasedBy - kto wydał
// ------------------------------------------------------------------------------------------------
type XMLReleasedBy struct {
	XMLHandle []XMLHandle `xml:"Handle"`
	XMLGroup  []XMLGroup  `xml:"Group"`
}

// XMLCredit - XMLCredit za produkcję
// ------------------------------------------------------------------------------------------------
type XMLCredit struct {
	CreditType string    `xml:"CreditType"`
	XMLHandle  XMLHandle `xml:"Handle"`
}

// XMLDownloadLink - download links
// ------------------------------------------------------------------------------------------------
type XMLDownloadLink struct {
	Link string `xml:"Link"`
}

// XMLRelease - wydanie produkcji na csdb
// ------------------------------------------------------------------------------------------------
type XMLRelease struct {
	ReleaseID         string            `xml:"Release>ID"`
	ReleaseName       string            `xml:"Release>Name"`
	ReleaseType       string            `xml:"Release>Type"`
	ReleaseScreenShot string            `xml:"Release>ScreenShot"`
	Rating            float32           `xml:"Release>Rating"`
	XMLReleasedBy     XMLReleasedBy     `xml:"Release>ReleasedBy"`
	Credits           []XMLCredit       `xml:"Release>Credits>Credit"`
	DownloadLinks     []XMLDownloadLink `xml:"Release>DownloadLinks>DownloadLink"`
}

// Release - wydanie produkcji na csdb
// ------------------------------------------------------------------------------------------------
type Release struct {
	ReleaseID         int
	ReleaseName       string
	ReleaseScreenShot string
	Rating            float32
	ReleasedBy        []string
	Credits           []string
	DownloadLinks     []string
}

// releases - glówna i globalna tablica z aktualnymi produkcjami
// ================================================================================================
var releases []Release

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

	// resp, errGet := http.Get("https://csdb.dk/rss/latestreleases.php")
	// ErrCheck(errGet)

	// data, errRead := ioutil.ReadAll(resp.Body)
	// ErrCheck(errRead)

	// dataString := string(data)

	// Info o wejściu do GET
	log.Println("CSDBGetLatestReleases()")
	// log.Println(dataString)

	mutex.Lock()
	releasesTemp := releases
	mutex.Unlock()

	c.JSON(http.StatusOK, releasesTemp)
}

// CSDBGetRelease - ostatnie release'y
// ================================================================================================
func CSDBGetRelease(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	id, _ := strconv.Atoi(c.Query("id"))

	// resp, errGet := http.Get("https://csdb.dk/webservice/?type=release&id=" + id)
	// ErrCheck(errGet)

	// data, errRead := ioutil.ReadAll(resp.Body)
	// ErrCheck(errRead)

	// dataString := string(data)

	// Info o wejściu do GET
	log.Println("CSDBGetRelease() nr ", id)
	// log.Println(dataString)

	mutex.Lock()
	releasesTemp := releases
	mutex.Unlock()

	c.JSON(http.StatusOK, releasesTemp[id])
}

// AudioGet - granie utworu
// ================================================================================================
func AudioGet(c *gin.Context) {

	volDown := false
	// const maxOffset int64 = 50000000 // ~ 10 min
	// const maxOffset int64 = 25000000 // ~ 5 min
	// const maxOffset int64 = 5000000 // ~ 1 min
	const maxOffset int64 = 44100*2*300 + 44 // = 5 min
	const maxVol float64 = 1.25
	var vol float64 = maxVol
	var loop = true

	// if posted {
	// 	posted = false

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

	log.Println("WAV filename = " + filenameWAV)

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

		czas := "-t333"
		// bits := "-p16"
		// freq := "-f44100"

		// Odpalenie sidplayfp
		if runtime.GOOS == "windows" {
			cmdName = "sidplayfp/sidplayfp.exe"
		} else {
			cmdName = "sidplayfp/sidplayfp" // zakładamy że jest zainstalowany
		}

		log.Println("Starting sidplayfp... cmdName(" + cmdName + " " + czas + " " + paramName + " " + filename + ")")
		cmd := exec.Command(cmdName, czas, paramName, filename)
		err := cmd.Start()
		ErrCheck(err)

		defer func() {
			loop = false

			var err error
			err = cmd.Process.Kill()
			ErrCheck(err)
			log.Println("Usuwam pliki")
			log.Println(filenameSID)
			err = os.Remove(filenameSID)
			ErrCheck(err)
			log.Println(filenamePRG)
			err = os.Remove(filenamePRG)
			ErrCheck(err)
			log.Println(filenameWAV)
			err = os.Remove(filenameWAV)
			ErrCheck(err)
		}()
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

		// log.Println("start cmd")
		err := cmd.Start()
		ErrCheck(err)

		// log.Println("Result: " + out.String())
		// log.Println("Errors: " + stderr.String())

		defer func() {
			loop = false

			var err error
			err = cmd.Process.Kill()
			ErrCheck(err)
			log.Println("Usuwam pliki")
			log.Println(filenameSID)
			err = os.Remove(filenameSID)
			ErrCheck(err)
			log.Println(filenamePRG)
			err = os.Remove(filenamePRG)
			ErrCheck(err)
			log.Println(filenameWAV)
			err = os.Remove(filenameWAV)
			ErrCheck(err)
		}()
	}

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

	for loop {

		// Jeżeli doszliśmy w pliku do 50MB to koniec
		if offset > maxOffset {

			// log.Println("Wyciszamy...")
			// break
			volDown = true
		}

		// Jeżeli stracimy kontekst to wychodzimy
		if c.Request.Context() == nil {
			log.Println("ERR! c.Request.Context() == nil")
			loop = false
		} else {

			// Otwieraamy plik - bez sprawdzania błędów
			file, _ := os.Open(filenameWAV)
			// ErrCheck(errOpen)

			// Czytamy z pliku kolejne dane do bufora
			readed, _ := file.ReadAt(p, offset)
			// ErrCheck(err)
			file.Close()

			// Jeżeli coś odczytaliśmy to wysyłamy
			if readed > 0 {

				// Modyfikacja sampli
				//
				if offset > 44 {
					// log.Print("readed " + strconv.Itoa(readed))
					var ix int
					for ix = 0; ix < readed; ix = ix + 2 {

						// Wyciszanie
						if volDown && vol > 0.0 {
							vol = maxVol - (float64(offset-maxOffset+int64(ix)) / 88.494 * 0.0002)
							if vol < 0 {
								vol = 0.0
								loop = false
								// break
							}
						}

						// Wzmocnienie głośności (domyślnie x 1.25)
						var valInt1 int16
						valInt1 = int16(p[ix]) + 256*int16(p[ix+1])

						var valFloat float64
						valFloat = float64(valInt1) * vol
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

		// Wysyłamy pakiet co 500 ms
		time.Sleep(500 * time.Millisecond)
	}

	// Feedback gdybyśmy wyszli z LOOP
	c.JSON(http.StatusOK, "Loop ended.")
	log.Println("Loop ended.")

	// } else {

	// 	// Przy powturzonym Get
	// 	c.JSON(http.StatusOK, "ERR! Repeated GET.")
	// 	log.Println("ERR! Repeated GET.")
	// }
}

// AudioPost - Odebranie linka do SID lub PRG
// ================================================================================================
func AudioPost(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	log.Println("AudioPost start with GlobalFileCnt = " + strconv.Itoa(GlobalFileCnt))

	GlobalFileCnt++
	// posted = true

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

// AudioPut - Odebranie pliku SID lub PRG - Drag&Drop
// ================================================================================================
func AudioPut(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	log.Println("AudioPut start with GlobalFileCnt = " + strconv.Itoa(GlobalFileCnt))

	file, fileErr := c.FormFile("file")
	ErrCheck(fileErr)
	log.Println("Odebrałem plik " + file.Filename)

	GlobalFileCnt++
	// posted = true

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

// makeCharsetReader - decode reader
// ================================================================================================
func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	return input, nil

	// if charset == "ISO-8859-1" {
	// 	// Windows-1252 is a superset of ISO-8859-1, so should do here
	// 	return charmap.Windows1252.NewDecoder().Reader(input), nil
	// }
	// return nil, fmt.Errorf("Unknown charset: %s", charset)
}

// // toUtf8 - konwersja kodowania
// // ================================================================================================
// func toUtf8(inputbuf []byte) string {
// 	buf := make([]rune, len(inputbuf))
// 	for i, b := range inputbuf {
// 		buf[i] = rune(b)
// 	}
// 	return string(buf)
// }

// insertRelease - Wstawienie release'u do slice
// ================================================================================================
func insertRelease(array []Release, value Release, index int) []Release {
	return append(array[:index], append([]Release{value}, array[index:]...)...)
}

// ReadLatestReleasesThread - Wątek odczygtujący dane z csdb
// ================================================================================================
func ReadLatestReleasesThread() {

	defer func() {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! Koniec watku ScannerThread !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}()

	netClient := &http.Client{Timeout: time.Second * 10}

	var foundNewReleases int

	for {
		resp, err := netClient.Get("https://csdb.dk/rss/latestreleases.php")

		if ErrCheck(err) {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			ErrCheck(err)
			// fmt.Println(string(body))
			resp.Body.Close()

			// Przerobienie na strukturę

			var latestReleases XMLRssFeed
			reader := bytes.NewReader(body)
			decoder := xml.NewDecoder(reader)
			decoder.CharsetReader = makeCharsetReader
			// err = xml.Unmarshal([]byte(body), &latestReleases)
			err = decoder.Decode(&latestReleases)
			ErrCheck(err)

			// fmt.Println("Odebrano: ", latestReleases)
			// fmt.Println("===================================")
			log.Println("Odebrano listę ostatnich releases...")
			// fmt.Println("===================================")

			foundNewReleases = 0

			var releasesTemp []Release

			for index := 0; index < len(latestReleases.Items); index++ {
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

					var entry XMLRelease
					reader := bytes.NewReader(body)
					decoder := xml.NewDecoder(reader)
					decoder.CharsetReader = makeCharsetReader
					err = decoder.Decode(&entry)
					ErrCheck(err)

					// Szukamy takiego release w naszej bazie
					//

					var relTypesAllowed = [...]string{"C64 Music", "C64 Demo", "C64 One-File Demo", "C64 Intro", "C64 4K Intro", "C64 Crack intro", "C64 Music Collection", "C64 Graphics Collection", "C64 Diskmag", "C64 Charts", "C64 Invitation", "C64 1K Intro", "C64 Fake Demo", "C128 Release"}
					found := false
					for _, rel := range releasesTemp {
						id, _ := strconv.Atoi(entry.ReleaseID)
						if rel.ReleaseID == id {
							found = true
						}
					}
					typeOK := false
					for _, relType := range relTypesAllowed {
						if relType == entry.ReleaseType {
							typeOK = true
							break
						}
					}

					// TODO zrobić update tych info (ktoś mógł uzupełnić potem dane lub pliki)
					// Jeżeli znaleźliśmy to sprawdzamy typ i dodajemy
					//
					if !found && typeOK {

						// Tworzymy nowy obiekt release który dodamy do slice
						//
						var newRelease Release
						id, _ := strconv.Atoi(entry.ReleaseID)
						newRelease.ReleaseID = id
						newRelease.ReleaseName = entry.ReleaseName
						newRelease.ReleaseScreenShot = entry.ReleaseScreenShot
						newRelease.Rating = entry.Rating

						// fmt.Println("Nazwa:  ", entry.ReleaseName)
						// fmt.Println("ID:     ", entry.ReleaseID)
						// fmt.Println("Typ:    ", entry.ReleaseType)
						for _, group := range entry.XMLReleasedBy.XMLGroup {
							// fmt.Println("XMLGroup:  ", group.Name)
							newRelease.ReleasedBy = append(newRelease.ReleasedBy, group.Name)
						}
						for _, handle := range entry.XMLReleasedBy.XMLHandle {
							// fmt.Println("XMLHandle: ", handle.XMLHandle)
							newRelease.ReleasedBy = append(newRelease.ReleasedBy, handle.XMLHandle)
						}
						// fmt.Println("-----------------------------------")
						for _, credit := range entry.Credits {

							creditHandle := "???"
							if len(credit.XMLHandle.XMLHandle) > 0 {
								// fmt.Println(credit.CreditType + ": " + credit.XMLHandle.XMLHandle + " [" + credit.XMLHandle.ID + "]")
								if credit.CreditType == "Music" {
									newRelease.Credits = append(newRelease.Credits, credit.XMLHandle.XMLHandle)
								}
							} else {
								found := false
								for _, releaseHandle := range entry.XMLReleasedBy.XMLHandle {
									if releaseHandle.ID == credit.XMLHandle.ID && releaseHandle.XMLHandle != "" {
										// fmt.Println(credit.CreditType + ": " + releaseHandle.XMLHandle + " [" + releaseHandle.ID + "]")
										creditHandle = releaseHandle.XMLHandle
										found = true
										break
									}
								}
								if !found {
									for _, releaseHandle := range entry.Credits {
										if releaseHandle.XMLHandle.ID == credit.XMLHandle.ID && releaseHandle.XMLHandle.XMLHandle != "" {
											// fmt.Println(credit.CreditType + ": " + releaseHandle.XMLHandle.XMLHandle + " [" + releaseHandle.XMLHandle.ID + "]")
											creditHandle = releaseHandle.XMLHandle.XMLHandle
											break
										}
									}
								}

								// Jeżeli mamy handle i type
								//
								if credit.CreditType == "Music" && creditHandle != "" {
									newRelease.Credits = append(newRelease.Credits, creditHandle)
								}
							}
						}
						// fmt.Println("===================================")

						// Linki dościągnięcia
						// Najpierw SIDy

						for _, link := range entry.DownloadLinks {
							if strings.Contains(link.Link, ".sid") {
								newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
							}
						}
						// Potem PRGs
						for _, link := range entry.DownloadLinks {
							if strings.Contains(link.Link, ".prg") {
								newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
							}
						}

						// Dodajemy new release
						// ale tylko jeżeli mamy niezbędne info o produkcji
						if len(newRelease.DownloadLinks) > 0 {
							releasesTemp = append(releasesTemp, newRelease)
							foundNewReleases++
						}
					}
				} else {
					log.Println("Błąd komunikacji z csdb.dk")
				}
			}

			// Wyświetlenie danych
			log.Println("Found", foundNewReleases, "new music releases.")
			// for _, rel := range releasesTemp {
			// 	// fmt.Println()
			// 	// fmt.Println(rel)
			// 	log.Println(rel)
			// }
			// fmt.Println("===============================================")

			// Przepisanie do zmiennej globalnej
			//

			mutex.Lock()
			releases = releasesTemp
			mutex.Unlock()

		} else {
			log.Println("Błąd komunikacji z csdb.dk")
		}

		// SLEEP
		// ----------------------------------------------------------------------------------------
		time.Sleep(300 * time.Second)
	}

}

// ================================================================================================
// MAIN()
// ================================================================================================
func main() {

	// args := os.Args[1:]

	go ReadLatestReleasesThread()

	// Logowanie do pliku
	//
	logFileApp, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	ErrCheck(err)
	log.SetOutput(io.MultiWriter(os.Stdout, logFileApp))

	gin.DisableConsoleColor()
	logFileGin, err := os.OpenFile("gin.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	ErrCheck(err)
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFileGin)

	log.Println("==========================================")
	log.Println("=======          APP START        ========")
	log.Println("==========================================")

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.Use(Options)

	r.LoadHTMLGlob("./dist/*.html")

	r.StaticFS("/fonts", http.Dir("./dist/fonts"))
	r.StaticFS("/css", http.Dir("./dist/css"))
	r.StaticFS("/js", http.Dir("./dist/js"))

	r.StaticFile("/", "./dist/index.html")
	r.StaticFile("favicon.ico", "./dist/favicon.ico")
	r.StaticFile("sign.png", "./dist/sign.png")

	r.GET("/api/v1/audio/:player", AudioGet)
	r.POST("/api/v1/audio", AudioPost)
	r.PUT("/api/v1/audio", AudioPut)
	r.GET("/api/v1/csdb_releases", CSDBGetLatestReleases)
	r.POST("/api/v1/csdb_release", CSDBGetRelease)

	log.Fatal(autotls.Run(r, "sidcloud.net"))

	// r.Run(":" + args[0])
}
