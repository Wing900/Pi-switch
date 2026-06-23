# Pi Switch 0.0.0.1 架构说明

这一版先交付一个可直接打开的前端原型，但接口边界按后续 Wails/Go 正式实现预留。

## 目录职责

- `frontend/src/config`
  - 主题与 Provider 预设。
- `frontend/src/services`
  - `PiSwitchApi` 作为前端唯一的数据入口，当前是本地模拟，后续可替换成 Wails 绑定。
- `frontend/src/state`
  - 极简状态容器，避免 UI 组件之间相互修改数据。
- `frontend/src/components`
  - 只负责把状态渲染成界面字符串。
- `frontend/src/app`
  - 事件绑定、状态编排、主题切换。

## 当前实现范围

- 中文单页桌面布局
- Provider 预设新增
- Provider 编辑与删除
- 连接测试结果弹层
- 获取模型与导入模型
- 设置默认模型
- 启动命令预览
- Pi 路径设置弹层
- 三套轻莫奈色系

## 后续接入 Go/Wails 的替换点

- 把 `frontend/src/services/pi-switch-api.js` 中的方法替换为文档定义的 Wails 暴露方法。
- 保持 `renderApp` 与 `store` 不变，可减少 UI 层返工。
- 后续正式工程可以直接落为：
  - `main.go`
  - `app.go`
  - `internal/config`
  - `internal/pi`
  - `internal/provider`
  - `frontend/src`
