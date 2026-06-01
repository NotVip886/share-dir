## ADDED Requirements

### Requirement: 桌面界面消息入口
共享者桌面界面应提供消息功能的入口。

#### Scenario: 显示消息图标
- **WHEN** 共享者打开应用
- **THEN** 界面头部显示消息图标按钮

#### Scenario: 消息图标角标
- **WHEN** 有消息存在
- **THEN** 消息图标显示消息数量角标

#### Scenario: 打开消息对话框
- **WHEN** 共享者点击消息图标
- **THEN** 弹出消息对话框

### Requirement: 桌面消息对话框
共享者桌面界面应提供消息对话框。

#### Scenario: 对话框内容
- **WHEN** 消息对话框打开
- **THEN** 显示消息列表、消息输入框、发送按钮、清空按钮

#### Scenario: 发送者标识
- **WHEN** 共享者发送消息
- **THEN** 消息发送者标识为"共享者"
