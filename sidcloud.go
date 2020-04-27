// ================================================================================================
// Sidcloud by DKT/Samar
// ================================================================================================

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
)

var csdbDataReady bool = false

const cacheDir = "cache/"
const wavSize = 29458844
const wavTime = "333"

const historyMaxEntries = 80

const historyMaxMonths = 3

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

// XMLEvent - kompo
// ------------------------------------------------------------------------------------------------
type XMLEvent struct {
	ID   string `xml:"ID"`
	Name string `xml:"Name"`
}

// XMLReleasedAt - kompa
// ------------------------------------------------------------------------------------------------
type XMLReleasedAt struct {
	XMLEvent XMLEvent `xml:"Event"`
}

// XMLUsedSID - SIDy
// ------------------------------------------------------------------------------------------------
type XMLUsedSID struct {
	ID       string `xml:"ID"`
	HVSCPath string `xml:"HVSCPath"`
	Name     string `xml:"Name"`
	Author   string `xml:"Author"`
}

// XMLRelease - wydanie produkcji na csdb
// ------------------------------------------------------------------------------------------------
type XMLRelease struct {
	ReleaseID         string            `xml:"Release>ID"`
	ReleaseName       string            `xml:"Release>Name"`
	ReleaseType       string            `xml:"Release>Type"`
	ReleaseYear       string            `xml:"Release>ReleaseYear"`
	ReleaseMonth      string            `xml:"Release>ReleaseMonth"`
	ReleaseDay        string            `xml:"Release>ReleaseDay"`
	ReleaseScreenShot string            `xml:"Release>ScreenShot"`
	Rating            float32           `xml:"Release>Rating"`
	XMLReleasedBy     XMLReleasedBy     `xml:"Release>ReleasedBy"`
	XMLReleasedAt     XMLReleasedAt     `xml:"Release>ReleasedAt"`
	Credits           []XMLCredit       `xml:"Release>Credits>Credit"`
	DownloadLinks     []XMLDownloadLink `xml:"Release>DownloadLinks>DownloadLink"`
	UsedSIDs          []XMLUsedSID      `xml:"Release>UsedSIDs>SID"`
}

// LatestRelease - najwyższy numer ID
// ------------------------------------------------------------------------------------------------
type LatestRelease struct {
	ID int `xml:"LatestReleaseId"`
}

// UsedSID - wydanie produkcji na csdb
// ------------------------------------------------------------------------------------------------
type UsedSID struct {
	ID       string
	HVSCPath string
	Name     string
	Author   string
}

// Release - wydanie produkcji na csdb
// ------------------------------------------------------------------------------------------------
type Release struct {
	ReleaseID         int
	ReleaseYear       int
	ReleaseMonth      int
	ReleaseDay        int
	ReleaseName       string
	ReleaseType       string
	ReleaseScreenShot string
	ReleasedAt        string
	SIDPath           string
	Rating            float32
	ReleasedBy        []string
	Credits           []string
	DownloadLinks     []string
	// UsedSIDs          []UsedSID
	SIDCached bool
	WAVCached bool
	Ext       string
}

// releases - glówna i globalna tablica z aktualnymi produkcjami
// ================================================================================================
var releases []Release

// allReleases - glówna i globalna tablica ze wszystkimi produkcjami
// ================================================================================================
var csdb []Release

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
		log.Println(errNr)
		return false
	}
	return true
}

