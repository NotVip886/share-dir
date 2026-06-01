<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime'
import {
  GetShareDir,
  SetShareDir,
  SelectFolder,
  StartSharing,
  StopSharing,
  IsSharing,
  GetShareURL,
  GetLocalIP,
  GenerateQRCode,
  CopyShareURL,
  OpenInBrowser,
  UploadFile,
  UploadFilesByPath,
  GetMessages,
  SendMessage,
  ClearMessages,
  GetMessageCount
} from '../wailsjs/go/main/App'

const shareDir = ref('D:\\共享')
const isSharing = ref(false)
const shareURL = ref('')
const localIP = ref('')
const qrCodeData = ref('')
const statusMessage = ref('未开启共享')
const errorMessage = ref('')
const successMessage = ref('')
const isDragOver = ref(false)
const isUploading = ref(false)
const showMessageDialog = ref(false)
const messages = ref([])
const messageInput = ref('')
const messageCount = ref(0)
let messagePollTimer = null

function handleDragEnter(e) {
  e.preventDefault()
  if (e.dataTransfer.types.includes('Files')) {
    isDragOver.value = true
  }
}

function handleDragLeave(e) {
  e.preventDefault()
  if (e.target === document || e.target === document.documentElement) {
    isDragOver.value = false
  }
}

function handleDragOver(e) {
  e.preventDefault()
}

onMounted(async () => {
  document.addEventListener('dragenter', handleDragEnter)
  document.addEventListener('dragleave', handleDragLeave)
  document.addEventListener('dragover', handleDragOver)

  try {
    const dir = await GetShareDir()
    if (dir) shareDir.value = dir
  } catch (e) {}

  OnFileDrop(async (x, y, files) => {
    isDragOver.value = false
    if (!isSharing.value) {
      errorMessage.value = '请先开启共享后再上传文件'
      setTimeout(() => { errorMessage.value = '' }, 3000)
      return
    }
    if (files && files.length > 0) {
      await handleDropFiles(files)
    }
  }, false)
  
  startMessagePoll()
})

onUnmounted(() => {
  document.removeEventListener('dragenter', handleDragEnter)
  document.removeEventListener('dragleave', handleDragLeave)
  document.removeEventListener('dragover', handleDragOver)
  OnFileDropOff()
  stopMessagePoll()
})

async function handleSelectFolder() {
  try {
    const dir = await SelectFolder()
    if (dir) {
      shareDir.value = dir
    }
  } catch (e) {
    errorMessage.value = '选择文件夹失败: ' + e
  }
}

async function handleToggleSharing() {
  errorMessage.value = ''
  successMessage.value = ''
  if (isSharing.value) {
    const err = await StopSharing()
    if (err) {
      errorMessage.value = err
      return
    }
    isSharing.value = false
    shareURL.value = ''
    localIP.value = ''
    qrCodeData.value = ''
    statusMessage.value = '未开启共享'
  } else {
    await SetShareDir(shareDir.value)
    const err = await StartSharing()
    if (err) {
      errorMessage.value = err
      return
    }
    isSharing.value = await IsSharing()
    shareURL.value = await GetShareURL()
    localIP.value = await GetLocalIP()
    qrCodeData.value = await GenerateQRCode()
    statusMessage.value = '共享已开启'
  }
}

async function handleCopyURL() {
  const result = await CopyShareURL()
  if (result === '复制成功') {
    successMessage.value = '链接已复制到剪贴板'
    errorMessage.value = ''
    setTimeout(() => { successMessage.value = '' }, 2000)
  } else if (result) {
    errorMessage.value = result
  }
}

async function handleOpenBrowser() {
  await OpenInBrowser()
}

async function handleUpload() {
  if (isUploading.value) return
  errorMessage.value = ''
  successMessage.value = ''
  isUploading.value = true
  try {
    const result = await UploadFile()
    if (result) {
      if (result.includes('上传完成')) {
        successMessage.value = result
        errorMessage.value = ''
        setTimeout(() => { successMessage.value = '' }, 3000)
      } else {
        errorMessage.value = result
      }
    }
  } finally {
    isUploading.value = false
  }
}

async function handleDropFiles(files) {
  if (isUploading.value) return
  errorMessage.value = ''
  successMessage.value = ''
  isUploading.value = true
  try {
    const result = await UploadFilesByPath(files)
    if (result) {
      if (result.includes('上传完成')) {
        successMessage.value = result
        errorMessage.value = ''
        setTimeout(() => { successMessage.value = '' }, 3000)
      } else {
        errorMessage.value = result
      }
    }
  } finally {
    isUploading.value = false
  }
}

