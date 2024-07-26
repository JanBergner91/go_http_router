package app_filebrowser

import (
	"httpr2/mw_session"
	"httpr2/mw_template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var AppName = "app_filebrowser"
var FilesSubfolder = "files"
var BaseURI = "files"
var RealBasePath = "./" + AppName + "/" + FilesSubfolder

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
		d.WebPath = strings.Replace(d.WebPath, RealBasePath, BaseURI, 1)
		d.WebPath = strings.Replace(d.WebPath, "\\", "/", -1)
		newDirs = append(newDirs, d)
		//fmt.Println(d)
	}

	for _, d := range files {
		d.FolderPath = strings.Replace(d.FolderPath, "\\", "/", -1)
		d.WebPath = strings.Replace(d.WebPath, RealBasePath, BaseURI, 1)
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
	appRouter.HandleFunc("/", fileHandler)
	return appRouter
}

func filterReplace(s string) string {
	st := s
	st = strings.Replace(st, "/"+AppName+"", "", 1)
	st = strings.Replace(st, "\\"+AppName+"", "", 1)
	st = strings.Replace(st, ""+AppName+"\\", "", 1)
	st = strings.Replace(st, ""+AppName+"/", "", 1)
	return st
}

func fileHandler(w http.ResponseWriter, r *http.Request) {

	sessionID := r.Context().Value(mw_session.SessionKey).(string)
	item := mw_session.SessionItem{Key: "fb-key", Value: "fb-value"}
	mw_session.AddOrUpdateSessionItem(sessionID, item)

	path := r.URL.Path[1:] // Remove the leading "/"
	spath := r.URL.Query().Get("p")
	query := r.URL.Query().Get("q")

	if spath != "" {
		path = spath

		var temppath = ""
		if path == "" {
			temppath = RealBasePath
		} else {
			temppath = RealBasePath + spath
		}

		fileInfo, err := os.Stat(path)
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		var dirs []FileInfo2
		var files []FileInfo2

		var pageData = PageData2{}

		if fileInfo.IsDir() {

			err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if p == path {
					return nil
				}

				if query == "" || strings.Contains(strings.ToLower(info.Name()), strings.ToLower(query)) {
					realName := strings.Replace(info.Name(), ""+AppName+"\\", ".\\", 1)
					realName = strings.Replace(realName, ""+AppName+"/", ".\\", 1)
					realPath := filterReplace(p)

					file := FileInfo2{
						Name:       realName,
						WebPath:    realPath,
						FolderPath: realPath,
						IsDir:      info.IsDir(),
					}
					if info.IsDir() {
						dirs = append(dirs, file)
					} else {
						files = append(files, file)
					}
				}
				return nil
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			pageData = GeneratePaths(temppath, "", dirs, files)

		} else {
			http.ServeFile(w, r, path)
			return
		}

		mw_template.ProcessTemplate(w, "filebrowser.html", "./html-templates", 200, pageData)
	} else {
		var temppath = ""
		if path == "" {
			temppath = RealBasePath
		} else {
			temppath = RealBasePath + "/" + path
		}
		fileInfo, err := os.Stat(temppath)
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		var dirs []FileInfo2
		var files []FileInfo2

		var pageData = PageData2{}

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

					realName := strings.Replace(info.Name(), ""+AppName+"\\", "", 1)
					realName = strings.Replace(realName, ""+AppName+"/", "", 1)

					realPath := filterReplace(temppath)
					realPath = strings.Replace(realPath, "\\", "/", -1)

					file := FileInfo2{
						Name:       realName,
						WebPath:    temppath + "/" + realName,
						FolderPath: realPath,
						IsDir:      info.IsDir(),
					}

					if info.IsDir() {
						dirs = append(dirs, file)
					} else {
						files = append(files, file)
					}
				}
			}

			pageData = GeneratePaths(temppath, "", dirs, files)

		} else {
			//fmt.Println("ServerFilePath::" + temppath)
			http.ServeFile(w, r, temppath)
			return
		}

		mw_template.ProcessTemplate(w, "filebrowser.html", "./html-templates", 200, pageData)

	}

}