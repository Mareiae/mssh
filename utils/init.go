package utils

import (
	"archive/zip"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type (
	//远程服务器主机信息
	ServerInfo struct {
		Host     string `yaml:'host'`
		Port     string `yaml:'port'`
		User     string `yaml:'user'`
		Password string `yaml:'password'`
	}
	//远程服务器sftp操作配置
	RcpInfo struct {
		SrcDir  string `yaml:'srcdir'`
		DestDir string `yaml:'destdir'`
		ZipName string `yaml:'zipname'`
	}
	//配置文件信息
	ConfigInfo struct {
		Server ServerInfo `yaml:'server'`
		Shell  []string   `yaml:'shell'`
		Rcp    RcpInfo    `yaml:'rcp'`
	}
)

//读取配置文件
func (c *ConfigInfo)LoadConfigInfo()error{
	curdir,_ := os.Getwd()
	file, err := ioutil.ReadFile(filepath.Join(curdir,"/config.yaml"))
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, c)
	if err != nil {
		return err
	}
	return nil
}

//ssh 连接远程服务器
func (c *ConfigInfo)HostDial() (*ssh.Client,*sftp.Client, error) {
	var (
		auth   []ssh.AuthMethod
		addr   string
		clientConfig *ssh.ClientConfig
		sshClient *ssh.Client
		sftpClient *sftp.Client
		err   error
	)
	// 将密码穿到验证方法切片里
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(c.Server.Password))
	//配置项
	clientConfig = &ssh.ClientConfig{
		User: c.Server.User,
		Auth: auth,
		Timeout: 30 * time.Second,
		//这各参数是验证服务端的，返回nil可以不做验证，如果不设置会报错
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	//连接ip和端口
	addr = fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
	//通过tcp协议,连接ssh
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil,nil, err
	}

	//创建sftp服务对象
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil,nil, err
	}
	//返回sftp服务对象
	return sshClient,sftpClient, nil
}

//转linux路径
func toLinuxPath(basePath string) string {
	return strings.ReplaceAll(basePath, "\\", "/")
}

//文件压缩
func (c *ConfigInfo)Zip(fp string, w io.ReadWriter) error {
	archive := zip.NewWriter(w)
	defer archive.Close()

	linuxFilePath := toLinuxPath(fp)
	filepath.Walk(linuxFilePath, func(path string, info os.FileInfo, err error) error {

		var linuxPath = toLinuxPath(path)
		if linuxPath == linuxFilePath {
			return nil
		}

		header, _ := zip.FileInfoHeader(info)
		header.Name = strings.TrimPrefix(linuxPath, linuxFilePath+"/")

		if info.IsDir() {
			header.Name += `/`
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}
		// 创建：压缩包头部信息
		writer, _ := archive.CreateHeader(header)
		if !info.IsDir() {
			file, _ := os.Open(linuxPath)
			defer file.Close()
			io.Copy(writer, file)
		}
		return nil
	})

	return nil
}

//文件解压
func (c *ConfigInfo)Unzip(basePath string, r io.Reader) error {
	/* 创建属于解压的缓存目录 */
	var dir = path.Join(os.TempDir(), "zip")
	os.MkdirAll(dir, 0666)

	/* 创建解压缓存文件 */
	f, e := ioutil.TempFile(dir, "zip")
	if nil != e { return e }
	defer func() {
		f.Close()
		os.RemoveAll(f.Name())
	}()

	_, e = io.Copy(f, r)
	if nil != e { return e }

	return unzip(basePath, f)
}

func unzip(basePath string, f *os.File) error {
	var reader *zip.Reader
	var stat, _ = f.Stat()
	reader, e := zip.NewReader(f, stat.Size())
	if nil != e { return e }
	os.MkdirAll(basePath, 0666) // 确保解压目录存在

	for _, info := range reader.File {
		var fp = toLinuxPath(path.Join(basePath, info.Name))
		if info.FileInfo().IsDir() {
			if e := os.MkdirAll(fp, info.FileInfo().Mode()); nil != e { return e }
			continue
		}

		readcloser, e := info.Open()
		if nil != e { return e }

		b, e := ioutil.ReadAll(readcloser)
		if nil != e { return e }
		readcloser.Close()

		if e := ioutil.WriteFile(fp, b, info.FileInfo().Mode()); nil != e { return e }
	}
	return nil
}