function startMessagePoll() {
  loadMessages()
  messagePollTimer = setInterval(loadMessages, 3000)
}

function stopMessagePoll() {
  if (messagePollTimer) {
    clearInterval(messagePollTimer)
    messagePollTimer = null
  }
}

async function loadMessages() {
  try {
    const data = await GetMessages()
    messages.value = JSON.parse(data)
    messageCount.value = await GetMessageCount()
  } catch (e) {}
}

async function handleSendMessage() {
  if (!messageInput.value.trim()) return
  await SendMessage(messageInput.value.trim())
  messageInput.value = ''
  await loadMessages()
}

async function handleClearMessages() {
  await ClearMessages()
  await loadMessages()
}

function openMessageDialog() {
  showMessageDialog.value = true
}

function closeMessageDialog() {
  showMessageDialog.value = false
}
</script>

<template>
  <div class="app-container">
    <Transition name="overlay">
      <div v-if="isDragOver" class="drag-overlay" :class="{ 'drag-overlay-disabled': !isSharing }">
        <div class="drag-overlay-content">
          <template v-if="isSharing">
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="17 8 12 3 7 8"/>
              <line x1="12" y1="3" x2="12" y2="15"/>
            </svg>
            <p class="drag-overlay-title">释放文件以上传</p>
            <p class="drag-overlay-hint">文件将上传到共享目录</p>
          </template>
          <template v-else>
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/>
              <line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/>
            </svg>
            <p class="drag-overlay-title">请先开启共享</p>
            <p class="drag-overlay-hint">开启共享后才能上传文件</p>
          </template>
        </div>
      </div>
    </Transition>

    <div class="header">
      <div class="icon-wrapper">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 6h16M4 12h16M4 18h16"/>
          <circle cx="8" cy="6" r="1" fill="currentColor"/>
          <circle cx="8" cy="12" r="1" fill="currentColor"/>
          <circle cx="8" cy="18" r="1" fill="currentColor"/>
        </svg>
      </div>
      <div class="header-text">
        <h1 class="title">局域网文件共享</h1>
        <p class="subtitle">支持二维码和链接分享</p>
      </div>
      <button class="msg-icon-btn" @click="openMessageDialog">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
        </svg>
        <span v-if="messageCount > 0" class="msg-badge">{{ messageCount > 99 ? '99+' : messageCount }}</span>
      </button>
      <div class="status-badge" :class="{ 'status-active': isSharing }">
        <div class="status-dot" :class="{ 'dot-active': isSharing }"></div>
        <span>{{ statusMessage }}</span>
      </div>
    </div>

    <div class="main-row">
      <div class="path-section">
        <input
          type="text"
          class="path-input"
          v-model="shareDir"
          :disabled="isSharing"
          placeholder="共享文件夹路径"
        />
        <button
          class="btn-browse"
          @click="handleSelectFolder"
          :disabled="isSharing"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
          </svg>
        </button>
      </div>
      <button
        class="btn-share"
        :class="{ 'btn-stop': isSharing }"
        @click="handleToggleSharing"
      >
        <span v-if="!isSharing" class="btn-icon">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
        </span>
        <span v-else class="btn-icon">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <rect x="6" y="6" width="12" height="12"/>
          </svg>
        </span>
        {{ isSharing ? '停止' : '开启' }}
      </button>
    </div>

    <div v-if="isSharing" class="share-info">
      <div class="info-row">
        <div class="qr-wrapper">
          <img :src="qrCodeData" alt="QR Code" class="qr-image" />
          <p class="qr-hint">扫码访问</p>
        </div>
        <div class="url-section">
          <div class="ip-display">
            <span class="ip-label">访问地址</span>
            <span class="ip-value">{{ shareURL }}</span>
          </div>
          <div class="url-actions">
            <button class="btn-action" @click="handleCopyURL">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
              </svg>
              复制链接
            </button>
            <button class="btn-action" @click="handleOpenBrowser">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
                <polyline points="15 3 21 3 21 9"/>
                <line x1="10" y1="14" x2="21" y2="3"/>
              </svg>
              浏览器打开
            </button>
          </div>
        </div>
      </div>

      <div
        class="upload-area"
        :class="{ 'dragover': isDragOver, 'uploading': isUploading }"
        @click="handleUpload"
      >
        <template v-if="isUploading">
          <div class="upload-spinner"></div>
          <p>正在上传中...</p>
        </template>
        <template v-else>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="17 8 12 3 7 8"/>
            <line x1="12" y1="3" x2="12" y2="15"/>
          </svg>
          <p>拖拽文件到此处，或 <span>点击选择文件</span></p>
        </template>
      </div>
    </div>

    <div v-if="successMessage" class="success-message">
      {{ successMessage }}
    </div>

    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>

    <Transition name="dialog">
      <div v-if="showMessageDialog" class="dialog-overlay" @click.self="closeMessageDialog">
        <div class="dialog-content">
          <div class="dialog-header">
            <h3>📢 消息板</h3>
            <button class="dialog-close" @click="closeMessageDialog">×</button>
          </div>
          <div class="dialog-body">
            <div class="message-list">
              <div v-if="messages.length === 0" class="no-messages">
                暂无消息
              </div>
              <div v-for="(msg, idx) in messages" :key="idx" class="message-item" :class="{ 'is-sharer': msg.isSharer }">
                <span class="msg-sender">{{ msg.sender }}</span>
                <span class="msg-content">{{ msg.content }}</span>
                <span class="msg-time">{{ msg.timestamp }}</span>
              </div>
            </div>
          </div>
          <div class="dialog-footer">
            <input
              type="text"
              class="msg-input"
              v-model="messageInput"
              placeholder="输入消息..."
              @keyup.enter="handleSendMessage"
            />
            <button class="msg-send-btn" @click="handleSendMessage">发送</button>
            <button class="msg-clear-btn" @click="handleClearMessages">清空</button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.app-container {
  min-height: 100vh;
  padding: 12px;
  background: linear-gradient(180deg, #f0f4ff 0%, #f5f7fa 100%);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
}

.drag-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(79, 110, 247, 0.08);
  backdrop-filter: blur(4px);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 3px dashed #4f6ef7;
  margin: 4px;
  border-radius: 8px;
  pointer-events: none;
}

