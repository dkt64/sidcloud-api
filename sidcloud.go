// ================================================================================================
// Sidcloud by DKT/Samar
// ================================================================================================

package main

import (
	"bytes"
	"encoding/json"
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

const cacheDir = "cache/"

// mutex - wielowątkowość na tablicy releases
// ------------------------------------------------------------------------------------------------
var mutex = &sync.Mutex{}

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
	SIDCached         bool
	WAVCached         bool
}

// releases - glówna i globalna tablica z aktualnymi produkcjami
// ================================================================================================
var releases []Release

// sidplayExe - nazwa EXE dla siplayfp
// ================================================================================================
var sidplayExe string

// ErrCheck - obsługa błedów
// ================================================================================================
func ErrCheck(errNr error) bool {
	if errNr != nil {
		fmt.Println(errNr)
		return false
	}
	return true
}

// ReadDb - Odczyt bazy
// ================================================================================================
func ReadDb() {
	file, _ := ioutil.ReadFile("releases.json")
	_ = json.Unmarshal([]byte(file), &releases)
}

// WriteDb - Zapis bazy
// ================================================================================================
func WriteDb() {
	file, _ := json.MarshalIndent(releases, "", " ")
	_ = ioutil.WriteFile("releases.json", file, 0666)
}

// fileExists - sprawdzenie czy plik istnieje
// ================================================================================================
func fileExists(filename string) bool {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
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
	log.Println("Ściągam plik '" + filepath + "' o rozmiarze " + strconv.Itoa(int(size)))

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

	return err
}

// CSDBGetLatestReleases - ostatnie release'y
// ================================================================================================
func CSDBGetLatestReleases(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	// Info o wejściu do GET
	log.Println("CSDBGetLatestReleases()")

	mutex.Lock()
	releasesTemp := releases
	mutex.Unlock()

	c.JSON(http.StatusOK, releasesTemp)
}

// CreateWAVFiles - Creating WAV files
// ================================================================================================
func CreateWAVFiles() {
	//

	log.Println("CreateWAVFiles()...")
	for index, rel := range releases {

		id := strconv.Itoa(rel.ReleaseID)
		filenameWAV := cacheDir + id + ".wav"

		var size int64

		if fileExists(filenameWAV) {
			file, err := os.Stat(filenameWAV)
			if err != nil {
				fmt.Println("Problem z odczytem rozmiaru pliku " + filenameWAV)
			}
			size = file.Size()
		}

		if !fileExists(filenameWAV) || size < 29458844 {

			log.Println("Tworzenie pliku " + filenameWAV)
			filenameSID := cacheDir + id + ".sid"
			filenamePRG := cacheDir + id + ".prg"
			// log.Println("WAV filename = " + filenameWAV)

			var filename = ""
			if strings.Contains(rel.DownloadLinks[0], ".sid") {
				filename = filenameSID
			} else if strings.Contains(rel.DownloadLinks[0], ".prg") {
				filename = filenamePRG
			}

			paramName := "-w" + cacheDir + id

			var cmdName string

			czas := "-t333"
			// bits := "-p16"
			// freq := "-f44100"

			// Odpalenie sidplayfp
			if runtime.GOOS == "windows" {
				cmdName = "sidplayfp/sidplayfp.exe"
			} else {
				cmdName = sidplayExe // zakładamy że jest zainstalowany
			}

			log.Println("Starting sidplayfp... cmdName(" + cmdName + " " + czas + " " + paramName + " " + filename + ")")
			cmd := exec.Command(cmdName, czas, paramName, filename)
			err := cmd.Run()
			if ErrCheck(err) {

				mutex.Lock()
				releases[index].WAVCached = true
				mutex.Unlock()
				log.Println(filenameWAV + " cached")
				WriteDb()
			}

		} else {
			log.Println("Plik " + filenameWAV + " już istnieje")
			mutex.Lock()
			releases[index].WAVCached = true
			mutex.Unlock()
			log.Println(filenameWAV + " cached")
			WriteDb()
		}

	}
}

// AudioGet - granie utworu
// ================================================================================================
func AudioGet(c *gin.Context) {

	// Typ połączania
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Connection", "Keep-Alive")
	c.Header("Transfer-Encoding", "chunked")

	// Odczytujemy parametr - typ playera
	id := c.Param("id")
	filenameWAV := cacheDir + id + ".wav"

	// const maxOffset int64 = 50000000 // ~ 10 min
	// const maxOffset int64 = 25000000 // ~ 5 min
	// const maxOffset int64 = 5000000 // ~ 1 min
	const maxOffset int64 = 44100*2*300 + 44 // = 5 min
	const maxVol float64 = 1.25

	var vol float64 = maxVol
	loop := true
	volDown := false

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
			defer file.Close()
			// ErrCheck(err)

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
	// }

	// Feedback gdybyśmy wyszli z LOOP
	c.JSON(http.StatusOK, "Loop ended.")
	log.Println("Loop ended.")
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

						// Potem ściągamy
						if len(newRelease.DownloadLinks) > 0 {
							filename := cacheDir
							if strings.Contains(newRelease.DownloadLinks[0], ".prg") {
								filename += strconv.Itoa(newRelease.ReleaseID) + ".prg"
							}
							if strings.Contains(newRelease.DownloadLinks[0], ".sid") {
								filename += strconv.Itoa(newRelease.ReleaseID) + ".sid"
							}

							// Dodajemy new release
							// ale tylko jeżeli mamy niezbędne info o produkcji
							if filename != "" {

								if !fileExists(filename) {
									err := DownloadFile(filename, newRelease.DownloadLinks[0])
									if ErrCheck(err) {
										newRelease.SIDCached = true
									}
								} else {
									newRelease.SIDCached = true
								}

								releasesTemp = append(releasesTemp, newRelease)
								foundNewReleases++
							}
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
			WriteDb()

			CreateWAVFiles()
			WriteDb()

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

	args := os.Args[1:]

	if len(args) == 0 || (args[0] != "http" && args[0] != "https") {
		log.Fatal("Podaj parametr 'http' lub 'https'!")
		os.Exit(1)
	}

	sidplayExe = "sidplayfp/sidplayfp"

	if args[0] == "http" {
		if len(args) > 2 {
			if args[2] == "arm" {
				sidplayExe = "sidplayfp"
			}
		}
	}

	ReadDb()
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

	if args[0] == "https" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(Options)

	r.LoadHTMLGlob("./dist/*.html")

	r.StaticFS("/fonts", http.Dir("./dist/fonts"))
	r.StaticFS("/css", http.Dir("./dist/css"))
	r.StaticFS("/js", http.Dir("./dist/js"))

	r.StaticFile("/", "./dist/index.html")
	r.StaticFile("favicon.ico", "./dist/favicon.ico")
	r.StaticFile("sign.png", "./dist/sign.png")

	r.GET("/api/v1/audio/:id", AudioGet)
	r.GET("/api/v1/csdb_releases", CSDBGetLatestReleases)

	if args[0] == "https" {
		log.Fatal(autotls.Run(r, "sidcloud.net", "www.sidcloud.net"))
	}
	if args[0] == "http" {
		if len(args) > 1 {
			r.Run(":" + args[1])
		} else {
			r.Run(":80")
		}
	}
}
