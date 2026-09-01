package files

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
	"image"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/spf13/afero"
)

var (
	reSubDirs = regexp.MustCompile("(?i)^sub(s|titles)$")
	reSubExts = regexp.MustCompile("(?i)(.vtt|.srt|.ass|.ssa)$")
)

// FileInfo describes a file.
type FileInfo struct {
	*Listing
	Fs         afero.Fs          `json:"-"`
	Path       string            `json:"path"`
	Name       string            `json:"name"`
	Size       int64             `json:"size"`
	Extension  string            `json:"extension"`
	ModTime    time.Time         `json:"modified"`
	Mode       os.FileMode       `json:"mode"`
	IsDir      bool              `json:"isDir"`
	IsSymlink  bool              `json:"isSymlink"`
	Type       string            `json:"type"`
	Subtitles  []string          `json:"subtitles,omitempty"`
	Content    string            `json:"content,omitempty"`
	Checksums  map[string]string `json:"checksums,omitempty"`
	Token      string            `json:"token,omitempty"`
	currentDir []os.FileInfo     `json:"-"`
	Resolution *ImageResolution  `json:"resolution,omitempty"`
	UploadOnly bool              `json:"uploadOnly,omitempty"`
}

// FileOptions are the options when getting a file info.
type FileOptions struct {
	Fs         afero.Fs
	Path       string
	Modify     bool
	Expand     bool
	ReadHeader bool
	CalcImgRes bool
	Token      string
	Checker    rules.Checker
	Content    bool
}

type ImageResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// NewFileInfo creates a File object from a path and a given user. This File
// object will be automatically filled depending on if it is a directory
// or a file. If it's a video file, it will also detect any subtitles.
func NewFileInfo(opts *FileOptions) (*FileInfo, error) {
	if !opts.Checker.Check(opts.Path) {
		return nil, os.ErrPermission
	}

	file, err := stat(opts)
	if err != nil {
		return nil, err
	}

	// Do not expose the name of root directory.
	if file.Path == "/" {
		file.Name = ""
	}

	if opts.Expand {
		if file.IsDir {
			if err := file.readListing(opts.Checker, opts.ReadHeader, opts.CalcImgRes); err != nil {
				return nil, err
			}
			return file, nil
		}

		err = file.detectType(opts.Modify, opts.Content, true, opts.CalcImgRes)
		if err != nil {
			return nil, err
		}
	}

	return file, err
}

func stat(opts *FileOptions) (*FileInfo, error) {
	var file *FileInfo

	if lstaterFs, ok := opts.Fs.(afero.Lstater); ok {
		info, _, err := lstaterFs.LstatIfPossible(opts.Path)
		if err != nil {
			return nil, err
		}
		file = &FileInfo{
			Fs:        opts.Fs,
			Path:      opts.Path,
			Name:      info.Name(),
			ModTime:   info.ModTime(),
			Mode:      info.Mode(),
			IsDir:     info.IsDir(),
			IsSymlink: IsSymlink(info.Mode()),
			Size:      info.Size(),
			Extension: filepath.Ext(info.Name()),
			Token:     opts.Token,
		}
	}

	// regular file
	if file != nil && !file.IsSymlink {
		return file, nil
	}

	// fs doesn't support afero.Lstater interface or the file is a symlink
	info, err := opts.Fs.Stat(opts.Path)
	if err != nil {
		// can't follow symlink
		if file != nil && file.IsSymlink {
			return file, nil
		}
		return nil, err
	}

	// set correct file size in case of symlink
	if file != nil && file.IsSymlink {
		file.Size = info.Size()
		file.IsDir = info.IsDir()
		return file, nil
	}

	file = &FileInfo{
		Fs:        opts.Fs,
		Path:      opts.Path,
		Name:      info.Name(),
		ModTime:   info.ModTime(),
		Mode:      info.Mode(),
		IsDir:     info.IsDir(),
		Size:      info.Size(),
		Extension: filepath.Ext(info.Name()),
		Token:     opts.Token,
	}

	return file, nil
}

