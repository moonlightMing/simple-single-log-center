# Simple-single-log-center

- 简易日志查看工具，用于查看远程机器上面的实时日志内容，当然也可查看一般文本文件
- 整个工具使用SSH协议进行远程文件的访问，无需插件。但是只允许密钥认证连接
- 提供了一个简易的WebTerminal（可选开启）
- 由于不带帐号认证，请不要在生产环境使用该工具

## 功能

- 左键双击资产节点浏览文件

![](https://github.com/moonlightMing/simple-single-log-center/blob/master/logwatch.gif)

- 右键单击资产节点开启WebTerminal（需在配置中开启）

![](https://github.com/moonlightMing/simple-single-log-center/blob/master/webterminal.gif)

## 构建运行

```
git clone https://github.com/moonlightMing/simple-log-center-face.git
cd ./simple-log-center-face
go build
./simple-log-center-face
```

## 配置

配置分为两个部分，一是资产清单文件，完全按照ansible-hosts文件的方式进行填写，但是只对IP部分生效，无视额外变量。二是config.json，详细说明如下：

```
# mv config.json.example config.json

# 按照注释说明修改对应内容
{
  "server": {
    "addr": "0.0.0.0",                      // 服务对外监听地址
    "port": 9090                            // 服务对外开放端口
  },
  "ssh": {
    "user": "root",                         // 远程SSH连接账户
    "port": 22,                             // 远程SSH连接端口
    "private_key": "/root/.ssh/id_rsa",     // SSH认证所需密钥
    "private_key_password": ""              // SSH密钥认证所需的组合密码，没有则置空
  },
  "allow": {
    "dir": ["/root", "/data"],              // 允许访问的路径
    "file_type": ["log", "sh"]              // 允许访问的文件类型，没有该后缀的文件不显示
  },
  "ansible": {
    "hosts_file": "/etc/ansible/hosts"      // Ansible资产清单文件
  },
  "web_terminal": {
      "open": false                         // WebTerminal开关
  }
}
```

## 后记

制作该工具的起因是公司在没有容器化自动化的场景下，方便开发人员在测试环境的远程调试。