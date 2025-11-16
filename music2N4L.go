
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
	"errors"
	"path/filepath"
//        "github.com/dhowden/tag"
	tag "github.com/unitnotes/audiotag"
	"github.com/go-flac/go-flac/v2"
//	"github.com/dmulholl/mp3lib"
	"github.com/hajimehoshi/go-mp3"
)

//******************************************************************

type Track struct {

	N int
	File string
	Title string
	Img string
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

var CURRENT_ALBUM string
var CURRENT_IMAGE string
var COLLECTION = make(map[string][]Track)
var IGNORE = []string{"Orchestra","Engineer","Producer","Conductor","Composer","Studio"}

//******************************************************************

func main() {

	name := "output.n4l"

	if FileExists(name) {
		fmt.Println("File exists - careful!\n")
		return
	}

	fp, err := os.Create(name)

	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer fp.Close()

	ScanDirectories(fp)

}

//******************************************************************

func ScanDirectories(fp io.Writer) {

	// Check Brahms, Andris // Yoyo ma.. Jurowski
	// Star Trek lala land, artist  "s"
	// Made In Japan

	root_path := "/mnt/Recordings/Kōhei Tanaka, Shirō Hamaguchi"

	ignore_prefix := "/mnt"

	off := len(ignore_prefix)
	
	err := filepath.WalkDir(root_path, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			// Handle errors that occur during directory traversal
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil // Return the error to stop further traversal in this branch
		}
		
		if d.IsDir() {
			if strings.Contains(path,"/.") {
				return nil
			}

			if strings.Contains(path,"@Recycle") {
				return nil
			}

			CURRENT_IMAGE = ""
			CURRENT_ALBUM = ""
		} else {
			file := filepath.Base(path) 

			if strings.HasPrefix(file,".") || strings.HasPrefix(file,":") || strings.HasSuffix(file,"pdf") {
				return nil
			}

			// Try to catch a falling album cover ...

			if  strings.HasSuffix(file,".png") || strings.HasSuffix(file,".jpg") {

				switch file {
				case "Folder.jpg", "folder.jpg","cover.jpg":
					CURRENT_IMAGE = path[off:]
				default:
					if CURRENT_IMAGE != "" {
						CURRENT_IMAGE = path[off:]
					}
				}

				AlbumCover(CURRENT_ALBUM,CURRENT_IMAGE)
				return nil
			}

			title,track := AnnotateFile(path)

			if len(title) > 0 {

				COLLECTION[title] = append(COLLECTION[title],track)

				if title != CURRENT_ALBUM {
					fmt.Println("New album:",title)
					CURRENT_ALBUM = title
					AlbumCover(CURRENT_ALBUM,CURRENT_IMAGE)
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

func AlbumCover(title,img string) {

	// Install album cover in an empty track, if we can catch it

	if title != "" && img != "" {
		var image Track
		image.Img = img
		COLLECTION[title] = append(COLLECTION[title],image)
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
			fmt.Println("Unable to read",err)
		} else {
			n,N := m.Track()
			t.N = n
			title,length,t.Year,t.Title = AnalyzeFLAC(path,f,m,n,N,t)

			if len(filepath.Ext(path)) > 1 {
				t.File = filepath.Ext(path)[1:]
			}
			f.Close()
		}
	}
	
	t.Duration = length
	
	return title,t
}

// ****************************************************************

func SummarizeAlbum(fp io.Writer,t []Track,title string) {

	fmt.Printf("Summarizing %s\n",title)

	fmt.Fprintln(fp,"\n",Esc(title))
	fmt.Fprintln(fp,"    \"    (release date) ",t[0].Year)

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
	var image string

	for i,_ := range t {

		if t[i].Img != "" {
			image = t[i].Img
			continue
		}

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

	if image != "" {
		fmt.Fprintf(fp,"    \"    (img) \"%s\"\n",image)
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
		if len(t[i].Title) == 0 {
			continue
		}

		fmt.Fprintln(fp,"\n",Esc(t[i].Title)," (track in) ",Esc(title))
		fmt.Fprintln(fp,"     \"     (duration) ",t[i].Duration)
		fmt.Fprintln(fp,"     \"     (encoding) ",t[i].File)
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

func AnalyzeFLAC(path string,f *os.File,m tag.Metadata,n,tot int,t Track) (string,string,int,string) {

	var album_title string
	var duration string

	// From the metadata

	//PrintRaw(m)

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

	track_length,_ := GetTrackLength(path,f)
	mins := track_length / time.Minute
	secs := track_length % time.Minute / time.Second

	if secs < 10 {
		duration = fmt.Sprintf("%d:0%d\n",mins,secs)
	} else {
		duration = fmt.Sprintf("%d:%d\n",mins,secs)
	}

	// Sampling quality

	sample,depth := GetSampleRate(path,f)
	key := fmt.Sprintf("%.1f KHz/%d bits",float64(sample)/1000.0,depth)
	t.Samplings[key]++

	// No try to decode the broken metadata

	Deconstruct(track_artist,t,"artist")
	Deconstruct(album_artist,t,"artist")
	Deconstruct(composer,t,"composer")

	return album_title,duration,year,track_name
}

// ****************************************************************

func GetSampleRate(path string,f *os.File) (int,int) {
	
	if filepath.Ext(path) == ".flac" {

		f, err := flac.ParseFile(path)

		if err != nil {
			// probably not a flac supported
			return 0,0
		}

		data, err := f.GetStreamInfo()
		
		if err != nil {
			// probably not a flac supported
			return 0,0
		}
		return data.SampleRate,data.BitDepth
	}

	if filepath.Ext(path) == ".mp3" {

		d,err := mp3.NewDecoder(f)
		
		if err != nil {
			return 0,0
		}
		
		return d.SampleRate(),16
	}

	return 0,0
}

// ****************************************************************

func GetTrackLength(path string,f *os.File) (time.Duration, error) {
	
	if filepath.Ext(path) == ".flac" {
		
		f, err := flac.ParseFile(path)

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
	
	if filepath.Ext(path) == ".mp3" {

		d,err := mp3.NewDecoder(f)

		if err != nil {
			return 0, nil
		}

		const sampleSize = 4                  // From documentation.
		samples := d.Length() / sampleSize    // Number of samples.
		durationSeconds := float64(samples) / float64(d.SampleRate())
		duration := time.Duration(durationSeconds * float64(time.Second))
		return duration, nil
	}
	
	return 0, nil
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

func Deconstruct(annotation string,t Track,intended string) {

	// This describes the logic of inference (i.e. hackery) to decode intent in metadata
	// First try to split on intentional packing, either \n or ;
	// Then we need to do 2 passes, splitting or not splitting on comma ,
	// because comma is used in multiple ways

	if SimpleEntry(annotation) {
		switch intended {
		case "artist":
			CheckFor("Artist",annotation,t.Performers)
			return
		case "composer":
			CheckFor("Composer",annotation,t.Composers)
			return
		}
	}

	//fmt.Println("\n( Deconstruct",annotation,")")

	annotation = strings.ReplaceAll(annotation,"MainArtist",";")
	annotation = strings.ReplaceAll(annotation,"AssociatedPerformer",";")
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

func SimpleEntry(entry string) bool {

	checkfor := []string{",",";",":","\n"}

	for _,c := range checkfor {
		if strings.Contains(entry,c) {
			return false
		}
	}
	return true
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

		if role == "Orchestra" {
			match = strings.Contains(item,"Philharmonic") || strings.Contains(item,"Symfon")
		} else if role != "Unknown" && role != "Artist" {
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

			if strings.Contains(item,",") {
				sub := strings.Split(item,",")
				item = sub[0]
			}

			item = strings.TrimSpace(item)

			if record[item] > 0 {
				return true
			}

			if len(item) > 0 && !Ignore(item) {
				record[item]++
				//fmt.Printf("  -- Extracted %s for %s\n",item,role)
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

// ****************************************************************

func PrintRaw(m tag.Metadata) {

	fmt.Println("-------------------")
	fmt.Println("ALBUM:",m.Album())
	fmt.Println("GENRE:", m.Genre())
	fmt.Println("YEAR:", m.Year())
	fmt.Println("TITLE:", m.Title())
	fmt.Println("ALBUM ARTIST:", m.AlbumArtist())
	fmt.Println("ARTIST:",m.Artist())
	fmt.Println("COMPOSER:",m.Composer())
	fmt.Println("-------------------")
}

// ****************************************************************

func FileExists(path string) bool {

	info, err := os.Stat(path)

	if err == nil {
		return !info.IsDir()
	}

	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}