// Checksum checksums a given File for a given User, using a specific
// algorithm. The checksums data is saved on File object.
func (i *FileInfo) Checksum(algo string) error {
	if i.IsDir {
		return fberrors.ErrIsDirectory
	}

	if i.Checksums == nil {
		i.Checksums = map[string]string{}
	}

	reader, err := i.Fs.Open(i.Path)
	if err != nil {
		return err
	}
	defer reader.Close()

	var h hash.Hash

	switch algo {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return fberrors.ErrInvalidOption
	}

	_, err = io.Copy(h, reader)
	if err != nil {
		return err
	}

	i.Checksums[algo] = hex.EncodeToString(h.Sum(nil))
	return nil
}

func (i *FileInfo) RealPath() string {
	if realPathFs, ok := i.Fs.(interface {
		RealPath(name string) (fPath string, err error)
	}); ok {
		realPath, err := realPathFs.RealPath(i.Path)
		if err == nil {
			return realPath
		}
	}

	return i.Path
}

func isKnownTextExtension(ext, name string) bool {
	switch ext {
	case ".json", ".json5", ".jsonc", ".jsonl", ".geojson",
		".php", ".phtml", ".php3", ".php4", ".php5", ".php7", ".phps",
		".html", ".htm", ".xhtml", ".vue", ".svelte", ".jsx", ".tsx",
		".js", ".mjs", ".cjs", ".ts", ".mts", ".cts",
		".css", ".scss", ".sass", ".less", ".styl",
		".yaml", ".yml", ".toml", ".xml", ".sql", ".env", ".ini", ".conf", ".cfg", ".cnf",
		".properties", ".prefs", ".log", ".txt", ".md", ".markdown", ".rst", ".tex",
		".py", ".pyw", ".go", ".rs", ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx",
		".java", ".kt", ".kts", ".cs", ".swift", ".rb", ".lua", ".sh", ".bash", ".zsh",
		".fish", ".bat", ".cmd", ".ps1", ".pl", ".pm", ".r", ".dart", ".scala", ".groovy",
		".erl", ".ex", ".exs", ".hs", ".clj", ".nim", ".v", ".zig", ".csv", ".tsv",
		".graphql", ".proto", ".diff", ".patch", ".dockerfile", ".makefile", ".lock":
		return true
	}
	switch name {
	case "dockerfile", "makefile", "rakefile", "gemfile", "procfile", "vagrantfile", "caddyfile",
		".gitignore", ".gitattributes", ".dockerignore", ".editorconfig", ".npmrc", ".nvmrc", ".env":
		return true
	}
	return false
}

func isKnownImageExtension(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".ico", ".bmp", ".avif",
		".heic", ".heif", ".tiff", ".tif", ".jfif", ".apng":
		return true
	}
	return false
}

func isKnownVideoExtension(ext string) bool {
	switch ext {
	case ".mp4", ".mkv", ".webm", ".avi", ".mov", ".wmv", ".flv", ".m4v", ".3gp", ".ogv":
		return true
	}
	return false
}

func isKnownAudioExtension(ext string) bool {
	switch ext {
	case ".mp3", ".wav", ".ogg", ".flac", ".m4a", ".aac", ".opus", ".wma", ".mid", ".midi":
		return true
	}
	return false
}

