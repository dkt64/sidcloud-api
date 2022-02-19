// ================================================================================================
// Sidcloud by DKT/Samar
// ================================================================================================

package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

var csdbDataReady bool = false

const cacheDir = "cache/"

const wavTime = "330"

const historyMaxEntries = 80

const historyMaxMonths = 3

const defaultBufferSize = 1024 * 32

const wavHeaderSize = 44 // rozmiar nagłówka WAV

const wavTime5minutes int64 = 44100 * 2 * 300 // = 5 min

const wavTime10seconds int64 = 44100 * 2 * 10 // = 10 sekund

const wavTime5seconds int64 = 44100 * 2 * 5 // = 5 sekund

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
	SrcCached         bool
	WAVCached         bool
	SrcExt            string
	Disabled          bool
	// UsedSIDs          []UsedSID
}

// releases - glówna i globalna tablica z aktualnymi produkcjami
// ================================================================================================
var releases []Release

// allReleases - glówna i globalna tablica ze wszystkimi produkcjami
// ================================================================================================
var csdb []Release

// allReleases - glówna i globalna tablica ze wszystkimi produkcjami
// ================================================================================================
var allReleases []Release

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
	foundPRG := false
	found0801 := false

	for loop {
		log.Println("[ExtractD64] Dir track " + strconv.Itoa(int(dirTrack)) + " and sector " + strconv.Itoa(int(dirSector)))
		sector := D64GetSector(file, dirTrack, dirSector)
		for ptr := 0; ptr < 8*0x20; ptr += 0x20 {
			if (sector[ptr+2] & 7) == 2 {

				foundPRG = true
				name := string(sector[ptr+5 : ptr+14])
				var fileTrack byte = sector[ptr+3]
				var fileSector byte = sector[ptr+4]

				// Najpierw sprawdzimy czy load address == $0801
				log.Println("[ExtractD64] Reading " + name + " TRACK:" + strconv.Itoa(int(fileTrack)) + " SECTOR:" + strconv.Itoa(int(fileSector)))
				prg := D64GetSector(file, fileTrack, fileSector)

				if (prg[2] == 1 || prg[2] == 0) && prg[3] == 8 {
					log.Println("[ExtractD64] Loading address is OK")
					found0801 = true
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

	log.Println("[ExtractD64] Trying without load address check...", loop, foundPRG)
	dirTrack = 18
	dirSector = 1

	// druga pętla bez restrykcji, pierwszy lepszy PRG bez $0801
	for loop && foundPRG && !found0801 {
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

				log.Println("[ExtractD64] Loading address different, but OK")
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

// Sortowanie datami i ID
// ================================================================================================

type byDateAndID []Release

func (s byDateAndID) Len() int {
	return len(s)
}
func (s byDateAndID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byDateAndID) Less(i, j int) bool {

	d1 := time.Date(s[i].ReleaseYear, time.Month(s[i].ReleaseMonth), s[i].ReleaseDay, 0, 0, 0, 0, time.Local)
	d2 := time.Date(s[j].ReleaseYear, time.Month(s[j].ReleaseMonth), s[j].ReleaseDay, 0, 0, 0, 0, time.Local)
	id1 := s[i].ReleaseID
	id2 := s[j].ReleaseID

	return d2.Before(d1) && id1 > id2
}

// fileExists - sprawdzenie czy plik istnieje
// ================================================================================================
func fileExists(filename string) bool {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}

	return true
}

// fileExists - sprawdzenie czy plik istnieje
// ================================================================================================
func fileSize(filename string) (int64, error) {
	// Sprawdzamy rozmiar pliku
	fileStat, err := os.Stat(filename)
	if ErrCheck(err) {
		return fileStat.Size(), err
	}
	return fileStat.Size(), err
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
		zipReader, err := zip.OpenReader(filepath)

		if ErrCheck(err) {
			//
			// Najpierw SIDy
			//
			for _, file := range zipReader.File {

				// log.Println(file.Name)

				if (strings.Contains(file.Name, ".sid") || strings.Contains(file.Name, ".SID")) && !file.FileInfo().IsDir() {

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
					if ErrCheck(err) {
						defer zippedFile.Close()

						_, err = io.Copy(outputFile, zippedFile)
						ErrCheck(err)
					}
					return ext, err
				}
			}

			//
			// Potem PRG
			//
			for _, file := range zipReader.File {

				// log.Println(file.Name)

				if (strings.Contains(file.Name, ".prg") || strings.Contains(file.Name, ".PRG")) && !file.FileInfo().IsDir() {

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
						if ErrCheck(err) {
							defer zippedFile.Close()

							_, err = io.Copy(outputFile, zippedFile)
							ErrCheck(err)

						}
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

				if (strings.Contains(file.Name, ".d64") || strings.Contains(file.Name, ".D64")) && !file.FileInfo().IsDir() {

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
					if ErrCheck(err) {
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
		}
	}

	return ext, err
}

// DownloadFiles - Wątek ściągający pliki produkcji z csdb
// ================================================================================================
func DownloadFiles() {

	log.Println("[DownloadFiles] LOOP")

	for index, rel := range releases {
		if len(rel.DownloadLinks) > 0 {
			filename := cacheDir
			foundSID := false

			var downloadLinkIndex int

			for i, dl := range rel.DownloadLinks {
				if strings.Contains(dl, ".sid") || strings.Contains(dl, ".SID") {
					filename += strconv.Itoa(rel.ReleaseID) + ".sid"
					foundSID = true
					downloadLinkIndex = i
					break
				}
			}
			if !foundSID {
				for i, dl := range rel.DownloadLinks {
					if strings.Contains(dl, ".prg") || strings.Contains(dl, ".PRG") {
						filename += strconv.Itoa(rel.ReleaseID) + ".prg"
						foundSID = true
						downloadLinkIndex = i
						break
					}
				}
			}
			if !foundSID {
				for i, dl := range rel.DownloadLinks {
					if strings.Contains(dl, ".zip") || strings.Contains(dl, ".ZIP") {
						filename += strconv.Itoa(rel.ReleaseID) + ".zip"
						foundSID = true
						downloadLinkIndex = i
						break
					}
				}
			}
			if !foundSID {
				for i, dl := range rel.DownloadLinks {
					if strings.Contains(dl, ".d64") || strings.Contains(dl, ".D64") {
						filename += strconv.Itoa(rel.ReleaseID) + ".d64"
						foundSID = true
						downloadLinkIndex = i
						break
					}
				}
			}

			// Dodajemy new release
			// ale tylko jeżeli mamy niezbędne info o produkcji
			if filename != "" {
				// if !rel.SrcCached || !(fileExists(cacheDir+strconv.Itoa(rel.ReleaseID)+".sid") || fileExists(cacheDir+strconv.Itoa(rel.ReleaseID)+".prg")) {
				if !rel.SrcCached {
					_, err := DownloadFile(filename, rel.DownloadLinks[downloadLinkIndex], rel.ReleaseID)

					if ErrCheck(err) {
						// Sprawdzay czy istnieje SID lub PRG
						if fileExists(cacheDir+strconv.Itoa(rel.ReleaseID)+".sid") || fileExists(cacheDir+strconv.Itoa(rel.ReleaseID)+".prg") {
							rel.SrcCached = true
							log.Println("[DownloadFiles] File cached")
						} else {
							log.Println("[DownloadFiles] File not cached")
							rel.SrcCached = false
							rel.WAVCached = false
						}
						// SendEmail("Nowa produkcja na CSDB.DK: " + rel.ReleaseName + " by " + rel.ReleasedBy[0])
					}

					if fileExists(cacheDir + strconv.Itoa(rel.ReleaseID) + ".prg") {
						rel.SrcExt = ".prg"
					}
					if fileExists(cacheDir + strconv.Itoa(rel.ReleaseID) + ".sid") {
						rel.SrcExt = ".sid"
					}

					releases[index] = rel

					WriteDb()
				}
			}
		}
	}
}

// WAVPrepare - Usuwa puste miejsca w pliku WAV, wycisza, wzmacnia
// ================================================================================================
func WAVPrepare(filename string, r Release) error {

	var size int64
	if fileExists(filename) {
		file, err := os.Stat(filename)
		if ErrCheck(err) {
			size = file.Size()
		} else {
			log.Println("[WAVPrepare] Problem z odczytem rozmiaru pliku " + filename)
			return err
		}
	}

	file, err := os.Open(filename)
	defer file.Close()

	if ErrCheck(err) {
		p := make([]byte, size)
		readed, err := file.ReadAt(p, 0)
		if (int64(readed) == size) && ErrCheck(err) {

			file.Close()
			os.Remove(filename)

			// Wycinamy początkowe śmieci
			p = append(p[:wavHeaderSize], p[0x2000+wavHeaderSize:]...)

			sil := 0

			if r.SrcExt == ".prg" {
				var i int
				for i = wavHeaderSize; i < (len(p) - 2); i = i + 2 {
					if (p[i] < 0xFA && p[i] > 5) || p[i+1] != 0 {
						// Wycinamy początkową ciszę
						log.Println("[WAVPrepare] Wycinam początkową ciszę do " + strconv.Itoa(i))
						p = append(p[:wavHeaderSize], p[i:]...)
						break
					}
				}

				const silTime5seconds = 44100 * 5 // = 5 sekund

				for i = int(wavTime10seconds) + wavHeaderSize; i < (len(p) - 2); i = i + 2 {
					if (p[i] >= 0xFA && p[i+1] == 0xFF) || (p[i] <= 5 && p[i+1] == 0) {
						sil++
						// log.Println("[WAVPrepare] Found zeroes at " + strconv.Itoa(i))
						if sil > silTime5seconds {
							// Wycinamy końcową ciszę
							log.Println("[WAVPrepare] Wycinam końcową ciszę od " + strconv.Itoa(i))
							p = append(p[:i], p[len(p):]...)
							break
						}
					} else {
						sil = 0
					}
				}

				// Przycięcie do max 5 minut

				if len(p) > int(wavTime5minutes+wavHeaderSize) {
					p = append(p[:wavTime5minutes+wavHeaderSize], p[len(p):]...)
				}

			}

			// Wzmocnienie i wyciszenie

			const maxVol float64 = 1.25
			var vol float64 = maxVol

			for ix := wavHeaderSize; ix < len(p); ix = ix + 2 {

				// Wyciszenie (tylko gdy nie było cięcia)
				if sil == 0 && ix > len(p)-int(wavTime5seconds) {
					vol = maxVol * (float64(len(p)-ix) / float64(wavTime5seconds))
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

			// Zmieniamy rozmiar chunks

			ChunkSize := len(p) - 8
			DataSize := len(p) - wavHeaderSize

			binary.LittleEndian.PutUint32(p[4:], uint32(ChunkSize))
			binary.LittleEndian.PutUint32(p[40:], uint32(DataSize))

			file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
			written, err := file.Write(p)
			defer file.Close()

			if ErrCheck(err) {
				log.Println("[WAVPrepare] Zapisałem plik " + filename + " o nowym rozmiarze " + strconv.Itoa(written))
				return nil
			}
			return err
		}
	} else {
		return err
	}

	return err
}

// CreateWAVFiles - Creating WAV files
// ================================================================================================
func CreateWAVFiles() {

	log.Println("[CreateWAVFiles] LOOP")
	for index, rel := range releases {

		if len(rel.SrcExt) == 4 && rel.SrcCached && !rel.Disabled {
			id := strconv.Itoa(rel.ReleaseID)
			filenameWAV := cacheDir + id + ".wav"

			// var size int64
			// if fileExists(filenameWAV) {
			// 	file, err := os.Stat(filenameWAV)
			// 	if err != nil {
			// 		log.Println("[CreateWAVFiles] Problem z odczytem rozmiaru pliku " + filenameWAV)
			// 	}
			// 	size = file.Size()
			// }

			// if !fileExists(filenameWAV) || size < wavSize || (fileExists(filenameWAV) && !rel.WAVCached) {
			if !fileExists(filenameWAV) || (fileExists(filenameWAV) && !rel.WAVCached) {

				log.Println("[CreateWAVFiles] Creating file " + filenameWAV)
				filenameSID := cacheDir + id + rel.SrcExt
				paramName := "-w" + cacheDir + id

				var cmdName string

				czas := "-t" + wavTime
				// bits := "-p16"
				// freq := "-fwavHeaderSize100"
				model := "-mn"

				// Odpalenie sidplayfp
				if runtime.GOOS == "windows" {
					cmdName = "sidplayfp/sidplayfp.exe"
				} else {
					cmdName = sidplayExe // zakładamy że jest zainstalowany
				}

				additionalSIDs := ""
				if rel.SrcExt == ".prg" {
					// drugi i trzeci sid gdy mamy tylko PRG
					additionalSIDs = "-ds0xd420 -ts0xd440"
				}

				// stereo
				stereo := "-s"

				var cmd exec.Cmd

				if rel.SrcExt == ".prg" {
					log.Println("[CreateWAVFiles PRG] Starting sidplayfp... cmdName(" + cmdName + " " + czas + " " + additionalSIDs + " " + stereo + " " + model + " " + paramName + " " + filenameSID + ")")
					cmd = *exec.Command(cmdName, czas, additionalSIDs, stereo, model, paramName, filenameSID)
				} else {
					log.Println("[CreateWAVFiles SID] Starting sidplayfp... cmdName(" + cmdName + " " + czas + " " + stereo + " " + model + " " + paramName + " " + filenameSID + ")")
					cmd = *exec.Command(cmdName, czas, stereo, model, paramName, filenameSID)
				}

				errStart := cmd.Start()

				if ErrCheck(errStart) {

					done := make(chan error)
					go func() { done <- cmd.Wait() }()
					select {
					case err := <-done:
						// exited

						// Jeszcze raz sprawdzamy czy plik powstał o odpowiedniej długości
						// if fileExists(filenameWAV) {
						// 	file, err := os.Stat(filenameWAV)
						// 	if err != nil {
						// 		log.Println("[CreateWAVFiles] Problem z odczytem rozmiaru pliku " + filenameWAV)
						// 	}
						// 	size = file.Size()
						// }

						// if fileExists(filenameWAV) && size >= wavSize {
						if fileExists(filenameWAV) && ErrCheck(err) {
							WAVPrepare(filenameWAV, rel)
							releases[index].WAVCached = true
							log.Println("[CreateWAVFiles] " + filenameWAV + " cached")
							WriteDb()
						} else {
							log.Println("[CreateWAVFiles] " + filenameWAV + " not cached")
						}

					case <-time.After(10 * time.Minute):
						// timed out
						log.Println("[CreateWAVFiles] Problem with sidplayfp and " + filenameWAV)
						os.Remove(filenameWAV)
						releases[index].Disabled = true
						cmd.Process.Kill()
						WriteDb()
					}
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
	releases[index].ReleaseType = newRelease.ReleaseType
	releases[index].Rating = newRelease.Rating
	releases[index].ReleaseScreenShot = newRelease.ReleaseScreenShot
	releases[index].ReleasedAt = newRelease.ReleasedAt
	releases[index].ReleasedBy = newRelease.ReleasedBy
	releases[index].ReleaseDay = newRelease.ReleaseDay
	releases[index].ReleaseMonth = newRelease.ReleaseMonth
	releases[index].ReleaseYear = newRelease.ReleaseYear

	if newRelease.DownloadLinks != nil {
		if releases[index].DownloadLinks == nil {
			for _, link := range newRelease.DownloadLinks {
				releases[index].DownloadLinks = append(releases[index].DownloadLinks, link)
			}
			// } else if len(releases[index].DownloadLinks) < len(newRelease.DownloadLinks) && releases[index].DownloadLinks[0] != ".sid" && newRelease.DownloadLinks[0] == ".sid" {
		} else if len(releases[index].DownloadLinks) < len(newRelease.DownloadLinks) {
			releases[index].DownloadLinks = newRelease.DownloadLinks
			releases[index].SrcCached = false
			releases[index].WAVCached = false
		}
	}
	if !fileExists(cacheDir + strconv.Itoa(releases[index].ReleaseID) + ".wav") {
		releases[index].WAVCached = false
	}

	// Jeżeli pojawił się SID a wcześniej był PRG to trzeba to przetworzyć
	var sidFilePresent bool
	sidFilePresent = false
	for _, link := range newRelease.DownloadLinks {
		if strings.Contains(link, ".sid") || strings.Contains(link, ".SID") {
			sidFilePresent = true
			break
		}
	}

	if (releases[index].SrcExt == ".prg" || releases[index].SrcExt == ".PRG") && sidFilePresent {
		releases[index].SrcCached = false
		releases[index].WAVCached = false
		releases[index].SrcExt = ""
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

				var relTypesAllowed = [...]string{"C64 Music", "C64 Graphics", "C64 Demo", "C64 One-File Demo", "C64 Intro", "C64 4K Intro", "C64 Crack Intro", "C64 Music Collection", "C64 Graphics Collection", "C64 Diskmag", "C64 Charts", "C64 Invitation", "C64 1K Intro", "C64 Fake Demo"}
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
					if len(newRelease.DownloadLinks) > 0 && len(newRelease.Credits) > 0 {
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
					// log.Println("[ReadLatestReleases] Update of", rel1.ReleaseID)
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
						// log.Println("[ReadLatestReleases] Update of", rel1.ReleaseID)
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

		if foundNewReleases > 0 {
			sort.Sort(byID(releases))
		}
		// sort.Sort(byID(releases))
		// sort.Sort(byDate(releases))

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

						var relTypesAllowed = [...]string{"C64 Music", "C64 Graphics", "C64 Demo", "C64 One-File Demo", "C64 Intro", "C64 4K Intro", "C64 Crack intro", "C64 Music Collection", "C64 Graphics Collection", "C64 Diskmag", "C64 Charts", "C64 Invitation", "C64 1K Intro", "C64 Fake Demo"}
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
							if len(newRelease.DownloadLinks) > 0 && len(newRelease.Credits) > 0 {
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

	ReadDb()

	releasesTemp := releases

	c.JSON(http.StatusOK, releasesTemp)
}

// CSDBGetRelease - jeden release
// ================================================================================================
func CSDBGetRelease(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	// Info o wejściu do GET
	log.Println("[GIN:CSDBGetRelease]")

	ReadDb()

	id := c.Param("id")

	nr, err := strconv.Atoi(id)
	if ErrCheck(err) {
		found := false
		if nr > 0 {
			for _, rel := range releases {
				if nr == rel.ReleaseID {
					var releasesTemp []Release
					releasesTemp = append(releasesTemp, rel)
					c.JSON(http.StatusOK, releasesTemp)
					found = true
					break
				}
			}
		}
		if !found {
			releasesTemp := releases
			c.JSON(http.StatusOK, releasesTemp)

		}
	} else {
		releasesTemp := releases
		c.JSON(http.StatusOK, releasesTemp)
	}

}

// AudioGetNew - granie utworu po nowemu
// https://stackoverflow.com/questions/61453199/html-audio-stream-dont-play-on-apples-ios-safari-and-iphone-https-sidclo
// ================================================================================================
func AudioGetNew(c *gin.Context) {
	id := c.Param("id")
	filenameWAV := cacheDir + id + ".wav"

	var size int64

	if fileExists(filenameWAV) {
		var err error
		size, err = fileSize(filenameWAV)
		if ErrCheck(err) {
			log.Println("[GIN:AudioGet] Size of file " + filenameWAV + " = " + strconv.Itoa(int(size)))
		} else {
			log.Println("[GIN:AudioGet] Can't read size of file " + filenameWAV)
			c.JSON(http.StatusInternalServerError, "Can't read size of file")
			return
		}
	} else {
		log.Println("[GIN:AudioGet] No WAV file " + filenameWAV)
		c.JSON(http.StatusInternalServerError, "No WAV file")
		return
	}

	if size > 0 {
		log.Println("[GIN:AudioGet] Sending " + id + " ...")
		http.ServeFile(c.Writer, c.Request, filenameWAV) // assuming filenameWAV is the location
		log.Println("[GIN:AudioGet] Sending " + id + " end.")
	}
}

// AudioGet - granie utworu
// ================================================================================================
func AudioGet(c *gin.Context) {

	// Typ połączania
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Connection", "Keep-Alive")
	c.Header("Transfer-Encoding", "identity")
	c.Header("Accept-Ranges", "bytes")

	// Odczytujemy parametr - numer muzy
	id := c.Param("id")
	filenameWAV := cacheDir + id + ".wav"

	var size int64
	if fileExists(filenameWAV) {
		s, err := fileSize(filenameWAV)
		if ErrCheck(err) {
			log.Println("[GIN:AudioGet] Size of file " + filenameWAV + " = " + strconv.Itoa(int(s)))
			size = s
		} else {
			log.Println("[GIN:AudioGet] Can't read size of file " + filenameWAV)
			c.JSON(http.StatusInternalServerError, "Can't read size of file")
			return
		}
	} else {
		log.Println("[GIN:AudioGet] No WAV file " + filenameWAV)
		c.JSON(http.StatusInternalServerError, "No WAV file")
		return
	}

	//
	// Analiza nagłówka - ile bajtów mamy wysłać
	//
	bytesToSend := 0
	bytesToSendStart := 0
	bytesToSendEnd := 0
	headerRange := c.GetHeader("Range")
	log.Println("[GIN:AudioGet] Header:Range = " + headerRange)
	if len(headerRange) > 0 {
		headerRangeSplitted1 := strings.Split(headerRange, "=")

		if len(headerRangeSplitted1) > 0 {
			log.Println("[GIN:AudioGet] range in " + headerRangeSplitted1[0])

			if len(headerRangeSplitted1) > 1 {
				headerRangeSplitted2 := strings.Split(headerRangeSplitted1[1], "-")
				if len(headerRangeSplitted2) > 0 {
					log.Println("[GIN:AudioGet] start = " + headerRangeSplitted2[0])
					if len(headerRangeSplitted2) > 1 {
						log.Println("[GIN:AudioGet] end = " + headerRangeSplitted2[1])
						bytesToSendStart, err := strconv.Atoi(headerRangeSplitted2[0])
						if ErrCheck2(err) {
							bytesToSendEnd, err := strconv.Atoi(headerRangeSplitted2[1])
							if ErrCheck2(err) {
								bytesToSend = bytesToSendEnd - bytesToSendStart + 1
							}
						}
					}
				}
			}
		}
	}

	log.Println("[GIN:AudioGet] Bytes to send " + strconv.Itoa(bytesToSend))
	log.Println("[GIN:AudioGet] From " + strconv.Itoa(bytesToSendStart) + " to " + strconv.Itoa(bytesToSendEnd))

	if bytesToSend > 0 {
		c.Header("Content-length", strconv.Itoa(bytesToSend))
		c.Header("Content-range", "bytes "+strconv.Itoa(bytesToSendStart)+"-"+strconv.Itoa(bytesToSendEnd)+"/"+strconv.Itoa(int(size)))
		size = int64(bytesToSend)
	}

	// Streaming LOOP...
	// ----------------------------------------------------------------------------------------------

	// Otwieraamy plik - bez sprawdzania błędów
	file, err := os.Open(filenameWAV)
	defer file.Close()
	if ErrCheck(err) {
		// Info o wejściu do GET
		log.Println("[GIN:AudioGet] Sending " + id + "...")

		p := make([]byte, size)
		file.ReadAt(p, int64(bytesToSendStart))
		file.Close()
		if bytesToSend > 0 {
			c.Data(http.StatusPartialContent, "audio/wav", p)
		} else {
			c.Data(http.StatusOK, "audio/wav", p)
		}
	} else {
		log.Println("[GIN:AudioGet] Can't open file " + filenameWAV)
	}

	log.Println("[GIN:AudioGet] Sending " + id + " ended.")
}

// // smtpServer - smtpServer data to smtp server
// // ================================================================================================
// type smtpServer struct {
// 	host string
// 	port string
// }

// // Address - URI to smtp server
// // ================================================================================================
// func (s *smtpServer) Address() string {
// 	return s.host + ":" + s.port
// }

// // SendEmail - decode reader
// // ================================================================================================
// func SendEmail(in string) {
// 	// Sender data.
// 	from := "sidcloud.net@gmail.com"
// 	password := "SidCloud1024!"

// 	// Receiver email address.
// 	to := []string{
// 		"b.apanasewicz@gmail.com",
// 		// "secondemail@gmail.com",
// 	}

// 	// smtp server configuration.
// 	smtpServer := smtpServer{host: "smtp.gmail.com", port: "587"}

// 	// Message.
// 	message := []byte(in)
// 	// Authentication.
// 	auth := smtp.PlainAuth("", from, password, smtpServer.host)
// 	// Sending email.
// 	err := smtp.SendMail(smtpServer.Address(), auth, from, to, message)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	log.Println("Email Sent!")
// }

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

// redirect - Przekierowanie http na https
// ================================================================================================
func redirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.RequestURI

	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

// ================================================================================================
// MAIN()
// ================================================================================================
func main() {
	//
	// Sprawdzamy argumenty
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
	sort.Sort(byID(releases))
	// log.Print(releases)

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

	r.GET("/api/v1/audio/:id", AudioGetNew)
	r.GET("/api/v1/csdb_releases", CSDBGetLatestReleases)
	r.GET("/api/v1/csdb_release/:id", CSDBGetRelease)
	r.GET("/api/v1/hvsc_filter/:id", GetHVSCFilter)

	//
	// Start serwera
	//
	if args[0] == "https" {

		// log.Fatal(autotls.Run(r, "sidcloud.net", "www.sidcloud.net"))

		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("sidcloud.net"),
			Cache:      autocert.DirCache("./" + cacheDir),
		}

		log.Fatal(autotls.RunWithManager(r, &m))

	}
	if args[0] == "http" {
		if len(args) > 1 {
			r.Run(":" + args[1])
		} else {
			r.Run(":80")
		}
	}
}
