
## Go语言编译

实现将Go语音代码编译为可执行程序的服务。

```shell
mkdir xxxxxxxx
cd xxxxxxxx
vim go.mod
vim main.go
export GOOS=windows && export GOARCH=amd64 && go mod tidy && go build
```

## Python语言

```shell
mkdir xxxxxxxx
cd xxxxxxxx
vim app.py
pigar generate
pip install -r requirements.txt
pyinstaller --onefile app.py
```