.drag-overlay-disabled {
  background: rgba(239, 68, 68, 0.08);
  border-color: #ef4444;
}

.drag-overlay-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #4f6ef7;
}

.drag-overlay-disabled .drag-overlay-content {
  color: #ef4444;
}

.drag-overlay-title {
  font-size: 16px;
  font-weight: 600;
}

.drag-overlay-hint {
  font-size: 12px;
  color: #9ca3af;
}

.overlay-enter-active {
  transition: opacity 0.15s ease;
}

.overlay-leave-active {
  transition: opacity 0.2s ease;
}

.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}

.header {
  width: 100%;
  max-width: 376px;
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.icon-wrapper {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #4f6ef7 0%, #6c5ce7 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.header-text {
  flex: 1;
}

.title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a2e;
  letter-spacing: 0.3px;
  margin-bottom: 1px;
}

.subtitle {
  font-size: 10px;
  color: #9ca3af;
  font-weight: 400;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 20px;
  background: #f3f4f6;
  font-size: 11px;
  color: #6b7280;
  transition: all 0.3s ease;
}

.status-badge.status-active {
  background: #dcfce7;
  color: #16a34a;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #9ca3af;
  transition: all 0.3s ease;
}

.status-dot.dot-active {
  background: #22c55e;
  box-shadow: 0 0 4px rgba(34, 197, 94, 0.5);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 3px rgba(34, 197, 94, 0.4); }
  50% { box-shadow: 0 0 8px rgba(34, 197, 94, 0.7); }
}

.main-row {
  width: 100%;
  max-width: 376px;
  display: flex;
  gap: 6px;
  margin-bottom: 10px;
}

.path-section {
  flex: 1;
  display: flex;
  gap: 6px;
}

.path-input {
  flex: 1;
  padding: 8px 10px;
  border: 1px solid #e2e5ec;
  border-radius: 8px;
  font-size: 12px;
  color: #374151;
  background: #ffffff;
  outline: none;
  transition: border-color 0.2s;
}

.path-input:focus {
  border-color: #4f6ef7;
  box-shadow: 0 0 0 2px rgba(79, 110, 247, 0.1);
}

.path-input:disabled {
  background: #f3f4f6;
  color: #9ca3af;
}

.btn-browse {
  padding: 8px 10px;
  border: 1px solid #e2e5ec;
  border-radius: 8px;
  background: #ffffff;
  color: #4f6ef7;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-browse:hover:not(:disabled) {
  background: #f0f3ff;
  border-color: #4f6ef7;
}

.btn-browse:disabled {
  color: #9ca3af;
  cursor: not-allowed;
}

.btn-share {
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #4f6ef7 0%, #6c5ce7 100%);
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.3s;
  box-shadow: 0 2px 8px rgba(79, 110, 247, 0.3);
}

.btn-share:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(79, 110, 247, 0.4);
}

.btn-share.btn-stop {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
}

