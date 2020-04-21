// ================================================================================================
// Sidcloud by DKT/Samar
// ================================================================================================

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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
	Ext               string
}

// releases - glówna i globalna tablica z aktualnymi produkcjami
// ================================================================================================
var releases []Release

// sidplayExe - nazwa EXE dla siplayfp
// ================================================================================================
var sidplayExe string

// SIDFile - opis pliku z HVSC
// ================================================================================================
type SIDFile struct {
	ID         int64
	Filepath   string
	Filename   string
	Author     string
	PlayLength int64
}

// hvsc - lista plików HVSC
// ================================================================================================
var hvsc []SIDFile

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

// ReadHVSCJson - Odczyt pliku HVSC
// ================================================================================================
func ReadHVSCJson() {
	file, _ := ioutil.ReadFile("hvsc.json")
	_ = json.Unmarshal([]byte(file), &hvsc)
}

// WriteHVSCJson - Zapis pliku HVSC
// ================================================================================================
func WriteHVSCJson() {
	file, _ := json.MarshalIndent(hvsc, "", " ")
	_ = ioutil.WriteFile("hvsc.json", file, 0666)
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
func DownloadFile(filepath string, url string, id int) (string, error) {

	var ext string = ""

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return ext, err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return ext, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	ErrCheck(err)
	out.Close()

	// Sprawdzzamy rozmiar pliku
	fi, err := os.Stat(filepath)
	if err != nil {
		return ext, err
	}
	// get the size
	size := fi.Size()
	log.Println("Ściągam plik '" + filepath + "' o rozmiarze " + strconv.Itoa(int(size)))

	// // if size < 8 || size > 65535 {
	// if size < 8 || size > 5*1024*1024 { // Może być ZIP z innymi większymi plikami więc ustalilem max na 5M
	// 	err := errors.New("Rozmiar pliku niewłaściwy")
	// 	return ext, err
	// }

	//
	// Rozpakowanie pliku D64
	//
	if strings.Contains(filepath, ".d64") {

		extractedPRG, found := ExtractD64(filepath)
		if found {

			ext = ".prg"

			// Create the file
			out, err := os.Create("cache/" + strconv.Itoa(id) + ext)
			ErrCheck(err)
			defer out.Close()

			// Write the body to file
			_, err = out.Write(extractedPRG)
			ErrCheck(err)
			out.Close()

			return ext, err
		}
	}

	//
	// Rozkakowanie pliku ZIP
	//
	if strings.Contains(filepath, ".zip") {
		zipReader, _ := zip.OpenReader(filepath)

		//
		// Najpierw SIDy
		//
		for _, file := range zipReader.File {

			log.Println(file.Name)

			if strings.Contains(file.Name, ".sid") && !file.FileInfo().IsDir() {

				log.Println("Found SID file")
				ext = ".sid"
				log.Println("File extracted: " + file.Name + " with ID " + strconv.Itoa(id))
				outputFile, err := os.OpenFile(
					"cache/"+strconv.Itoa(id)+ext,
					os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
					file.Mode(),
				)
				ErrCheck(err)
				defer outputFile.Close()

				zippedFile, err := file.Open()
				ErrCheck(err)
				defer zippedFile.Close()

				_, err = io.Copy(outputFile, zippedFile)
				ErrCheck(err)

				return ext, err
			}
		}

		//
		// Potem PRG
		//
		for _, file := range zipReader.File {

			log.Println(file.Name)

			if strings.Contains(file.Name, ".prg") && !file.FileInfo().IsDir() {

				// Sprawdzamy czy PRG ładuje się pod $0801
				zippedFile, err := file.Open()
				ErrCheck(err)
				defer zippedFile.Close()
				p := make([]byte, 2)
				zippedFile.Read(p)
				zippedFile.Close()

				if p[0] == 1 && p[1] == 8 {

					log.Println("Found PRG file")
					ext = ".prg"

					log.Println("File extracted: " + file.Name + " with ID " + strconv.Itoa(id))
					outputFile, err := os.OpenFile(
						"cache/"+strconv.Itoa(id)+ext,
						os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
						file.Mode(),
					)
					ErrCheck(err)
					defer outputFile.Close()

					zippedFile, err := file.Open()
					ErrCheck(err)
					defer zippedFile.Close()

					_, err = io.Copy(outputFile, zippedFile)
					ErrCheck(err)

					return ext, err

				}
				log.Println("PRG file load address != $0801")

			}
		}

		//
		// Potem D64
		//
		for _, file := range zipReader.File {

			log.Println(file.Name)

			if strings.Contains(file.Name, ".d64") && !file.FileInfo().IsDir() {

				log.Println("Found D64 file")
				log.Println("File extracted: " + file.Name + " with ID " + strconv.Itoa(id))
				outputFile, err := os.OpenFile(
					"cache/"+strconv.Itoa(id)+".d64",
					os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
					file.Mode(),
				)
				ErrCheck(err)
				defer outputFile.Close()

				zippedFile, err := file.Open()
				ErrCheck(err)
				defer zippedFile.Close()

				_, err = io.Copy(outputFile, zippedFile)
				ErrCheck(err)

				extractedPRG, found := ExtractD64("cache/" + strconv.Itoa(id) + ".d64")
				if found {

					ext = ".prg"

					// Create the file
					out, err := os.Create("cache/" + strconv.Itoa(id) + ext)
					ErrCheck(err)
					defer out.Close()

					// Write the body to file
					_, err = out.Write(extractedPRG)
					ErrCheck(err)
					out.Close()

					return ext, err
				}

			}
		}
	}

	return ext, err
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
			filenameSID := cacheDir + id + rel.Ext
			paramName := "-w" + cacheDir + id

			var cmdName string

			czas := "-t333"
			// bits := "-p16"
			// freq := "-f44100"
			model := "-mn"

			// Odpalenie sidplayfp
			if runtime.GOOS == "windows" {
				cmdName = "sidplayfp/sidplayfp.exe"
			} else {
				cmdName = sidplayExe // zakładamy że jest zainstalowany
			}

			log.Println("Starting sidplayfp... cmdName(" + cmdName + " " + czas + " " + model + " " + paramName + " " + filenameSID + ")")
			cmd := exec.Command(cmdName, czas, model, paramName, filenameSID)
			err := cmd.Run()
			if ErrCheck(err) {

				mutex.Lock()
				releases[index].WAVCached = true
				mutex.Unlock()
				log.Println(filenameWAV + " cached")
				WriteDb()
			}

		} else {
			// log.Println("Plik " + filenameWAV + " już istnieje")
			mutex.Lock()
			releases[index].WAVCached = true
			mutex.Unlock()
			// log.Println(filenameWAV + " cached")
			WriteDb()
		}

	}
}

