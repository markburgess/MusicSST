
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
	"io"
	"io/fs"
	"log"
	"time"
	"strings"
	"sort"
	"path/filepath"
        "github.com/dhowden/tag"
	"github.com/go-flac/go-flac/v2"
)

//******************************************************************

type Track struct {

	N int
	Title string
	Duration string
	Year int
	Samplings map[string]int
	Composers map[string]int
	Conductors map[string]int
	Performers map[string]int
	Orchestra map[string]int
	Engineer  map[string]int
	Producer  map[string]int
	Choir  map[string]int
	Genres map[string]int
	Unknowns map[string]int
}

var ALBUM_COUNTER int
var CURRENT_ALBUM string
var COLLECTION = make(map[string][]Track)
var IGNORE = []string{"Orchestra","Engineer","Producer","Conductor","Composer","Studio"}

//******************************************************************

func main() {

	//AnnotateFile("/home/mark/TESTFLAC1.flac")
	//AnnotateFile("/home/mark/TESTFLAC2.flac")
	//AnnotateFile("/home/mark/TESTFLAC3.flac")

	fp, err := os.Create("output.txt")

	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer fp.Close()

	ScanDirectories(fp)

}

//******************************************************************

func ScanDirectories(fp io.Writer) {

	rootPath := "/mnt/Recordings/Ralph-Vaughan-Williams/"
	
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
			fmt.Printf("Entering: %s\n",d.Name())
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

		fmt.Fprintln(fp,"\n  ###################################\n ")
		SummarizeAlbum(fp,COLLECTION[all],all)
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
	t.Orchestra = make(map[string]int)
	t.Engineer  = make(map[string]int)
	t.Producer  = make(map[string]int)
	t.Choir   = make(map[string]int)
	t.Genres = make(map[string]int)
	t.Unknowns = make(map[string]int)

	f, err := os.Open(path)
	
	if err != nil {
		fmt.Print(err)
	} else {
		m, err := tag.ReadFrom(f)

		if err != nil {
			fmt.Print(err)
		} else {
			n,N := m.Track()
			t.N = n
			title,length,t.Year,t.Title = AnalyzeFLAC(path,m,n,N,t)
			f.Close()
		}
	}

	t.Duration = length

	return title,t
}

// ****************************************************************

func SummarizeAlbum(fp io.Writer,t []Track,title string) {

	fmt.Fprintln(fp,"\n",Esc(title))
	fmt.Fprintln(fp,"     \"    (release date) ",t[0].Year)

	var allcomposers = make(map[string]int)
	var allsample = make(map[string]int)
	var allconduct = make(map[string]int)
	var allorch = make(map[string]int)
	var allperf = make(map[string]int)
	var alleng = make(map[string]int)
	var allprod = make(map[string]int)
	var allchoir = make(map[string]int)
	var allgenre = make(map[string]int)
	var allknow = make(map[string]int)

	for i,_ := range t {
		MergeMaps(allcomposers,t[i].Composers)
		MergeMaps(allsample,t[i].Samplings)
		MergeMaps(allconduct,t[i].Conductors)
		MergeMaps(allorch,t[i].Orchestra)
		MergeMaps(allchoir,t[i].Choir)
		MergeMaps(allperf,t[i].Performers)
		MergeMaps(alleng,t[i].Engineer)
		MergeMaps(allprod,t[i].Producer)
		MergeMaps(allgenre,t[i].Genres)
		MergeMaps(allknow,t[i].Unknowns)
	}

	Add(fp,0,allsample,"sample rate")
	Add(fp,0,allcomposers,"composer")
	Add(fp,0,allconduct,"conductor")
	Add(fp,0,allorch,"orchestra")
	Add(fp,0,allchoir,"choir")
	Add(fp,0,allperf,"performer")
	Add(fp,0,alleng,"engineer")
	Add(fp,0,allprod,"producer")
	Add(fp,0,allgenre,"genre")
	Add(fp,0,allknow,"undecipherable role")

	// ******

	fmt.Fprintln(fp,"\n  +:: _sequence_ ::\n")

	sort.Slice(t, func(i, j int) bool {
		return t[i].N < t[j].N
	})

	for i,_ := range t {
		fmt.Fprintln(fp,"\n",Esc(t[i].Title)," (track in) ",Esc(title))
		fmt.Fprintln(fp,"     \"     (duration) ",t[i].Duration)
		Add(fp,1,allsample,"sample rate")
		Add(fp,1,allcomposers,"composer")
		Add(fp,1,allconduct,"conductor")
		Add(fp,1,allorch,"orchestra")
		Add(fp,1,allchoir,"choir")
		Add(fp,1,allperf,"performer")
		Add(fp,1,alleng,"engineer")
		Add(fp,1,allprod,"producer")
		Add(fp,1,allgenre,"genre")
		Add(fp,1,allknow,"undecipherable role")
	}

	fmt.Fprintln(fp,"\n  -:: _sequence_ ::\n")

	// *******
}

