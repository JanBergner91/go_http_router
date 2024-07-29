package app_filebrowser

import (
	"httpr2/mw_template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var appID = "files"
var root = "./" + appID
var BaseURI = "files"

func ConvertURI(uri string) {

}

type FileInfo2 struct {
	Name       string
	WebPath    string
	FolderPath string
	IsDir      bool
	Prefix     string
}

func GeneratePaths(path, query string, directories, files []FileInfo2) PageData2 {
	var newDirs = []FileInfo2{}
	var newFiles = []FileInfo2{}
	for _, d := range directories {
		d.FolderPath = strings.Replace(d.FolderPath, "\\", "/", -1)
		d.WebPath = strings.Replace(d.WebPath, root, BaseURI, 1)
		d.WebPath = strings.Replace(d.WebPath, "\\", "/", -1)
		newDirs = append(newDirs, d)
		//fmt.Println(d)
	}

	for _, d := range files {
		d.FolderPath = strings.Replace(d.FolderPath, "\\", "/", -1)
		d.WebPath = strings.Replace(d.WebPath, root, BaseURI, 1)
		d.WebPath = strings.Replace(d.WebPath, "\\", "/", -1)
		newFiles = append(newFiles, d)
	}

	temp := PageData2{
		Path:  path,
		Dirs:  newDirs,
		Files: newFiles,
		Query: query,
	}
	return temp
}

type PageData2 struct {
	Path  string
	Dirs  []FileInfo2
	Files []FileInfo2
	Query string
}

func Main() *http.ServeMux {
	appRouter := http.NewServeMux()
	appRouter.HandleFunc("/", fileHandler2)
	return appRouter
}

func fileHandler2(w http.ResponseWriter, r *http.Request) {
	//path := r.URL.Path[1:] // Remove the leading "/"
	path := root + r.URL.Path // Remove the leading "/"
	spath := r.URL.Query().Get("p")
	query := r.URL.Query().Get("q")
	if spath != "" {
		path = spath
	}

	//fmt.Println("path: ", path)
	//fmt.Println("spath: ", spath)
	//fmt.Println("query: ", query)

	var temppath = ""
	if path == "" {
		temppath = (root)
	} else {
		if strings.HasSuffix(path, "/") {
			temppath = path
		} else {
			temppath = path + "/"
		}

	}
	//fmt.Println("temppath", temppath)

	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	var dirs []FileInfo2
	var files []FileInfo2

	var pageData = PageData2{}

	if query != "" {
		if fileInfo.IsDir() {
			err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if p == path {
					return nil
				}
				//fmt.Println("P::", p)
				file := FileInfo2{
					Name:       info.Name(),
					WebPath:    p,
					FolderPath: p,
					IsDir:      info.IsDir(),
				}
				if info.IsDir() {
					//dirs = append(dirs, file)
				} else {
					files = append(files, file)
				}
				return nil
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.ServeFile(w, r, path)
			return
		}
	} else {
		if fileInfo.IsDir() {
			entries, err := os.ReadDir(temppath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, entry := range entries {

				info, err := entry.Info()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if query == "" || strings.Contains(info.Name(), query) {
					file := FileInfo2{
						Name:       info.Name(),
						WebPath:    temppath + info.Name(),
						FolderPath: temppath + info.Name(),
						IsDir:      info.IsDir(),
					}

					if info.IsDir() {
						dirs = append(dirs, file)
					} else {
						files = append(files, file)
					}
				}
			}
		} else {
			http.ServeFile(w, r, temppath)
			return
		}
	}
	pageData = GeneratePaths(temppath, query, dirs, files)
	//fmt.Println("PD0:", pageData)
	mw_template.ProcessTemplate(w, "filebrowser.html", "./html-templates", 200, pageData)
}
