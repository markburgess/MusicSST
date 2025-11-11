
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

//******************************************************************

type Track struct {

	Title string
	Duration string
	Year int
	Samplings map[string]int
	Composers map[string]int
	Conductors map[string]int
	Performers map[string]int
	Genres map[string]int
}

var ALBUM_COUNTER int
var CURRENT_ALBUM string
var COLLECTION = make(map[string][]Track)

//******************************************************************

func main() {

	//AnnotateFile("/home/mark/TESTFLAC1.flac")
	//AnnotateFile("/home/mark/TESTFLAC2.flac")
	//AnnotateFile("/home/mark/TESTFLAC3.flac")

	//return

	rootPath := "/mnt/Recordings/Ralph-Vaughan-Williams/" // Vaughan-Williams-Symphony-No2-A-London-Symphony-and-Symphony-No8/"
	
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
			fmt.Printf("\nEntering: %s\n",d.Name())
		} else {
			file := filepath.Base(path) 

			if strings.HasPrefix(file,".") || strings.HasPrefix(file,":") || strings.HasSuffix(file,"pdf") || strings.HasSuffix(file,"png") || strings.HasSuffix(file,"jpg")  {
				return nil
			}

			title,track := AnnotateFile(path)

			if len(title) > 0 {
				COLLECTION[title] = append(COLLECTION[title],track)

				if title != CURRENT_ALBUM {
					fmt.Println("New album:",title)
					CURRENT_ALBUM = title
					ALBUM_COUNTER++
				}
			}
		}
		return nil // Return nil to continue traversal
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}


	for all := range COLLECTION {
		fmt.Println("--",all)
		fmt.Println(COLLECTION[all])
	}
}

//******************************************************************

func AnnotateFile(path string) (string,Track) {

	var t Track
	var title string = "errors"
	var length string

	t.Samplings = make(map[string]int)
	t.Composers = make(map[string]int)
	t.Conductors = make(map[string]int)
	t.Performers = make(map[string]int)
	t.Genres = make(map[string]int)

	f, err := os.Open(path)
	
	if err != nil {
		fmt.Print(err)
	} else {
		m, err := tag.ReadFrom(f)

		if err != nil {
			fmt.Print(err)
		} else {
			n,N := m.Track()
			
			title,length,t.Year,t.Title = AnalyzeFLAC(path,m,n,N,t.Composers,t.Conductors,t.Performers,t.Genres,t.Samplings)
			
			f.Close()
		}
	}

	t.Duration = length

	return title,t
}

// ****************************************************************

func SummarizeAlbum() {

/*	fmt.Println("\n",PrintMap(titles))
	fmt.Println("    \"   ",PrintMap(genres))
	fmt.Println("    \"   (music by)",PrintMap(composers))
	fmt.Println("    \"   (sample rate) ",PrintMap(samplings))

	fmt.Println("\n  +:: _sequence_ ::\n")

	for _,s := range album_level_notes {
		fmt.Println(s)
	}

	fmt.Println("\n  -:: _sequence_ ::\n")

	// *******

	stanza += fmt.Sprintln(track_name," (track_in) ",album)
	stanza += fmt.Sprintln(track_name," (release date) ",year)

	stanza += fmt.Sprintf("    \"    (composer) %%'%s'\n",c)
	stanza += fmt.Sprintf("    \"    (performed by) %%'%s'\n",c)
	stanza += fmt.Sprintf("    \"    (genre) %%'%s'\n",strings.TrimSpace(c))

*/


}

// ****************************************************************

func AnalyzeFLAC(path string,m tag.Metadata,n,tot int,composers,conductors,performers,genres,samplings map[string]int) (string,string,int,string) {

	var album_title string
	var duration string

	// From the metadata

	genre := m.Genre()
	year := m.Year()
	track := m.Title()
	album_artist := m.AlbumArtist()
	track_artist := m.Artist()
	composer := m.Composer()
	album := m.Album()
	album_title = strings.TrimSpace(album)	

	// Make sure we encapsulate tracks with number, since tracks may collide with album

	track_name := fmt.Sprintf("%d. '%s'",n,track)

	track_length,_ := getTrackLength(path)
	mins := track_length / time.Minute
	secs := track_length % time.Minute / time.Second

	if secs < 10 {
		duration = fmt.Sprintf("%d:0%d\n",mins,secs)
	} else {
		duration = fmt.Sprintf("%d:%d\n",mins,secs)
	}

	sample,depth := getSampleRate(path)
	key := fmt.Sprintf("%.1f KHz/%d bits",float64(sample)/1000.0,depth)
	samplings[key]++

	Deconstruct(track_artist,composers,conductors,performers)
	Deconstruct(album_artist,composers,conductors,performers)
	Deconstruct(composer,composers,conductors,performers)

// producer/engineer/artist/choir/solo/orchestra


	genres[genre]++


	return album_title,duration,year,track_name
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

func Deconstruct(annotation string,composers,conductors,performers map[string]int) {

	fmt.Printf("DECON: (%s)\n",annotation)
	
	// First try to split on intentional packing, either \n or ;
	annotation = strings.ReplaceAll(annotation," and ",",")
	annotation = strings.ReplaceAll(annotation,"\n",";")
	annotation = strings.ReplaceAll(annotation,"&",";")
	annotation = strings.ReplaceAll(annotation,".",";")
	item := strings.Split(annotation,";")

	for _,c := range item {

		c = strings.TrimSpace(c)
		c = strings.ReplaceAll(c,",","")

		if len(c) > 0 {
			if strings.Contains(strings.ToLower(c),"conductor") {
				c = strings.ReplaceAll(c,"conductor","")
				c = strings.ReplaceAll(c,"Conductor","")
				c = strings.TrimSpace(c)
				conductors[c]++
				continue
			}

			if strings.Contains(strings.ToLower(c),"composer") {
				c = strings.ReplaceAll(c,"composer","")
				c = strings.ReplaceAll(c,"Composer","")
				c = strings.TrimSpace(c)
				if len(c) > 0 {
					composers[c]++
				}
				continue
			}

			if len(c) > 0 {
				performers[c]++
			}
		}
	}
}