// GetHVSCFilter - lista przefiltrowanych SIDów
// ================================================================================================
func GetHVSCFilter(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	// Info o wejściu do GET
	log.Println("GetHVSCFilter()")

	// Odczytujemy parametr - filtr
	id := c.Param("id")

	id = strings.ToLower(id)

	var hvscTemp []SIDFile

	for _, sid := range hvsc {
		searchAuthor := strings.ToLower(sid.Author)
		searchFilename := strings.ToLower(sid.Filename)
		if strings.Contains(searchAuthor, id) || strings.Contains(searchFilename, id) {
			hvscTemp = append(hvscTemp, sid)
		}
	}

	c.JSON(http.StatusOK, hvscTemp)
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

	var sum float64
	silenceCnt := 0

	for loop {

		sum = 0

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
			readed, err := file.ReadAt(p, offset)
			file.Close()

			if ErrCheck(err) {

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

							sum += math.Abs(valFloat)
						}
					}

					sum = sum / float64(readed)

					if sum >= 5.0 || offset < 44 || offset > 44100*60 {
						c.Data(http.StatusOK, "audio/wav", p)
						// log.Print(".")
					}

					if sum < 5.0 && offset > 44100*60 { // po 30 sekundach
						silenceCnt++
						if silenceCnt >= 5 {
							log.Println("Silence at " + strconv.FormatInt(offset, 10))
							loop = false
						}
					}

					offset += int64(readed)
				}
			}
		}

		// Wysyłamy pakiet co 250 ms
		if sum >= 5.0 || offset > 44100*60 {
			// if sum >= 5.0 {
			time.Sleep(250 * time.Millisecond)
		}

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
						// Potem ZIPy
						for _, link := range entry.DownloadLinks {
							if strings.Contains(link.Link, ".zip") {
								newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
							}
						}
						// Potem D64y
						for _, link := range entry.DownloadLinks {
							if strings.Contains(link.Link, ".d64") {
								newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
							}
						}

						// Potem ściągamy
						if len(newRelease.DownloadLinks) > 0 {
							filename := cacheDir
							if strings.Contains(newRelease.DownloadLinks[0], ".sid") {
								filename += strconv.Itoa(newRelease.ReleaseID) + ".sid"
							}
							if strings.Contains(newRelease.DownloadLinks[0], ".prg") {
								filename += strconv.Itoa(newRelease.ReleaseID) + ".prg"
							}
							if strings.Contains(newRelease.DownloadLinks[0], ".zip") {
								filename += strconv.Itoa(newRelease.ReleaseID) + ".zip"
							}
							if strings.Contains(newRelease.DownloadLinks[0], ".d64") {
								filename += strconv.Itoa(newRelease.ReleaseID) + ".d64"
							}

							// Dodajemy new release
							// ale tylko jeżeli mamy niezbędne info o produkcji
							if filename != "" {

								if !fileExists(filename) {
									_, err := DownloadFile(filename, newRelease.DownloadLinks[0], newRelease.ReleaseID)
									if ErrCheck(err) {
										newRelease.SIDCached = true
									}
								} else {
									newRelease.SIDCached = true
								}

								if fileExists(cacheDir + strconv.Itoa(newRelease.ReleaseID) + ".sid") {
									newRelease.Ext = ".sid"
								}
								if fileExists(cacheDir + strconv.Itoa(newRelease.ReleaseID) + ".prg") {
									newRelease.Ext = ".prg"
								}
								if fileExists(cacheDir + strconv.Itoa(newRelease.ReleaseID) + ".wav") {
									newRelease.WAVCached = true
								}

								if len(newRelease.Ext) > 0 {
									releasesTemp = append(releasesTemp, newRelease)
									foundNewReleases++
								}
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

// HVSCPrepareData - Wątek odczygtujący dane z HVSC
// ================================================================================================
func HVSCPrepareData() {

	log.Println("HVSC start")
	var id int64

	root := "./C64Music/Games"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".sid" {
			return nil
		}

		split := "/"
		if runtime.GOOS == "windows" {
			split = "\\"
		}

		pathSlice := strings.Split(path, split)
		var newSIDFile SIDFile
		newSIDFile.ID = id
		newSIDFile.Filepath = path
		newSIDFile.Author = "Games"
		newSIDFile.Filename = strings.ReplaceAll(strings.TrimSuffix(pathSlice[len(pathSlice)-1], ".sid"), "_", " ")
		hvsc = append(hvsc, newSIDFile)
		id++
		return nil
	})
	ErrCheck(err)

	root = "./C64Music/Musicians"
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".sid" {
			return nil
		}

		split := "/"
		if runtime.GOOS == "windows" {
			split = "\\"
		}

		pathSlice := strings.Split(path, split)
		var newSIDFile SIDFile
		newSIDFile.ID = id
		newSIDFile.Filepath = path
		newSIDFile.Author = strings.ReplaceAll(pathSlice[len(pathSlice)-2], "_", " ")
		newSIDFile.Filename = strings.ReplaceAll(strings.TrimSuffix(pathSlice[len(pathSlice)-1], ".sid"), "_", " ")
		hvsc = append(hvsc, newSIDFile)
		id++
		return nil
	})
	ErrCheck(err)

	WriteHVSCJson()
	log.Println("HVSC stop")
}