func (i *FileInfo) detectType(modify, saveContent, readHeader bool, calcImgRes bool) error {
	if IsNamedPipe(i.Mode) {
		i.Type = "blob"
		return nil
	}

	ext := strings.ToLower(i.Extension)
	name := strings.ToLower(i.Name)

	if isKnownImageExtension(ext) {
		i.Type = "image"
		if calcImgRes {
			resolution, err := calculateImageResolution(i.Fs, i.Path)
			if err != nil {
				log.Printf("Error calculating image resolution: %v", err)
			} else {
				i.Resolution = resolution
			}
		}
		return nil
	}

	if isKnownVideoExtension(ext) {
		i.Type = "video"
		i.detectSubtitles()
		return nil
	}

	if isKnownAudioExtension(ext) {
		i.Type = "audio"
		return nil
	}

	if ext == ".pdf" {
		i.Type = "pdf"
		return nil
	}

	if ext == ".epub" {
		i.Type = "epub"
		return nil
	}

	if isKnownTextExtension(ext, name) && i.Size <= 25*1024*1024 { // Up to 25 MB for text/code
		i.Type = "text"
		if !modify {
			i.Type = "textImmutable"
		}
		if saveContent {
			afs := &afero.Afero{Fs: i.Fs}
			content, err := afs.ReadFile(i.Path)
			if err != nil {
				return err
			}
			i.Content = string(content)
		}
		return nil
	}

	mimetype := mime.TypeByExtension(i.Extension)

	var buffer []byte
	if readHeader {
		buffer = i.readFirstBytes()

		if mimetype == "" {
			mimetype = http.DetectContentType(buffer)
		}
	}

	switch {
	case strings.HasPrefix(mimetype, "video"):
		i.Type = "video"
		i.detectSubtitles()
		return nil
	case strings.HasPrefix(mimetype, "audio"):
		i.Type = "audio"
		return nil
	case strings.HasPrefix(mimetype, "image"):
		i.Type = "image"
		if calcImgRes {
			resolution, err := calculateImageResolution(i.Fs, i.Path)
			if err != nil {
				log.Printf("Error calculating image resolution: %v", err)
			} else {
				i.Resolution = resolution
			}
		}
		return nil
	case strings.HasSuffix(mimetype, "pdf"):
		i.Type = "pdf"
		return nil
	case (strings.HasPrefix(mimetype, "text") || (len(buffer) > 0 && !isBinary(buffer))) && i.Size <= 25*1024*1024:
		i.Type = "text"

		if !modify {
			i.Type = "textImmutable"
		}

		if saveContent {
			afs := &afero.Afero{Fs: i.Fs}
			content, err := afs.ReadFile(i.Path)
			if err != nil {
				return err
			}

			i.Content = string(content)
		}
		return nil
	default:
		i.Type = "blob"
	}

	return nil
}

func calculateImageResolution(fSys afero.Fs, filePath string) (*ImageResolution, error) {
	file, err := fSys.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			log.Printf("Failed to close file: %v", cErr)
		}
	}()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	return &ImageResolution{
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

func (i *FileInfo) readFirstBytes() []byte {
	reader, err := i.Fs.Open(i.Path)
	if err != nil {
		log.Print(err)
		i.Type = "blob"
		return nil
	}
	defer reader.Close()

	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Print(err)
		i.Type = "blob"
		return nil
	}

	return buffer[:n]
}

func (i *FileInfo) detectSubtitles() {
	if i.Type != "video" {
		return
	}

	i.Subtitles = []string{}
	ext := filepath.Ext(i.Path)

	// detect multiple languages. Base*.vtt
	parentDir := strings.TrimRight(i.Path, i.Name)
	var dir []os.FileInfo
	if len(i.currentDir) > 0 {
		dir = i.currentDir
	} else {
		var err error
		dir, err = afero.ReadDir(i.Fs, parentDir)
		if err != nil {
			return
		}
	}

	base := strings.TrimSuffix(i.Name, ext)
	for _, f := range dir {
		// load all supported subtitles from subs directories
		// should cover all instances of subtitle distributions
		// like tv-shows with multiple episodes in single dir
		if f.IsDir() && reSubDirs.MatchString(f.Name()) {
			subsDir := path.Join(parentDir, f.Name())
			i.loadSubtitles(subsDir, base, true)
		} else if isSubtitleMatch(f, base) {
			i.addSubtitle(path.Join(parentDir, f.Name()))
		}
	}
}

func (i *FileInfo) loadSubtitles(subsPath, baseName string, recursive bool) {
	dir, err := afero.ReadDir(i.Fs, subsPath)
	if err == nil {
		for _, f := range dir {
			if isSubtitleMatch(f, "") {
				i.addSubtitle(path.Join(subsPath, f.Name()))
			} else if f.IsDir() && recursive && strings.HasPrefix(f.Name(), baseName) {
				subsDir := path.Join(subsPath, f.Name())
				i.loadSubtitles(subsDir, baseName, false)
			}
		}
	}
}