// ****************************************************************

func Esc(s string) string {

	s = strings.ReplaceAll(s,"(","[")
	s = strings.ReplaceAll(s,")","]")
	return s
}

// ****************************************************************

func Add(fp io.Writer,lim int,attrib map[string]int,relation string) {

	if len(attrib) > lim {

		for p := range attrib {
			fmt.Fprintf(fp,"    \"    (%s) '%s'\n",relation,p)
		}
	}
}

// ****************************************************************

func AnalyzeFLAC(path string,m tag.Metadata,n,tot int,t Track) (string,string,int,string) {

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

	t.Genres[genre]++

	// Make sure we encapsulate tracks with number, since tracks may collide with album

	track_name := fmt.Sprintf("%d. '%s'",n,track)

	// Calculate the duration

	track_length,_ := GetTrackLength(path)
	mins := track_length / time.Minute
	secs := track_length % time.Minute / time.Second

	if secs < 10 {
		duration = fmt.Sprintf("%d:0%d\n",mins,secs)
	} else {
		duration = fmt.Sprintf("%d:%d\n",mins,secs)
	}

	// Sampling quality

	sample,depth := GetSampleRate(path)
	key := fmt.Sprintf("%.1f KHz/%d bits",float64(sample)/1000.0,depth)
	t.Samplings[key]++

	// No try to decode the broken metadata

	Deconstruct(track_artist,t)
	Deconstruct(album_artist,t)
	Deconstruct(composer,t)

	return album_title,duration,year,track_name
}

// ****************************************************************

func GetSampleRate(fileName string) (int,int) {

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

func GetTrackLength(fileName string) (time.Duration, error) {

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

func Deconstruct(annotation string,t Track) {

	// This describes the logic of inference (i.e. hackery) to decode intent in metadata
	// First try to split on intentional packing, either \n or ;
	// Then we need to do 2 passes, splitting or not splitting on comma ,
	// because comma is used in multiple ways

	fmt.Println("\n( Deconstruct",annotation,")")

	annotation = strings.ReplaceAll(annotation," and ",";")
	annotation = strings.ReplaceAll(annotation,"\n",";")
	annotation = strings.ReplaceAll(annotation,"&",";")
	annotation = strings.ReplaceAll(annotation,".",";")

	items := strings.Split(annotation,";")

	// Split commas first

	for _,item := range items {

		var done bool = false

		subs := strings.Split(item,",")

		for _,it := range subs {
			done = DoChecks(it,t) || done
		}

		if done {
			continue
		}

		if !DoChecks(item,t) {
			
			// If we still have commma separated, split them as they are unlabelled
			
			subi := strings.Split(item,",")
			
			for _,i := range subi {				
				i = strings.TrimSpace(i)
				CheckFor("Unknown",i,t.Unknowns)
			}
			
		}
	}
}

// ****************************************************************

func DoChecks(item string,t Track) bool {

	strings.TrimSpace(item)

	// Order matters, due to overriding

	if CheckFor("Orchestra",item,t.Orchestra) {
		return true
	}

	if CheckFor("Engineer",item,t.Engineer) {
		return true
	}
	
	if CheckFor("Producer",item,t.Producer) {
		return true
	}

	if CheckFor("Conductor",item,t.Conductors) {
		return true
	}
	
	if CheckFor("Composer",item,t.Composers) {
		return true
	}

	if CheckFor("Unknown",item,t.Unknowns) {
		return false
	}

	return false
}

// ****************************************************************

func CheckFor(role string,item string,record map[string]int) bool {

	var match bool

	item = strings.TrimSpace(item)
	
	if len(item) > 0 {
		
		if !match && role == "Orchestra" {
			match = strings.Contains(item,"Philharmonic") || strings.Contains(item,"Symfon")
		} else if role != "Unknown" {
			match = strings.Contains(item,role) || strings.Contains(strings.ToLower(item),role)
		} else {
			match = true
		}

		if match {
			item = strings.ReplaceAll(item,role,"")
			item = strings.ReplaceAll(item,strings.ToLower(role),"")
			item = strings.ReplaceAll(item,"MainArtist","")
			item = strings.ReplaceAll(item,"Artist","")
			item = strings.ReplaceAll(item,"  "," ")

			if strings.Contains(item," ,") {
				sub := strings.Split(item," ,")
				item = sub[0]
			}

			item = strings.TrimSpace(item)

			if record[item] > 0 {
				return true
			}

			if len(item) > 0 && !Ignore(item) {
				record[item]++
				fmt.Printf("  -- Extracted %s for %s\n",item,role)
				return true
			}
		}
	}

	return false
}

// ****************************************************************

func Ignore(str string) bool {

	for _,s := range IGNORE {
		if s == str {
			return true
		}
	}
	return false
}

// ****************************************************************

func MergeMaps(target,source map[string]int) {

	for key := range source {
		target[key]++
	}
}