// D64GetSector - Read sector from D64
// ================================================================================================
func D64GetSector(file *os.File, track byte, sector byte) []byte {

	var track2Address = [...]int64{0x00000, 0x00000, 0x01500, 0x02A00, 0x03F00, 0x05400, 0x06900,
		0x07E00, 0x09300, 0x0A800, 0x0BD00, 0x0D200, 0x0E700, 0x0FC00, 0x11100, 0x12600, 0x13B00,
		0x15000, 0x16500, 0x17800, 0x18B00, 0x19E00, 0x1B100, 0x1C400, 0x1D700, 0x1EA00, 0x1FC00,
		0x20E00, 0x22000, 0x23200, 0x24400, 0x25600, 0x26700, 0x27800, 0x28900, 0x29A00, 0x2AB00,
		0x2BC00, 0x2CD00, 0x2DE00, 0x2EF00}

	file.Seek(track2Address[int64(track)]+256*int64(sector), 0)
	p := make([]byte, 256)
	file.Read(p)

	return p
}

// ExtractD64 - Extract PRG from D64
// ================================================================================================
func ExtractD64(filename string) ([]byte, bool) {

	file, err := os.Open(filename)
	ErrCheck(err)
	defer file.Close()

	var dirTrack byte = 18
	var dirSector byte = 1
	var outfile []byte
	loop := true

	for loop {
		log.Println("I'm on dir track " + strconv.Itoa(int(dirTrack)) + " and sector " + strconv.Itoa(int(dirSector)))
		sector := D64GetSector(file, dirTrack, dirSector)
		for ptr := 0; ptr < 8*0x20; ptr += 0x20 {
			if (sector[ptr+2] & 7) == 2 {
				log.Println("Found PRG file, ptr " + strconv.Itoa(int(ptr)))
				name := string(sector[ptr+5 : ptr+14])
				var fileTrack byte = sector[ptr+3]
				var fileSector byte = sector[ptr+4]
				log.Println(name + " T:" + strconv.Itoa(int(fileTrack)) + " S:" + strconv.Itoa(int(fileSector)))

				// Najpierw sprawdzimy czy load address == $0801
				prg := D64GetSector(file, fileTrack, fileSector)

				if prg[2] == 1 && prg[3] == 8 {
					log.Println("Loading address is OK")

					fileTrack = prg[0]
					fileSector = prg[1]
					fileloop := true
					outfile = append(outfile, prg[2:]...)

					for fileloop && fileTrack != 0 {
						prg = D64GetSector(file, fileTrack, fileSector)
						fileTrack = prg[0]
						fileSector = prg[1]
						outfile = append(outfile, prg[2:]...)
					}
					log.Println("Koniec pliku")
					return outfile, true

					// fileloop = false
					// loop = false
					// break
				}
				log.Println("Loading address is NOK")
			}
		}
		dirTrack = sector[0]
		if dirTrack == 0 {
			break
		}
		dirSector = sector[1]
	}

	return outfile, false

}

// ================================================================================================
// MAIN()
// ================================================================================================
func main() {

	// HVSCPrepareData()

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
	r.GET("/api/v1/hvsc_filter/:id", GetHVSCFilter)

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