// ErrCheck2 - obsługa błedów bez komunikatu
// ================================================================================================
func ErrCheck2(errNr error) bool {
	if errNr != nil {
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

// ReadCSDb - Odczyt bazy
// ================================================================================================
func ReadCSDb() {
	file, _ := ioutil.ReadFile("csdb.json")
	_ = json.Unmarshal([]byte(file), &csdb)
}

// WriteCSDb - Zapis bazy
// ================================================================================================
func WriteCSDb() {
	file, _ := json.MarshalIndent(csdb, "", " ")
	_ = ioutil.WriteFile("csdb.json", file, 0666)
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

// insertRelease - Wstawienie release'u do slice
// ================================================================================================
func insertRelease(array []Release, value Release, index int) []Release {
	return append(array[:index], append([]Release{value}, array[index:]...)...)
}

// Difference - Różnica pomiędzy dwoma slice
// ================================================================================================
func Difference(a, b []Release) (diff []Release) {

	for _, itema := range b {
		found := false
		for _, itemb := range a {
			if itema.ReleaseID == itemb.ReleaseID &&
				len(itema.DownloadLinks) == len(itemb.DownloadLinks) {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, itema)
		}
	}

	return diff
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
		log.Println("[ExtractD64] Dir track " + strconv.Itoa(int(dirTrack)) + " and sector " + strconv.Itoa(int(dirSector)))
		sector := D64GetSector(file, dirTrack, dirSector)
		for ptr := 0; ptr < 8*0x20; ptr += 0x20 {
			if (sector[ptr+2] & 7) == 2 {
				name := string(sector[ptr+5 : ptr+14])
				var fileTrack byte = sector[ptr+3]
				var fileSector byte = sector[ptr+4]

				// Najpierw sprawdzimy czy load address == $0801
				log.Println("[ExtractD64] Reading " + name + " TRACK:" + strconv.Itoa(int(fileTrack)) + " SECTOR:" + strconv.Itoa(int(fileSector)))
				prg := D64GetSector(file, fileTrack, fileSector)

				if (prg[2] == 1 || prg[2] == 0) && prg[3] == 8 {
					log.Println("[ExtractD64] Loading address is OK")
					// log.Println("[ExtractD64] First PRG file in D64: " + name + " T:" + strconv.Itoa(int(fileTrack)) + " S:" + strconv.Itoa(int(fileSector)))

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
					// log.Println("Koniec pliku")
					return outfile, true

					// fileloop = false
					// loop = false
					// break
				}
				log.Println("[ExtractD64] Loading address is NOK")
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

// Sortowanie datami
// ================================================================================================

type byDate []Release

func (s byDate) Len() int {
	return len(s)
}
func (s byDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byDate) Less(i, j int) bool {

	d1 := time.Date(s[i].ReleaseYear, time.Month(s[i].ReleaseMonth), s[i].ReleaseDay, 0, 0, 0, 0, time.Local)
	d2 := time.Date(s[j].ReleaseYear, time.Month(s[j].ReleaseMonth), s[j].ReleaseDay, 0, 0, 0, 0, time.Local)

	return d2.Before(d1)
}

// Sortowanie byID
// ================================================================================================

type byID []Release

func (s byID) Len() int {
	return len(s)
}
func (s byID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byID) Less(i, j int) bool {
	return s[i].ReleaseID > s[j].ReleaseID
}

// fileExists - sprawdzenie czy plik istnieje
// ================================================================================================
func fileExists(filename string) bool {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}

	return true
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
	log.Println("[DownloadFile] Downloading file '" + filepath + "' size " + strconv.Itoa(int(size)))

	// // if size < 8 || size > 65535 {
	// if size < 8 || size > 5*1024*1024 { // Może być ZIP z innymi większymi plikami więc ustalilem max na 5M
	// 	err := errors.New("Rozmiar pliku niewłaściwy")
	// 	return ext, err
	// }

	//
	// Rozpakowanie pliku D64
	//
	if strings.Contains(filepath, ".d64") || strings.Contains(filepath, ".D64") {

		extractedPRG, found := ExtractD64(filepath)
		if found {

			ext = ".prg"

			// Create the file
			out, err := os.Create(cacheDir + strconv.Itoa(id) + ext)
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

			// log.Println(file.Name)

			if strings.Contains(file.Name, ".sid") && !file.FileInfo().IsDir() {

				log.Println("[DownloadFile] Found SID file")
				ext = ".sid"
				log.Println("[DownloadFile] File extracted: " + file.Name + " with ID " + strconv.Itoa(id))
				outputFile, err := os.OpenFile(
					cacheDir+strconv.Itoa(id)+ext,
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

			// log.Println(file.Name)

			if strings.Contains(file.Name, ".prg") && !file.FileInfo().IsDir() {

				// Sprawdzamy czy PRG ładuje się pod $0801
				zippedFile, err := file.Open()
				ErrCheck(err)
				defer zippedFile.Close()
				p := make([]byte, 2)
				zippedFile.Read(p)
				zippedFile.Close()

				if p[0] == 1 && p[1] == 8 {

					// log.Println("[DownloadFile] Found PRG file")
					ext = ".prg"

					log.Println("[DownloadFile] File extracted: " + file.Name + " with ID " + strconv.Itoa(id))
					outputFile, err := os.OpenFile(
						cacheDir+strconv.Itoa(id)+ext,
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
				// log.Println("[DownloadFile] PRG file load address != $0801")
			}
		}

		//
		// Potem D64
		//
		for _, file := range zipReader.File {

			// log.Println(file.Name)

			if strings.Contains(file.Name, ".d64") && !file.FileInfo().IsDir() {

				// log.Println("[DownloadFile] Found D64 file")
				log.Println("[DownloadFile] File extracted: " + file.Name + " with ID " + strconv.Itoa(id))
				outputFile, err := os.OpenFile(
					cacheDir+strconv.Itoa(id)+".d64",
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

				extractedPRG, found := ExtractD64(cacheDir + strconv.Itoa(id) + ".d64")
				if found {

					ext = ".prg"

					// Create the file
					out, err := os.Create(cacheDir + strconv.Itoa(id) + ext)
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

// DownloadFiles - Wątek ściągający pliki produkcji z csdb
// ================================================================================================
func DownloadFiles() {

	log.Println("[DownloadFiles] LOOP")

	for index, newRelease := range releases {
		if len(newRelease.DownloadLinks) > 0 {
			filename := cacheDir
			if strings.Contains(newRelease.DownloadLinks[0], ".sid") || strings.Contains(newRelease.DownloadLinks[0], ".SID") {
				filename += strconv.Itoa(newRelease.ReleaseID) + ".sid"
			}
			if strings.Contains(newRelease.DownloadLinks[0], ".prg") || strings.Contains(newRelease.DownloadLinks[0], ".PRG") {
				filename += strconv.Itoa(newRelease.ReleaseID) + ".prg"
			}
			if strings.Contains(newRelease.DownloadLinks[0], ".zip") || strings.Contains(newRelease.DownloadLinks[0], ".ZIP") {
				filename += strconv.Itoa(newRelease.ReleaseID) + ".zip"
			}
			if strings.Contains(newRelease.DownloadLinks[0], ".d64") || strings.Contains(newRelease.DownloadLinks[0], ".D64") {
				filename += strconv.Itoa(newRelease.ReleaseID) + ".d64"
			}

			// Dodajemy new release
			// ale tylko jeżeli mamy niezbędne info o produkcji
			if filename != "" {
				if !newRelease.SIDCached || !(fileExists(cacheDir+strconv.Itoa(newRelease.ReleaseID)+".sid") || fileExists(cacheDir+strconv.Itoa(newRelease.ReleaseID)+".prg")) {
					_, err := DownloadFile(filename, newRelease.DownloadLinks[0], newRelease.ReleaseID)

					if ErrCheck(err) {
						// Sprawdzay czy istnieje SID lub PRG
						if fileExists(cacheDir+strconv.Itoa(newRelease.ReleaseID)+".sid") || fileExists(cacheDir+strconv.Itoa(newRelease.ReleaseID)+".prg") {
							newRelease.SIDCached = true
							log.Println("[DownloadFiles] File cached")
						} else {
							log.Println("[DownloadFiles] File not cached")
							newRelease.SIDCached = false
							newRelease.WAVCached = false
						}
						// SendEmail("Nowa produkcja na CSDB.DK: " + newRelease.ReleaseName + " by " + newRelease.ReleasedBy[0])
					}

					if fileExists(cacheDir + strconv.Itoa(newRelease.ReleaseID) + ".sid") {
						newRelease.Ext = ".sid"
					}
					if fileExists(cacheDir + strconv.Itoa(newRelease.ReleaseID) + ".prg") {
						newRelease.Ext = ".prg"
					}

					releases[index] = newRelease

					WriteDb()
				}
			}
		}
	}
}

// CreateWAVFiles - Creating WAV files
// ================================================================================================
func CreateWAVFiles() {

	log.Println("[CreateWAVFiles] LOOP")
	for index, rel := range releases {

		if len(rel.Ext) == 4 && rel.SIDCached {
			id := strconv.Itoa(rel.ReleaseID)
			filenameWAV := cacheDir + id + ".wav"

			var size int64

			if fileExists(filenameWAV) {
				file, err := os.Stat(filenameWAV)
				if err != nil {
					log.Println("[CreateWAVFiles] Problem z odczytem rozmiaru pliku " + filenameWAV)
				}
				size = file.Size()
			}

			if !fileExists(filenameWAV) || size < wavSize || (fileExists(filenameWAV) && !rel.WAVCached) {

				log.Println("[CreateWAVFiles] Creating file " + filenameWAV)
				filenameSID := cacheDir + id + rel.Ext
				paramName := "-w" + cacheDir + id

				var cmdName string

				czas := "-t" + wavTime
				// bits := "-p16"
				// freq := "-f44100"
				model := "-mn"

				// Odpalenie sidplayfp
				if runtime.GOOS == "windows" {
					cmdName = "sidplayfp/sidplayfp.exe"
				} else {
					cmdName = sidplayExe // zakładamy że jest zainstalowany
				}

				log.Println("[CreateWAVFiles] Starting sidplayfp... cmdName(" + cmdName + " " + czas + " " + model + " " + paramName + " " + filenameSID + ")")
				cmd := exec.Command(cmdName, czas, model, paramName, filenameSID)
				err := cmd.Run()
				if ErrCheck(err) {

					// Jeszcze raz sprawdzamy czy plik powstał o odpowiedniej długości
					if fileExists(filenameWAV) {
						file, err := os.Stat(filenameWAV)
						if err != nil {
							log.Println("[CreateWAVFiles] Problem z odczytem rozmiaru pliku " + filenameWAV)
						}
						size = file.Size()
					}

					if fileExists(filenameWAV) && size >= wavSize {
						releases[index].WAVCached = true
						log.Println("[CreateWAVFiles] " + filenameWAV + " cached")
						WriteDb()
					} else {
						log.Println("[CreateWAVFiles] " + filenameWAV + " not cached")
					}
				} else {
					log.Println("[CreateWAVFiles] Problem with sidplayfp and " + filenameWAV)
				}

			} else {
				// log.Println("Plik " + filenameWAV + " już istnieje")
				releases[index].WAVCached = true
				// log.Println(filenameWAV + " cached")
				WriteDb()
			}
		}

	}

}

// updateReleaseInfo - Dane do zmiany w releases
// ================================================================================================
func updateReleaseInfo(index int, newRelease Release) {
	releases[index].Credits = newRelease.Credits
	releases[index].Rating = newRelease.Rating
	releases[index].ReleaseScreenShot = newRelease.ReleaseScreenShot
	releases[index].ReleasedAt = newRelease.ReleasedAt
	releases[index].ReleasedBy = newRelease.ReleasedBy

	if newRelease.DownloadLinks != nil {
		if releases[index].DownloadLinks == nil {
			for _, link := range newRelease.DownloadLinks {
				releases[index].DownloadLinks = append(releases[index].DownloadLinks, link)
			}
		} else if len(releases[index].DownloadLinks) < len(newRelease.DownloadLinks) && releases[index].DownloadLinks[0] != ".sid" && newRelease.DownloadLinks[0] == ".sid" {
			releases[index].DownloadLinks = newRelease.DownloadLinks
			releases[index].SIDCached = false
			releases[index].WAVCached = false
		}
	}
}

// ReadLatestReleases - Wątek odczygtujący dane z csdb
// ================================================================================================
func ReadLatestReleases() {

	netClient := &http.Client{Timeout: time.Second * 10}

	var foundNewReleases int

	log.Println("[ReadLatestReleases] LOOP")

	resp, err := netClient.Get("https://csdb.dk/rss/latestreleases.php")

	if ErrCheck(err) {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		ErrCheck(err)
		// log.Println(string(body))
		resp.Body.Close()

		// Przerobienie na strukturę

		var latestReleases XMLRssFeed
		reader := bytes.NewReader(body)
		decoder := xml.NewDecoder(reader)
		decoder.CharsetReader = makeCharsetReader
		// err = xml.Unmarshal([]byte(body), &latestReleases)
		err = decoder.Decode(&latestReleases)
		ErrCheck(err)

		// log.Println("Odebrano: ", latestReleases)
		// log.Println("===================================")
		log.Println("[ReadLatestReleases] Got latest releases RSS")
		// log.Println("===================================")

		foundNewReleases = 0

		var releasesTemp []Release

		for index := 0; index < len(latestReleases.Items); index++ {
			rssItem := latestReleases.Items[index]
			// log.Println(rssItem.Title)
			url, err := url.Parse(rssItem.GUID)
			ErrCheck(err)
			q := url.Query()
			// log.Println(q.Get("id"))

			resp, err := netClient.Get("https://csdb.dk/webservice/?type=release&id=" + q.Get("id"))

			if ErrCheck(err) {
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				ErrCheck(err)
				// log.Println(string(body))
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

				id, _ := strconv.Atoi(entry.ReleaseID)

				var relTypesAllowed = [...]string{"C64 Music", "C64 Demo", "C64 One-File Demo", "C64 Intro", "C64 4K Intro", "C64 Crack intro", "C64 Music Collection", "C64 Graphics Collection", "C64 Diskmag", "C64 Charts", "C64 Invitation", "C64 1K Intro", "C64 Fake Demo"}
				found := false
				for _, rel := range releasesTemp {
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
					newRelease.ReleaseYear, _ = strconv.Atoi(entry.ReleaseYear)
					newRelease.ReleaseMonth, _ = strconv.Atoi(entry.ReleaseMonth)
					newRelease.ReleaseDay, _ = strconv.Atoi(entry.ReleaseDay)
					newRelease.ReleaseType = entry.ReleaseType
					newRelease.ReleasedAt = entry.XMLReleasedAt.XMLEvent.Name

					if len(entry.UsedSIDs) == 1 {
						newRelease.SIDPath = entry.UsedSIDs[0].HVSCPath
					}

					// log.Println("Nazwa:  ", entry.ReleaseName)
					// log.Println("ID:     ", entry.ReleaseID)
					// log.Println("Typ:    ", entry.ReleaseType)
					// log.Println("Event:  ", entry.XMLReleasedAt.XMLEvent.Name)

					for _, group := range entry.XMLReleasedBy.XMLGroup {
						// log.Println("XMLGroup:  ", group.Name)
						newRelease.ReleasedBy = append(newRelease.ReleasedBy, group.Name)
					}
					for _, handle := range entry.XMLReleasedBy.XMLHandle {
						// log.Println("XMLHandle: ", handle.XMLHandle)
						newRelease.ReleasedBy = append(newRelease.ReleasedBy, handle.XMLHandle)
					}
					// for _, entrySid := range entry.UsedSIDs {
					// 	var sid UsedSID
					// 	sid.Author = entrySid.Author
					// 	sid.HVSCPath = entrySid.HVSCPath
					// 	sid.ID = entrySid.ID
					// 	sid.Name = entrySid.Name
					// 	newRelease.UsedSIDs = append(newRelease.UsedSIDs, sid)
					// }
					// log.Println("-----------------------------------")
					for _, credit := range entry.Credits {

						creditHandle := "???"
						if len(credit.XMLHandle.XMLHandle) > 0 {
							// log.Println(credit.CreditType + ": " + credit.XMLHandle.XMLHandle + " [" + credit.XMLHandle.ID + "]")
							if credit.CreditType == "Music" {
								newRelease.Credits = append(newRelease.Credits, credit.XMLHandle.XMLHandle)
							}
						} else {
							found := false
							for _, releaseHandle := range entry.XMLReleasedBy.XMLHandle {
								if releaseHandle.ID == credit.XMLHandle.ID && releaseHandle.XMLHandle != "" {
									// log.Println(credit.CreditType + ": " + releaseHandle.XMLHandle + " [" + releaseHandle.ID + "]")
									creditHandle = releaseHandle.XMLHandle
									found = true
									break
								}
							}
							if !found {
								for _, releaseHandle := range entry.Credits {
									if releaseHandle.XMLHandle.ID == credit.XMLHandle.ID && releaseHandle.XMLHandle.XMLHandle != "" {
										// log.Println(credit.CreditType + ": " + releaseHandle.XMLHandle.XMLHandle + " [" + releaseHandle.XMLHandle.ID + "]")
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
					// log.Println("===================================")

					// Linki dościągnięcia
					// Najpierw SIDy

					for _, link := range entry.DownloadLinks {
						if strings.Contains(link.Link, ".sid") || strings.Contains(link.Link, ".SID") {
							newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
						}
					}
					// Potem PRGs
					for _, link := range entry.DownloadLinks {
						if strings.Contains(link.Link, ".prg") || strings.Contains(link.Link, ".PRG") {
							newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
						}
					}
					// Potem ZIPy
					for _, link := range entry.DownloadLinks {
						if strings.Contains(link.Link, ".zip") || strings.Contains(link.Link, ".ZIP") {
							newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
						}
					}
					// Potem D64y
					for _, link := range entry.DownloadLinks {
						if strings.Contains(link.Link, ".d64") || strings.Contains(link.Link, ".D64") {
							newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
						}
					}

					//
					// Dodajemy
					//
					if len(newRelease.DownloadLinks) > 0 {
						releasesTemp = append(releasesTemp, newRelease)
					}
				}
			} else {
				log.Println("[ReadLatestReleases] Błąd komunikacji z csdb.dk")
			}
		}

		//
		// Dodanie do globalnej tablicy
		//

		for _, rel1 := range releasesTemp {

			found := false
			for index, rel2 := range releases {
				if rel2.ReleaseID == rel1.ReleaseID {
					found = true
					updateReleaseInfo(index, rel1)
					break
				}
			}

			if !found {
				releases = append(releases, rel1)
				foundNewReleases++
			}
		}

		if csdbDataReady {
			for _, rel1 := range csdb {

				found := false
				for index, rel2 := range releases {
					if rel2.ReleaseID == rel1.ReleaseID {
						found = true
						updateReleaseInfo(index, rel1)
						break
					}
				}

				if !found {
					releases = append(releases, rel1)
					foundNewReleases++
				}
			}
		}

		// Wyświetlenie danych
		log.Println("[ReadLatestReleases] Found " + strconv.Itoa(foundNewReleases) + " new releases")

		sort.Sort(byID(releases))
		sort.Sort(byDate(releases))

		if len(releases) > historyMaxEntries {
			releases = releases[0:historyMaxEntries]
		}

		WriteDb()

	} else {
		log.Println("[ReadLatestReleases] Błąd komunikacji z csdb.dk")
	}

}

// CSDBPrepareData - Wątek odczygtujący wszystkie releasy z csdb
// ================================================================================================
func CSDBPrepareData() {

	lastDate := time.Now().AddDate(0, -historyMaxMonths, 0)

	netClient := &http.Client{Timeout: time.Second * 10}

	resp, err := netClient.Get("https://csdb.dk/webservice/?type=release&id=0")

	log.Println("[CSDBPrepareData] LOOP")

	csdbDataReady = false

	if ErrCheck(err) {

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		ErrCheck(err)
		// log.Println(string(body))
		resp.Body.Close()

		// Przerobienie na strukturę

		var entry LatestRelease
		reader := bytes.NewReader(body)
		decoder := xml.NewDecoder(reader)
		decoder.CharsetReader = makeCharsetReader
		err = decoder.Decode(&entry)
		ErrCheck(err)

		// log.Println("===================================")
		log.Println("[CSDBPrepareData] Najwyższy numer ID wynosi " + strconv.Itoa(entry.ID))

		var csdbTemp []Release

		foundNewReleases := 0
		id := entry.ID

		for foundNewReleases < historyMaxEntries {

			resp, err := netClient.Get("https://csdb.dk/webservice/?type=release&id=" + strconv.Itoa(id))

			// log.Println("ID " + strconv.Itoa(id))

			id--

			if ErrCheck(err) {
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				defer resp.Body.Close()

				if ErrCheck(err) {
					resp.Body.Close()
					// log.Println(string(body))

					// Przerobienie na strukturę

					var entry XMLRelease
					reader := bytes.NewReader(body)
					decoder := xml.NewDecoder(reader)
					decoder.CharsetReader = makeCharsetReader
					err = decoder.Decode(&entry)
					if ErrCheck2(err) {

						// Szukamy takiego release w naszej bazie
						//

						var relTypesAllowed = [...]string{"C64 Music", "C64 Demo", "C64 One-File Demo", "C64 Intro", "C64 4K Intro", "C64 Crack intro", "C64 Music Collection", "C64 Graphics Collection", "C64 Diskmag", "C64 Charts", "C64 Invitation", "C64 1K Intro", "C64 Fake Demo"}
						typeOK := false
						for _, relType := range relTypesAllowed {
							if relType == entry.ReleaseType {
								typeOK = true
								break
							}
						}

						prodYear, _ := strconv.Atoi(entry.ReleaseYear)
						prodMonth, _ := strconv.Atoi(entry.ReleaseMonth)
						prodDay, _ := strconv.Atoi(entry.ReleaseDay)
						prodTime := time.Date(prodYear, time.Month(prodMonth), prodDay, 0, 0, 0, 0, time.Local)

						// TODO zrobić update tych info (ktoś mógł uzupełnić potem dane lub pliki)
						// Jeżeli znaleźliśmy to sprawdzamy typ i dodajemy
						//
						if typeOK && prodTime.After(lastDate) {

							// Tworzymy nowy obiekt release który dodamy do slice
							//
							var newRelease Release
							id, _ := strconv.Atoi(entry.ReleaseID)
							newRelease.ReleaseID = id
							newRelease.ReleaseName = entry.ReleaseName
							newRelease.ReleaseScreenShot = entry.ReleaseScreenShot
							newRelease.Rating = entry.Rating
							newRelease.ReleaseYear, _ = strconv.Atoi(entry.ReleaseYear)
							newRelease.ReleaseMonth, _ = strconv.Atoi(entry.ReleaseMonth)
							newRelease.ReleaseDay, _ = strconv.Atoi(entry.ReleaseDay)
							newRelease.ReleaseType = entry.ReleaseType
							newRelease.ReleasedAt = entry.XMLReleasedAt.XMLEvent.Name

							if len(entry.UsedSIDs) == 1 {
								newRelease.SIDPath = entry.UsedSIDs[0].HVSCPath
							}

							// log.Println("[CSDBPrepareData] Entry name: " + entry.ReleaseName)
							// log.Println("ID:     ", entry.ReleaseID)
							// log.Println("Typ:    ", entry.ReleaseType)
							// log.Println("Event:  ", entry.XMLReleasedAt.XMLEvent.Name)

							for _, group := range entry.XMLReleasedBy.XMLGroup {
								// log.Println("XMLGroup:  ", group.Name)
								newRelease.ReleasedBy = append(newRelease.ReleasedBy, group.Name)
							}
							for _, handle := range entry.XMLReleasedBy.XMLHandle {
								// log.Println("XMLHandle: ", handle.XMLHandle)
								newRelease.ReleasedBy = append(newRelease.ReleasedBy, handle.XMLHandle)
							}
							// for _, entrySid := range entry.UsedSIDs {
							// 	var sid UsedSID
							// 	sid.Author = entrySid.Author
							// 	sid.HVSCPath = entrySid.HVSCPath
							// 	sid.ID = entrySid.ID
							// 	sid.Name = entrySid.Name
							// 	newRelease.UsedSIDs = append(newRelease.UsedSIDs, sid)
							// }
							// log.Println("-----------------------------------")
							for _, credit := range entry.Credits {

								creditHandle := "???"
								if len(credit.XMLHandle.XMLHandle) > 0 {
									// log.Println(credit.CreditType + ": " + credit.XMLHandle.XMLHandle + " [" + credit.XMLHandle.ID + "]")
									if credit.CreditType == "Music" {
										newRelease.Credits = append(newRelease.Credits, credit.XMLHandle.XMLHandle)
									}
								} else {
									found := false
									for _, releaseHandle := range entry.XMLReleasedBy.XMLHandle {
										if releaseHandle.ID == credit.XMLHandle.ID && releaseHandle.XMLHandle != "" {
											// log.Println(credit.CreditType + ": " + releaseHandle.XMLHandle + " [" + releaseHandle.ID + "]")
											creditHandle = releaseHandle.XMLHandle
											found = true
											break
										}
									}
									if !found {
										for _, releaseHandle := range entry.Credits {
											if releaseHandle.XMLHandle.ID == credit.XMLHandle.ID && releaseHandle.XMLHandle.XMLHandle != "" {
												// log.Println(credit.CreditType + ": " + releaseHandle.XMLHandle.XMLHandle + " [" + releaseHandle.XMLHandle.ID + "]")
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
							// log.Println("===================================")

							// Linki dościągnięcia
							// Najpierw SIDy

							for _, link := range entry.DownloadLinks {
								if strings.Contains(link.Link, ".sid") || strings.Contains(link.Link, ".SID") {
									newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
								}
							}
							// Potem PRGs
							for _, link := range entry.DownloadLinks {
								if strings.Contains(link.Link, ".prg") || strings.Contains(link.Link, ".PRG") {
									newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
								}
							}
							// Potem ZIPy
							for _, link := range entry.DownloadLinks {
								if strings.Contains(link.Link, ".zip") || strings.Contains(link.Link, ".ZIP") {
									newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
								}
							}
							// Potem D64y
							for _, link := range entry.DownloadLinks {
								if strings.Contains(link.Link, ".d64") || strings.Contains(link.Link, ".D64") {
									newRelease.DownloadLinks = append(newRelease.DownloadLinks, link.Link)
								}
							}

							//
							// Dodajemy
							//
							if len(newRelease.DownloadLinks) > 0 {
								csdbTemp = append(csdbTemp, newRelease)
								foundNewReleases++
								log.Println("[CSDBPrepareData] " + strconv.Itoa(foundNewReleases) + ") Entry name: " + entry.ReleaseName + ", Entry ID: " + entry.ReleaseID)
							}
						}
					}
				} else {
					log.Println("[CSDBPrepareData] Błąd komunikacji z csdb.dk")
					break
				}
			} else {
				log.Println("[CSDBPrepareData] Błąd komunikacji z csdb.dk")
				break
			}

		}
		// sort.Sort(byID(csdbTemp))
		// sort.Sort(byDate(csdbTemp))
		csdb = csdbTemp
		WriteCSDb()

		log.Println("[CSDBPrepareData] Finish")

		csdbDataReady = true

		// log.Println("[CSDBPrepareData] Amount of " + strconv.Itoa(len(csdb)) + " releases from last " + strconv.Itoa(historyMaxMonths) + " month(s)")
		// log.Println("[CSDBPrepareData] Amount of " + strconv.Itoa(len(csdb)) + " releases from last " + strconv.Itoa(historyMaxMonths) + " month(s)")

	} else {
		log.Println("[CSDBPrepareData] Błąd komunikacji z csdb.dk")
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

// CSDBGetLatestReleases - ostatnie release'y
// ================================================================================================
func CSDBGetLatestReleases(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	// Info o wejściu do GET
	log.Println("[GIN:CSDBGetLatestReleases]")

	releasesTemp := releases

	c.JSON(http.StatusOK, releasesTemp)
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

	if fileExists(filenameWAV) {

		// Info o wejściu do GET
		log.Println("[GIN:AudioGet] " + id)

		// const maxOffset int64 = 50000000 // ~ 10 min
		// const maxOffset int64 = 25000000 // ~ 5 min
		// const maxOffset int64 = 5000000 // ~ 1 min
		const maxOffset int64 = 44100 * 2 * 300 // = 5 min
		const maxVol float64 = 1.25

		var vol float64 = maxVol
		loop := true
		volDown := false

		// Przygotowanie bufora do streamingu
		const bufferSize = 1024 * 64
		var offset int64
		p := make([]byte, bufferSize)

		log.Println("[GIN:AudioGet] Sending " + id + "...")

		// Streaming LOOP...
		// ----------------------------------------------------------------------------------------------

		var sum float64
		var dataSent int64 = 0
		silenceCnt := 0

		for loop {

			sum = 0

			// Jeżeli doszliśmy w pliku do 50MB to koniec
			if dataSent > maxOffset {

				// log.Println("Wyciszamy...")
				// break
				volDown = true
			}

			// Jeżeli stracimy kontekst to wychodzimy
			if c.Request.Context() == nil {
				log.Println("[GIN:AudioGet] ERR! c.Request.Context() == nil")
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
									vol = maxVol - (float64(dataSent-maxOffset+int64(ix)) / 88.494 * 0.0002)
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
							dataSent += int64(len(p))
							// log.Print(".")
						}

						if sum < 5.0 && offset > 44100*60 { // po 30 sekundach
							silenceCnt++
							if silenceCnt >= 5 {
								log.Println("[GIN:AudioGet] Silence at " + strconv.FormatInt(offset, 10))
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
	} else {
		log.Println("[GIN:AudioGet] WAV file doesn't exists")
	}
	// }

	// Feedback gdybyśmy wyszli z LOOP
	c.JSON(http.StatusOK, "[GIN:AudioGet] Loop ended")
	log.Println("[GIN:AudioGet] Loop ended")
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

// smtpServer - smtpServer data to smtp server
// ================================================================================================
type smtpServer struct {
	host string
	port string
}

// Address - URI to smtp server
// ================================================================================================
func (s *smtpServer) Address() string {
	return s.host + ":" + s.port
}

// SendEmail - decode reader
// ================================================================================================
func SendEmail(in string) {
	// Sender data.
	from := "sidcloud.net@gmail.com"
	password := "SidCloud1024!"

	// Receiver email address.
	to := []string{
		"b.apanasewicz@gmail.com",
		// "secondemail@gmail.com",
	}

	// smtp server configuration.
	smtpServer := smtpServer{host: "smtp.gmail.com", port: "587"}

	// Message.
	message := []byte(in)
	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpServer.host)
	// Sending email.
	err := smtp.SendMail(smtpServer.Address(), auth, from, to, message)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Email Sent!")
}

// CSDBWebServices - Obsługa CSDB Web Services w osobnym wątku
// ================================================================================================
func CSDBWebServices() {

	// const CSDBPrepareDataCycle time.Duration = 60 * 60 * 24
	const CSDBWebServicesCycle time.Duration = 60 * 5

	// time1 := time.Now()
	// firstRun := true

	go CSDBPrepareData()

	for {
		// time0 := time.Now()

		// if (time0.Sub(time1)) > (CSDBPrepareDataCycle*time.Second) || firstRun {
		// if firstRun {
		// 	// time1 = time.Now()
		// 	firstRun = false
		// }

		ReadLatestReleases()
		DownloadFiles()
		CreateWAVFiles()

		time.Sleep(CSDBWebServicesCycle * time.Second)
	}
}

// ================================================================================================
// MAIN()
// ================================================================================================
func main() {
	//
	// Spraedzamy argumenty
	//
	args := os.Args[1:]

	if len(args) == 0 || (args[0] != "http" && args[0] != "https") {
		log.Fatal("Podaj parametr 'http' lub 'https'!")
		os.Exit(1)
	}

	//
	// Nazwa programu SIDPLAYFP zależna od OS
	//
	sidplayExe = "sidplayfp/sidplayfp"

	if args[0] == "http" {
		if len(args) > 2 {
			if args[2] == "arm" {
				sidplayExe = "sidplayfp"
			}
		}
	}

	//
	// Logowanie do pliku
	//
	logFileApp, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	ErrCheck(err)
	log.SetOutput(io.MultiWriter(os.Stdout, logFileApp))

	gin.DisableConsoleColor()
	logFileGin, err := os.OpenFile("gin.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	ErrCheck(err)
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFileGin)

	// Info powitalne
	//
	log.Println("==========================================")
	log.Println("=======          APP START        ========")
	log.Println("==========================================")

	//
	// Odczyt JSONów
	//
	ReadCSDb()
	ReadDb()

	//
	// Uruchomienie wątków
	//
	go CSDBWebServices()

	//
	// Tryb serwera, dla https tryb Rel, dla http tryb Dev
	//
	if args[0] == "https" {
		gin.SetMode(gin.ReleaseMode)
	}

	//
	// Konfiguracja serwera
	//
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

	//
	// Start serwera
	//
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
