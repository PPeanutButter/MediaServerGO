package main

import (
	"os"
	"path"
	"path/filepath"
)

//var diskWeight = []int{3, 3, 1, 1}

const (
	DiskManagerDir   = "F:/_disk_manager_dir"
	Root             = DiskManagerDir
	PreviewCacheDir  = "_preview_cache_dir"
	BookmarkCacheDir = "_bookmark_cache_dir"
	BitRateCacheFile = "bit_rate_cache.json"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

type DiskManager struct {
	MountPoints []string
	DiskNames   []string
}

func NewDiskManager(mountPoints []string) *DiskManager {
	//init
	if PathExists(DiskManagerDir) {
		_ = os.RemoveAll(DiskManagerDir)
	}
	if err := os.MkdirAll(DiskManagerDir, 0777); err != nil { //os.ModePerm
		panic("init DiskManager Failed, check permission:" + err.Error())
	}
	if !PathExists(PreviewCacheDir) {
		_ = os.MkdirAll(PreviewCacheDir, 0777)
	}
	if !PathExists(BookmarkCacheDir) {
		_ = os.MkdirAll(BookmarkCacheDir, 0777)
	}
	var diskNames = make([]string, 0, len(mountPoints))
	for _, mountPoint := range mountPoints {
		// ToSlash Converts \\ to /
		diskNames = append(diskNames, path.Base(filepath.ToSlash(mountPoint)))
		err := os.Symlink(mountPoint, path.Join(DiskManagerDir, path.Base(filepath.ToSlash(mountPoint))))
		if err != nil {
			panic(err)
		}
	}
	//done
	return &DiskManager{
		MountPoints: mountPoints,
		DiskNames:   diskNames,
	}
}

func (m *DiskManager) listDir(relativePath string) ([]string, error) {
	var roots []string
	if relativePath == "/" {
		roots = m.DiskNames
	} else {
		roots = []string{relativePath}
	}
	var result []string
	for _, root := range roots {
		//检查软连接是否存在并防止目录穿越
		if iRoot := path.Join(DiskManagerDir, root); PathExists(iRoot) && isAllowedPath(iRoot) {
			dirs, err := os.ReadDir(iRoot)
			if err != nil {
				return nil, err
			}
			tmp := make([]string, len(dirs))
			for i, dir := range dirs {
				tmp[i] = path.Join(root, dir.Name())
			}
			result = append(result, tmp...)
		}
	}
	return result, nil
}
