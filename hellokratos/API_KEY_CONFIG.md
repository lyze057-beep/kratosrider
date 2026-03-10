# API Key 配置说明

## 百度地图 API

### 问题
API 返回错误: `APP 服务被禁用 (status: 240)`

### 原因
您提供的 API Key `gsgC3aiW4r6W7WTjb5vRVVzclnbt17Bi` 在百度地图开发者平台中，对应的 APP 服务没有启用路线规划、附近搜索等服务。

### 解决方案
1. 访问 [百度地图开发者平台](https://lbsyun.baidu.com/)
2. 登录您的账号
3. 进入【应用管理】→【我的应用】
4. 找到您使用的应用，点击【编辑】
5. 在【服务列表】中启用以下服务：
   - 路线规划
   - 附近搜索
   - 距离计算
6. 保存配置
7. 等待 5-10 分钟生效后，API 即可正常使用

### 验证 API
启用服务后，可以运行测试脚本验证：
```bash
go run test_api.go
```

## 和风天气 API

### 问题
API 返回错误: `403 Invalid Host`

### 原因
和风天气 API 需要配置允许访问的域名/IP，当前请求来源不在白名单中。

### 解决方案
1. 访问 [和风天气开发者平台](https://dev.qweather.com/)
2. 登录您的账号
3. 进入【管理后台】→【我的应用】
4. 找到您使用的应用，点击【编辑】
5. 在【白名单】中添加允许访问的域名/IP：
   - 本地开发: `localhost`, `127.0.0.1`
   - 生产环境: 您的服务器域名或 IP
6. 保存配置
7. 等待配置生效

### 或者使用免费测试 API Key
如果只是测试，可以使用和风天气提供的免费 API Key：
- 注册账号后，系统会自动分配一个测试 Key
- 测试 Key 有调用次数限制（通常每天 1000 次）

## 当前配置

### 导航 Skill
```go
apiKey:  "gsgC3aiW4r6W7WTjb5vRVVzclnbt17Bi"
baseURL: "https://api.map.baidu.com"
```

### 天气 Skill
```go
apiKey:  "eb5ee80bd47f49c089682d19bd1e19f4"
baseURL: "https://devapi.qweather.com/v7"
```

## 临时解决方案

如果暂时无法配置 API Key，系统会返回友好的错误提示：
- 导航功能: "路线规划失败: API error: APP 服务被禁用"
- 天气功能: "实时天气查询失败: API error: 403"

您可以：
1. 先使用模拟数据进行开发和测试
2. 等待 API Key 配置完成后再启用真实 API 调用
3. 或者使用其他地图和天气服务（如高德地图、OpenWeather）

## 测试状态

当前测试结果：
- ✅ 代码编译通过
- ❌ 百度地图 API 需要启用服务
- ❌ 和风天气 API 需要配置白名单

建议先配置好 API Key，然后运行测试验证功能。