func IsSupportedSubtitle(fileName string) bool {
	return reSubExts.MatchString(fileName)
}

func isSubtitleMatch(f fs.FileInfo, baseName string) bool {
	return !f.IsDir() && strings.HasPrefix(f.Name(), baseName) &&
		IsSupportedSubtitle(f.Name())
}

func (i *FileInfo) addSubtitle(fPath string) {
	i.Subtitles = append(i.Subtitles, fPath)
}

func (i *FileInfo) readListing(checker rules.Checker, readHeader bool, calcImgRes bool) error {
	dir, err := readDir(i.Fs, i.Path)
	if err != nil {
		return err
	}

	listing := &Listing{
		Items:    []*FileInfo{},
		NumDirs:  0,
		NumFiles: 0,
	}

	for _, f := range dir {
		name := f.Name()
		if name == ".filebrowser_trash" {
			continue
		}
		fPath := path.Join(i.Path, name)

		if !checker.Check(fPath) {
			continue
		}

		isSymlink, isInvalidLink := false, false
		if IsSymlink(f.Mode()) {
			isSymlink = true
			// It's a symbolic link. We try to follow it. The scoped filesystem
			// refuses to dereference a link whose target escapes the scope
			// (permission error); such a link is omitted from the listing
			// entirely so it cannot leak the target's metadata. Any other
			// failure means a broken link, which we surface as an invalid link
			// rather than the target's information.
			info, err := i.Fs.Stat(fPath)
			switch {
			case err == nil:
				f = info
			case errors.Is(err, os.ErrPermission):
				continue
			default:
				isInvalidLink = true
			}
		}

		file := &FileInfo{
			Fs:         i.Fs,
			Name:       name,
			Size:       f.Size(),
			ModTime:    f.ModTime(),
			Mode:       f.Mode(),
			IsDir:      f.IsDir(),
			IsSymlink:  isSymlink,
			Extension:  filepath.Ext(name),
			Path:       fPath,
			currentDir: dir,
		}

		if !file.IsDir && strings.HasPrefix(mime.TypeByExtension(file.Extension), "image/") && calcImgRes {
			resolution, err := calculateImageResolution(file.Fs, file.Path)
			if err != nil {
				log.Printf("Error calculating resolution for image %s: %v", file.Path, err)
			} else {
				file.Resolution = resolution
			}
		}

		if file.IsDir {
			listing.NumDirs++
		} else {
			listing.NumFiles++

			if isInvalidLink {
				file.Type = "invalid_link"
			} else {
				err := file.detectType(true, false, readHeader, calcImgRes)
				if err != nil {
					return err
				}
			}
		}

		listing.Items = append(listing.Items, file)
	}

	i.Listing = listing
	return nil
}

func readDir(afs afero.Fs, dirname string) ([]os.FileInfo, error) {
	dir, err := afero.ReadDir(afs, dirname)
	if err == nil {
		return dir, nil
	}

	dir, fallbackErr := readDirNames(afs, dirname)
	if fallbackErr != nil {
		return nil, err
	}

	return dir, nil
}

func readDirNames(afs afero.Fs, dirname string) ([]os.FileInfo, error) {
	file, err := afs.Open(dirname)
	if err != nil {
		return nil, err
	}

	names, err := file.Readdirnames(-1)
	if closeErr := file.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}

	sort.Strings(names)
	dir := make([]os.FileInfo, 0, len(names))
	for _, name := range names {
		fPath := path.Join(dirname, name)
		info, err := lstatIfPossible(afs, fPath)
		if err != nil {
			log.Printf("Skipping inaccessible file %s: %v", fPath, err)
			continue
		}

		dir = append(dir, info)
	}

	return dir, nil
}

func lstatIfPossible(afs afero.Fs, name string) (os.FileInfo, error) {
	if lstaterFs, ok := afs.(afero.Lstater); ok {
		info, _, err := lstaterFs.LstatIfPossible(name)
		return info, err
	}

	return afs.Stat(name)
}
