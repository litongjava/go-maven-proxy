# go-maven-proxy

`go-maven-proxy` 是一个用 Golang 编写的轻量级 Maven 代理程序。它接收客户端的 HTTP 请求，将其转发到 Maven 中央仓库（`https://repo.maven.apache.org`），并将响应返回给客户端。为了提高效率和减少网络请求次数，它还实现了本地缓存功能，可以在本地保存远程资源供后续请求使用。
## 功能特性

- **HTTP 到 HTTPS 转发**：将客户端的 HTTP 请求转发到 Maven 中央仓库的 HTTPS 地址，确保数据传输的安全性。
- **本地缓存**：将从远程服务器获取的资源缓存到本地，以便在后续请求相同资源时直接从本地读取，提高响应速度。
- **目录结构保持**：缓存文件的路径保持与原始请求一致，在本地磁盘上保留请求的目录结构。
- **可配置端口**：在运行时可以指定代理服务器的端口，默认为 `10010`。

## 使用指南

### 1. 克隆项目

```bash
git clone https://github.com/litongjava/go-maven-proxy.git
cd go-maven-proxy
```

### 2. 构建和运行

确保你已经安装了 Go 开发环境，然后运行以下命令：

```bash
go build -o go-maven-proxy
./go-maven-proxy
```

### 3. 运行代理

#### 使用默认端口

默认情况下，代理将监听 `10010` 端口。你可以直接运行以下命令：

```bash
./go-maven-proxy
```

输出：
```
Starting HTTP proxy with caching on :10010
```

访问代理服务器：
```bash
curl http://localhost:10010/maven2/com/google/guava/guava/30.1.1-jre/guava-30.1.1-jre.pom
```

#### 指定自定义端口

你可以通过 `-port` 参数在运行时指定代理服务器的端口。例如，使用 `8080` 端口运行：

```bash
./go-maven-proxy -port 8080
```

输出：
```
Starting HTTP proxy with caching on :8080
```

访问代理服务器：
```bash
curl http://localhost:8080/maven2/com/google/guava/guava/30.1.1-jre/guava-30.1.1-jre.pom
```

### 4. 缓存文件

所有缓存的文件将被存储在项目根目录下的 `./cache` 文件夹中。缓存文件的目录结构与请求的 URI 保持一致。例如：

- 请求 URL: `http://localhost:10010/maven2/com/google/guava/guava/30.1.1-jre/guava-30.1.1-jre.pom`
- 缓存路径: `./cache/maven2/com/google/guava/guava/30.1.1-jre/guava-30.1.1-jre.pom`
### 5. nginx
```nginx
server {
  listen 80;
  server_name maven.example.com;

  location / {
    proxy_pass http://127.0.0.1:10010;
    proxy_pass_header Set-Cookie;
    proxy_set_header Host $host:$server_port;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    error_log  /var/log/nginx/backend.error.log;
    access_log  /var/log/nginx/backend.access.log;
  }
}
```
### 6. systemd
```
vi /etc/systemd/system/go-maven-proxy.service
```

```
[Unit]
Description=go-maven-proxy
After=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
WorkingDirectory = /opt/go-maven-proxy
ExecStart=/opt/go-maven-proxy/go-maven-proxy

[Install]
WantedBy=multi-user.target
```

```
systemctl start go-maven-proxy
systemctl status go-maven-proxy
systemctl enable go-maven-proxy
```
## 配置

- **缓存目录**：默认缓存目录为 `./cache`，可以在代码中修改 `cacheDir` 常量来更改缓存目录。
- **端口号**：默认监听端口为 `10010`。你可以在运行时通过 `-port` 参数指定其他端口，例如 `./go-maven-proxy -port 8080`。

## 未来改进

- **缓存策略**：添加缓存过期和失效策略，以确保缓存的数据保持最新。
- **日志增强**：添加更详细的日志记录，以便更好地监控代理的运行状态。
- **错误处理**：增强错误处理逻辑，处理更多的异常情况。

## 许可证

`go-maven-proxy` 使用 MIT 许可证，详细信息请参见 [LICENSE](LICENSE)。


