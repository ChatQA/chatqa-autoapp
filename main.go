package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}))

	e.GET("/", hello)
	e.POST("/users", saveUser)
	e.POST("/buildApp", buildApp)
	e.GET("/users/:id", getUser)
	e.Logger.Fatal(e.Start(":9000"))
}

// e.GET("/", hello)
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

// e.GET("/users/:id", getUser)
func getUser(c echo.Context) error {
	// User ID from path `users/:id`
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

// e.POST("/users", saveUser)
func saveUser(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, u)
}

type User struct {
	Name  string `json:"name" xml:"name" form:"name" query:"name"`
	Email string `json:"email" xml:"email" form:"email" query:"email"`
}

func buildApp(c echo.Context) error {
	props := new(BuildAppProps)
	if err := c.Bind(props); err != nil {
		return err
	}

	cacheID := uuid.NewV4().String()
	fmt.Printf("Generated UUID: %s\n", cacheID)

	cacheDir := "/tmp/" + cacheID
	if err := os.MkdirAll(cacheDir, 0644); err != nil {
		log.Println(err)
	}

	// 文件名
	filename := cacheDir + "/main.go"
	// 要写入的内容
	content := []byte(props.Code)

	// 创建或打开文件（如果已存在则追加）
	err := os.WriteFile(filename, content, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}

	// 文件名
	goModFilename := cacheDir + "/go.mod"
	// 要写入的内容
	goModContent := []byte(`module chatqa.cloud/` + props.Bin + `

go 1.18
`)

	// 创建或打开文件（如果已存在则追加）
	err = os.WriteFile(goModFilename, goModContent, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}

	fileList, err := listFilesInDir(cacheDir)
	if err != nil {
		fmt.Println("Failed to get files in directory:", err)
	}

	fmt.Println("Files in directory:", fileList)

	cmd := fmt.Sprintf(
		"cd %s && export GOOS=%s && export GOARCH=%s && go mod tidy && go build",
		cacheDir, props.OS, props.Arch)

	command := exec.Command("bash", "-lc", cmd)
	output, err := command.CombinedOutput()
	//output, err := exec.Command("bash", "-lc", cmd).CombinedOutput()
	if err != nil {
		log.Println("编译失败")
		log.Println(string(output))
		//return err
		return c.JSON(http.StatusUnprocessableEntity, BuildAppResult{
			Success:      false,
			ErrorMessage: string(output),
		})
	}
	log.Println(string(output))

	ext := ""
	if props.OS == "windows" {
		ext = ".exe"
	}

	binFilename := cacheDir + "/" + props.Bin + ext

	if _, err := os.Stat(binFilename); err != nil && !os.IsExist(err) {
		log.Println("新程序不存在")
		return err
	}

	props.ID = cacheID

	return c.File(binFilename)
	//return c.JSON(http.StatusCreated, props)
}

type BuildAppProps struct {
	Code string `json:"code" xml:"code" form:"code" query:"code"`
	OS   string `json:"os" xml:"os" form:"os" query:"os"`
	Arch string `json:"arch" xml:"arch" form:"arch" query:"arch"`
	Bin  string `json:"bin" xml:"bin" form:"bin" query:"bin"`
	ID   string `json:"id" xml:"id" form:"id" query:"id"`
}

type BuildAppResult struct {
	Success      bool   `json:"success" xml:"success" form:"success" query:"success"`
	ErrorMessage string `json:"errorMessage" xml:"errorMessage" form:"errorMessage" query:"errorMessage"`
}

func listFilesInDir(dirPath string) ([]string, error) {
	// 获取指定目录下的所有文件和子目录
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var filePaths []string

	// 遍历文件列表，筛选出普通文件
	for _, file := range files {
		if !file.IsDir() { // 筛选非目录（即普通文件）
			filePaths = append(filePaths, filepath.Join(dirPath, file.Name()))
		}
	}

	return filePaths, nil
}
