package main

import (
  "flag"
  "fmt"
  "io"
  "log"
  "net/http"
  "os"
  "path/filepath"
)

// 缓存目录
const cacheDir = "./cache"

// 默认端口
const defaultPort = 10010

// 代理处理函数
func proxyHandler(w http.ResponseWriter, r *http.Request) {
  // 构建目标 URL
  targetURL := "https://repo.maven.apache.org/maven2" + r.RequestURI

  // 构建缓存文件路径
  cacheFilePath := getCacheFilePath(r.RequestURI)

  // 检查缓存是否存在
  if cacheExists(cacheFilePath) {
    // 如果缓存存在，直接从缓存读取
    log.Printf("Serving from cache: %s", cacheFilePath)
    http.ServeFile(w, r, cacheFilePath)
    return
  }

  // 如果缓存不存在，则从远程服务器获取资源
  log.Printf("Fetching from remote: %s", targetURL)
  req, err := http.NewRequest(r.Method, targetURL, r.Body)
  if err != nil {
    http.Error(w, "Error creating request:"+targetURL+",err:"+err.Error(), http.StatusInternalServerError)
    return
  }

  // 复制请求头
  req.Header = r.Header
  // 移除可能导致问题的请求头
  req.Header.Del("Host")
  req.Header.Del("Connection")
  req.Header.Del("Upgrade")
  req.Header.Del("Keep-Alive")
  req.Header.Del("Proxy-Connection")

  // 强制使用 HTTP/1.1
  transport := &http.Transport{
    ForceAttemptHTTP2: false,
  }
  client := &http.Client{Transport: transport}
  resp, err := client.Do(req)
  if err != nil {
    http.Error(w, "Error forwarding request:"+targetURL+",err:"+err.Error(), http.StatusInternalServerError)
    return
  }
  defer resp.Body.Close()

  // 将远程服务的响应头复制到本地响应
  for key, values := range resp.Header {
    for _, value := range values {
      w.Header().Add(key, value)
    }
  }

  // 设置状态码
  w.WriteHeader(resp.StatusCode)

  // 将远程服务的响应体写入本地响应并保存到缓存
  err = saveToCacheAndWriteResponse(w, resp.Body, cacheFilePath)
  if err != nil {
    log.Printf("Error saving to cache: %v", err)
  }
}

// 获取缓存文件路径
func getCacheFilePath(requestURI string) string {
  // 组合缓存路径
  fullPath := filepath.Join(cacheDir, requestURI)

  // 获取文件夹路径
  dirPath := filepath.Dir(fullPath)

  // 创建文件夹（如果不存在）
  err := os.MkdirAll(dirPath, os.ModePerm)
  if err != nil {
    log.Printf("Error creating directory: %v", err)
  }

  return fullPath
}

// 检查缓存文件是否存在
func cacheExists(filePath string) bool {
  _, err := os.Stat(filePath)
  return err == nil
}

// 保存响应到缓存并写入到响应中
func saveToCacheAndWriteResponse(w http.ResponseWriter, src io.Reader, cacheFilePath string) error {
  // 创建缓存文件
  cacheFile, err := os.Create(cacheFilePath)
  if err != nil {
    return err
  }
  defer cacheFile.Close()

  // 创建一个多路输出，将数据同时写入缓存文件和响应
  multiWriter := io.MultiWriter(w, cacheFile)
  _, err = io.Copy(multiWriter, src)
  return err
}

func main() {
  // 使用 flag 包解析命令行参数
  port := flag.Int("port", defaultPort, "Port to run the proxy server on")
  flag.Parse()

  // 将所有请求转发到 proxyHandler
  http.HandleFunc("/", proxyHandler)

  // 监听指定端口
  address := fmt.Sprintf(":%d", *port)
  log.Printf("Starting HTTP proxy with caching on %s", address)
  err := http.ListenAndServe(address, nil)
  if err != nil {
    log.Fatalf("Failed to start server: %v", err)
  }
}
