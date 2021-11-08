package main

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"mssh/utils"
	"os"
)

func main(){
	var (
		config 			utils.ConfigInfo
		sshClient   	*ssh.Client
		sftpClient  	*sftp.Client
		err         	error
		isFilesCopy 	bool
		isExecCommand 	bool
	)

	//解析命令行参数
	if len(os.Args) == 2{
		switch os.Args[1] {
		case "-a":isExecCommand = true;isFilesCopy = true; break
		case "-c":isExecCommand = false;isFilesCopy = true;break
		case "-e":isExecCommand = true;isFilesCopy = false;break
		case "-help":{
			fmt.Println("参数详情：")
			fmt.Println("\t-a：执行内容：远程文件拷贝；远程批命令执行")
			fmt.Println("\t-c：执行内容：远程文件拷贝")
			fmt.Println("\t-e：执行内容：远程批命令执行")
		};return
		default:
			isExecCommand = true;isFilesCopy = true;break
		}
	}else if len(os.Args) == 1{
		fmt.Println("缺少参数，请使用-hlep查看")
		return
	}

	//读取配置文件
	if err = config.LoadConfigInfo();err != nil{
		log.Println("配置文件加载失败...：",err.Error())
		utils.Logs(utils.Error,"配置文件加载失败...","ssh->HostConn")
		return
	}

	//连接远程服务器
	if sshClient,sftpClient,err  = config.HostDial();err != nil{
		log.Println("远程服务器连接失败...：",err.Error())
		utils.Logs(utils.Error,"远程服务器连接失败...","ssh->HostConn")
		return
	}
	defer sshClient.Close()
	defer sftpClient.Close()
	log.Println("远程服务器连接成功...")

	//将指定文件拷贝到远程服务器
	if isFilesCopy == true{
		if err  = utils.CopyFilesToHost(sftpClient,&config);err != nil{
			log.Println("文件拷贝失败...：",err.Error())
			utils.Logs(utils.Error,"文件拷贝失败...","ssh->HostConn")
			return
		}
		log.Println("文件拷贝至远程服务器成功...")
	}

	//远程服务器shell命令批处理
	if isExecCommand == true{
		if err  = utils.RemoteHostShellBatching(sshClient,&config);err != nil{
			log.Println("远程服务器批命令执行失败...：",err.Error())
			utils.Logs(utils.Error,"远程服务器批命令执行失败...","ssh->HostConn")
			return
		}
		log.Println("远程服务器批命令执行完成...")
	}
}
