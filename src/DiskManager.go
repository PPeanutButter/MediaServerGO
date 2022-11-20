package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"log"
	"os"
	"path"
	"path/filepath"
)

var DiskWeight = []uint64{2, 2, 1, 1}

const (
	DiskManagerDir   = "_disk_manager_dir"
	Root             = DiskManagerDir
	PreviewCacheDir  = "_preview_cache_dir"
	BookmarkCacheDir = "_bookmark_cache_dir"
	Ass2SrtCacheDir  = "_a2s_cache_dir"
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
	if !PathExists(Ass2SrtCacheDir) {
		_ = os.MkdirAll(Ass2SrtCacheDir, 0777)
	}
	var diskNames = make([]string, 0, len(mountPoints))
	for _, mountPoint := range mountPoints {
		// ToSlash Converts \\ to /
		diskNames = append(diskNames, path.Base(filepath.ToSlash(mountPoint)))
		err := os.Symlink(mountPoint, path.Join(DiskManagerDir, path.Base(filepath.ToSlash(mountPoint))))
		if err != nil {
			log.Println("NewDiskManager", "初始化DiskManager-创建软连接失败", err)
			panic(err)
		}
	}
	//done
	return &DiskManager{
		MountPoints: mountPoints,
		DiskNames:   diskNames,
	}
}

func (this *DiskManager) listDir(relativePath string) ([]string, error) {
	var roots []string
	if relativePath == "/" {
		roots = this.DiskNames
	} else {
		roots = []string{relativePath}
	}
	var result []string
	for _, root := range roots {
		//检查软连接是否存在并防止目录穿越
		if iRoot := path.Join(DiskManagerDir, root); PathExists(iRoot) && isAllowedPath(iRoot, Root) {
			dirs, err := os.ReadDir(iRoot)
			if err != nil {
				log.Println("DiskManager.listDir", "系统调用失败", err)
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

func (this *DiskManager) getMaxAvailableDisk(folder string) string {
	var maxSize uint64 = 0
	var maxDisk = ""
	for i, mountPoint := range this.MountPoints {
		if PathExists(path.Join(mountPoint, folder)) {
			return path.Base(mountPoint)
		}
		stat, err := disk.Usage(mountPoint)
		if err != nil {
			log.Println("DiskManager.getMaxAvailableDisk", "获取磁盘大小失败", err)
			return path.Base(mountPoint)
		}
		free := stat.Free * DiskWeight[i]
		if free > maxSize {
			maxSize = free
			maxDisk = path.Base(mountPoint)
		}
	}
	return maxDisk
}

func (this *DiskManager) string() string {
	var re = ""
	for _, mountPoint := range this.MountPoints {
		stat, err := disk.Usage(mountPoint)
		if err == nil {
			re = re + fmt.Sprintf("%s\t%dused\t%dfree\n", mountPoint, stat.Used, stat.Free)
		}
	}
	return re
}
