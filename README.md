# MediaServerGO
- 🎉Linux、Windows
# api
- `/login`
- `/remote_download`
- `/getAssets`
- `/userLogin`
- `/remote_download`
- 🔐`/`
- 🔐`/getDeviceName`
- 🔐`/getFileList`
- 🔐`/getCover`
- 🔐`/getFile/:name`
- 🔐`/getFile2/:name`
- 🔐`/getVideoPreview`
- 🔐`/toggleBookmark`
- 🔐`/getDeviceInfo`
# install
- 解压后配置好config.json后运行`server`即可
```json
{
  "port": 80,//监听端口
  "webPath": "F:\\MediaClientWeb",//前端项目路径、见MediaClientWeb项目
  "Aria2": {//Aria2跳板支持，用于迅雷网盘抓包，见XunleiVapture项目
    "RPC": "http://localhost:6800/jsonrpc",//RPC路径
    "Token": ""//token，未测试留空，安全起见建议设置
  },
  "mountPoints": [//磁盘挂载点
    "F:\\media\\NAS500",
    "F:\\media\\NAS600"
  ],
  "JWT": {//用于用户认证
    "algorithm": "HS256",//加密算法
    "secret": "your_key",//密钥
    "durationHours": 168//有效期，7天 == 7 * 24 = 168h
  },
  "users": [//注册用户，目前仅可手动注册
    {
      "name": "pan",//用户名
      "hash": "246DCD487EF18B08F36DEC3AE43029EA"//密码的MD5值
    },{
      "name": "tao",
      "hash": "19A6A0B9360519FE82B5B06B3F79D62C"
    }
  ]
}
```
