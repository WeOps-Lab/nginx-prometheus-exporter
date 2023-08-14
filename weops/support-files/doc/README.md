## 嘉为蓝鲸nginx插件使用说明

## 使用说明

### 插件功能  

Nginx Exporter通过解析 Nginx 的状态页面和其他可访问的信息源, 从中提取出有价值的监控数据。 该工具能够将从 Nginx 收集到的数据转换为易于理解和分析的监控指标，使用户能够更轻松地监视和评估 Nginx 实例的性能。收集多种关键的 Nginx 指标，从而实现性能优化和故障排查。

### 版本支持

操作系统支持: linux, windows

是否支持arm: 支持

**是否支持远程采集:**

是

### 参数说明

| **参数名**                       | **含义**                         | **是否必填** | **使用举例**                          |
|-------------------------------|--------------------------------|----------|-----------------------------------|
| -nginx.common                 | nginx stub采集开关(**开关参数**), 默认打开 | 是        |                                   |
| NGINX_SCRAPE_URI              | nginx stub访问地址(**环境变量**)       | 否        | http://127.0.0.1:8080/stub_status |
| -nginx.vts                    | nginx vts采集开关(**开关参数**), 默认关闭            | 是        |                                   |  |
| NGINX_VTS_SCRAPE_URI          | nginx vts访问地址(**环境变量**)                  | 否        | http://127.0.0.1:8080/vts_status  |
| -nginx.rtmp                   | nginx rtmp采集开关(**开关参数**), 默认关闭           | 是        |                                   |
| NGINX_RTMP_SCRAPE_URI         | nginx rtmp访问地址(**环境变量**)                 | 否        | http://127.0.0.1:8080/rtmp_status |
| NGINX_RTMP_REGEX_STREAM | nginx rtmp音视频流正则过滤(**环境变量**), 默认采集所有     | 否        | .*                                |
| -nginx.timeout                | nginx采集超时时间, 默认为5s             | 是        | 5s                                |
| -log.level                    | 日志级别                           | 否        | info                              |
| --web.listen-address          | exporter监听id及端口地址              | 否        | 127.0.0.1:9601                    |



### 使用指引
**注意：采集器所在的服务器需要能够正常访问对应模块功能的地址。**  

1. nginx配置
   采集nginx基础指标需要开启stub_status模块。
   采集vts类指标需要开启vts模块, 提供对 nginx 虚拟机主机状态数据的访问，可将数据输出格式为json、html、jsonp、prometheus。
   采集rtmp类指标需要开启rtmp模块。

2. 检查模块配置  
   通过 `nginx -V` 查看模块是否添加成功, 可看到示例中已安装stub_status和vts模块
   ```
   nginx -V

   # 返回结果示例
   nginx version: nginx/1.25.1
   built with OpenSSL 1.1.1n  15 Mar 2022
   TLS SNI support enabled
   configure arguments: --prefix=/opt/bitnami/nginx --with-http_stub_status_module --with-stream --with-http_gzip_static_module --with-mail --with-http_realip_module --with-http_stub_status_module --with-http_v2_module --with-http_ssl_module --with-mail_ssl_module --with-http_gunzip_module --with-threads --with-http_auth_request_module --with-http_sub_module --with-http_geoip_module --with-compat --with-stream_realip_module --with-stream_ssl_module --with-cc-opt=-Wno-stringop-overread --add-module=/bitnami/blacksmith-sandox/nginx-module-vts-0.2.2 --add-dynamic-module=/bitnami/blacksmith-sandox/nginx-module-geoip2-3.4.0 --add-module=/bitnami/blacksmith-sandox/nginx-module-substitutions-filter-0.20220124.0 --add-dynamic-module=/bitnami/blacksmith-sandox/nginx-module-brotli-0.20220429.0
   ```

3. nginx配置文件
   文件内容示例(一般为nginx.conf)
   ```
   # 开启 upstram zones 
   upstream backend{
     server 127.0.0.1:80;
   }
   
   vhost_traffic_status_zone;  # 开启vts统计模块
   vhost_traffic_status_filter_by_host on;  # 打开vts vhost过滤
   vhost_traffic_status_filter_by_set_key $status $server_name;  # 开启vts详细状态码统计
   
   server {
     server_name *.example.org;
   
     listen 8080;
   
     # vts访问路径
     location /vts_status {  
       vhost_traffic_status_display;   # 开启vts展示
       vhost_traffic_status_display_format html;
     }
   
     # stub_status访问路径
     location /stub_status { 
       stub_status on;   # 开启stub_status模块
       access_log   off;
       allow 127.0.0.1;    # 只允许本地IP访问
       deny all;        # 禁止任何IP访问
     }
   }
   ```

   nginx访问控制需要自行配置，`allow` 和 `deny` 后内容按照实际情况填写。

   vts除了状态码统计, 还有基于地理信息的统计，根据访问量或访问流量对nginx做访问限制，详细使用见文档: https://github.com/vozlt/nginx-module-vts#installation

4. 重新加载配置  
  `sudo nginx -t && sudo nginx -s reload`  

5. 检查配置  
   `
   nginx -t

   # 返回结果示例
   nginx: the configuration file /opt/bitnami/nginx/conf/nginx.conf syntax is ok
   nginx: configuration file /opt/bitnami/nginx/conf/nginx.conf test is successful
  `

