## 1. 后端数据结构与初始化

- [x] 1.1 在 app.go 中定义 Message 结构体（Sender, Content, Timestamp, IsSharer）
- [x] 1.2 在 App 结构体中添加 messages 切片和 maxMessages 常量
- [x] 1.3 在 NewApp() 中初始化 messages 和 maxMessages(50)

## 2. 后端 API 接口

- [x] 2.1 实现 GetMessages() 方法返回所有消息
- [x] 2.2 实现 SendMessage(content string) 方法供共享者调用
- [x] 2.3 实现 ClearMessages() 方法清空所有消息
- [x] 2.4 实现 GetMessageCount() 方法返回消息数量
- [x] 2.5 添加 /api/messages 处理函数（GET 请求）
- [x] 2.6 添加 /api/message 处理函数（POST 请求，供访问者发送）
- [x] 2.7 添加 /api/myip 处理函数（返回访问者IP）
- [x] 2.8 在 StartSharing() 中注册新的路由

## 3. 前端 Vue（共享者界面）

- [x] 3.1 在 App.vue 头部添加消息图标按钮
- [x] 3.2 实现消息图标角标显示（消息数量）
- [x] 3.3 创建消息对话框组件（MessageDialog.vue）
- [x] 3.4 实现对话框打开/关闭逻辑
- [x] 3.5 在对话框中展示消息列表
- [x] 3.6 添加消息输入框和发送按钮
- [x] 3.7 添加清空消息按钮
- [x] 3.8 实现轮询刷新消息（每3秒）
- [x] 3.9 绑定 Wails Go 方法（GetMessages, SendMessage, ClearMessages, GetMessageCount）

## 4. 前端 HTML（访问者网页）

- [x] 4.1 在 indexHTML 头部添加消息图标按钮
- [x] 4.2 实现消息图标角标显示
- [x] 4.3 添加消息对话框 HTML 结构
- [x] 4.4 添加名字输入框（从 localStorage 读取）
- [x] 4.5 添加消息列表展示区域
- [x] 4.6 添加消息输入框和发送按钮
- [x] 4.7 实现对话框打开/关闭逻辑
- [x] 4.8 实现 /api/messages 请求获取消息
- [x] 4.9 实现 /api/message POST 发送消息
- [x] 4.10 实现 /api/myip 获取访问者IP
- [x] 4.11 实现名字保存到 localStorage
- [x] 4.12 实现轮询刷新消息（每3秒）
- [x] 4.13 添加对话框 CSS 样式

## 5. 测试与验证

- [x] 5.1 验证共享者发送消息功能
- [x] 5.2 验证访问者发送消息功能
- [x] 5.3 验证消息列表正确展示
- [x] 5.4 验证消息数量超过50条自动删除最早的
- [x] 5.5 验证访问者名字持久化（localStorage）
- [x] 5.6 验证轮询刷新正常工作
- [x] 5.7 验证清空消息功能
