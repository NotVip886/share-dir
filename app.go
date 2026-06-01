package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skip2/go-qrcode"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	server     *http.Server
	shareDir   string
	isSharing  bool
	shareURL   string
	localIP    string
}

func NewApp() *App {
	return &App{
		shareDir: `D:\共享`,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetShareDir() string {
	return a.shareDir
}

func (a *App) SetShareDir(dir string) {
	a.shareDir = dir
}

func (a *App) IsSharing() bool {
	return a.isSharing
}

func (a *App) GetShareURL() string {
	return a.shareURL
}

func (a *App) GetLocalIP() string {
	return a.localIP
}

func (a *App) SelectFolder() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择共享文件夹",
	})
	if err != nil {
		return ""
	}
	if dir != "" {
		a.shareDir = dir
	}
	return dir
}

func (a *App) getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String(), nil
}

func (a *App) StartSharing() string {
	if a.isSharing {
		return "共享已在运行中"
	}

	if _, err := os.Stat(a.shareDir); os.IsNotExist(err) {
		err := os.MkdirAll(a.shareDir, 0755)
		if err != nil {
			return fmt.Sprintf("创建目录失败: %v", err)
		}
	}

	ip, err := a.getLocalIP()
	if err != nil {
		return fmt.Sprintf("获取本地IP失败: %v", err)
	}
	a.localIP = ip

	port := 8765
	a.shareURL = fmt.Sprintf("http://%s:%d", ip, port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/files/", a.handleFileDownload)
	mux.HandleFunc("/upload", a.handleUpload)
	mux.HandleFunc("/api/list", a.handleFileList)

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	a.isSharing = true
	return ""
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func (a *App) handleFileList(w http.ResponseWriter, r *http.Request) {
	files, err := a.getFileList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(files))
}

func (a *App) getFileList() (string, error) {
	entries, err := os.ReadDir(a.shareDir)
	if err != nil {
		return "[]", nil
	}

	type FileInfo struct {
		Name    string `json:"name"`
		IsDir   bool   `json:"isDir"`
		Size    int64  `json:"size"`
		ModTime string `json:"modTime"`
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	result := "["
	for i, f := range files {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`{"name":"%s","isDir":%v,"size":%d,"modTime":"%s"}`,
			f.Name, f.IsDir, f.Size, f.ModTime)
	}
	result += "]"
	return result, nil
}

func (a *App) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/files/")
	filepath := filepath.Join(a.shareDir, filename)
	http.ServeFile(w, r, filepath)
}

func (a *App) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	maxSize := int64(100 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	err := r.ParseMultipartForm(maxSize)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()
		a.saveFile(handler, file)
	} else {
		for _, handler := range files {
			file, err := handler.Open()
			if err != nil {
				continue
			}
			a.saveFile(handler, file)
			file.Close()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

func (a *App) saveFile(handler *multipart.FileHeader, file multipart.File) {
	dstPath := filepath.Join(a.shareDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return
	}
	defer dst.Close()
	io.Copy(dst, file)
}

func (a *App) StopSharing() string {
	if !a.isSharing || a.server == nil {
		return "共享未在运行"
	}
	err := a.server.Close()
	if err != nil {
		return fmt.Sprintf("停止共享失败: %v", err)
	}
	a.server = nil
	a.isSharing = false
	a.shareURL = ""
	a.localIP = ""
	return ""
}

func (a *App) GenerateQRCode() string {
	if a.shareURL == "" {
		return ""
	}
	png, err := qrcode.Encode(a.shareURL, qrcode.Medium, 256)
	if err != nil {
		return ""
	}
	base64Str := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + base64Str
}

func (a *App) CopyShareURL() string {
	if a.shareURL == "" {
		return ""
	}
	err := runtime.ClipboardSetText(a.ctx, a.shareURL)
	if err != nil {
		return fmt.Sprintf("复制失败: %v", err)
	}
	return "复制成功"
}

func (a *App) OpenInBrowser() string {
	if a.shareURL == "" {
		return ""
	}
	url := a.shareURL
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	runtime.BrowserOpenURL(a.ctx, url)
	return ""
}

func (a *App) UploadFile() string {
	if !a.isSharing {
		return "请先开启共享"
	}

	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要上传的文件",
	})
	if err != nil {
		return fmt.Sprintf("选择文件失败: %v", err)
	}
	if len(files) == 0 {
		return ""
	}

	return a.UploadFilesByPath(files)
}

