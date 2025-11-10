
// go mod init github.com/markburgess/MusicSST
// go mod tidy
// sudo mount -t nfs 192.168.0.250:/Recordings /mnt/Recordings

//******************************************************************
//
// Demo of node by node addition, assuming that the arrows are predefined
//
//******************************************************************

package main

import (
	"fmt"
	"os"
	"io/fs"
	"log"
	"time"
	"strings"
	"path/filepath"
        "github.com/dhowden/tag"
	"github.com/go-flac/go-flac/v2"
)

var ALBUM_COUNTER int

		/*
		fmt.Println("TRACK-NAME:",m.Title()) // string
		fmt.Println("ALBUM-NAME:",m.Album()) // string
		fmt.Println("TRACK-ARTIST:",m.Artist()) // string
		fmt.Println("ALBUM-ARTIST:",m.AlbumArtist()) // string
		fmt.Println("COMPOSER:",m.Composer()) // string
		fmt.Println("YEAR:",m.Year()) // int		
		n,N := m.Track() // (int, int) // Number, Total
		fmt.Println("TRACK:",n,"of",N) */

		/* These fields are typically noise
		fmt.Println("COMMENT:",m.Comment()) // string 
		fmt.Println("GENRE:",m.Genre()) // string
		fmt.Println(Disc()) // (int, int) // Number, Total
		fmt.Println("IMG",m.Picture()) // *Picture // Artwork
		fmt.Println("LYRIC:",m.Lyrics()) // string
		t := m.Raw()
		for i,j := range t {
			fmt.Printf("RAW - Key: %s = %s\n",i,j)
		}*/

//******************************************************************

func main() {

	AnnotateFile("/home/mark/TESTFLAC1.flac")

	AnnotateFile("/home/mark/TESTFLAC2.flac")

	AnnotateFile("/home/mark/TESTFLAC3.flac")

	return

	rootPath := "/mnt/Recordings"
	
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			// Handle errors that occur during directory traversal
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return err // Return the error to stop further traversal in this branch
		}
		
		if d.IsDir() {
			if strings.Contains(path,"/.") {
				return nil
			}
			fmt.Printf("Entering: %s----> %s\n",path,d.Name())
		} else {
			file := filepath.Base(path) 
			if strings.HasPrefix(file,".") || strings.HasPrefix(file,":") {
				return nil
			}

			ALBUM_COUNTER++
			AnnotateFile(path)

		}
		return nil // Return nil to continue traversal
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}
}

//******************************************************************

func AnnotateFile(path string) {

	var stanzas []string
	var sampling = make(map[string]int)
	var composers = make(map[string]int)
	var performers = make(map[string]int)
	var genres = make(map[string]int)
	var titles  = make(map[string]int)

	f, err := os.Open(path)

	m, err := tag.ReadFrom(f)

	if err != nil {
		fmt.Print(err)
	} else {

		n,N := m.Track()

		stanzas = append(stanzas,AnalyzeFLAC(path,m.Genre(),m.Album(),m.Year(),m.Title(),n,N,m.AlbumArtist(),m.Artist(),m.Composer(),sampling,composers,performers,genres,titles))

	f.Close()
	}

	fmt.Println("\n",PrintMap(titles))
	fmt.Println("    \"   ",PrintMap(genres))
	fmt.Println("    \"   (music by)",PrintMap(composers))
	fmt.Println("    \"   (sample rate) ",PrintMap(sampling))

	fmt.Println("\n  +:: _sequence_ ::\n")

	for _,s := range stanzas {
		fmt.Println(s)
	}

	fmt.Println("\n  -:: _sequence_ ::\n")
}

// ****************************************************************

func AnalyzeFLAC(path,genre,album string,year int,track string,n,tot int,album_artist,track_artist,composer string,sampling,composers,performers,genres,titles map[string]int) string {

	var stanza string
	//var title string
	//var conductor string
	//var performer string

	track_name := fmt.Sprintf("%d. '%s'",n,track)
	track_length,_ := getTrackLength(path)
	sample,depth := getSampleRate(path)

	// *******

	stanza += fmt.Sprintln(track_name," (track_in) ",album)

	titles[album]++
	
	mins := track_length / time.Minute
	secs := track_length % time.Minute / time.Second

	stanza += fmt.Sprintf("    \"    (length) %d:%d\n",mins,secs)

	if len(composer) > 0 {
		composer = strings.ReplaceAll(composer," and ",",")
		cmps := strings.Split(composer,",")
		for _,c := range cmps {
			if len(c) > 0 {
				stanza += fmt.Sprintf("    \"    (composer) %%'%s'\n",strings.TrimSpace(c))
				composers[c]++
			}
		}
	}

	// *******

	TryToExtract(track_artist)

	if len(track_artist) > 0 {
		track_artist = strings.ReplaceAll(track_artist," and ",",")
		cmps := strings.Split(track_artist,",")
		for _,c := range cmps {
			if len(c) > 0 {
				stanza += fmt.Sprintf("    \"    (performed by) %%'%s'\n",strings.TrimSpace(c))
				performers[c]++
			}
		}
	}

	// *******

	if len(genre) > 0 {
		genre = strings.ReplaceAll(genre," and ",",")
		cmps := strings.Split(genre,",")
		for _,c := range cmps {
			if len(c) > 0 {
				stanza += fmt.Sprintf("    \"    (genre) %%'%s'\n",strings.TrimSpace(c))
				genres[c]++
			}
		}
	}

	key := fmt.Sprintf("%.1f KHz/%d bits",float64(sample)/1000.0,depth)
	sampling[key]++

        stanza += fmt.Sprintln("    \"    (conductor)    ?")
        stanza += fmt.Sprintln("    \"    (performance)  ?")

	return stanza
}

// ****************************************************************

func getSampleRate(fileName string) (int,int) {

	f, err := flac.ParseFile(fileName)
	if err != nil {
		panic(err)
	}
	data, err := f.GetStreamInfo()
	if err != nil {
		panic(err)
	}
	return data.SampleRate,data.BitDepth
}

// ****************************************************************

func getTrackLength(fileName string) (time.Duration, error) {

	f, err := flac.ParseFile(fileName)
	if err != nil {
		return 0, fmt.Errorf("failed to parse FLAC file: %w", err)
	}
	defer f.Close()

	streamInfo, err := f.GetStreamInfo()
	if err != nil {
		return 0, fmt.Errorf("failed to get stream info: %w", err)
	}

	// Calculate duration in seconds
	durationSeconds := float64(streamInfo.SampleCount) / float64(streamInfo.SampleRate)

	// Convert to time.Duration
	duration := time.Duration(durationSeconds * float64(time.Second))

	return duration, nil
}

// ****************************************************************

func PrintMap(m map[string]int) string {

	var s string
	for p := range m {
		s += p + ","
	}

	return strings.Trim(s,",")
}

// ****************************************************************

func TryToExtract(s string) {

	fmt.Println("ANALYZE",s)

}

