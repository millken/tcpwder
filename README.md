# tcpwder
Simple tcp/udp 4-layer forwarding

simple config.toml
```
[logging]
level = "info"
output = "stdout"

[defaults]
max_connections = 0    
client_idle_timeout = "0" 
backend_idle_timeout = "0" 
backend_connection_timeout = "0"

[servers]

[servers.sample]
protocol = "tcp"
bind = "localhost:3000"
upstream = [
      "localhost:8000",
      "localhost:8001"
  ]

[servers.dns]
protocol = "udp"
bind = "localhost:53"
balance = "roundrobin"
upstream = [
      "8.8.8.8:53",
      "8.8.4.4:53"
  ]
```