.btn-share.btn-stop:hover {
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.btn-icon {
  display: flex;
  align-items: center;
}

.share-info {
  width: 100%;
  max-width: 376px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.info-row {
  display: flex;
  gap: 10px;
  align-items: stretch;
  margin-bottom: 10px;
}

.qr-wrapper {
  padding: 10px;
  background: #ffffff;
  border-radius: 10px;
  border: 1px solid #e8ecf3;
  box-shadow: 0 1px 6px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
}

.qr-image {
  width: 110px;
  height: 110px;
  display: block;
}

.qr-hint {
  margin-top: 6px;
  font-size: 10px;
  color: #9ca3af;
}

.url-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ip-display {
  flex: 1;
  padding: 10px 12px;
  background: #ffffff;
  border-radius: 8px;
  border: 1px solid #e8ecf3;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.ip-label {
  display: block;
  font-size: 10px;
  color: #9ca3af;
  margin-bottom: 4px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.ip-value {
  font-size: 13px;
  color: #4f6ef7;
  font-weight: 600;
  font-family: "SF Mono", "Consolas", monospace;
  word-break: break-all;
}

.url-actions {
  display: flex;
  gap: 6px;
}

.btn-action {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 10px 8px;
  border: 1px solid #e2e5ec;
  border-radius: 8px;
  background: #ffffff;
  color: #4b5563;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-action:hover {
  background: #f0f3ff;
  border-color: #4f6ef7;
  color: #4f6ef7;
}

.upload-area {
  width: 100%;
  border: 2px dashed #d1d5db;
  border-radius: 10px;
  padding: 14px;
  background: #fafbfc;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}

.upload-area:hover,
.upload-area.dragover {
  border-color: #4f6ef7;
  background: #f0f3ff;
}

.upload-area.uploading {
  border-color: #4f6ef7;
  background: #f0f3ff;
  cursor: wait;
}

.upload-area svg {
  color: #9ca3af;
}

.upload-area:hover svg,
.upload-area.dragover svg {
  color: #4f6ef7;
}

.upload-area p {
  font-size: 12px;
  color: #6b7280;
}

.upload-area span {
  color: #4f6ef7;
  font-weight: 500;
}

.upload-spinner {
  width: 24px;
  height: 24px;
  border: 3px solid #e0e7ff;
  border-top-color: #4f6ef7;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.success-message {
  position: fixed;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  padding: 8px 16px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  color: #16a34a;
  font-size: 12px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  z-index: 100;
  animation: slideUp 0.3s ease;
}

.error-message {
  position: fixed;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  padding: 8px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  font-size: 12px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  z-index: 100;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

.msg-icon-btn {
  position: relative;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: #ffffff;
  color: #6b7280;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.msg-icon-btn:hover {
  background: #f0f3ff;
  color: #4f6ef7;
}

.msg-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  background: #ef4444;
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 300;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-content {
  width: 340px;
  max-height: 380px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dialog-header {
  padding: 12px 16px;
  border-bottom: 1px solid #e8ecf3;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dialog-header h3 {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a2e;
}

.dialog-close {
  width: 24px;
  height: 24px;
  border: none;
  background: none;
  color: #9ca3af;
  font-size: 18px;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-close:hover {
  background: #f3f4f6;
  color: #374151;
}

.dialog-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.no-messages {
  text-align: center;
  color: #9ca3af;
  font-size: 12px;
  padding: 20px;
}

.message-item {
  padding: 8px 10px;
  background: #f8fafc;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.message-item.is-sharer {
  background: #f0f3ff;
}

.msg-sender {
  font-size: 11px;
  font-weight: 600;
  color: #4f6ef7;
}

.msg-content {
  font-size: 12px;
  color: #374151;
  word-break: break-all;
}

.msg-time {
  font-size: 10px;
  color: #9ca3af;
  align-self: flex-end;
}

.dialog-footer {
  padding: 12px;
  border-top: 1px solid #e8ecf3;
  display: flex;
  gap: 6px;
}

.msg-input {
  flex: 1;
  padding: 8px 10px;
  border: 1px solid #e2e5ec;
  border-radius: 6px;
  font-size: 12px;
  outline: none;
}

.msg-input:focus {
  border-color: #4f6ef7;
}

.msg-send-btn {
  padding: 8px 12px;
  border: none;
  border-radius: 6px;
  background: #4f6ef7;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.msg-send-btn:hover {
  background: #4358d9;
}

.msg-clear-btn {
  padding: 8px 10px;
  border: 1px solid #e2e5ec;
  border-radius: 6px;
  background: #fff;
  color: #6b7280;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.msg-clear-btn:hover {
  background: #fef2f2;
  border-color: #fecaca;
  color: #dc2626;
}

.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 0.2s ease;
}

.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}
</style>
