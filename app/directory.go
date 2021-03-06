package app

import (
	"io/ioutil"
	"log"
	"mime"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
)

// Types
const Pages = "Pages document"
const TAR = "Tarball"
const ZIP = "Zipball"
const PDF = "PDF Document"
const MP3 = "MP3 Audio"
const OGG = "OGG Audio"
const FLAC = "FLAC High-Definition"
const MIDI = "MIDI Synth Audio"
const WAV = "WAV Audio"
const ASD = "Ableton Analysis File"
const AVI = "AVI Video"
const DMG = "Apple Disk Image"
const MKV = "Matroska High-Definition Video"
const MP4 = "MP4 Video"
const TXT = "Plain Text"
const SRT = "Subtitle File"

const Other = "Other"

// Categories
const Dir = "Directory"
const Document = "Document"
const Archive = "Archive"
const Audio = "Audio"
const Video = "Video"
const Program = "Program"

type Type struct {
	Cat  Category
	Type string
}

type File struct {
	Name     string
	Type     *Type
	Category Category
	Ext      string
	IsDir    bool
	Size     string
}

func getKnownExt(ext string) *Type {
	known := make(map[string]*Type)
	known[".pages"] = &Type{Document, Pages}
	known[".asd"] = &Type{Audio, ASD}
	known[".srt"] = &Type{Document, SRT}
	known[".txt"] = &Type{Document, TXT}
	t := known[ext]
	if t == nil {
		return &Type{Other, ext}
	}
	return t
}

func getKnownMime(mime string) *Type {
	var t *Type

	s := strings.Split(mime, "/")
	known_app := make(map[string]*Type)
	known_app["pdf"] = &Type{Document, PDF}
	known_app["x-tar"] = &Type{Archive, TAR}
	known_app["x-apple-diskimage"] = &Type{Program, DMG}
	known_app["zip"] = &Type{Archive, ZIP}
	known_app["x-subrip"] = &Type{Document, SRT}
	known_audio := make(map[string]*Type)
	known_audio["mpeg"] = &Type{Audio, MP3}
	known_audio["mid"] = &Type{Audio, MIDI}
	known_audio["x-wav"] = &Type{Audio, WAV}
	known_audio["x-flac"] = &Type{Audio, FLAC}
	known_audio["ogg"] = &Type{Audio, OGG}
	known_video := make(map[string]*Type)
	known_video["x-msvideo"] = &Type{Video, AVI}
	known_video["x-matroska"] = &Type{Video, MKV}
	known_video["mp4"] = &Type{Video, MP4}
	switch s[0] {
	case "application":
		t = known_app[s[1]]
	case "audio":
		t = known_audio[s[1]]
	case "video":
		t = known_video[s[1]]
	default:
		tmpCat := strings.Title(s[0])
		tmpSub := strings.ToUpper(s[1])
		t = &Type{Category(tmpCat), tmpSub}
	}
	if t == nil {
		return &Type{Other, s[1]}
	}
	return t
}

func getKnownType(f string) *Type {
	var ft *Type
	ext := filepath.Ext(f)
	mime := mime.TypeByExtension(ext)
	if mime == "" {
		ft = getKnownExt(ext)
	} else {
		ft = getKnownMime(mime)
	}
	return ft
}

func cleanName(fp string) string {
	b := filepath.Base(fp)
	c := filepath.Clean(b)
	return c
}

func CountFilesInDir(dir string) (uint64, error) {
	var count uint64
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	for _, file := range files {
		cleanedName := cleanName(file.Name())
		if !strings.HasPrefix(cleanedName, ".") {
			count++
		}
	}
	return count, nil
}

func ProcessDir(dir string) (map[Category]interface{}, error) {
	tmp := make(map[Category]interface{})
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		var f *File
		var t *Type

		originalName := file.Name()
		cleanedName := cleanName(originalName)
		if strings.HasPrefix(cleanedName, ".") {
			log.Println("Skipping ", cleanedName)
		} else {
			if file.IsDir() {
				if tmp[Dir] == nil {
					tmp[Dir] = make([]*File, 0)
				}
				f = &File{
					Name:     file.Name(),
					IsDir:    true,
					Type:     nil,
					Category: Dir,
				}
				tmp[Dir] = append(tmp[Dir].([]*File), f)
			} else {
				t = getKnownType(cleanedName)
				f = &File{
					Name:     file.Name(),
					Type:     t,
					Ext:      filepath.Ext(file.Name()),
					Category: t.Cat,
					Size:     humanize.Bytes(uint64(file.Size())),
				}
				if t.Type != "" {
					ptr := tmp[t.Cat]
					if ptr == nil { // IF doesn't exist
						tmp[t.Cat] = make(map[string][]*File, 0)
						ptr = tmp[t.Cat]
					}
					subCat := ptr.(map[string][]*File)
					subCatPtr := subCat[t.Type]
					if subCatPtr == nil {
						subCat[t.Type] = make([]*File, 0)
					}
					subCat[t.Type] = append(subCat[t.Type], f)
				}
			}
		}
	}
	// Print debug
	// for cat, entity := range tmp {
	// 	log.Println(cat)
	// 	t := reflect.TypeOf(entity)
	// 	if t.Kind() == reflect.Map { // If sub-cat
	// 		subCats := entity.(map[string][]*File)
	// 		for subCat, files := range subCats { // List sub-cat files
	// 			log.Println("\t", subCat)
	// 			for _, file := range files {
	// 				log.Println("\t\t", file.Name)
	// 			}
	// 		}
	// 	} else { // List files
	// 		files := entity.([]*File)
	// 		for _, file := range files {
	// 			log.Println("\t", file.Name)
	// 		}
	// 	}
	// }
	return tmp, nil
}
