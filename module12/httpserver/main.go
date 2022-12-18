package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/Crytoll/homework/module12/httpserver/metrics"
)

// 定义config类型
// 类型里的属性，全是配置文件里的属性
type Config struct {
	filename       string
	lastModifyTime int64
	Http           Http `yaml:"http"`
	Log            Log  `yaml:"log"`
}

type Http struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type Log struct {
	Level string `yaml:"level"`
}

// 定义全局变量
var (
	log        *logrus.Logger
	config     *Config
	configLock = new(sync.RWMutex)
)

// 定义健康检测
func healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}

// 定义 web 根
func webRoot(w http.ResponseWriter, r *http.Request) {
	log.Debug("entering web root handler")
	timer := metrics.NewTimer()
	defer timer.ObserveTotal()
	user := r.URL.Query().Get("user")
	delay := randInt(0, 2000)
	time.Sleep(time.Microsecond * time.Duration(delay))

	// 请求 header
	h := r.Header
	// 获取环境变量 VERSION
	version := os.Getenv("VERSION")
	// header 中加入环境变量
	h.Add("VERSION", version)
	// 将请求 header 转成 返回 header
	for k, v := range h {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	clientIP := ClientIP(r)
	// 记录日志
	log.Infof("agent ip: %v, status code: %d", clientIP, 200)

	if user != "" {
		io.WriteString(w, fmt.Sprintf("hello [%s]\n", user))
	} else {
		io.WriteString(w, "hello [stranger]\n")
	}

	io.WriteString(w, "===================Details of the http request header:============\n")
	for k, v := range r.Header {
		io.WriteString(w, fmt.Sprintf("%s=%s\n", k, v))
	}

	log.Infof("Respond in %d ms", delay)
}

// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

func ExitFunc() {
	fmt.Println("开始退出...")
	fmt.Println("执行清理...")
	fmt.Println("结束退出...")
	os.Exit(0)
}

func startHttpServer(listenServer string) {
	metrics.Register()
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/", webRoot)
	mux.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(listenServer, mux); err != nil {
		log.Fatalf("start http server failed, error: %s\n", err.Error())
	}
}

func LogInit(logLevel string) {
	log = logrus.New()

	// 设置日志输出
	log.Out = os.Stdout

	level := logrus.InfoLevel
	switch {
	case logLevel == "debug":
		level = logrus.DebugLevel
	case logLevel == "info":
		level = logrus.InfoLevel
	case logLevel == "warn":
		level = logrus.WarnLevel
	case logLevel == "error":
		level = logrus.ErrorLevel
	default:
		level = logrus.InfoLevel
	}
	// log.Formatter = &logrus.JSONFormatter{}
	log.SetLevel(level)
}

// 获取可执行文件目录，便于读取配置文件
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Get current path fail\n", err.Error())
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func (config *Config) reload() {
	// 定时器
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		// 打开文件
		func() {
			f, err := os.Open(GetConfig().filename)
			if err != nil {
				log.Fatalf("open file error:%s\n", err)
				return
			}
			defer f.Close()

			fileInfo, err := f.Stat()
			if err != nil {
				log.Fatalf("stat file error:%s\n", err)
				return
			}
			// 或取当前文件修改时间
			curModifyTime := fileInfo.ModTime().Unix()
			if curModifyTime > GetConfig().lastModifyTime {
				// 重新解析时，要考虑应用程序正在读取这个配置因此应该加锁
				// 使用了 configLock　全局锁
				log.Info("cfg change, load new cfg ...")
				loadConfig()
				GetConfig().lastModifyTime = curModifyTime
			}
		}()
	}
}

func loadConfig() bool {
	log.Println("Load cfg ...")

	f, err := ioutil.ReadFile(config.filename)
	if err != nil {
		log.Fatalln("load config error: ", err)
		return false
	}

	//不同的配置规则，解析复杂度不同
	temp := new(Config)
	err = yaml.Unmarshal(f, &temp)
	if err != nil {
		log.Fatalln("Para config failed: ", err)
		return false
	}

	temp.filename = GetConfig().filename
	temp.lastModifyTime = GetConfig().lastModifyTime
	log.Debugf("now cfg:%#v\n", temp)
	configLock.Lock()
	config = temp
	configLock.Unlock()
	return true
}

// GetConfig
func GetConfig() *Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func IsIpv4(ipv4 string) bool {
	address := net.ParseIP(ipv4)
	if address != nil {
		log.Infof("%s is a legal ipv4 address\n", ipv4)
		return true
	} else {
		log.Infof("%s is not a legal ipv4 address\n", ipv4)
		return false
	}
}

func CheckPortRange(port int) bool {
	if 1 <= port && port <= 65535 {
		return true
	}
	return false
}

func CheckConfig() (listenServer, logLevel string) {
	var allConfig = GetConfig()
	var httpConfig = allConfig.Http
	var httpPort = httpConfig.Port
	var httpHost = httpConfig.Host

	if !IsIpv4(httpHost) {
		httpHost = "0.0.0.0"
	}

	if port, err := strconv.Atoi(httpPort); err == nil {
		if !CheckPortRange(port) {
			httpPort = "8080"
		}
	}
	listenServer = httpHost + ":" + httpPort
	logLevel = allConfig.Log.Level
	return listenServer, logLevel
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func init() {
	// 定义默认日志级别
	LogInit("info")

	config = new(Config)
	pwd := getCurrentDirectory()
	config.filename = pwd + "/conf/config.yaml"
	if !loadConfig() {
		os.Exit(1)
	}
	go config.reload()
}

func main() {
	listenServer, logLevel := CheckConfig()

	// 重新载入日志级别
	LogInit(logLevel)

	// 创建监听退出chan
	c := make(chan os.Signal)
	// 监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Infoln("退出", s)
				ExitFunc()
			case syscall.SIGUSR1:
				log.Infoln("usr1", s)
			case syscall.SIGUSR2:
				log.Infoln("usr2", s)
			default:
				log.Infoln("other", s)
			}
		}
	}()

	startHttpServer(listenServer)
}