func (a *App) UploadFilesByPath(files []string) string {
	if !a.isSharing {
		return "请先开启共享"
	}

	if len(files) == 0 {
		return ""
	}

	uploadedCount := 0
	for _, srcPath := range files {
		srcFile, err := os.Open(srcPath)
		if err != nil {
			continue
		}

		fileName := filepath.Base(srcPath)
		dstPath := filepath.Join(a.shareDir, fileName)

		dstFile, err := os.Create(dstPath)
		if err != nil {
			srcFile.Close()
			continue
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()

		if err == nil {
			uploadedCount++
		}
	}

	if uploadedCount == 0 {
		return "上传失败"
	}
	return fmt.Sprintf("上传完成，共 %d 个文件", uploadedCount)
}

func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
}

var indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>局域网文件共享</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
            background: linear-gradient(180deg, #f0f4ff 0%, #f5f7fa 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
        }
        .header {
            text-align: center;
            margin-bottom: 24px;
        }
        .header h1 {
            color: #1a1a2e;
            font-size: 24px;
            margin-bottom: 8px;
        }
        .header p {
            color: #8c8ca1;
            font-size: 14px;
        }
        .upload-section {
            background: #fff;
            border-radius: 16px;
            padding: 24px;
            margin-bottom: 20px;
            border: 1px solid #e8ecf3;
            box-shadow: 0 2px 12px rgba(0,0,0,0.04);
        }
        .upload-area {
            border: 2px dashed #d1d5db;
            border-radius: 12px;
            padding: 40px 20px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s;
            background: #fafbfc;
        }
        .upload-area:hover, .upload-area.dragover {
            border-color: #4f6ef7;
            background: #f0f3ff;
        }
        .upload-area svg {
            width: 48px;
            height: 48px;
            color: #9ca3af;
            margin-bottom: 12px;
        }
        .upload-area p {
            color: #6b7280;
            font-size: 14px;
        }
        .upload-area span {
            color: #4f6ef7;
            font-weight: 500;
        }
        #fileInput {
            display: none;
        }
        .file-list-section {
            background: #fff;
            border-radius: 16px;
            border: 1px solid #e8ecf3;
            box-shadow: 0 2px 12px rgba(0,0,0,0.04);
            overflow: hidden;
        }
        .file-list-header {
            padding: 16px 20px;
            border-bottom: 1px solid #e8ecf3;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .file-list-header h2 {
            font-size: 16px;
            color: #1a1a2e;
            font-weight: 600;
        }
        .refresh-btn {
            background: none;
            border: none;
            color: #4f6ef7;
            cursor: pointer;
            padding: 6px 12px;
            border-radius: 6px;
            font-size: 13px;
            display: flex;
            align-items: center;
            gap: 4px;
            transition: background 0.2s;
        }
        .refresh-btn:hover {
            background: #f0f3ff;
        }
        .file-list {
            list-style: none;
        }
        .file-item {
            display: flex;
            align-items: center;
            padding: 14px 20px;
            border-bottom: 1px solid #f3f4f6;
            transition: background 0.2s;
        }
        .file-item:last-child {
            border-bottom: none;
        }
        .file-item:hover {
            background: #f8fafc;
        }
        .file-icon {
            width: 36px;
            height: 36px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 14px;
            flex-shrink: 0;
        }
        .file-icon.folder {
            background: #fef3c7;
            color: #f59e0b;
        }
        .file-icon.file {
            background: #e0e7ff;
            color: #4f6ef7;
        }
        .file-info {
            flex: 1;
            min-width: 0;
        }
        .file-name {
            font-size: 14px;
            color: #1f2937;
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        .file-meta {
            font-size: 12px;
            color: #9ca3af;
            margin-top: 2px;
        }
        .file-download {
            padding: 6px 14px;
            background: #f0f3ff;
            border: 1px solid #e0e7ff;
            border-radius: 6px;
            color: #4f6ef7;
            font-size: 13px;
            text-decoration: none;
            transition: all 0.2s;
        }
        .file-download:hover {
            background: #4f6ef7;
            color: #fff;
            border-color: #4f6ef7;
        }
        .empty-state {
            padding: 60px 20px;
            text-align: center;
            color: #9ca3af;
        }
        .empty-state svg {
            width: 64px;
            height: 64px;
            margin-bottom: 16px;
            opacity: 0.5;
        }
        .empty-state p {
            font-size: 14px;
        }
        .upload-progress {
            margin-top: 16px;
            display: none;
        }
        .upload-progress.show {
            display: block;
        }
        .progress-bar {
            height: 6px;
            background: #e5e7eb;
            border-radius: 3px;
            overflow: hidden;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #4f6ef7, #6c5ce7);
            width: 0%;
            transition: width 0.3s;
        }
        .progress-text {
            font-size: 13px;
            color: #6b7280;
            margin-top: 8px;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📁 局域网文件共享</h1>
            <p>点击文件下载，或拖拽文件到上传区域</p>
        </div>

        <div class="upload-section">
            <div class="upload-area" id="uploadArea">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                    <polyline points="17 8 12 3 7 8"/>
                    <line x1="12" y1="3" x2="12" y2="15"/>
                </svg>
                <p>拖拽文件到此处，或 <span>点击选择文件</span></p>
            </div>
            <input type="file" id="fileInput" multiple>
            <div class="upload-progress" id="uploadProgress">
                <div class="progress-bar">
                    <div class="progress-fill" id="progressFill"></div>
                </div>
                <p class="progress-text" id="progressText">上传中...</p>
            </div>
        </div>

        <div class="file-list-section">
            <div class="file-list-header">
                <h2>文件列表</h2>
                <button class="refresh-btn" onclick="loadFiles()">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M23 4v6h-6M1 20v-6h6"/>
                        <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
                    </svg>
                    刷新
                </button>
            </div>
            <ul class="file-list" id="fileList"></ul>
        </div>
    </div>

    <script>
        const uploadArea = document.getElementById('uploadArea');
        const fileInput = document.getElementById('fileInput');
        const fileList = document.getElementById('fileList');
        const uploadProgress = document.getElementById('uploadProgress');
        const progressFill = document.getElementById('progressFill');
        const progressText = document.getElementById('progressText');

        uploadArea.addEventListener('click', () => fileInput.click());

        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });

        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('dragover');
        });

        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                uploadFiles(files);
            }
        });

        fileInput.addEventListener('change', () => {
            if (fileInput.files.length > 0) {
                uploadFiles(fileInput.files);
            }
        });

        async function uploadFiles(files) {
            const formData = new FormData();
            for (let i = 0; i < files.length; i++) {
                formData.append('files', files[i]);
            }

            uploadProgress.classList.add('show');
            progressFill.style.width = '0%';
            progressText.textContent = '正在上传...';

            try {
                const xhr = new XMLHttpRequest();
                xhr.upload.onprogress = (e) => {
                    if (e.lengthComputable) {
                        const percent = Math.round((e.loaded / e.total) * 100);
                        progressFill.style.width = percent + '%';
                        progressText.textContent = '上传进度: ' + percent + '%';
                    }
                };
                xhr.onload = () => {
                    if (xhr.status === 200) {
                        progressFill.style.width = '100%';
                        progressText.textContent = '上传成功！';
                        fileInput.value = '';
                        setTimeout(() => {
                            uploadProgress.classList.remove('show');
                            loadFiles();
                        }, 1000);
                    } else {
                        progressText.textContent = '上传失败';
                        setTimeout(() => uploadProgress.classList.remove('show'), 2000);
                    }
                };
                xhr.onerror = () => {
                    progressText.textContent = '上传失败';
                    setTimeout(() => uploadProgress.classList.remove('show'), 2000);
                };
                xhr.open('POST', '/upload');
                xhr.send(formData);
            } catch (err) {
                progressText.textContent = '上传失败: ' + err.message;
                setTimeout(() => uploadProgress.classList.remove('show'), 2000);
            }
        }

        function formatSize(size) {
            if (size < 1024) return size + ' B';
            if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB';
            if (size < 1024 * 1024 * 1024) return (size / 1024 / 1024).toFixed(1) + ' MB';
            return (size / 1024 / 1024 / 1024).toFixed(1) + ' GB';
        }

        async function loadFiles() {
            try {
                const res = await fetch('/api/list');
                const files = await res.json();
                renderFiles(files);
            } catch (err) {
                fileList.innerHTML = '<li class="empty-state"><p>加载失败，请刷新重试</p></li>';
            }
        }

        function renderFiles(files) {
            if (files.length === 0) {
                fileList.innerHTML = '<li class="empty-state"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg><p>暂无文件，上传一些文件吧</p></li>';
                return;
            }

            fileList.innerHTML = files.map(f => {
                const icon = f.isDir
                    ? '<div class="file-icon folder"><svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-8l-2-2z"/></svg></div>'
                    : '<div class="file-icon file"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg></div>';
                const size = f.isDir ? '文件夹' : formatSize(f.size);
                const download = f.isDir ? '' : '<a href="/files/' + encodeURIComponent(f.name) + '" class="file-download" download>下载</a>';
                return '<li class="file-item">' + icon + '<div class="file-info"><div class="file-name">' + f.name + '</div><div class="file-meta">' + size + ' · ' + f.modTime + '</div></div>' + download + '</li>';
            }).join('');
        }

        loadFiles();
        setInterval(loadFiles, 5000);
    </script>
</body>
</html>`
