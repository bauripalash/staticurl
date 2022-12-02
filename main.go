package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed template.html
var tmplsrc string

//go:embed front.html
var frontsrc string

//go:embed config.json
var defaultConf string

var (
	Warnlog *log.Logger
	Errlog  *log.Logger
)

func init() {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lmsgprefix
	Warnlog = log.New(os.Stdout, "WARN: ", flags)
	Errlog = log.New(os.Stdout, "ERROR: ", flags)
}

type UrlData struct {
	Url  string
	Code string
}

type Config struct {
	Output string `json:"output"`
	Urldir string `json:"urldir"`
}

func checkDir(targedir string) error {
	return os.MkdirAll(filepath.Join(".", targedir), os.ModePerm)
}

func initNewSite(sitename string) {
	urlsdir := filepath.Join(sitename, "urls")
	checkDir(urlsdir)
	f, err := os.Create(filepath.Join(sitename, "config.json"))

	if err != nil {
		Warnlog.Println("Failed to create config.json; project will be using default values")
	}

	defer f.Close()

	_, errr := f.WriteString(defaultConf)

	if errr != nil {
		Warnlog.Println("Failed to write to default config.json; project will be using default values")
	}

}

func readConfig() (Config, bool) {
	configpath := filepath.Join(".", "config.json")
	var config Config
	f, err := os.ReadFile(configpath)

	if err != nil {
		Warnlog.Println("config file cannot be found!")
		return config, false
	}

	//	var config Config

	errr := json.Unmarshal(f, &config)

	if errr != nil {
		Warnlog.Println("Failed to parse config at ->" + configpath)
		return config, false
	}

	return config, true

}

func createFrontpage(outdir string) bool {
	if err := checkDir(outdir); err != nil {
		Warnlog.Println("Failed to create front page")
		return false
	}

	fpath := filepath.Join(".", outdir, "index.html")

	f, err := os.Create(fpath)

	if err != nil {
		Warnlog.Println("Failed to create front page at -> " + fpath)
		return false
	}

	defer f.Close()

	if _, err := f.WriteString(frontsrc); err != nil {
		Warnlog.Println("Failed to write to front page at -> " + fpath)
		return false
	}

	return true

}

func (u *UrlData) createFile(outdir string, data string) bool {
	if err := checkDir(outdir); err != nil {
		Errlog.Fatalln("Failed to create output directory")
		return false
	}

	if err := checkDir(filepath.Join(outdir, u.Code)); err != nil {
		Errlog.Fatalln("Failed to create url directory for `", u.Code, "`")
		return false
	}

	outputpath := filepath.Join(".", outdir, u.Code, "index.html")

	f, err := os.Create(outputpath)

	if err != nil {
		Errlog.Fatalln("Failed to create file for code -> ", u.Code)
		return false
	}

	defer f.Close()

	if _, err := f.WriteString(data); err != nil {
		Errlog.Fatalln("Failed to write to file for code -> ", u.Code)
		return false
	}

	return true
}

func processUrl(fpath string, fname string) (UrlData, bool) {
	f, err := os.Open(fpath)

	if err != nil {
		Errlog.Fatalln("Failed to open file " + fpath)
	}

	defer f.Close()
	rdr := bufio.NewScanner(f)
	rdr.Scan()

	if err := rdr.Err(); err != nil {
		Errlog.Fatalln("Failed to read file " + fpath)
	} else {
		return UrlData{Url: rdr.Text(), Code: fname}, true
	}

	return UrlData{}, false
}

func build() {
	conf, isOk := readConfig()
	var OUT string
	var CDIR string

	if !isOk {
		Warnlog.Println("Failed to read config file using default directories")
		OUT = "public"
		CDIR = "urls"
	} else {
		OUT = conf.Output
		CDIR = conf.Urldir
	}

	curdir, _ := os.Getwd()
	contentDir := curdir + string(os.PathSeparator) + CDIR + string(os.PathSeparator)
	files, err := ioutil.ReadDir(contentDir)

	tmpl, err := template.New("url").Parse(tmplsrc)

	if err != nil {
		Errlog.Fatalln("Failed to read Template file")
	}

	if err != nil {
		Errlog.Fatalln("Err : cannot read url directory")
		return
	}

	if len(files) < 1 {
		//fmt.Println("Urls directory is empty")
		Errlog.Fatalln("URLs directory is empty")
		return
	}

	for _, f := range files {

		if !f.IsDir() {
			data, isOk := processUrl(contentDir+f.Name(), f.Name())

			if isOk {
				//fmt.Println(data)
				tempbuf := &bytes.Buffer{}
				tmpl.Execute(tempbuf, data)
				data.createFile(OUT, tempbuf.String())

			}
		}
	}

	createFrontpage(OUT)

}

func main() {
	//	args := os.Args
	helpmsg := "staticurl v0.1.0\nsimple and fast flat-file based url shortener without any database\n"

	buildflag := flag.Bool("b", false, "Build Current Project")
	sitename := flag.String("n", "", "Create New staticurl Site")
	flag.Parse()

	if *buildflag {
		build()
		return
	}

	if len(*sitename) >= 1 {
		initNewSite(*sitename)
		return
	}

	println(helpmsg)
	flag.PrintDefaults()

}
