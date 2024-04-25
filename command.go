package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	minify "github.com/MasaeProject/StaticDeployment/plugin/minify"
	zhcodeconv "github.com/MasaeProject/StaticDeployment/plugin/zhcodeconv"
)

func runExec(run Run, srcPath string, names Names) bool {
	var cmds [][]string = [][]string{}
	if osName == "windows" && run.Windows != nil && len(*run.Windows) > 0 {
		cmds = *run.Windows
	} else if osName == "linux" && run.Linux != nil && len(*run.Linux) > 0 {
		cmds = *run.Linux
	} else if osName == "darwin" && run.Darwin != nil && len(*run.Darwin) > 0 {
		cmds = *run.Darwin
	} else if run.Default != nil && len(*run.Default) > 0 {
		cmds = *run.Default
	} else {
		log.Println("错误: 未找到适用于当前操作系统的命令。")
		return false
	}
	srcPath = CleanPath(srcPath)
	var cmdsLen int = len(cmds)
	totalCMD++
	var dir string = ""
	for cmdsI, cmd := range cmds {
		var noEmbCmd = false
		var cmdLen int = len(cmd)
		for cmdI, c := range cmd {
			var nKey string = "$SRC"
			c = CleanPath(c)
			var pathArr []string = strings.Split(srcPath, string(filepath.Separator))
			var pathArrLen = len(pathArr)
			var fileFullName string = pathArr[pathArrLen-1]
			var dirPath string = ""
			if pathArrLen > 1 {
				fileFullName = pathArr[pathArrLen-1]
				pathArr = pathArr[:pathArrLen-1]
				dirPath = strings.Join(pathArr, string(filepath.Separator))
			}
			var fileNameArr []string = strings.Split(fileFullName, ".")
			var fileNameArrLen = len(fileNameArr)
			var extName string = fileNameArr[fileNameArrLen-1]
			var fileName string = fileNameArr[0]
			if fileNameArrLen > 2 {
				fileNameArr = fileNameArr[:fileNameArrLen-1]
				fileName = strings.Join(fileNameArr, ".")
			} else if fileNameArrLen == 1 {
				extName = ""
			}
			nKey = "$SOLUTION"
			if strings.Contains(c, nKey) {
				c = strings.ReplaceAll(c, nKey, names.Solution)
			}
			nKey = "$PROJECT"
			if strings.Contains(c, nKey) {
				c = strings.ReplaceAll(c, nKey, names.Project)
			}
			nKey = "$JOBNAME"
			if strings.Contains(c, nKey) {
				c = strings.ReplaceAll(c, nKey, names.Replace)
			}
			nKey = "$SRCFILE"
			if strings.Contains(c, nKey) {
				c = strings.ReplaceAll(c, nKey, fileFullName)
			}
			nKey = "$SRCNAME"
			if strings.Contains(c, nKey) {
				c = strings.ReplaceAll(c, nKey, fileName)
			}
			nKey = "$SRCEXT"
			if strings.Contains(c, nKey) {
				if IsDirectory(srcPath) == 0 {
					c = strings.ReplaceAll(c, nKey, extName)
				} else {
					c = strings.ReplaceAll(c, nKey, "")
				}
			}
			nKey = "$SRCDIRNAME"
			if strings.Contains(c, nKey) {
				c = strings.ReplaceAll(c, nKey, pathArr[len(pathArr)-1])
			}
			nKey = "$SRCDIR"
			if strings.Contains(c, nKey) {
				c = strings.ReplaceAll(c, nKey, dirPath)
			}
			nKey = "$SRC"
			if strings.Contains(c, nKey) {
				c = strings.ReplaceAll(c, nKey, srcPath)
			}
			// fmt.Println("SRCFILE", fileFullName, "SRCNAME", fileName, "SRCEXT", extName, "SRCDIRNAME", pathArr[len(pathArr)-1], "SRCDIR", dirPath, "SRC", srcPath)
			cmd[cmdI] = CleanPath(c)
		}

		if len(cmd) == 0 {
			continue
		}
		log.Printf("运行命令 %d / %d : %s\n", cmdsI+1, cmdsLen, strings.Join(cmd, " "))
		var err error = nil
		var isOK bool = true
		switch cmd[0] {
		case "$CMDDIR":
			if cmdLen == 1 {
				dir = ""
			} else if cmdLen >= 2 {
				dir = cmd[1]
			}
		case "$BAK":
			if cmdLen == 1 || (cmdLen == 2 && len(cmd[1]) == 0) {
				isOK = backup(srcPath, names)
			}
			if cmdLen >= 2 {
				isOK = backup(cmd[1], names)
			}
		case "$RES":
			var resCmd Names = names
			if cmdLen >= 2 && len(cmd[1]) > 0 {
				resCmd.Solution = cmd[1]
			}
			if cmdLen >= 3 && len(cmd[2]) > 0 {
				resCmd.Project = cmd[2]
			}
			if cmdLen >= 4 && len(cmd[3]) > 0 {
				resCmd.Replace = cmd[3]
			}
			if cmdLen == 2 {
				isOK = restoreSolution(resCmd.Solution)
			} else if cmdLen >= 3 {
				isOK = restoreProject(resCmd.Solution, resCmd.Project)
			} else if cmdLen == 1 || cmdLen == 4 {
				isOK = restoreJob(resCmd.Solution, resCmd.Project, resCmd.Replace)
			}
		case "$CP":
			if cmdLen >= 3 {
				err = Copy(cmd[1], cmd[2])
			} else if cmdLen == 2 {
				err = Copy(srcPath, cmd[1])
			}
		case "$MV":
			if cmdLen >= 3 {
				err = Move(cmd[1], cmd[2])
			} else if cmdLen == 2 {
				err = Move(srcPath, cmd[1])
			}
		case "$SMV":
			if cmdLen >= 3 {
				err = MoveSecure(cmd[1], cmd[2])
			} else if cmdLen == 2 {
				err = MoveSecure(srcPath, cmd[1])
			}
		case "$RM":
			err = Remove(cmd[1])
		case "$SRM":
			err = RemoveSecure(cmd[1])
		case "$REN":
			if cmdLen >= 3 {
				err = RenamePath(cmd[1], cmd[2])
			} else if cmdLen == 2 {
				err = RenamePath(srcPath, cmd[1])
			}
		case "$ZHCODECONV":
			var lenCh [2]int
			lenCh, err = zhcodeconv.InitWithCmd(cmd, srcPath)
			log.Printf("非 ASCII 变量和函数名转换: %s ( %d -> %d )\n", srcPath, lenCh[0], lenCh[1])
		case "$MINIFY":
			var lenCh [2]int
			lenCh, err = minify.InitWithCmd(cmd, srcPath)
			log.Printf("代码压缩: %s ( %d -> %d )\n", srcPath, lenCh[0], lenCh[1])
		default:
			noEmbCmd = true
		}
		if err != nil {
			log.Printf("错误: 文件操作 %s 失败: %s\n", cmd[0], err)
			return false
		} else if !isOK {
			return false
		}

		if noEmbCmd {
			if !runCMD(cmd, dir) {
				return false
			}
		}
	}
	return true
}

func runCMD(cmd []string, dir string) bool {
	totalEXE++
	ex := exec.Command(cmd[0], cmd[1:]...)
	if len(dir) > 0 {
		ex.Dir = dir
	}
	ex.Stdout = os.Stdout
	ex.Stderr = os.Stderr
	err := ex.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			fmt.Printf("错误: 命令退出代码: %d\n", status.ExitStatus())
			return false
		}
	} else if err != nil {
		log.Println("错误：执行命令失败：", err)
		return false
	} else {
		log.Println("命令运行成功。")
		return true
	}
	return true
}
