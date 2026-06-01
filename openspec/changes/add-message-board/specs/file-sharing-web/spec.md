## ADDED Requirements

### Requirement: 访问者网页消息入口
访问者网页应提供消息功能的入口。

#### Scenario: 显示消息图标
- **WHEN** 访问者打开共享网页
- **THEN** 页面头部显示消息图标按钮

#### Scenario: 消息图标角标
- **WHEN** 有消息存在
- **THEN** 消息图标显示消息数量角标

#### Scenario: 打开消息对话框
- **WHEN** 访问者点击消息图标
- **THEN** 弹出消息对话框

### Requirement: 访问者消息对话框
访问者网页应提供消息对话框。

#### Scenario: 对话框内容
- **WHEN** 消息对话框打开
- **THEN** 显示消息列表、名字输入框、消息输入框、发送按钮

#### Scenario: 名字输入框
- **WHEN** 对话框显示名字输入框
- **THEN** 输入框预填用户上次使用的名字（如有），否则为空

#### Scenario: 名字持久化
- **WHEN** 访问者输入名字并发送消息
- **THEN** 名字保存到 localStorage，下次自动填充
