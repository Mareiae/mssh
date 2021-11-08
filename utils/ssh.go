package utils

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"path"
)


//将文件拷贝到远程服务器
func CopyFilesToHost(ftp *sftp.Client,config *ConfigInfo)error{
	var (
		fp 	*os.File
		err error
	)

	//压缩文件
	curDir,_ := os.Getwd()
	if fp,err = os.Create(path.Join(curDir,config.Rcp.ZipName));err != nil{
		return err
	}
	defer fp.Close()
	if err = config.Zip(config.Rcp.SrcDir,fp);err != nil{
		return err
	}

	//远程文件上传
	srcFile, err := os.Open(path.Join(curDir,config.Rcp.ZipName))
	if err != nil {
		return err
	}
	defer srcFile.Close()
	var remoteFileName = path.Base(config.Rcp.ZipName)
	dstFile, err := ftp.Create(path.Join(config.Rcp.DestDir, remoteFileName))
	if err != nil {
		return err
	}
	defer dstFile.Close()
	bytes, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return err
	}
	dstFile.Write(bytes)
	return nil
}

//远程服务器批命令处理
func RemoteHostShellBatching(sshClient *ssh.Client,config *ConfigInfo)error{
	var (
		sshSession *ssh.Session
		err error
		outBytes []byte
	)

	for _,command := range config.Shell{
		//创建session会话
		if sshSession, err = sshClient.NewSession(); err != nil {
			return err
		}
		defer sshSession.Close()
		log.Println("开始执行命令："+command)
		if outBytes,err = sshSession.Output(command);err != nil{
			return err
		}
		fmt.Println(string(outBytes))
	}
	return nil
}