6. 重启服务 
  如果是改变 `Nginx` 的编译参数、添加新的模块, 通常需要重新编译和安装, 然后重启服务。

### 指标简介
| **指标ID**                               | **指标中文名**             | **维度ID**                      | **维度含义**               | **单位**       |
|----------------------------------------|-----------------------|-------------------------------|------------------------|--------------|
| nginx_up                               | Nginx运行状态             | -                             | -                      | -            |
| nginx_rtmp_up                          | Nginx rtmp运行状态        | -                             | -                      | -            |
| nginx_vts_up                           | Nginx vts运行状态         | -                             | -                      | -            |
| nginx_connections_accepted             | Nginx已接受的客户端连接数       | -                             | -                      | -            |
| nginx_connections_active               | Nginx活动中的客户端连接数       | -                             | -                      | -            |
| nginx_connections_handled              | Nginx已处理的客户端连接数       | -                             | -                      | -            |
| nginx_connections_reading              | Nginx正在读取请求头的连接数      | -                             | -                      | -            |
| nginx_connections_waiting              | Nginx空闲的客户端连接数        | -                             | -                      | -            |
| nginx_connections_writing              | Nginx正在将响应写回客户端的连接数   | -                             | -                      | -            |
| nginx_http_requests_total              | Nginx所有 HTTP 请求的总数    | -                             | -                      | -            |
| nginx_vts_server_uptime                | Nginx vts服务已运行时间      | hostName, nginxVersion        | 主机名称, 版本               | s            |
| nginx_vts_server_connections           | Nginx vts服务连接状态       | status                        | 状态                     | -            |
| nginx_vts_server_requests              | Nginx vts服务区域请求数量     | code, host                    | 状态码, 主机名称              | -            |
| nginx_vts_server_bytes                 | Nginx vts服务区域数据流量     | direction, host               | 流量方向, 主机名称             | bytes        |
| nginx_vts_server_cache                 | Nginx vts服务区域缓存信息     | status, host                  | 状态, 主机名称               | -            |
| nginx_vts_server_requestMsec           | Nginx vts服务区域请求处理时间   | host                          | 主机名称                   | ms           |
| nginx_vts_filter_requests              | Nginx vts过滤器请求数量      | code, filter, filterName      | 状态码, 过滤器, 过滤名          | -            |
| nginx_vts_filter_bytes                 | Nginx vts过滤器数据流量      | direction, filter, filterName | 流量方向, 过滤器, 过滤名         | bytes        |
| nginx_vts_filter_responseMsec          | Nginx vts过滤器响应时间      | filter, filterName            | 过滤器, 过滤名               | ms           |
| nginx_vts_filter_requestMsec           | Nginx vts过滤器请求时间      | filter, filterName            | 过滤器, 过滤名               | ms           |
| nginx_vts_upstream_requests            | Nginx vts上游服务请求数量     | backend, code, upstream       | 上游服务器地址, 状态码, 上游服务器名称  | -            |
| nginx_vts_upstream_bytes               | Nginx vts上游服务数据流量     | backend, direction, upstream  | 上游服务器地址, 流量方向, 上游服务器名称 | bytes        |
| nginx_vts_upstream_responseMsec        | Nginx vts上游服务响应时间     | backend, upstream             | 上游服务器地址, 上游服务器名称       | ms           |
| nginx_vts_upstream_requestMsec         | Nginx vts上游服务请求时间     | backend, upstream             | 上游服务器地址, 上游服务器名称       | ms           |
| nginx_rtmp_server_current_streams      | Nginx rtmp当前活跃的音视频流数量 | -                             | -                      | -            |
| nginx_rtmp_server_incoming_bytes_total | Nginx rtmp总接收数据字节     | -                             | -                      | bytes        |
| nginx_rtmp_server_outgoing_bytes_total | Nginx rtmp总输出数据字节     | -                             | -                      | bytes        |
| nginx_rtmp_server_receive_bytes        | Nginx rtmp服务接收带宽      | -                             | -                      | bytes/second |
| nginx_rtmp_server_transmit_bytes       | Nginx rtmp服务传输带宽      | -                             | -                      | bytes/second |
| nginx_rtmp_server_uptime_seconds_total | Nginx rtmp服务已运行时间     | -                             | -                      | s            |
| nginx_rtmp_stream_incoming_bytes_total | Nginx rtmp音视频流总接收字节   | stream                        | 音视频流名称                 | bytes        |
| nginx_rtmp_stream_outgoing_bytes_total | Nginx rtmp音视频流总输出字节   | stream                        | 音视频流名称                 | bytes        |
| nginx_rtmp_stream_receive_bytes        | Nginx rtmp音视频流接收带宽    | stream                        | 音视频流名称                 | bytes/second |
| nginx_rtmp_stream_transmit_bytes       | Nginx rtmp音视频流传输带宽    | stream                        | 音视频流名称                 | bytes/second |
| nginx_rtmp_stream_uptime_seconds_total | Nginx rtmp音视频流已运行时间   | stream                        | 音视频流名称                 | s            |


### 版本日志

#### weops_nginx_exporter 7.3.3

- weops调整

添加“小嘉”微信即可获取nginx监控指标最佳实践礼包，其他更多问题欢迎咨询

<img src="https://wedoc.canway.net/imgs/img/小嘉.jpg" width="50%" height="50%">
